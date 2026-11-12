package finder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCanRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !canRead(path) {
		t.Error("canRead() = false for a file it can open")
	}
	if canRead(filepath.Join(t.TempDir(), "missing")) {
		t.Error("canRead() = true for a file that does not exist")
	}
	if canRead("") {
		t.Error("canRead(\"\") = true")
	}
}
