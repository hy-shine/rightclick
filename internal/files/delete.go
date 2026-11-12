package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hy-shine/rightclick/internal/locales"
)

// Protected reports whether path must not be deleted: a relative path, the root
// folder, the home folder, or the top folder of a mounted volume (/Volumes/Name).
// The folders are compared on disk rather than by spelling, so a symbolic link or a
// capitalised spelling of the same folder is still caught. home is the user's home
// folder.
func Protected(path, home string) bool {
	if !filepath.IsAbs(filepath.Clean(path)) {
		return true
	}
	item := resolvedPath(path)
	if sameFolder(item, "/") || sameFolder(item, "/Volumes") || sameFolder(item, home) {
		return true
	}
	return sameFolder(filepath.Dir(item), "/Volumes")
}

// RemoveAll deletes each path with everything in it, for good. A protected path is
// refused and the other paths are still deleted. It returns how many paths were
// deleted and the errors of the ones that were not, joined.
func RemoveAll(paths []string, home string) (int, error) {
	removed := 0
	var errs []error
	for _, p := range paths {
		if Protected(p, home) {
			errs = append(errs, errors.New(locales.DeleteRefused(p)))
			continue
		}
		// The item is deleted where it really is, so that no link in its folders can
		// lead os.RemoveAll somewhere else.
		target, err := resolveParent(p)
		if errors.Is(err, fs.ErrNotExist) {
			// The folder above is gone, so the item is gone too.
			removed++
			continue
		}
		if err == nil {
			err = os.RemoveAll(target)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", Name(p), err))
			continue
		}
		removed++
	}
	return removed, errors.Join(errs...)
}

// resolvedPath is resolveParent without the error. When the folders above path cannot
// be resolved, the cleaned path is used. Then no link can lead into a folder that
// exists, so the lexical check is enough.
func resolvedPath(path string) string {
	resolved, err := resolveParent(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return resolved
}

// resolveParent returns the absolute path of path with the symbolic links in its
// parent folders followed. The last name is kept as it is: a link is the item that
// is deleted, not the folder it points to.
func resolveParent(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(abs)), nil
}
