package settings

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"rightclick/config"
	"rightclick/files"
	"rightclick/locales"
)

func TestFileTypeShortcutCanBeAddedAndRemovedFromList(t *testing.T) {
	model := readyApp(t)
	model.selectPane(paneTemplates)
	tester := ui.NewTester(model.view, minWidth, minHeight)
	label := locales.ShortcutFor(locales.FileTypeWord())
	if err := tester.Click(label); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Load(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(saved.MenuItems, func(item config.MenuItem) bool { return item.ID == "type:word" && item.Location == config.FileTypeMain }) {
		t.Fatal("file type shortcut was not saved")
	}
	if !previewHas(model, ctxBlank, locales.NewShortcut(locales.FileTypeWord())) || previewHas(model, ctxFile, locales.NewShortcut(locales.FileTypeWord())) {
		t.Fatal("shortcut preview does not follow the creation context")
	}
	if err := tester.Click(label); err != nil {
		t.Fatal(err)
	}
	if model.hasShortcut("type:word") || !model.cfg.FileTypes[0].Enabled {
		t.Fatal("removing a shortcut should preserve its enabled file type")
	}
}

func TestMenuReordersAcrossCategoriesAndShortcuts(t *testing.T) {
	model := readyApp(t)
	model.setShortcut("app:Cursor", true)
	last := len(model.cfg.MenuItems) - 1
	model.reorderGroup(-1, []int{last}, 0)
	if model.cfg.MenuItems[0].ID != "app:Cursor" {
		t.Fatal("shortcut cannot move across the built-in categories")
	}
	model.cfg.MenuItems[0].Label = "工作编辑器"
	model.saveConfig()
	saved, err := config.Load(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.MenuItems[0].Label != "工作编辑器" || model.previewTopItems(ctxBlank)[0].label != "工作编辑器" {
		t.Fatal("saved name or preview order differs from menu order")
	}
}

func TestFavoriteFoldersDoNotChangeWatchScope(t *testing.T) {
	model := readyApp(t)
	before := slices.Clone(model.cfg.WatchFolders)
	path := t.TempDir()
	if err := model.addFavorite(path); err != nil {
		t.Fatal(err)
	}
	if err := model.addFavorite(path); err != nil {
		t.Fatal(err)
	}
	if len(model.cfg.FavoriteFolders) != 1 || !slices.Equal(before, model.cfg.WatchFolders) {
		t.Fatal("favorite duplicated or watch scope changed")
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
	if model.hasShortcut("folder:"+folder.ID) || !slices.Equal(before, model.cfg.WatchFolders) {
		t.Fatal("removing favorite did not remove its shortcut or changed scope")
	}
}

func TestTemplateRenameKeepsShortcut(t *testing.T) {
	model := readyApp(t)
	model.templates = []files.Template{{Name: "original.md", Content: "# Notes"}}
	model.saveTemplates()
	model.setShortcut("template:original.md", true)
	model.openCreationEditor(creationRow{index: 0, template: true})
	model.editCreation.DefaultName = "renamed"
	model.saveCreationEditor()
	if !model.hasShortcut("template:renamed.md") || model.hasShortcut("template:original.md") {
		t.Fatal("renaming template left a stale shortcut")
	}
	model.tplSel = 0
	model.deleteTemplate()
	if model.hasShortcut("template:renamed.md") {
		t.Fatal("deleted template left a stale shortcut")
	}
}

func TestFavoritePaneAndRename(t *testing.T) {
	model := readyApp(t)
	path := t.TempDir()
	if err := model.addFavorite(path); err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneFavorites)
	tester := ui.NewTester(model.view, minWidth, minHeight)
	if _, found := tester.Find(filepath.Base(path)); !found {
		t.Fatal("favorite is missing from its pane")
	}
	if err := tester.Click(locales.EditMenuName()); err != nil {
		t.Fatal(err)
	}
	model.favoriteName = "项目归档"
	tester.Frame()
	if err := tester.Click(locales.SaveChanges()); err != nil {
		t.Fatal(err)
	}
	if model.cfg.FavoriteFolders[0].Label != "项目归档" {
		t.Fatal("favorite rename was not saved")
	}
}
