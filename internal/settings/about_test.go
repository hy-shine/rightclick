package settings

import (
	"bytes"
	"os"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

func TestGeneralPaneIncludesAboutInformation(t *testing.T) {
	for _, locale := range []string{"en", "zh-Hans", "zh-Hant"} {
		t.Run(locale, func(t *testing.T) {
			useLocale(t, locale)
			for _, size := range []struct {
				name   string
				width  int
				height int
			}{
				{"minimum", minWidth, minHeight},
				{"default", mainWidth, mainHeight},
			} {
				t.Run(size.name, func(t *testing.T) {
					model := readyApp(t)
					screen := ui.NewTester(model.view, size.width, size.height)
					if _, found := screen.Find(locales.About()); found {
						t.Fatal("sidebar still has a separate about entry")
					}
					if err := screen.Click(locales.PaneGeneral()); err != nil {
						t.Fatal(err)
					}
					if model.currentPane() != paneGeneral {
						t.Fatalf("selected pane = %q, want %q", model.currentPane(), paneGeneral)
					}
					for _, label := range []string{locales.About(), locales.VersionLine(appVersion()), locales.AboutBlurb(), locales.OpenGitHub(), locales.CheckUpdate()} {
						bounds, found := screen.Find(label)
						if !found || bounds.X < sidebarWidth {
							t.Errorf("about content %q is outside the pane: bounds=%+v, found=%v", label, bounds, found)
						}
						if size.name == "default" && (bounds.Y < 0 || bounds.Y+bounds.H > float32(size.height)) {
							t.Errorf("about content %q is outside the default pane: bounds=%+v", label, bounds)
						}
					}
					if err := screen.Click(locales.PaneDefaults()); err != nil {
						t.Fatal(err)
					}
					if model.currentPane() != paneDefaults {
						t.Fatalf("selected pane = %q, want %q", model.currentPane(), paneDefaults)
					}
					if _, found := screen.Find(locales.VersionLine(appVersion())); found {
						t.Fatal("about content remains after opening default actions")
					}
				})
			}
		})
	}
}

func TestAboutLinkOpensItsDestination(t *testing.T) {
	var opened []string
	environment := testEnv()
	environment.openURL = func(url string) error {
		opened = append(opened, url)
		return nil
	}
	model, err := newApp(testStore(t), environment)
	if err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneGeneral)
	screen := ui.NewTester(model.view, mainWidth, mainHeight)
	if err := screen.Click(locales.OpenGitHub()); err != nil {
		t.Fatal(err)
	}
	if want := []string{repositoryURL}; !slices.Equal(opened, want) {
		t.Fatalf("opened URLs = %q, want %q", opened, want)
	}
}

func TestSettingsNavigationPreservesConfig(t *testing.T) {
	model := readyApp(t)
	before, err := os.ReadFile(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	screen := ui.NewTester(model.view, mainWidth, mainHeight)
	for _, pane := range []string{locales.PaneGeneral(), locales.PaneDefaults(), locales.PaneGeneral()} {
		if err := screen.Click(pane); err != nil {
			t.Fatal(err)
		}
	}
	after, err := os.ReadFile(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("navigating settings pages changed the saved settings")
	}
}
