package settings

import (
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/files"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestCompareModelMatchAndCopy(t *testing.T) {
	oldWrite, oldPause := writeClipboard, pauseAfterCopy
	t.Cleanup(func() { writeClipboard, pauseAfterCopy = oldWrite, oldPause })
	var copied string
	writeClipboard = func(text string) error { copied = text; return nil }
	pauseAfterCopy = func() {}

	m := &compareModel{
		algo: files.SHA256,
		slots: [2]compareSlot{
			{path: "/tmp/a.bin", size: 1024, sized: true, digest: "abcd1234abcd1234"},
			{path: "/tmp/b.bin", size: 1024, sized: true, digest: "abcd1234abcd1234"},
		},
	}
	decided, identical := m.match()
	if !decided || !identical {
		t.Fatalf("match() = (%v, %v), want (true, true)", decided, identical)
	}
	m.copySlot(0)
	if copied != "abcd1234abcd1234" || !m.slots[0].copied {
		t.Fatalf("copySlot copied %q, flag=%v", copied, m.slots[0].copied)
	}
	m.copySummary()
	if !strings.Contains(copied, "a.bin (SHA-256): abcd1234abcd1234") || !strings.Contains(copied, "b.bin (SHA-256): abcd1234abcd1234") {
		t.Fatalf("copySummary = %q", copied)
	}

	m.slots[1].digest = "ffff0000ffff0000"
	decided, identical = m.match()
	if !decided || identical {
		t.Fatalf("mismatch match() = (%v, %v), want (true, false)", decided, identical)
	}
}

func TestCompareWindowLayoutAcrossLocales(t *testing.T) {
	for _, lang := range []string{"en", "zh-Hans", "zh-Hant"} {
		t.Run(lang, func(t *testing.T) {
			useLocale(t, lang)
			m := &compareModel{
				algo: files.SHA256,
				slots: [2]compareSlot{
					{path: "/tmp/release-arm64.dmg", size: 12_400_000, sized: true, digest: strings.Repeat("a1b2c3d4", 8)},
					{path: "/tmp/release-x64.dmg", size: 12_800_000, sized: true, digest: strings.Repeat("e5f60718", 8)},
				},
			}
			screen := ui.NewTester(m.view, compareWidth, compareHeight)
			texts := screen.Texts()
			if !slices.Contains(texts, locales.CompareMismatch()) {
				t.Fatalf("%s: mismatch pill missing; texts=%q", lang, texts)
			}
			copyBtn, ok := screen.Find(locales.CompareCopyBoth())
			if !ok || copyBtn.Y+copyBtn.H > compareHeight || copyBtn.X+copyBtn.W > compareWidth {
				t.Fatalf("%s: copy button clipped: %+v", lang, copyBtn)
			}
		})
	}
}
