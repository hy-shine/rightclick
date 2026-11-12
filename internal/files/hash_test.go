package files

import (
	"path/filepath"
	"testing"
)

func TestHashKnownDigests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.txt")
	writeFile(t, path, "abc")
	tests := []struct {
		algo Algorithm
		want string
	}{
		{SHA256, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{SHA1, "a9993e364706816aba3e25717850c26c9cd0d89d"},
		{MD5, "900150983cd24fb0d6963f7d28e17f72"},
	}
	for _, tt := range tests {
		t.Run(string(tt.algo), func(t *testing.T) {
			got, err := Hash(path, tt.algo)
			if err != nil {
				t.Fatalf("Hash(%q, %s): %v", path, tt.algo, err)
			}
			if got != tt.want {
				t.Errorf("Hash(%q, %s) = %s, want %s", path, tt.algo, got, tt.want)
			}
		})
	}
}

func TestCompareHashes(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	c := filepath.Join(dir, "c.txt")
	writeFile(t, a, "same")
	writeFile(t, b, "same")
	writeFile(t, c, "different")

	same, _, _, err := CompareHashes(a, b, SHA256)
	if err != nil {
		t.Fatalf("CompareHashes(a, b): %v", err)
	}
	if !same {
		t.Errorf("CompareHashes(a, b) same = false, want true for equal content")
	}
	same, _, _, err = CompareHashes(a, c, SHA256)
	if err != nil {
		t.Fatalf("CompareHashes(a, c): %v", err)
	}
	if same {
		t.Errorf("CompareHashes(a, c) same = true, want false for different content")
	}
}

func TestHashRefusesFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "folder")
	mustMkdir(t, dir)
	if _, err := Hash(dir, SHA256); err == nil {
		t.Errorf("Hash(%q) error = nil, want error for a folder", dir)
	}
}

func TestHashRejectsUnknownAlgorithm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.txt")
	writeFile(t, path, "abc")
	if _, err := Hash(path, Algorithm("SHA-512")); err == nil {
		t.Errorf("Hash(%q, SHA-512) error = nil, want error for unknown algorithm", path)
	}
}
