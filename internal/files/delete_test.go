package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtected(t *testing.T) {
	const home = "/Users/a"
	tests := []struct {
		path string
		want bool
	}{
		{"/", true},
		{"/Volumes", true},
		{"/Volumes/Data", true},
		{"/Volumes/Data/", true},
		{"/Users/a", true},
		{"/Users/a/", true},
		{"/Users/a/Desktop/notes.txt", false},
		{"/Volumes/Data/Projects", false},
		{"/Users/b", false},
		{"relative/path", true},
		{"", true},
	}
	for _, tt := range tests {
		if got := Protected(tt.path, home); got != tt.want {
			t.Errorf("Protected(%q, %q) = %v, want %v", tt.path, home, got, tt.want)
		}
	}
}

func TestRemoveAllKeepsProtectedPaths(t *testing.T) {
	home := t.TempDir()
	file := filepath.Join(home, "old.txt")
	folder := filepath.Join(home, "old")
	nested := filepath.Join(folder, "inner.txt")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{file, nested} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	removed, err := RemoveAll([]string{home, file, folder}, home)
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}
	if err == nil {
		t.Error("the home folder was not refused")
	}
	if _, statErr := os.Stat(home); statErr != nil {
		t.Errorf("home folder was removed: %v", statErr)
	}
	for _, p := range []string{file, folder, nested} {
		if _, statErr := os.Stat(p); !os.IsNotExist(statErr) {
			t.Errorf("%s still exists", p)
		}
	}
}

func TestRemoveAllContinuesPastMissingPaths(t *testing.T) {
	home := t.TempDir()
	file := filepath.Join(home, "kept-going.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	removed, err := RemoveAll([]string{filepath.Join(home, "missing"), file}, home)
	if removed != 2 || err != nil {
		t.Errorf("RemoveAll = %d, %v; want 2 and no error for a missing path", removed, err)
	}
}
