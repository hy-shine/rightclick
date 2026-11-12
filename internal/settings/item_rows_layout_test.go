package settings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// TestItemListsShareOneRowHeight checks that the favorite folders and the programs put
// their rows at the same pitch, so the two lists look alike.
func TestItemListsShareOneRowHeight(t *testing.T) {
	dir := t.TempDir()
	first, second := filepath.Join(dir, "alpha"), filepath.Join(dir, "beta")
	for _, p := range []string{first, second} {
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	m := readyApp(t)
	m.cfg.Programs = []string{"Cursor", "Warp"}
	m.cfg.FavoriteFolders = []config.FavoriteFolder{
		{ID: "f1", Path: first, Label: "Folder 01", Enabled: true},
		{ID: "f2", Path: second, Label: "Folder 02", Enabled: true},
	}
	panes := []struct {
		pane  string
		pairs [][2]string
	}{
		{paneFolders, [][2]string{{"Folder 01", "Folder 02"}}},
		{paneOpeners, [][2]string{{"Cursor", "Warp"}}},
	}
	for _, p := range panes {
		m.selectPane(p.pane)
		screen := ui.NewTester(m.view, mainWidth, mainHeight)
		for _, pair := range p.pairs {
			a, okA := screen.Find(pair[0])
			b, okB := screen.Find(pair[1])
			if !okA || !okB {
				t.Fatalf("%s: both rows must be on the pane", p.pane)
			}
			if pitch := b.Y - a.Y; pitch != itemRowHeight {
				t.Errorf("%s: row pitch = %.1f pt, want %d pt", pair[0], pitch, itemRowHeight)
			}
		}
	}
}

// TestFavoriteRowActionsAreIconTargets checks that the edit and delete actions of a
// favorite folder are the icon buttons every list uses, at the same size.
func TestFavoriteRowActionsAreIconTargets(t *testing.T) {
	m := readyApp(t)
	m.cfg.FavoriteFolders = []config.FavoriteFolder{{ID: "f1", Path: "/tmp", Label: "Folder 01", Enabled: true}}
	m.selectPane(paneFolders)
	screen := ui.NewTester(m.view, mainWidth, mainHeight)
	for _, label := range []string{locales.EditMenuName(), locales.Delete()} {
		target, ok := screen.Find(label)
		if !ok {
			t.Fatalf("%q button is missing", label)
		}
		if target.W != 32 || target.H != 26 {
			t.Errorf("%q target = %.0fx%.0f pt, want 32x26", label, target.W, target.H)
		}
	}
}

// TestFolderToolbarSitsCloseToItsList keeps the toolbar of the folder pane from floating
// away from the list it acts on: the space above it is the pane's usual 20 points, and
// the list starts close under it. The distances are measured between texts, which sit a few
// points inside their buttons, so the limits are 32.
func TestFolderToolbarSitsCloseToItsList(t *testing.T) {
	m := readyApp(t)
	m.cfg.FavoriteFolders = []config.FavoriteFolder{{ID: "f1", Path: "/tmp", Label: "Folder 01", Enabled: true}}
	m.selectPane(paneFolders)
	screen := ui.NewTester(m.view, mainWidth, mainHeight)
	_, okTitle := screen.Find(locales.PaneFolders())
	intro, okIntro := screen.Find(locales.FoldersIntro())
	add, okAdd := screen.Find(locales.AddFolder())
	row, okRow := screen.Find("Folder 01")
	if !okTitle || !okIntro || !okAdd || !okRow {
		t.Fatal("title, intro, add action and first row must be on the pane")
	}
	if above := add.Y - (intro.Y + intro.H); above > 32 {
		t.Errorf("%.1f pt between the intro and the toolbar, want at most 32", above)
	}
	if below := row.Y - (add.Y + add.H); below > 32 {
		t.Errorf("%.1f pt between the toolbar and the list, want at most 32", below)
	}
}
