package files

import "testing"

func TestExt(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"b.txt", "txt"},
		{"archive.tar.gz", "gz"},
		{".bashrc", ""},
		{"README", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := Ext(tt.path); got != tt.want {
				t.Errorf("Ext(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestBaseName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"b.txt", "b"},
		{"archive.tar.gz", "archive.tar"},
		{".bashrc", ".bashrc"},
		{"README", "README"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := BaseName(tt.path); got != tt.want {
				t.Errorf("BaseName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
