package settings

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestMenuReordersShortcutsToTheTop(t *testing.T) {
	model := readyApp(t)
	model.cfg.Programs = []string{"Cursor"}
	model.setShortcut("app:Cursor", true)
	at := slices.IndexFunc(model.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == "app:Cursor" })
	model.reorderGroup(-1, []int{at}, 0)
	if model.cfg.MenuItems[0].ID != "app:Cursor" {
		t.Fatal("shortcut did not reach the top of the main entries")
	}
	model.cfg.MenuItems[0].Label = "Work Editor"
	model.saveConfig()
	saved, err := config.Load(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.MenuItems[0].Label != "Work Editor" || model.previewTopItems(ctxBlank)[0].label != "Work Editor" {
		t.Fatal("saved name or preview order differs from menu order")
	}
}

func TestFavoriteFoldersShortcutsFollowTheEntry(t *testing.T) {
	model := readyApp(t)
	path := t.TempDir()
	if err := model.addFavorite(path); err != nil {
		t.Fatal(err)
	}
	if err := model.addFavorite(path); err != nil {
		t.Fatal(err)
	}
	if len(model.cfg.FavoriteFolders) != 1 {
		t.Fatal("favorite duplicated")
	}
	folder := model.cfg.FavoriteFolders[0]
	model.setShortcut("folder:"+folder.ID, true)
	if !previewHas(model, ctxFile, folder.Label) {
		t.Fatal("favorite shortcut missing from preview")
	}
	model.cfg.FavoriteFolders[0].Enabled = false
	if previewHas(model, ctxFile, folder.Label) || previewHas(model, ctxFile, locales.FavoriteFolders()) {
		t.Fatal("disabled favorites remain in preview")
	}
	model.removeFavorite(folder.ID)
	if model.hasShortcut("folder:" + folder.ID) {
		t.Fatal("removing favorite did not remove its shortcut")
	}
}

func TestFavoritePaneAndRename(t *testing.T) {
	model := readyApp(t)
	path := t.TempDir()
	if err := model.addFavorite(path); err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneFolders)
	tester := ui.NewTester(model.view, minWidth, minHeight)
	if _, found := tester.Find(filepath.Base(path)); !found {
		t.Fatal("favorite is missing from its pane")
	}
	if err := tester.Click(locales.EditMenuName()); err != nil {
		t.Fatal(err)
	}
	model.favoriteName = "Project Archive"
	tester.Frame()
	if err := tester.Click(locales.SaveChanges()); err != nil {
		t.Fatal(err)
	}
	if model.cfg.FavoriteFolders[0].Label != "Project Archive" {
		t.Fatal("favorite rename was not saved")
	}
}

func TestFavoriteSwitchesSaveIndependently(t *testing.T) {
	model := readyApp(t)
	if err := model.addFavorite(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	folder := model.cfg.FavoriteFolders[0]
	model.selectPane(paneFolders)
	screen := ui.NewTester(model.view, minWidth, minHeight)
	for _, step := range []struct {
		label    string
		enabled  bool
		shortcut bool
	}{
		{locales.ShortcutFor(folder.Label), true, true},
		{locales.EnableLabel(folder.Label), false, true},
		{locales.ShortcutFor(folder.Label), false, false},
		{locales.EnableLabel(folder.Label), true, false},
	} {
		if err := screen.Click(step.label); err != nil {
			t.Fatal(err)
		}
		saved, err := config.Load(model.store.ConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		shortcut := slices.ContainsFunc(saved.MenuItems, func(item config.MenuItem) bool {
			return item.ID == "folder:"+folder.ID
		})
		if saved.FavoriteFolders[0].Enabled != step.enabled || shortcut != step.shortcut {
			t.Fatalf("after %q: enabled=%v shortcut=%v, want %v %v", step.label, saved.FavoriteFolders[0].Enabled, shortcut, step.enabled, step.shortcut)
		}
	}
}
