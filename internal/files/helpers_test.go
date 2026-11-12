package files

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes content to path, creating its parent folders first.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// readFile returns the content of the file at path.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile %s: %v", path, err)
	}
	return string(data)
}

// mustMkdir creates the folder at path, along with any missing parents.
func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mustMkdir %s: %v", path, err)
	}
}

// exists reports whether anything, including a symbolic link, is at path.
func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
