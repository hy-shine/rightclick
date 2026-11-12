package main

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hy-shine/rightclick/config"
)

func TestStemNames(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  []string
	}{
		{name: "file with extension", paths: []string{"/a/notes.md"}, want: []string{"notes"}},
		{name: "several files", paths: []string{"/a/notes.md", "/b/archive.tar.gz"}, want: []string{"notes", "archive.tar"}},
		{name: "no extension", paths: []string{"/a/README"}, want: []string{"README"}},
		{name: "dot file", paths: []string{"/a/.bashrc"}, want: []string{".bashrc"}},
		{name: "folder", paths: []string{"/a/folder.name"}, want: []string{"folder"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stemNames(tt.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stemNames(%v) = %v, want %v", tt.paths, got, tt.want)
			}
		})
	}
}

func TestExtensionNamesSkipsItemsWithoutOne(t *testing.T) {
	paths := []string{"/a/notes.md", "/a/README", "/b/.bashrc", "/c/archive.tar.gz"}
	want := []string{"md", "gz"}
	if got := extensionNames(paths); !reflect.DeepEqual(got, want) {
		t.Errorf("extensionNames(%v) = %v, want %v", paths, got, want)
	}
}

// TestCopyExtensionRefusesWhenNothingHasOne keeps the no-extension case out of
// the clipboard: the action reports an error instead of writing an empty string.
func TestCopyExtensionRefusesWhenNothingHasOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), "README")
	err := runAction("copy-extension", nil, []string{path}, config.Default())
	if err == nil {
		t.Fatal("copy-extension of a name without an extension returned no error")
	}
}
