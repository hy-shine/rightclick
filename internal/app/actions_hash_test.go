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

// TestHashCopiesTheChosenAlgorithm copies the digest of the algorithm the menu
// sent, and names that algorithm in the notification.
func TestHashCopiesTheChosenAlgorithm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hello.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	var copied, title string
	oldCopy, oldNotify := copyText, notify
	t.Cleanup(func() { copyText, notify = oldCopy, oldNotify })
	copyText = func(text string) error { copied = text; return nil }
	notify = func(gotTitle, _ string) { title = gotTitle }

	for algo, digest := range helloDigests {
		copied, title = "", ""
		if err := hashItem(algo, []string{path}); err != nil {
			t.Fatalf("hashItem(%q) error = %v", algo, err)
		}
		if copied != digest {
			t.Errorf("%s copied %q, want %q", algo, copied, digest)
		}
		if !strings.Contains(title, algo) {
			t.Errorf("%s notification title %q does not name the algorithm", algo, title)
		}
	}
}

// TestHashDefaultsToSHA256AndRefusesUnknownAlgorithms keeps menu links without an
// algorithm working as SHA-256, and copies nothing for a name it does not know.
func TestHashDefaultsToSHA256AndRefusesUnknownAlgorithms(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hello.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	var copied string
	oldCopy, oldNotify := copyText, notify
	t.Cleanup(func() { copyText, notify = oldCopy, oldNotify })
	copyText = func(text string) error { copied = text; return nil }
	notify = func(string, string) {}

	if err := hashItem("", []string{path}); err != nil {
		t.Fatal(err)
	}
	if copied != helloDigests["SHA-256"] {
		t.Errorf("empty algorithm copied %q, want the SHA-256 digest", copied)
	}
	copied = ""
	if err := hashItem("SHA-512", []string{path}); err == nil {
		t.Error("an unknown algorithm was accepted")
	}
	if copied != "" {
		t.Errorf("an unknown algorithm copied %q", copied)
	}
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
