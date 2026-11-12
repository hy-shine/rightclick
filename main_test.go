package main

import (
	"os"
	"testing"

	"github.com/egoist/mygo/i18n"

	"rightclick/config"
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

// TestErrorsFollowTheLanguage checks that a message shown to the user is in the
// language the app shows.
func TestErrorsFollowTheLanguage(t *testing.T) {
	useLocale(t, "en")
	err := runAction("bogus", nil, nil, config.Config{})
	if err == nil || err.Error() != "Unknown action: bogus" {
		t.Errorf("runAction(bogus) error = %v, want %q", err, "Unknown action: bogus")
	}
}
