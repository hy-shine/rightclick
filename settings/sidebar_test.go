package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSidebarHasOnlyPaneLabels(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tester := ui.NewTester(func(context *ui.Context) { context.SetTheme(paperTheme(true)); model.sidebar(context) }, 176, 560)
	texts := tester.Texts()
	for _, tab := range paneTabs() {
		if !slices.Contains(texts, tab.label) {
			t.Errorf("missing pane %s", tab.label)
		}
		if summary := model.navSummary(tab.id); summary != "" && slices.Contains(texts, summary) {
			t.Errorf("sidebar still shows summary %q", summary)
		}
	}
	if err := tester.Click("新建"); err != nil {
		t.Fatal(err)
	}
	if model.currentPane() != paneTemplates {
		t.Fatal("sidebar navigation failed")
	}
}
