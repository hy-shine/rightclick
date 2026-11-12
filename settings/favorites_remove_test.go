package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/config"
)

// TestRemoveFavoriteButtonDeletesIt clicks the icon remove button of a favorite
// folder. The removal must wait until the list is built: removing inside the row
// made the list read a row that no longer existed and crash.
func TestRemoveFavoriteButtonDeletesIt(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.cfg.FavoriteFolders = []config.FavoriteFolder{{ID: "fav-a", Path: "/tmp", Label: "Only", Enabled: true}}
	m.selectPane(paneFavorites)
	if err := ui.NewTester(m.view, mainWidth, mainHeight).Click("Remove"); err != nil {
		t.Fatal(err)
	}
	if len(m.cfg.FavoriteFolders) != 0 {
		t.Fatalf("clicking the remove button left %d favorites", len(m.cfg.FavoriteFolders))
	}
}
