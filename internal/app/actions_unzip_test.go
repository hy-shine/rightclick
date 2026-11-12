package app

import (
	"archive/zip"
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
