package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hy-shine/rightclick/internal/locales"
)

// Copy copies the file, symbolic link or folder at src into dir and returns the
// new path. Folders are copied recursively. A name already taken in dir gets
// " 2", " 3", ...; nothing is overwritten.
func Copy(src, dir string) (string, error) {
	info, err := os.Lstat(src)
	if err != nil {
		return "", err
	}
	if err := refuseInside(src, dir, info.IsDir()); err != nil {
		return "", err
	}
	return copyEntry(src, info, dir, filepath.Base(src))
}

// Move moves the file or folder at src into dir and returns the new path, with
// the same naming as Copy. When src is on another volume, Move copies it and
// then removes src.
func Move(src, dir string) (string, error) {
	info, err := os.Lstat(src)
	if err != nil {
		return "", err
	}
	if err := refuseInside(src, dir, info.IsDir()); err != nil {
		return "", err
	}
	// Moving an item into the folder it already sits in would rename it to a
	// numbered name, so it is refused. The folders are compared on disk.
	if sameFolder(filepath.Dir(src), dir) {
		return "", errors.New(locales.AlreadyInFolder())
	}
	name := filepath.Base(src)
	target, err := moveTarget(dir, name, info.IsDir())
	if err != nil {
		return "", err
	}
	err = os.Rename(src, target)
	if err == nil {
		return target, nil
	}
	if !info.IsDir() {
		// The placeholder is still empty, so removing it loses nothing.
		if rmErr := os.Remove(target); rmErr != nil {
			return "", errors.Join(err, rmErr)
		}
	}
	if !errors.Is(err, syscall.EXDEV) {
		return "", err
	}
	dst, err := copyEntry(src, info, dir, name)
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(src); err != nil {
		return dst, fmt.Errorf("copied to %s, but could not remove %s: %w", dst, src, err)
	}
	return dst, nil
}

// moveTarget returns the path that Move renames src onto. A file gets a
// placeholder created with O_EXCL, so the rename replaces only that placeholder.
// A folder gets the first name that is free at the check; see freeFolderName.
func moveTarget(dir, name string, isDir bool) (string, error) {
	if isDir {
		return freeFolderName(dir, name)
	}
	return reserveFile(dir, name)
}

// copyEntry copies one entry into dir under a free name. A folder is reserved
// first, then its children are copied into it.
func copyEntry(src string, info fs.FileInfo, dir, name string) (string, error) {
	switch {
	case info.Mode()&fs.ModeSymlink != 0:
		target, err := os.Readlink(src)
		if err != nil {
			return "", err
		}
		return reserve(dir, name, false, func(path string) error {
			return os.Symlink(target, path)
		})
	case info.IsDir():
		dst, err := reserve(dir, name, true, func(path string) error {
			// Owner rwx is added so that a read-only folder can still receive its children.
			return os.Mkdir(path, info.Mode().Perm()|0o700)
		})
		if err != nil {
			return "", err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return dst, err
		}
		for _, e := range entries {
			child := filepath.Join(src, e.Name())
			childInfo, err := os.Lstat(child)
			if err != nil {
				return dst, err
			}
			if _, err := copyEntry(child, childInfo, dst, e.Name()); err != nil {
				return dst, err
			}
		}
		// The folder keeps its own quarantine, as a downloaded app bundle has it.
		if err := copyQuarantine(src, dst); err != nil {
			return dst, err
		}
		return dst, nil
	case info.Mode().IsRegular():
		return reserve(dir, name, false, copyFileTo(src, info.Mode().Perm()))
	default:
		return "", fmt.Errorf("cannot copy %s: not a file, link or folder", src)
	}
}

// copyFileTo returns a create function that copies src into a new file at the
// path it is given. The file is opened with O_EXCL, so it is never overwritten.
// The copy keeps the quarantine of src, so a downloaded file is still checked
// by macOS the first time it runs.
func copyFileTo(src string, perm fs.FileMode) func(path string) error {
	return func(path string) error {
		in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()
		// The file is created with owner write access, so that its quarantine can be
		// set after the bytes are written. The permissions of src are restored last.
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm|0o600)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			os.Remove(path)
			return err
		}
		if err := out.Close(); err != nil {
			os.Remove(path)
			return err
		}
		if err := copyQuarantine(src, path); err != nil {
			os.Remove(path)
			return err
		}
		if err := os.Chmod(path, perm); err != nil {
			os.Remove(path)
			return err
		}
		return nil
	}
}

// refuseInside rejects copying or moving a folder into itself, which would never
// finish. The check compares folders on disk, so a link into src counts as src.
func refuseInside(src, dir string, isDir bool) error {
	if !isDir {
		return nil
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	inside, err := isInside(dir, info)
	if err != nil {
		return err
	}
	if inside {
		return errors.New(locales.FolderIntoItself())
	}
	return nil
}
