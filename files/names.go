// Package files holds the file actions behind the right-click menu: naming,
// creating files, copying and moving, archives, hashes, and opening apps.
// It does not know about Finder or the UI.
package files

import (
	"path/filepath"
	"strings"
)

// Name returns the last element of path: "b.txt" for "/a/b.txt".
func Name(path string) string { return filepath.Base(path) }

// Ext returns the extension without its dot: "txt" for "b.txt". A name that
// starts with a dot and has no other dot, such as ".bashrc", has no extension.
func Ext(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	if ext == name {
		return ""
	}
	return strings.TrimPrefix(ext, ".")
}

// BaseName returns the name without its extension: "b" for "b.txt".
func BaseName(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	if ext == "" || ext == name {
		return name
	}
	return strings.TrimSuffix(name, ext)
}
