package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFolderAccess(t *testing.T) {
	root := t.TempDir()
	readable := filepath.Join(root, "readable")
	if err := os.Mkdir(readable, 0o755); err != nil {
		t.Fatal(err)
	}
	// A folder the user refused to share reads as a permission error.
	unreadable := filepath.Join(root, "unreadable")
	if err := os.Mkdir(unreadable, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o755) })
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		dir  string
		want Access
	}{
		{"readable folder", readable, AccessOK},
		{"unreadable folder", unreadable, AccessDenied},
		{"missing folder", filepath.Join(root, "missing"), AccessMissing},
		{"not a folder", file, AccessMissing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FolderAccess(tt.dir); got != tt.want {
				t.Errorf("FolderAccess(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}
