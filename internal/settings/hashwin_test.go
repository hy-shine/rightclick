package settings

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/files"
)

func TestParseAlgorithm(t *testing.T) {
	for name, want := range map[string]files.Algorithm{
		"SHA-256": files.SHA256,
		"sha-256": files.SHA256,
		"SHA-1":   files.SHA1,
		"MD5":     files.MD5,
	} {
		got, ok := parseAlgorithm(name)
		if !ok || got != want {
			t.Errorf("parseAlgorithm(%q) = %q, %v; want %q", name, got, ok, want)
		}
	}
	if _, ok := parseAlgorithm("SHA-512"); ok {
		t.Error("an algorithm the menu does not offer was accepted")
	}
}

func TestWindowAlgorithmDefaultsToSHA256(t *testing.T) {
	if got, ok := windowAlgorithm(""); !ok || got != files.SHA256 {
		t.Errorf("windowAlgorithm(%q) = %q, %v; want SHA-256, true", "", got, ok)
	}
	if got, ok := windowAlgorithm("MD5"); !ok || got != files.MD5 {
		t.Errorf("windowAlgorithm(%q) = %q, %v; want MD5, true", "MD5", got, ok)
	}
	if _, ok := windowAlgorithm("SHA-512"); ok {
		t.Error("an algorithm the menu does not offer was accepted")
	}
}

func TestDisplayDirWritesTheHomeFolderAsTilde(t *testing.T) {
	t.Setenv("HOME", "/Users/someone")
	for path, want := range map[string]string{
		"/Users/someone/report.pdf":         "~",
		"/Users/someone/Desktop/report.pdf": "~/Desktop",
		"/Users/someone-else/report.pdf":    "/Users/someone-else",
		"/Volumes/Disk/report.pdf":          "/Volumes/Disk",
	} {
		if got := displayDir(path); got != want {
			t.Errorf("displayDir(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestShortDirKeepsTheFoldersNearestTheFile(t *testing.T) {
	short := "~/Desktop"
	if got := shortDir(short, 68); got != short {
		t.Errorf("shortDir(%q) = %q, want it unchanged", short, got)
	}
	long := "~/Library/Mobile Documents/com~apple~CloudDocs/工作/合同/2026年第三季度/采购合同扫描件/已签字版本"
	got := shortDir(long, 68)
	if !utf8.ValidString(got) {
		t.Errorf("shortDir cut inside a character: %q", got)
	}
	if !strings.HasPrefix(got, "…/") || !strings.HasSuffix(got, "/采购合同扫描件/已签字版本") {
		t.Errorf("shortDir(%q) = %q, want an ellipsis and then the last folders", long, got)
	}
	if textWidth(got) > 68 {
		t.Errorf("shortDir left a width of %d, want at most 68: %q", textWidth(got), got)
	}
}

func TestFormatSizeUsesDecimalUnits(t *testing.T) {
	for bytes, want := range map[int64]string{
		0:             "0 B",
		999:           "999 B",
		12_000:        "12 KB",
		2_400_000:     "2.4 MB",
		1_250_000_000: "1.25 GB",
	} {
		if got := formatSize(bytes); got != want {
			t.Errorf("formatSize(%d) = %q, want %q", bytes, got, want)
		}
	}
}

func TestDigestSpansKeepTheWholeDigest(t *testing.T) {
	digest := "3f786850e387550fdab836ed7e6dc881de23001b4f2b6f9e5f1c7a0d9b2e4c6a"
	var joined strings.Builder
	for _, span := range digestSpans(digest, ui.Hex("#000000"), ui.Hex("#888888")) {
		joined.WriteString(span.Text)
	}
	if joined.String() != digest {
		t.Errorf("the spans read %q, want the digest unchanged", joined.String())
	}
}

func TestCopyDigestWritesTheClipboard(t *testing.T) {
	oldWrite, oldPause := writeClipboard, pauseAfterCopy
	t.Cleanup(func() { writeClipboard, pauseAfterCopy = oldWrite, oldPause })
	var copied string
	writeClipboard = func(text string) error { copied = text; return nil }
	pauseAfterCopy = func() {}

	m := &hashModel{digest: "abc123"}
	m.copyDigest()
	if copied != "abc123" {
		t.Errorf("copied %q, want the digest", copied)
	}
	if !m.copied {
		t.Error("the model did not report the copy")
	}
}

func TestCopyDigestDoesNothingWhileComputing(t *testing.T) {
	oldWrite := writeClipboard
	t.Cleanup(func() { writeClipboard = oldWrite })
	called := false
	writeClipboard = func(string) error { called = true; return nil }

	m := &hashModel{busy: true}
	m.copyDigest()
	if called {
		t.Error("a digest that is still computing was copied")
	}
}
