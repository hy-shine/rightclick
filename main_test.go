package main

import (
	"os"
	"testing"

	"github.com/egoist/mygo/i18n"

	"github.com/hy-shine/rightclick/config"
)

// TestMain shows the catalogs in English, the source language that tests assert
// against. Tests that check another language call useLocale.
func TestMain(m *testing.M) {
	i18n.SetLocaleFunc(func() string { return "en" })
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
	err := runAction("bogus", nil, nil, config.Config{})
	if err == nil || err.Error() != "Unknown action: bogus" {
		t.Errorf("runAction(bogus) error = %v, want %q", err, "Unknown action: bogus")
	}
}

func TestShouldOpenWindow(t *testing.T) {
	tests := []struct {
		name           string
		openedAtLogin  bool
		launchedForURL bool
		want           bool
	}{
		{"manual start", false, false, true},
		{"started at login", true, false, false},
		{"started by a link", false, true, false},
		{"login item asked to open a link", true, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldOpenWindow(tt.openedAtLogin, tt.launchedForURL); got != tt.want {
				t.Errorf("shouldOpenWindow(%v, %v) = %v, want %v",
					tt.openedAtLogin, tt.launchedForURL, got, tt.want)
			}
		})
	}
}
