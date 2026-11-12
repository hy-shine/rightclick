package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateFileNumbersTakenNames(t *testing.T) {
	dir := t.TempDir()
	names := []string{"Untitled.md", "Untitled 2.md", "Untitled 3.md"}
	contents := []string{"first", "second", "third"}
	for i, want := range names {
		got, err := CreateFile(dir, "Untitled.md", []byte(contents[i]))
		if err != nil {
			t.Fatalf("CreateFile attempt %d: %v", i+1, err)
		}
		if wantPath := filepath.Join(dir, want); got != wantPath {
			t.Errorf("CreateFile attempt %d = %q, want %q", i+1, got, wantPath)
		}
	}
	for i, name := range names {
		if got := readFile(t, filepath.Join(dir, name)); got != contents[i] {
			t.Errorf("%s content = %q, want %q", name, got, contents[i])
		}
	}
}

func TestCreateFolderNumbersTakenNames(t *testing.T) {
	dir := t.TempDir()
	names := []string{"New Folder", "New Folder 2"}
	for _, want := range names {
		got, err := CreateFolder(dir, "New Folder")
		if err != nil {
			t.Fatalf("CreateFolder: %v", err)
		}
		if wantPath := filepath.Join(dir, want); got != wantPath {
			t.Errorf("CreateFolder = %q, want %q", got, wantPath)
		}
		info, err := os.Lstat(got)
		if err != nil {
			t.Fatalf("Lstat %s: %v", got, err)
		}
		if !info.IsDir() {
			t.Errorf("%s is not a folder, mode %v", got, info.Mode())
		}
	}
}
