package app

import (
	"archive/zip"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// TestUnzipActionExtractsAndReportsTheCount extracts the archive next to itself and
// reports how many files it wrote.
func TestUnzipActionExtractsAndReportsTheCount(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bundle.zip")
	writeZip(t, archive, map[string]string{"hello.txt": "hi"})
	var title, body string
	oldNotify := notify
	t.Cleanup(func() { notify = oldNotify })
	notify = func(gotTitle, gotBody string) { title, body = gotTitle, gotBody }

	if err := unzipAction(nil, []string{archive}, config.Default()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil || string(data) != "hi" {
		t.Fatalf("extracted file = %q, error = %v", data, err)
	}
	if title != locales.NotifyUnzipped() || body != locales.UnzipWritten(1) {
		t.Errorf("notification = %q / %q", title, body)
	}
}

func TestExtractToFolderActionReportsDestination(test *testing.T) {
	root := test.TempDir()
	archive := filepath.Join(root, "bundle.zip")
	writeZip(test, archive, map[string]string{"nested/notes.md": "hello"})
	previous := notify
	test.Cleanup(func() { notify = previous })
	var title, body string
	notify = func(nextTitle, nextBody string) { title, body = nextTitle, nextBody }
	if err := runAction("unzip-to-folder", url.Values{}, []string{archive}, config.Default()); err != nil {
		test.Fatal(err)
	}
	folder := filepath.Join(root, "bundle")
	if title != locales.NotifyUnzipped() || body != locales.UnzipFolderWritten(1, folder) {
		test.Fatalf("notification = %q / %q", title, body)
	}
	if data, err := os.ReadFile(filepath.Join(folder, "nested", "notes.md")); err != nil || string(data) != "hello" {
		test.Fatal("output file missing")
	}
}

func TestZipActionUsesSavedExclusions(test *testing.T) {
	root := test.TempDir()
	project := filepath.Join(root, "project")
	if err := os.Mkdir(project, 0o755); err != nil {
		test.Fatal(err)
	}
	for _, name := range []string{".DS_Store", "notes.md"} {
		if err := os.WriteFile(filepath.Join(project, name), []byte("content"), 0o600); err != nil {
			test.Fatal(err)
		}
	}
	if err := runAction("zip", nil, []string{project}, config.Default()); err != nil {
		test.Fatal(err)
	}
	reader, err := zip.OpenReader(filepath.Join(root, "project.zip"))
	if err != nil {
		test.Fatal(err)
	}
	defer reader.Close()
	found := false
	for _, entry := range reader.File {
		if entry.Name == "project/.DS_Store" {
			test.Fatal("saved exclusion was ignored by action")
		}
		if entry.Name == "project/notes.md" {
			found = true
		}
	}
	if !found {
		test.Fatal("regular file missing from archive")
	}
}

// writeZip creates an archive at path with one file per entry.
func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zw := zip.NewWriter(file)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}
