package files

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hy-shine/rightclick/internal/locales"
)

// Limits on one extraction. An archive is checked against the declared sizes before
// anything is written. The bytes actually written are counted as they go, so a size
// that a header understates cannot get past the total. They are variables so that
// tests can lower them.
var (
	// maxUnzipBytes is the most an archive may unpack to, in total.
	maxUnzipBytes int64 = 2 << 30
	// maxUnzipEntries is the most files and folders an archive may hold.
	maxUnzipEntries = 10000
)

const (
	// ratioCheckSize is the size from which the compression ratio of an entry is
	// checked. Smaller files may compress well without being a risk.
	ratioCheckSize = 10 << 20
	// maxUnzipRatio is the most an entry may unpack to, relative to its packed size.
	// Deflate can reach about 1000 to 1; a declared ratio above this is refused.
	maxUnzipRatio = 200
)

// Zip archives the files and folders in paths into a new archive called name in
// dir and returns its path. Folders are archived recursively. The archive is
// built in a temporary file and renamed only when complete, so a failure leaves
// no partial archive behind. When an archived file or folder is quarantined
// (downloaded), the archive is quarantined too, so that Unzip can pass it on.
func Zip(paths []string, dir, name string) (string, error) {
	target, err := reserveFile(dir, name)
	if err != nil {
		return "", err
	}
	var origin quarantineOrigin
	err = writeAtomic(target, 0o644, func(w io.Writer) error {
		zw := zip.NewWriter(w)
		used := map[string]bool{}
		for _, p := range paths {
			info, err := os.Lstat(p)
			if err != nil {
				return err
			}
			top := uniqueTop(filepath.Base(p), info.IsDir(), used)
			if err := addToZip(zw, p, top, &origin); err != nil {
				return err
			}
		}
		return zw.Close()
	})
	if err == nil && origin.found {
		err = writeQuarantine(target, origin.value)
	}
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
// name inside the archive. A folder's contents follow its own entry. origin notes
// the quarantine of the folders and files it adds.
func addToZip(zw *zip.Writer, path, entry string, origin *quarantineOrigin) error {
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
		if err := origin.note(path); err != nil {
			return err
		}
		header.Name += "/"
		if _, err := zw.CreateHeader(header); err != nil {
			return err
		}
		children, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, c := range children {
			if err := addToZip(zw, filepath.Join(path, c.Name()), entry+"/"+c.Name(), origin); err != nil {
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
		if err := origin.note(path); err != nil {
			return err
		}
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
// entry whose name would leave the folder stops the extraction with an error. An
// archive that is too large or holds too many items is refused before anything is
// written. If the bytes written reach the size limit during extraction, the
// extraction stops there. In both error cases, files written before the error stay
// in place. Nothing is written through a symbolic link in the folder.
func Unzip(path string) (written, skipped int, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return 0, 0, err
	}
	defer r.Close()
	if err := checkUnzipLimits(r.File); err != nil {
		return 0, 0, err
	}
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return 0, 0, err
	}
	// An archive from the internet passes its quarantine to every folder and file it
	// unpacks, as Archive Utility does.
	var origin quarantineOrigin
	if err := origin.note(path); err != nil {
		return 0, 0, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return 0, 0, err
	}
	defer root.Close()
	left := maxUnzipBytes
	for _, f := range r.File {
		name := filepath.FromSlash(f.Name)
		if !filepath.IsLocal(name) {
			return written, skipped, fmt.Errorf("archive entry leaves the folder: %s", f.Name)
		}
		mode := f.Mode()
		switch {
		case mode.IsDir():
			if err := mkdirAllIn(root, dir, name, &origin); err != nil {
				return written, skipped, err
			}
		case mode.IsRegular():
			rc, err := f.Open()
			if err != nil {
				return written, skipped, err
			}
			err = extractFile(root, dir, f.Mode().Perm(), rc, name, &left, &origin)
			rc.Close()
			if err != nil {
				return written, skipped, err
			}
			written++
		default:
			skipped++
		}
	}
	return written, skipped, nil
}

// checkUnzipLimits refuses an archive that is too large or has too many items,
// judged from the sizes its entries declare. It runs before anything is written.
func checkUnzipLimits(entries []*zip.File) error {
	if len(entries) > maxUnzipEntries {
		return errors.New(locales.ArchiveTooManyItems())
	}
	var total uint64
	limit := uint64(maxUnzipBytes)
	for _, f := range entries {
		if f.UncompressedSize64 > limit-total {
			return errors.New(locales.ArchiveTooLarge())
		}
		total += f.UncompressedSize64
		if f.UncompressedSize64 >= ratioCheckSize && f.CompressedSize64 > 0 &&
			f.UncompressedSize64/f.CompressedSize64 > maxUnzipRatio {
			return errors.New(locales.ArchiveRatioTooHigh())
		}
	}
	return nil
}

// mkdirAllIn creates the folders of name inside root, as os.MkdirAll does, where
// dir is the absolute path of root. A part that is a symbolic link is refused rather
// than followed, so that nothing is created through a link that the archive could
// have planted. A folder that is created gets the quarantine of the archive, if it
// has one.
func mkdirAllIn(root *os.Root, dir, name string, origin *quarantineOrigin) error {
	name = filepath.Clean(name)
	if name == "." {
		return nil
	}
	built := ""
	for _, part := range strings.Split(name, string(filepath.Separator)) {
		built = filepath.Join(built, part)
		info, err := root.Lstat(built)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if err := root.Mkdir(built, 0o755); err != nil {
				return err
			}
			if origin.found {
				if err := writeQuarantine(filepath.Join(dir, built), origin.value); err != nil {
					return err
				}
			}
		case err != nil:
			return err
		case info.Mode()&fs.ModeSymlink != 0:
			return errors.New(locales.ArchiveThroughLink(built))
		case !info.IsDir():
			return fmt.Errorf("%s is not a folder", built)
		}
	}
	return nil
}

// extractFile writes the data of one archive entry to name in root, under a free
// name. dir is the absolute path of root. Its folders are created first. The bytes
// it writes come from left. If they run out, the file is removed and the extraction
// stops with a message saying so. When the archive has a quarantine, the file gets
// it too.
func extractFile(root *os.Root, dir string, perm fs.FileMode, src io.Reader, name string, left *int64, origin *quarantineOrigin) error {
	parent := filepath.Dir(name)
	if err := mkdirAllIn(root, dir, parent, origin); err != nil {
		return err
	}
	// reserve names the file relative to root, so every write goes through root.
	rel, err := reserve(parent, filepath.Base(name), false, func(relPath string) error {
		out, err := root.OpenFile(relPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm|0o600)
		if err != nil {
			return err
		}
		// One byte past the budget is read, so that an overrun is seen.
		n, err := io.Copy(out, io.LimitReader(src, *left+1))
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
		if err == nil && n > *left {
			err = errors.New(locales.ArchiveStoppedAtLimit())
		}
		if err != nil {
			root.Remove(relPath)
			return err
		}
		*left -= n
		return nil
	})
	if err != nil {
		return err
	}
	if origin.found {
		if err := writeQuarantine(filepath.Join(dir, rel), origin.value); err != nil {
			root.Remove(rel)
			return err
		}
	}
	return nil
}
