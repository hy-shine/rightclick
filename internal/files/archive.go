package files

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Zip archives the files and folders in paths into a new archive called name in
// dir and returns its path. Folders are archived recursively. The archive is
// built in a temporary file and renamed only when complete, so a failure leaves
// no partial archive behind.
func Zip(paths []string, dir, name string) (string, error) {
	target, err := reserveFile(dir, name)
	if err != nil {
		return "", err
	}
	err = writeAtomic(target, 0o644, func(w io.Writer) error {
		zw := zip.NewWriter(w)
		used := map[string]bool{}
		for _, p := range paths {
			info, err := os.Lstat(p)
			if err != nil {
				return err
			}
			top := uniqueTop(filepath.Base(p), info.IsDir(), used)
			if err := addToZip(zw, p, top); err != nil {
				return err
			}
		}
		return zw.Close()
	})
	if err != nil {
		// Remove the placeholder reserved above, so a failure leaves nothing behind.
		if rmErr := os.Remove(target); rmErr != nil {
			return "", errors.Join(err, rmErr)
		}
		return "", err
	}
	return target, nil
}

// uniqueTop returns name, or name with " 2", " 3", ... added, so that two items
// with the same name get separate entries in the archive. isDir keeps a folder
// name such as "v1.2" whole.
func uniqueTop(name string, isDir bool, used map[string]bool) string {
	entry := name
	for n := 2; used[entry]; n++ {
		entry = candidate(name, isDir, n)
	}
	used[entry] = true
	return entry
}

// addToZip adds the file, symbolic link or folder at path to zw under entry, its
// name inside the archive. A folder's contents follow its own entry.
func addToZip(zw *zip.Writer, path, entry string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = entry
	switch {
	case info.IsDir():
		header.Name += "/"
		if _, err := zw.CreateHeader(header); err != nil {
			return err
		}
		children, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, c := range children {
			if err := addToZip(zw, filepath.Join(path, c.Name()), entry+"/"+c.Name()); err != nil {
				return err
			}
		}
		return nil
	case info.Mode()&fs.ModeSymlink != 0:
		// A link is stored as its target text, the way zip tools store it.
		target, err := os.Readlink(path)
		if err != nil {
			return err
		}
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, target)
		return err
	case info.Mode().IsRegular():
		header.Method = zip.Deflate
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	default:
		return fmt.Errorf("cannot archive %s: not a file, link or folder", path)
	}
}

// Unzip extracts the archive at path into the folder that holds it. It returns
// how many files were written and how many entries were skipped. Symbolic links
// are skipped. A file whose name is taken gets " 2", " 3", ... as Copy does. An
// entry whose name would leave the folder stops the extraction with an error;
// files written before it stay in place.
func Unzip(path string) (written, skipped int, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return 0, 0, err
	}
	defer r.Close()
	dir := filepath.Dir(path)
	for _, f := range r.File {
		name := filepath.FromSlash(f.Name)
		if !filepath.IsLocal(name) {
			return written, skipped, fmt.Errorf("archive entry leaves the folder: %s", f.Name)
		}
		mode := f.Mode()
		switch {
		case mode.IsDir():
			if err := os.MkdirAll(filepath.Join(dir, name), 0o755); err != nil {
				return written, skipped, err
			}
		case mode.IsRegular():
			if err := extractFile(f, dir, name); err != nil {
				return written, skipped, err
			}
			written++
		default:
			skipped++
		}
	}
	return written, skipped, nil
}

// extractFile writes one archive entry to dir/name, under a free name.
func extractFile(f *zip.File, dir, name string) error {
	parent := filepath.Join(dir, filepath.Dir(name))
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = reserve(parent, filepath.Base(name), false, func(path string) error {
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, f.Mode().Perm()|0o600)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			os.Remove(path)
			return err
		}
		return out.Close()
	})
	return err
}
