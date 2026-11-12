package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/locales"
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
	for _, label := range []string{"Open in Terminal", locales.MenuOpenEditor(), "Compress to ZIP", "Compare File Hashes", "Show or Hide Hidden Files (⌘⇧.)"} {
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

// TestMenuRuleKeepsTheRowPitch checks that the rule between the main entries and
// the submenu entries adds no row: the first submenu entry sits one row below the
// last main entry, within half a point of the rule's own width.
func TestMenuRuleKeepsTheRowPitch(t *testing.T) {
	model := readyApp(t)
	model.selectPane(paneMenu)
	screen := ui.NewTester(model.view, minWidth, minHeight)
	screen.Frame()
	rows := model.menuDisplayRows(-1)
	split := slices.IndexFunc(rows, func(i int) bool { return model.cfg.MenuItems[i].Location != config.FileTypeMain })
	if split <= 0 {
		t.Fatal("the main and the submenu entries are not split")
	}
	last, ok := screen.Find(model.menuEntry(model.cfg.MenuItems[rows[split-1]]).label)
	if !ok {
		t.Fatal("the last main entry is missing")
	}
	first, ok := screen.Find(model.menuEntry(model.cfg.MenuItems[rows[split]]).label)
	if !ok {
		t.Fatal("the first submenu entry is missing")
	}
	if pitch := first.Y - last.Y; pitch < menuRowHeight || pitch > menuRowHeight+1 {
		t.Errorf("the rule moves the next entry %.1f pt, want %d to %d", pitch, menuRowHeight, menuRowHeight+1)
	}
}
