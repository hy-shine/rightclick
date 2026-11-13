package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestArchiveExclusionSwitchesSave(test *testing.T) {
	model := readyApp(test)
	screen := ui.NewTester(model.view, mainWidth, mainHeight)
	if err := screen.Click(locales.PaneDefaults()); err != nil {
		test.Fatal(err)
	}
	if model.currentPane() != paneDefaults {
		test.Fatalf("selected pane = %q, want %q", model.currentPane(), paneDefaults)
	}
	for _, name := range []string{".DS_Store", ".git", "node_modules"} {
		if err := screen.Click(locales.ZipExcludeToggle(name)); err != nil {
			test.Fatal(err)
		}
	}
	saved, err := config.Load(model.store.ConfigPath)
	if err != nil {
		test.Fatal(err)
	}
	if saved.ZipExclusions != (config.ZipExclusions{Git: true, NodeModules: true}) {
		test.Fatalf("exclusion switches not saved: %+v", saved.ZipExclusions)
	}
}
