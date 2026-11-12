package files

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

var testQuarantine = []byte("0083;6af52dfe;Safari;")

func TestCopyKeepsQuarantine(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "tool.bin")
	writeFile(t, src, "data")
	if err := writeQuarantine(src, testQuarantine); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	mustMkdir(t, out)
	got, err := Copy(src, out)
	if err != nil {
		t.Fatal(err)
	}
	value, ok, err := readQuarantine(got)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !bytes.Equal(value, testQuarantine) {
		t.Errorf("copy quarantine = %q, %v; want %q", value, ok, testQuarantine)
	}
}

func TestCopyOfReadOnlyQuarantinedFileKeepsModeAndQuarantine(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "readonly.bin")
	writeFile(t, src, "data")
	if err := writeQuarantine(src, testQuarantine); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(src, 0o444); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	mustMkdir(t, out)
	got, err := Copy(src, out)
	if err != nil {
		t.Fatalf("Copy of a read-only file failed: %v", err)
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o444 {
		t.Errorf("copy mode = %v, want 0444", info.Mode().Perm())
	}
	value, ok, err := readQuarantine(got)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !bytes.Equal(value, testQuarantine) {
		t.Errorf("copy quarantine = %q, %v; want %q", value, ok, testQuarantine)
	}
}

func TestCopyWithoutQuarantineStaysClean(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "plain.txt")
	writeFile(t, src, "data")
	out := filepath.Join(dir, "out")
	mustMkdir(t, out)
	got, err := Copy(src, out)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := readQuarantine(got); err != nil || ok {
		t.Errorf("copy of a clean file: quarantine present = %v, err = %v", ok, err)
	}
}

func TestZipAndUnzipKeepQuarantine(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "tool.bin")
	writeFile(t, src, "data")
	if err := writeQuarantine(src, testQuarantine); err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(dir, "zip")
	mustMkdir(t, folder)
	zipPath, err := Zip([]string{src}, folder, "bundle.zip")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := readQuarantine(zipPath); err != nil || !ok {
		t.Fatalf("archive of a quarantined file: quarantine present = %v, err = %v", ok, err)
	}
	written, _, err := Unzip(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if written != 1 {
		t.Fatalf("Unzip wrote %d files, want 1", written)
	}
	value, ok, err := readQuarantine(filepath.Join(folder, "tool.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !bytes.Equal(value, testQuarantine) {
		t.Errorf("unpacked quarantine = %q, %v; want %q", value, ok, testQuarantine)
	}
}

func TestUnzipOfPlainArchiveAddsNoQuarantine(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "in.zip")
	writeZip(t, archive, []zipEntry{{name: "a.txt", body: "text"}})
	if _, _, err := Unzip(archive); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := readQuarantine(filepath.Join(dir, "a.txt")); err != nil || ok {
		t.Errorf("unpacked clean file: quarantine present = %v, err = %v", ok, err)
	}
}
