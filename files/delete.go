package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"rightclick/locales"
)

// Protected reports whether path must not be deleted: a relative path, the root
// folder, the home folder, or the top folder of a mounted volume (/Volumes/Name).
// home is the user's home folder.
func Protected(path, home string) bool {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return true
	}
	if clean == "/" || clean == "/Volumes" || clean == filepath.Clean(home) {
		return true
	}
	return filepath.Dir(clean) == "/Volumes"
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
		if err := os.RemoveAll(filepath.Clean(p)); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", Name(p), err))
			continue
		}
		removed++
	}
	return removed, errors.Join(errs...)
}
