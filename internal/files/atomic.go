package files

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// writeAtomic writes a file at path through a temporary file beside it, and
// renames the temporary file onto path only after write and close succeed. A
// failure leaves no partial file at path.
func writeAtomic(path string, perm fs.FileMode, write func(w io.Writer) error) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.part")
	if err != nil {
		return err
	}
	// After a successful rename the temporary name no longer exists, so this
	// only removes the file when something failed.
	defer os.Remove(tmp.Name())
	if err := write(tmp); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
