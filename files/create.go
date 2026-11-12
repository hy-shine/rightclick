package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// maxAttempts bounds the " 2", " 3", ... suffixes tried for one name.
const maxAttempts = 9999

// splitName splits name into a base and an extension that keeps its dot.
// Folders are never split, so "v1.2" stays whole.
func splitName(name string, isDir bool) (base, ext string) {
	if isDir {
		return name, ""
	}
	ext = filepath.Ext(name)
	if ext == name {
		return name, ""
	}
	return strings.TrimSuffix(name, ext), ext
}

// candidate returns the name to try on attempt n. Attempt 1 is the name itself;
// later attempts add " 2", " 3", ... before the extension.
func candidate(name string, isDir bool, n int) string {
	if n == 1 {
		return name
	}
	base, ext := splitName(name, isDir)
	return base + " " + strconv.Itoa(n) + ext
}

// freeFolderName returns the first path in dir for the folder name that does not
// exist yet. It only looks, so it is a check, not a reservation. The caller must
// rename a folder onto the path: rename refuses to replace an existing folder, so
// a name taken after the check makes the rename fail instead of overwriting.
func freeFolderName(dir, name string) (string, error) {
	for n := 1; n <= maxAttempts; n++ {
		path := filepath.Join(dir, candidate(name, true, n))
		_, err := os.Lstat(path)
		if errors.Is(err, fs.ErrNotExist) {
			return path, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("no free name for %q in %s", name, dir)
}

// reserve makes a new entry named after name in dir. create makes the entry at
// the path it is given. If that path already exists, reserve tries the next
// suffix. Creating the entry is the reservation, so two callers cannot take
// the same name.
func reserve(dir, name string, isDir bool, create func(path string) error) (string, error) {
	for n := 1; n <= maxAttempts; n++ {
		path := filepath.Join(dir, candidate(name, isDir, n))
		err := create(path)
		if err == nil {
			return path, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("no free name for %q in %s", name, dir)
}

// reserveFile reserves a free file name in dir for a file that a later rename
// will fill. The empty file it creates with O_EXCL is the reservation: a rename
// onto it replaces only that placeholder, never another file.
func reserveFile(dir, name string) (string, error) {
	return reserve(dir, name, false, func(path string) error {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return err
		}
		return f.Close()
	})
}

// CreateFile creates a file called name in dir with content and returns its
// path. If the name is taken, it adds " 2", " 3", ... before the extension.
func CreateFile(dir, name string, content []byte) (string, error) {
	return reserve(dir, name, false, func(path string) error {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		if _, err := f.Write(content); err != nil {
			f.Close()
			os.Remove(path)
			return err
		}
		return f.Close()
	})
}

// CreateFolder creates a folder called name in dir and returns its path. Taken
// names are numbered the same way as in CreateFile.
func CreateFolder(dir, name string) (string, error) {
	return reserve(dir, name, true, func(path string) error {
		return os.Mkdir(path, 0o755)
	})
}
