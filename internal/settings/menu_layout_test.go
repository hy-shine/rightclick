package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestMenuNamesStayOnOneLineInNarrowEditor(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tester := ui.NewTester(func(context *ui.Context) {
		context.SetTheme(paperTheme(true))
		ui.Column(context).Width(680).Children(func() {
			for group := 2; group < len(menuGroups); group++ {
				model.menuGroup(context, group, model.groupRows(group))
			}
		})
	}, 680, 800)
	for _, label := range []string{"Open with Terminal", "Compress to ZIP", "Compare File Hashes", "Show or Hide Hidden Files (⌘⇧.)"} {
		bounds, found := tester.Find(label)
		if !found || bounds.H > 22 || bounds.W < 100 {
			t.Errorf("menu name %q is not a readable single line: found=%v bounds=%+v", label, found, bounds)
		}
	}
}

// TestMenuRowsKeepTheConditionInItsOwnColumn checks the compact row: one line
// per entry, with the condition beside the name instead of under it.
func TestMenuRowsKeepTheConditionInItsOwnColumn(t *testing.T) {
	model := readyApp(t)
	model.selectPane(paneMenu)
	screen := ui.NewTester(model.view, minWidth, minHeight)
	screen.Frame()
	name, ok := screen.Find(locales.MenuNewFile())
	if !ok {
		t.Fatal("the first menu entry is missing")
	}
	condition, ok := screen.Find(locales.WhereBlankFolder())
	if !ok {
		t.Fatal("the condition column is missing")
	}
	if condition.X < name.X+name.W {
		t.Errorf("the condition at x=%.1f overlaps the name ending at %.1f", condition.X, name.X+name.W)
	}
	nameMiddle := name.Y + name.H/2
	conditionMiddle := condition.Y + condition.H/2
	if diff := nameMiddle - conditionMiddle; diff > 2 || diff < -2 {
		t.Errorf("name and condition sit %.1f pt apart vertically", diff)
	}
	rows := model.menuDisplayRows(-1)
	second, ok := screen.Find(model.menuEntry(model.cfg.MenuItems[rows[1]]).label)
	if !ok {
		t.Fatal("the second menu entry is missing")
	}
	if pitch := second.Y - name.Y; pitch > menuRowHeight {
		t.Errorf("row pitch = %.1f pt, want at most %d", pitch, menuRowHeight)
	}
}

// TestMenuSectionsSeparateMainAndSubmenu checks that the Finder Menu and
// RightClick Submenu sit in separate labeled sections rather than sharing one flat list.
func TestMenuSectionsSeparateMainAndSubmenu(t *testing.T) {
	model := readyApp(t)
	model.selectPane(paneMenu)
	screen := ui.NewTester(model.view, minWidth, minHeight)
	screen.Frame()
	mainHeader, okMain := screen.Find(locales.MenuSectionMain())
	subHeader, okSub := screen.Find(locales.MenuSectionSub())
	if !okMain || !okSub {
		t.Fatalf("missing section headers: main=%v sub=%v", okMain, okSub)
	}
	if subHeader.Y <= mainHeader.Y+mainHeader.H {
		t.Errorf("submenu section at y=%.1f does not sit below main section ending at %.1f", subHeader.Y, mainHeader.Y+mainHeader.H)
	}
	rows := model.locationRows(config.FileTypeMain)
	if len(rows) < 2 {
		t.Fatal("expected at least 2 main-menu items")
	}
	first, okFirst := screen.Find(model.menuEntry(model.cfg.MenuItems[rows[0]]).label)
	second, okSecond := screen.Find(model.menuEntry(model.cfg.MenuItems[rows[1]]).label)
	if !okFirst || !okSecond {
		t.Fatal("main-menu entries missing")
	}
	if pitch := second.Y - first.Y; pitch != menuRowHeight {
		t.Errorf("row pitch = %.1f pt, want %d", pitch, menuRowHeight)
	}
}
