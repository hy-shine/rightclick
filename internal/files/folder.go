package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// sameFolder reports whether a and b are the same folder on disk. The folders are
// compared by identity, not by spelling: APFS is case-insensitive by default, so
// /USERS/JESSY is the same folder as /Users/jessy. When either folder does not exist,
// the paths are compared as text without case. That can only refuse more, never less.
func sameFolder(a, b string) bool {
	if infoA, errA := os.Stat(a); errA == nil {
		if infoB, errB := os.Stat(b); errB == nil {
			return os.SameFile(infoA, infoB)
		}
	}
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// isInside reports whether dir is the folder described by folder, or lies inside it.
// Every folder from dir up to the root is compared with folder on disk, so a link or
// another spelling of the same folder cannot hide it.
func isInside(dir string, folder fs.FileInfo) (bool, error) {
	cur, err := resolvedFolder(dir)
	if err != nil {
		return false, err
	}
	for {
		info, err := os.Stat(cur)
		if err != nil {
			return false, err
		}
		if os.SameFile(info, folder) {
			return true, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return false, nil
		}
		cur = parent
	}
}

// resolvedFolder returns the absolute path of folder with its symbolic links
// followed.
func resolvedFolder(folder string) (string, error) {
	abs, err := filepath.Abs(folder)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}
