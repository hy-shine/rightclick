package settings

import (
	"os"
	"testing"

	"github.com/egoist/mygo/i18n"
)

// TestMain shows the catalogs in Simplified Chinese, the language that most
// assertions in this package check. Tests that check English call useLocale.
func TestMain(m *testing.M) {
	i18n.SetLocaleFunc(func() string { return "zh-Hans" })
	os.Exit(m.Run())
}

// useLocale shows the catalogs in tag until the test ends, then restores the
// language that was in effect before.
func useLocale(t *testing.T, tag string) {
	t.Helper()
	previous := i18n.Locale()
	i18n.SetLocaleFunc(func() string { return tag })
	t.Cleanup(func() { i18n.SetLocaleFunc(func() string { return previous }) })
}
