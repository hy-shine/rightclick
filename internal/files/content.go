package files

import (
	"bytes"
	"errors"
	"io"
	"os"
	"unicode/utf8"
)

var (
	ErrSymlinkNotSupported = errors.New("symbolic links are not supported")
	ErrNotRegularFile      = errors.New("not a regular file")
	ErrFileTooLarge        = errors.New("file is too large")
	ErrNotTextFile         = errors.New("file is not text")
	ErrNotUTF8Text         = errors.New("file content is not valid UTF-8")
)

// ReadTextContent reads up to maxBytes from a regular file at path.
// It rejects symbolic links, non-regular files, files exceeding maxBytes,
// and files containing NUL bytes or invalid UTF-8 sequences.
func ReadTextContent(path string, maxBytes int64) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", ErrSymlinkNotSupported
	}
	if !info.Mode().IsRegular() {
		return "", ErrNotRegularFile
	}
	if info.Size() > maxBytes {
		return "", ErrFileTooLarge
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Limit reader to maxBytes + 1 to detect concurrently growing files.
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxBytes {
		return "", ErrFileTooLarge
	}
	if len(data) == 0 {
		return "", nil
	}
	if bytes.IndexByte(data, 0) != -1 {
		return "", ErrNotTextFile
	}
	if !utf8.Valid(data) {
		return "", ErrNotUTF8Text
	}
	return string(data), nil
}
