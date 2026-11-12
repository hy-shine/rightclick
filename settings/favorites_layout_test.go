package settings

import (
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/locales"
)

// TestFavoriteRowsStayCompact keeps the folder list usable with an archive of
// folders: the rows must stay at the shared height and ten of them plus the add
// button must still fit the default window.
func TestFavoriteRowsStayCompact(t *testing.T) {
	model := readyApp(t)
	model.cfg.FavoriteFolders = nil
	for i := 0; i < 10; i++ {
		label := fmt.Sprintf("Folder %02d", i+1)
		model.cfg.FavoriteFolders = append(model.cfg.FavoriteFolders, config.FavoriteFolder{
			ID: fmt.Sprintf("favorite-%d", i), Path: "/tmp/" + label, Label: label, Enabled: true,
		})
	}
	model.selectPane(paneFavorites)
	screen := ui.NewTester(model.view, mainWidth, mainHeight)

	first, ok := screen.Find("Folder 01")
	if !ok {
		t.Fatal("the first folder row is missing")
	}
	second, ok := screen.Find("Folder 02")
	if !ok {
		t.Fatal("the second folder row is missing")
	}
	if pitch := second.Y - first.Y; pitch > favoriteRowHeight {
		t.Errorf("row pitch = %.1f pt, want at most %d", pitch, favoriteRowHeight)
	}
	if !slices.Contains(screen.Texts(), locales.QuickAccess()) {
		t.Error("the direct-entry check box has no visible label")
	}
	hint, ok := screen.Find(locales.FavoriteArrangeHint())
	if !ok {
		t.Fatal("the arrange hint is missing")
	}
	if hint.Y > first.Y {
		t.Error("the hint explains the controls below the list instead of above it")
	}
	add, ok := screen.Find(locales.AddFavorite())
	if !ok {
		t.Fatal("the Add Favorite Folder button is missing")
	}
	if add.Y+add.H > mainHeight {
		t.Errorf("the add button ends at %.1f pt, past the window height %d", add.Y+add.H, mainHeight)
	}
}
