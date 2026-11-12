package app

import (
	"reflect"
	"testing"
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
