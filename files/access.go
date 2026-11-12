package files

import (
	"errors"
	"io/fs"
	"os"
)

// Access is what a folder probe found. The folders pane shows one badge per
// state, so "the user refused" and "the folder is gone" stay apart.
type Access int

const (
	AccessOK Access = iota
	AccessDenied
	AccessMissing
)

// FolderAccess reports whether dir can be listed. macOS asks for permission the
// first time an app reads Desktop, Documents, or Downloads; a refusal surfaces
// as a permission error. A path that is gone, or is not a folder, is missing.
func FolderAccess(dir string) Access {
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return AccessMissing
		}
		return AccessDenied
	}
	if !info.IsDir() {
		return AccessMissing
	}
	if _, err := os.ReadDir(dir); err != nil {
		return AccessDenied
	}
	return AccessOK
}
