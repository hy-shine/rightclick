package files

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// zipEntry is one entry for writeZip. A zero mode means a regular file.
type zipEntry struct {
	name string
	body string
	mode fs.FileMode
}

// writeZip writes an archive at path with the given entries. It uses archive/zip
// directly so that tests can build entries that Zip never creates.
func writeZip(t *testing.T, path string, entries []zipEntry) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.mode != 0 {
			hdr.SetMode(e.mode)
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatalf("CreateHeader %s: %v", e.name, err)
		}
		if _, err := io.WriteString(w, e.body); err != nil {
			t.Fatalf("write %s: %v", e.name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("writeZip %s: %v", path, err)
	}
}

func TestZipRoundTrip(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	folder := filepath.Join(srcDir, "photos")
	writeFile(t, filepath.Join(folder, "day1", "a.txt"), "alpha")
	file := filepath.Join(srcDir, "notes.txt")
	writeFile(t, file, "beta")

	archive, err := Zip([]string{folder, file}, srcDir, "bundle.zip")
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}
	dst := filepath.Join(root, "dst")
	mustMkdir(t, dst)
	moved, err := Move(archive, dst)
	if err != nil {
		t.Fatalf("Move(%q, %q): %v", archive, dst, err)
	}
	written, skipped, err := Unzip(moved)
	if err != nil {
		t.Fatalf("Unzip(%q): %v", moved, err)
	}
	if written != 2 || skipped != 0 {
		t.Errorf("Unzip = written %d, skipped %d; want written 2, skipped 0", written, skipped)
	}
	if content := readFile(t, filepath.Join(dst, "photos", "day1", "a.txt")); content != "alpha" {
		t.Errorf("photos/day1/a.txt content = %q, want %q", content, "alpha")
	}
	if content := readFile(t, filepath.Join(dst, "notes.txt")); content != "beta" {
		t.Errorf("notes.txt content = %q, want %q", content, "beta")
	}
}

func TestUniqueTopKeepsFolderDots(t *testing.T) {
	used := map[string]bool{}
	first := uniqueTop("v1.2", true, used)
	second := uniqueTop("v1.2", true, used)
	if first != "v1.2" || second != "v1.2 2" {
		t.Fatalf("uniqueTop folder names = %q and %q, want %q and %q", first, second, "v1.2", "v1.2 2")
	}
}

func TestZipTakesFreeNameWhenTargetExists(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "notes.txt")
	writeFile(t, file, "beta")
	out := filepath.Join(root, "out")
	mustMkdir(t, out)
	writeFile(t, filepath.Join(out, "bundle.zip"), "old")

	got, err := Zip([]string{file}, out, "bundle.zip")
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}
	if want := filepath.Join(out, "bundle 2.zip"); got != want {
		t.Errorf("Zip = %q, want %q", got, want)
	}
	if content := readFile(t, filepath.Join(out, "bundle.zip")); content != "old" {
		t.Errorf("existing bundle.zip content = %q, want %q", content, "old")
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", out, err)
	}
	if len(entries) != 2 {
		t.Errorf("output folder has %d entries, want the old archive and the new one only", len(entries))
	}
}

func TestZipMissingInputLeavesNoArchive(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "out")
	mustMkdir(t, out)

	_, err := Zip([]string{filepath.Join(root, "missing.txt")}, out, "bundle.zip")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Zip error = %v, want fs.ErrNotExist", err)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", out, err)
	}
	for _, e := range entries {
		t.Errorf("Zip left %q behind after failing", e.Name())
	}
}

func TestUnzipRejectsEntryOutsideFolder(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dst")
	mustMkdir(t, dir)
	archive := filepath.Join(dir, "evil.zip")
	writeZip(t, archive, []zipEntry{{name: "../escape.txt", body: "escaped"}})

	if _, _, err := Unzip(archive); err == nil {
		t.Fatal("Unzip error = nil, want error for entry ../escape.txt")
	}
	if exists(filepath.Join(root, "escape.txt")) {
		t.Error("Unzip wrote escape.txt outside the archive's folder")
	}
}

func TestUnzipTakesFreeNameForTakenFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "old")
	archive := filepath.Join(dir, "in.zip")
	writeZip(t, archive, []zipEntry{{name: "a.txt", body: "new"}})

	written, skipped, err := Unzip(archive)
	if err != nil {
		t.Fatalf("Unzip(%q): %v", archive, err)
	}
	if written != 1 || skipped != 0 {
		t.Errorf("Unzip = written %d, skipped %d; want written 1, skipped 0", written, skipped)
	}
	if content := readFile(t, filepath.Join(dir, "a.txt")); content != "old" {
		t.Errorf("existing a.txt content = %q, want %q", content, "old")
	}
	if content := readFile(t, filepath.Join(dir, "a 2.txt")); content != "new" {
		t.Errorf("a 2.txt content = %q, want %q", content, "new")
	}
}

func TestUnzipSkipsSymlinkEntries(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "in.zip")
	writeZip(t, archive, []zipEntry{
		{name: "a.txt", body: "text"},
		{name: "link", body: "a.txt", mode: fs.ModeSymlink | 0o777},
	})

	written, skipped, err := Unzip(archive)
	if err != nil {
		t.Fatalf("Unzip(%q): %v", archive, err)
	}
	if written != 1 || skipped != 1 {
		t.Errorf("Unzip = written %d, skipped %d; want written 1, skipped 1", written, skipped)
	}
	if exists(filepath.Join(dir, "link")) {
		t.Error("Unzip created a link for a symlink entry, want it skipped")
	}
}
