package main

import (
	"net/url"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantAction string
		wantPaths  []string
		wantErr    bool
	}{
		{
			name:       "plain path",
			raw:        "rightclick://copy-path?path=%2FUsers%2Fa%2Fb.txt",
			wantAction: "copy-path",
			wantPaths:  []string{"/Users/a/b.txt"},
		},
		{
			name:       "spaces and unicode",
			raw:        "rightclick://copy-path?path=%2FUsers%2Fa%20b%2F%E6%96%87%E4%BB%B6.txt",
			wantAction: "copy-path",
			wantPaths:  []string{"/Users/a b/文件.txt"},
		},
		{
			// The extension must encode + and ; as %2B and %3B. An unencoded +
			// is read as a space, and url.Query drops a pair with an unencoded ;.
			name:       "plus, ampersand and semicolon",
			raw:        "rightclick://copy-path?path=%2Ftmp%2Fa%2Bb%26c%3Bd.txt",
			wantAction: "copy-path",
			wantPaths:  []string{"/tmp/a+b&c;d.txt"},
		},
		{
			name:       "several paths keep their order",
			raw:        "rightclick://zip?path=%2Ftmp%2Fb.txt&path=%2Ftmp%2Fa.txt",
			wantAction: "zip",
			wantPaths:  []string{"/tmp/b.txt", "/tmp/a.txt"},
		},
		{
			name:       "toggle hidden files",
			raw:        "rightclick://toggle-hidden?path=%2Ftmp",
			wantAction: "toggle-hidden",
			wantPaths:  []string{"/tmp"},
		},
		{
			name:       "new configured file type",
			raw:        "rightclick://new-file?type=json&path=%2Ftmp",
			wantAction: "new-file",
			wantPaths:  []string{"/tmp"},
		},
		{
			name:    "wrong scheme",
			raw:     "https://copy-path?path=%2Ftmp",
			wantErr: true,
		},
		{
			name:    "unknown action",
			raw:     "rightclick://other?path=%2Ftmp",
			wantErr: true,
		},
		{
			name:    "missing path",
			raw:     "rightclick://copy-path",
			wantErr: true,
		},
		{
			name:    "relative path",
			raw:     "rightclick://copy-path?path=tmp%2Fa.txt",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, _, paths, err := parseURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseURL(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if action != tt.wantAction {
				t.Fatalf("parseURL(%q) action = %q, want %q", tt.raw, action, tt.wantAction)
			}
			if !reflect.DeepEqual(paths, tt.wantPaths) {
				t.Fatalf("parseURL(%q) paths = %q, want %q", tt.raw, paths, tt.wantPaths)
			}
		})
	}
}

func TestParseURLParams(t *testing.T) {
	raw := "rightclick://open-with?path=%2Ftmp%2Fa.txt&app=Visual%20Studio%20Code&algo=MD5&name=" +
		url.QueryEscape("周报.md")
	_, params, _, err := parseURL(raw)
	if err != nil {
		t.Fatalf("parseURL(%q) error = %v", raw, err)
	}
	want := map[string]string{"app": "Visual Studio Code", "algo": "MD5", "name": "周报.md"}
	for key, value := range want {
		if got := params.Get(key); got != value {
			t.Fatalf("param %s = %q, want %q", key, got, value)
		}
	}
}

func TestRequireInWatchFolders(t *testing.T) {
	root := t.TempDir()
	if err := requireInWatchFolders([]string{filepath.Join(root, "a.txt")}, []string{root}); err != nil {
		t.Errorf("a path inside the whitelist was refused: %v", err)
	}
	if err := requireInWatchFolders([]string{filepath.Join(root, "..", "outside.txt")}, []string{root}); err == nil {
		t.Error("a path outside the whitelist was accepted")
	}
	if err := requireInWatchFolders([]string{filepath.Join(root, "a.txt")}, nil); err == nil {
		t.Error("a path with an empty whitelist was accepted")
	}
}
