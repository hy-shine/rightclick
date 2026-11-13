package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTextContent(t *testing.T) {
	dir := t.TempDir()

	t.Run("NormalText", func(t *testing.T) {
		path := filepath.Join(dir, "normal.txt")
		if err := os.WriteFile(path, []byte("Hello, World!\nSecond line."), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := ReadTextContent(path, 256*1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "Hello, World!\nSecond line." {
			t.Errorf("got %q, want %q", got, "Hello, World!\nSecond line.")
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := ReadTextContent(path, 256*1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	t.Run("SizeLimitStatic", func(t *testing.T) {
		path := filepath.Join(dir, "large.txt")
		content := strings.Repeat("A", 100)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := ReadTextContent(path, 50)
		if !errors.Is(err, ErrFileTooLarge) {
			t.Errorf("expected ErrFileTooLarge, got: %v", err)
		}
	})

	t.Run("NULByteRejection", func(t *testing.T) {
		path := filepath.Join(dir, "binary.bin")
		// Text at first, then NUL in middle
		data := []byte("Valid text here\x00and binary following")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := ReadTextContent(path, 256*1024)
		if !errors.Is(err, ErrNotTextFile) {
			t.Errorf("expected ErrNotTextFile, got: %v", err)
		}
	})

	t.Run("InvalidUTF8Rejection", func(t *testing.T) {
		path := filepath.Join(dir, "invalid_utf8.bin")
		// 0xff 0xfe is invalid UTF-8
		data := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0xff, 0xfe}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := ReadTextContent(path, 256*1024)
		if !errors.Is(err, ErrNotUTF8Text) {
			t.Errorf("expected ErrNotUTF8Text, got: %v", err)
		}
	})

	t.Run("SymlinkRejection", func(t *testing.T) {
		target := filepath.Join(dir, "target.txt")
		if err := os.WriteFile(target, []byte("text"), 0o644); err != nil {
			t.Fatal(err)
		}
		symlink := filepath.Join(dir, "link.txt")
		if err := os.Symlink(target, symlink); err != nil {
			t.Fatal(err)
		}
		_, err := ReadTextContent(symlink, 256*1024)
		if !errors.Is(err, ErrSymlinkNotSupported) {
			t.Errorf("expected ErrSymlinkNotSupported, got: %v", err)
		}
	})

	t.Run("DirectoryRejection", func(t *testing.T) {
		sub := filepath.Join(dir, "sub")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		_, err := ReadTextContent(sub, 256*1024)
		if !errors.Is(err, ErrNotRegularFile) {
			t.Errorf("expected ErrNotRegularFile, got: %v", err)
		}
	})
}
