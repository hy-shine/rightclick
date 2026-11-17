package settings

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/files"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestMenuRenameButtonSavesTheVisibleEntry(t *testing.T) {
	model := readyApp(t)
	screen := ui.NewTester(model.view, minWidth, minHeight)
	if err := screen.Click(locales.EditItem(locales.MenuCopyPath())); err != nil {
		t.Fatal(err)
	}
	if !model.menuNameOpen || model.menuNameID != "copy-path" {
		t.Fatalf("rename opened for %q, want copy-path", model.menuNameID)
	}
	model.menuName = "Copy Full Path"
	screen.Frame()
	if err := screen.Click(locales.SaveChanges()); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Load(model.store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range saved.MenuItems {
		if item.ID == "copy-path" && item.Label != "Copy Full Path" {
			t.Fatalf("saved menu name = %q", item.Label)
		}
	}
	if model.menuNameOpen || !screen.HasText("Copy Full Path") {
		t.Fatal("the editor did not show the saved menu name")
	}
}

func TestCreationEmptySearchReturnsToTheList(t *testing.T) {
	model := readyApp(t)
	model.selectPane(paneTemplates)
	model.creationSearch = "unmatched-format"
	screen := ui.NewTester(model.view, minWidth, minHeight)
	if _, found := screen.Find(locales.FileTypesList()); found {
		t.Fatal("empty results still show a blank table")
	}
	if err := screen.Click(locales.ClearSearch()); err != nil {
		t.Fatal(err)
	}
	if model.creationSearch != "" {
		t.Fatal("the search was not cleared")
	}
	if _, found := screen.Find(locales.EditItem(locales.FileTypeWord())); !found {
		t.Fatal("clearing the search did not restore editable file types")
	}
}

func TestMenuColumnHeadingsAlignWithTheirControls(t *testing.T) {
	for _, locale := range []string{"en", "zh-Hans", "zh-Hant"} {
		t.Run(locale, func(t *testing.T) {
			useLocale(t, locale)
			model := readyApp(t)
			screen := ui.NewTester(model.view, minWidth, minHeight)
			for _, pair := range [][2]string{
				{locales.ColContext(), locales.WhereBlankFolder()},
				{locales.ColLocation(), locales.LocationLabel(locales.MenuNewFile())},
			} {
				heading, headingOK := screen.Find(pair[0])
				cell, cellOK := screen.Find(pair[1])
				if !headingOK || !cellOK || math.Abs(float64(heading.X-cell.X)) > 1 {
					t.Errorf("heading %q at %+v does not align with %q at %+v", pair[0], heading, pair[1], cell)
				}
			}
		})
	}
}

func TestPreviewSubmenuTracksTheParentEntry(t *testing.T) {
	for _, extra := range []bool{false, true} {
		model := readyApp(t)
		if extra {
			model.cfg.Programs = []string{"Example Editor"}
			model.cfg.MenuItems = append(model.cfg.MenuItems, config.MenuItem{ID: "app:Example Editor", Enabled: true, Location: config.FileTypeMain})
		}
		screen := ui.NewTester(model.menuStage, 520, 640)
		parent, parentOK := screen.Find(locales.AppName())
		groups := model.previewGroups(model.previewCtx)
		child, childOK := screen.Find(groups[0][0].label)
		if !parentOK || !childOK {
			t.Fatal("the parent or submenu entry is missing")
		}
		if delta := parent.Y + parent.H/2 - child.Y - child.H/2; math.Abs(float64(delta)) > 1 {
			t.Errorf("extra entry=%v: submenu starts %.1fpt away from its parent", extra, delta)
		}
	}
}

func TestPreviewContextStaysAvailableWhileScrolling(t *testing.T) {
	model := readyApp(t)
	model.menuPreviewOpen = true
	screen := ui.NewTester(model.view, minWidth, minHeight)
	before, found := screen.Find(locales.ContextBlank())
	if !found {
		t.Fatal("preview context is missing")
	}
	screen.Scroll(minWidth/2, minHeight/2, 0, 1000)
	after, found := screen.Find(locales.ContextBlank())
	if !found || before.Y != after.Y {
		t.Fatal("scrolling the menu hides the context picker")
	}
	if err := screen.Click(locales.ContextBlank()); err != nil {
		t.Fatal(err)
	}
	if model.previewCtx != ctxBlank {
		t.Fatal("the pinned context picker did not change the preview")
	}
}

func TestTemplateEditorKeepsFooterOnOneLine(t *testing.T) {
	for _, locale := range []string{"en", "zh-Hans", "zh-Hant"} {
		t.Run(locale, func(t *testing.T) {
			useLocale(t, locale)
			model := readyApp(t)
			model.selectPane(paneTemplates)
			model.templates = []files.Template{{Name: "Meeting Notes.md", Content: "# Notes"}}
			model.openCreationEditor(creationRow{index: 0, template: true})
			screen := ui.NewTester(model.view, minWidth, minHeight)
			save, saveOK := screen.Find(locales.SaveChanges())
			cancel, cancelOK := screen.Find(locales.Cancel())
			if !saveOK || !cancelOK || save.Y != cancel.Y || save.Y+save.H > minHeight-16 {
				t.Fatalf("footer wrapped or clipped: save=%+v cancel=%+v", save, cancel)
			}
		})
	}
}

func TestCreationErrorStaysVisibleInShortWindow(t *testing.T) {
	model := readyApp(t)
	model.selectPane(paneTemplates)
	model.newFileType()
	screen := ui.NewTester(model.view, minWidth, minHeight)
	if err := screen.Click(locales.SaveChanges()); err != nil {
		t.Fatal(err)
	}
	message, found := screen.Find(locales.InvalidName())
	save, saveOK := screen.Find(locales.SaveChanges())
	if !found || !saveOK || message.Y < 0 || message.Y+message.H > save.Y || save.Y+save.H > minHeight {
		t.Fatalf("error is not visible above save: message=%+v save=%+v", message, save)
	}
}
