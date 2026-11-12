package settings

import (
	"fmt"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// TestFavoriteRowsStayCompact keeps the favorite folders usable with an archive of
// folders: the rows stay in their list and the add button stays on the pane.
func TestFavoriteRowsStayCompact(t *testing.T) {
	model := readyApp(t)
	model.cfg.FavoriteFolders = nil
	for i := 0; i < 10; i++ {
		label := fmt.Sprintf("Folder %02d", i+1)
		model.cfg.FavoriteFolders = append(model.cfg.FavoriteFolders, config.FavoriteFolder{
			ID: fmt.Sprintf("favorite-%d", i), Path: "/tmp/" + label, Label: label, Enabled: true,
		})
	}
	model.selectPane(paneFolders)
	screen := ui.NewTester(model.view, mainWidth, mainHeight)

	first, ok := screen.Find("Folder 01")
	if !ok {
		t.Fatal("the first folder row is missing")
	}
	if !slices.Contains(screen.Texts(), locales.QuickAccess()) {
		t.Error("the direct-entry check box has no visible label")
	}
	hint, ok := screen.Find(locales.FavoriteFoldersHint())
	if !ok {
		t.Fatal("the arrange hint is missing")
	}
	if hint.Y < first.Y {
		t.Error("the hint sits above the list; it is a note under the list, as on the Open With pane")
	}
	if _, ok := screen.Find(locales.AddFolder()); !ok {
		t.Fatal("the Add Favorite Folder button is missing")
	}
}
