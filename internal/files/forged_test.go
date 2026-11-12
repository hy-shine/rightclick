package files

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hy-shine/rightclick/internal/locales"
)

// TestExtractFileStopsAtBudget gives extractFile more data than the budget left. The
// write is cut at the limit, the file is removed, and the error says where the
// extraction stopped. This is the second line of defense: archive/zip already refuses
// data that is longer than its header declares, so this path is reached only by a
// reader that does not check.
func TestExtractFileStopsAtBudget(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	left := int64(10)
	var origin quarantineOrigin
	err = extractFile(root, dir, 0o644, strings.NewReader("0123456789ABCDEF"), "big.txt", &left, &origin)
	if err == nil || err.Error() != locales.ArchiveStoppedAtLimit() {
		t.Fatalf("extractFile error = %v, want the stopped-at-limit message", err)
	}
	if exists(filepath.Join(dir, "big.txt")) {
		t.Error("big.txt was kept after the limit was reached")
	}
	if left != 10 {
		t.Errorf("budget left = %d after a refused file, want 10", left)
	}
}

// TestUnzipRefusesDataLongerThanHeader writes an entry whose header declares 10 bytes
// while its data holds 96. archive/zip rejects the surplus, so the extraction fails
// and leaves no file behind.
func TestUnzipRefusesDataLongerThanHeader(t *testing.T) {
	dir := t.TempDir()
	body := bytes.Repeat([]byte("abcdefgh"), 12)
	var packed bytes.Buffer
	fw, err := flate.NewWriter(&packed, flate.DefaultCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := fw.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.CreateRaw(&zip.FileHeader{
		Name:               "forged.txt",
		Method:             zip.Deflate,
		CompressedSize64:   uint64(packed.Len()),
		UncompressedSize64: 10,
		CRC32:              crc32.ChecksumIEEE(body),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(packed.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "forged.zip")
	if err := os.WriteFile(archive, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Unzip(archive); err == nil {
		t.Fatal("Unzip accepted data longer than the header declares")
	}
	if exists(filepath.Join(dir, "forged.txt")) {
		t.Error("forged.txt was left behind after the failed extraction")
	}
}
