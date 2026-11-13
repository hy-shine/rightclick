package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestAppearancePickerSavesAndAppliesChoice(t *testing.T) {
	for _, locale := range []string{"en", "zh-Hans", "zh-Hant"} {
		t.Run(locale, func(t *testing.T) {
			useLocale(t, locale)
			store := testStore(t)
			var applied []string
			environment := testEnv()
			environment.setAppearance = func(mode string) {
				saved, err := config.Load(store.ConfigPath)
				if err != nil {
					t.Fatal(err)
				}
				if saved.Appearance != mode {
					t.Fatalf("applied appearance = %q before saved appearance = %q", mode, saved.Appearance)
				}
				applied = append(applied, mode)
			}
			model, err := newApp(store, environment)
			if err != nil {
				t.Fatal(err)
			}
			model.cfg.Language = "en"
			model.cfg.DefaultHashAlgorithm = "MD5"
			expectedZip := config.ZipExclusions{Git: true, NodeModules: true}
			expectedBase := t.TempDir()
			model.cfg.ZipExclusions = expectedZip
			model.cfg.CopyBaseFolder = expectedBase
			if err := config.Save(store.ConfigPath, model.cfg); err != nil {
				t.Fatal(err)
			}
			model.selectPane(paneGeneral)
			screen := ui.NewTester(model.view, mainWidth, mainHeight)
			selectedLabel := locales.AppearanceSystem()
			for _, choice := range []struct {
				label string
				mode  string
			}{
				{locales.AppearanceDark(), "dark"},
				{locales.AppearanceLight(), "light"},
				{locales.AppearanceSystem(), ""},
				{locales.AppearanceDark(), "dark"},
			} {
				if err := screen.Click(selectedLabel); err != nil {
					t.Fatal(err)
				}
				if err := screen.Click(choice.label); err != nil {
					t.Fatal(err)
				}
				selectedLabel = choice.label
				saved, err := config.Load(store.ConfigPath)
				if err != nil {
					t.Fatal(err)
				}
				if saved.Appearance != choice.mode {
					t.Fatalf("saved appearance = %q, want %q", saved.Appearance, choice.mode)
				}
				if saved.DefaultHashAlgorithm != "MD5" || saved.ZipExclusions != expectedZip || saved.CopyBaseFolder != expectedBase {
					t.Fatalf("changing appearance changed action preferences: %+v", saved)
				}
			}
			if want := []string{"dark", "light", "", "dark"}; !slices.Equal(applied, want) {
				t.Fatalf("applied appearances = %q, want %q", applied, want)
			}
			reopened, err := newApp(store, testEnv())
			if err != nil {
				t.Fatal(err)
			}
			if reopened.cfg.Appearance != "dark" {
				t.Fatalf("reopened appearance = %q, want dark", reopened.cfg.Appearance)
			}
		})
	}
}

func TestDefaultHashSelectionSavesAndPreservesPreferences(t *testing.T) {
	model := readyApp(t)
	model.cfg.Appearance = "dark"
	model.cfg.ZipExclusions = config.ZipExclusions{DSStore: true, Git: true}
	model.cfg.CopyBaseFolder = t.TempDir()
	if err := config.Save(model.store.ConfigPath, model.cfg); err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneDefaults)
	screen := ui.NewTester(model.view, mainWidth, mainHeight)
	latest, err := config.Load(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	latest.CopyBaseFolder = t.TempDir()
	if err := config.Save(model.store.ConfigPath, latest); err != nil {
		t.Fatal(err)
	}
	selectedAlgorithm := model.cfg.DefaultHashAlgorithm
	for _, algorithm := range []string{"SHA-1", "MD5", "SHA-256", "MD5"} {
		if err := screen.Click(selectedAlgorithm); err != nil {
			t.Fatal(err)
		}
		if err := screen.Click(algorithm); err != nil {
			t.Fatal(err)
		}
		selectedAlgorithm = algorithm
		saved, err := config.Load(model.store.ConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		if saved.DefaultHashAlgorithm != algorithm {
			t.Fatalf("saved algorithm = %q, want %q", saved.DefaultHashAlgorithm, algorithm)
		}
		if saved.Appearance != "dark" || saved.ZipExclusions != latest.ZipExclusions || saved.CopyBaseFolder != latest.CopyBaseFolder {
			t.Fatalf("changing hash algorithm changed other preferences: %+v", saved)
		}
	}
	for _, pane := range []string{locales.PaneGeneral(), locales.PaneDefaults()} {
		if err := screen.Click(pane); err != nil {
			t.Fatal(err)
		}
	}
	if model.cfg.DefaultHashAlgorithm != "MD5" {
		t.Fatalf("algorithm after navigation = %q, want MD5", model.cfg.DefaultHashAlgorithm)
	}
	reopened, err := newApp(model.store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	if reopened.cfg.DefaultHashAlgorithm != "MD5" {
		t.Fatalf("reopened algorithm = %q, want MD5", reopened.cfg.DefaultHashAlgorithm)
	}
	if reopened.cfg.Appearance != "dark" || reopened.cfg.ZipExclusions != latest.ZipExclusions || reopened.cfg.CopyBaseFolder != latest.CopyBaseFolder {
		t.Fatalf("reopened action preferences differ: %+v", reopened.cfg)
	}
}
