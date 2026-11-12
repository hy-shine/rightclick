package files

import (
	"bytes"
	"path/filepath"
	"testing"
)

// A downloaded app bundle carries its quarantine on the folder itself, so the
// folder must keep it through a copy, an archive and an extraction.

func TestCopyKeepsFolderQuarantine(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "Tool.app")
	writeFile(t, filepath.Join(src, "Contents", "Info.plist"), "plist")
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
		t.Errorf("copied folder quarantine = %q, %v; want %q", value, ok, testQuarantine)
	}
}

func TestZipAndUnzipKeepFolderQuarantine(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "Tool.app")
	writeFile(t, filepath.Join(src, "Contents", "Info.plist"), "plist")
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
		t.Fatalf("archive of a quarantined folder: quarantine present = %v, err = %v", ok, err)
	}
	if _, _, err := Unzip(zipPath); err != nil {
		t.Fatal(err)
	}
	value, ok, err := readQuarantine(filepath.Join(folder, "Tool.app"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !bytes.Equal(value, testQuarantine) {
		t.Errorf("unpacked folder quarantine = %q, %v; want %q", value, ok, testQuarantine)
	}
}
