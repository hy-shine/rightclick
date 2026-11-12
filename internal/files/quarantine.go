package files

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// quarantineAttr is the extended attribute that macOS puts on a file downloaded from
// the internet. Gatekeeper checks such a file before it runs for the first time.
const quarantineAttr = "com.apple.quarantine"

// xattrTool is the system tool that reads and writes extended attributes. Using it
// keeps the module free of extra dependencies.
const xattrTool = "/usr/bin/xattr"

// quarantineOrigin keeps the quarantine attribute of the first file that has one.
// A copy, an archive made from such a file, or a file unpacked from a downloaded
// archive keeps that attribute too, so that macOS still checks it before it runs.
type quarantineOrigin struct {
	value []byte
	found bool
}

// note reads the quarantine attribute of path, unless one has been found already.
func (o *quarantineOrigin) note(path string) error {
	if o.found {
		return nil
	}
	value, ok, err := readQuarantine(path)
	if err != nil {
		return err
	}
	o.value, o.found = value, ok
	return nil
}

// copyQuarantine gives dst the quarantine attribute of src, if src has one. A file
// without the attribute leaves dst without it.
func copyQuarantine(src, dst string) error {
	value, ok, err := readQuarantine(src)
	if err != nil || !ok {
		return err
	}
	return writeQuarantine(dst, value)
}

// readQuarantine returns the quarantine attribute of path. ok is false when path has
// none. Any other failure, such as a file that cannot be read, is an error, so that
// a missing attribute is never taken for a clean file by mistake.
func readQuarantine(path string) (value []byte, ok bool, err error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, false, err
	}
	var stderr bytes.Buffer
	cmd := exec.Command(xattrTool, "-px", quarantineAttr, abs)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(stderr.String(), "No such xattr") {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read the quarantine of %s: %w: %s", path, err, strings.TrimSpace(stderr.String()))
	}
	// xattr prints the value as hex pairs, split over lines.
	value, err = hex.DecodeString(strings.Join(strings.Fields(string(out)), ""))
	if err != nil {
		return nil, false, fmt.Errorf("read the quarantine of %s: %w", path, err)
	}
	return value, true, nil
}

// writeQuarantine sets the quarantine attribute of path to value.
func writeQuarantine(path string, value []byte) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	out, err := exec.Command(xattrTool, "-wx", quarantineAttr, hex.EncodeToString(value), abs).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set the quarantine of %s: %w: %s", path, err, strings.TrimSpace(string(out)))
	}
	return nil
}
