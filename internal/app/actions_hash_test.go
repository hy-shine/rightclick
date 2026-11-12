package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hy-shine/rightclick/internal/locales"
)

// helloDigests are the digests of the text "hello" under each algorithm.
var helloDigests = map[string]string{
	"SHA-256": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
	"SHA-1":   "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
	"MD5":     "5d41402abc4b2a76b9719d911017c592",
}

// TestCompareFilesUsesTheChosenAlgorithm reports whether two files match under the
// chosen algorithm, lists both digests, and refuses anything but two files.
func TestCompareFilesUsesTheChosenAlgorithm(t *testing.T) {
	dir := t.TempDir()
	first, second, other := filepath.Join(dir, "a.txt"), filepath.Join(dir, "b.txt"), filepath.Join(dir, "c.txt")
	for path, content := range map[string]string{first: "hello", second: "hello", other: "world"} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	title, body, err := compareFiles("MD5", []string{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if want := locales.HashCompareTitle("MD5", locales.VerdictSame()); title != want {
		t.Errorf("matching files: title %q, want %q", title, want)
	}
	if !strings.Contains(body, helloDigests["MD5"]) {
		t.Errorf("body %q does not list the MD5 digest", body)
	}
	if title, _, err = compareFiles("MD5", []string{first, other}); err != nil {
		t.Fatal(err)
	}
	if want := locales.HashCompareTitle("MD5", locales.VerdictDifferent()); title != want {
		t.Errorf("different files: title %q, want %q", title, want)
	}
	if _, _, err := compareFiles("MD5", []string{first}); err == nil {
		t.Error("one file was accepted for a comparison")
	}
}
