package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestSidebarHasOnlyPaneLabels(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tester := ui.NewTester(func(context *ui.Context) { context.SetTheme(paperTheme(true)); model.sidebar(context) }, sidebarWidth, 560)
	texts := tester.Texts()
	want := []string{locales.AppName()}
	for _, tab := range paneTabs() {
		if !slices.Contains(texts, tab.label) {
			t.Errorf("missing pane %s", tab.label)
		}
		want = append(want, tab.label)
	}
	for _, text := range texts {
		if !slices.Contains(want, text) {
			t.Errorf("unexpected sidebar text %q", text)
		}
	}
	if err := tester.Click("Files & Templates"); err != nil {
		t.Fatal(err)
	}
	if model.currentPane() != paneTemplates {
		t.Fatal("sidebar navigation failed")
	}
}

func TestVibrancySidebarConsistentAcrossPanes(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.finder = finderOn

	for _, dark := range []bool{false, true} {
		tester := ui.NewTester(model.view, mainWidth, mainHeight)
		tester.SetDark(dark)
		tester.SetVibrancy(true)
		for _, tab := range paneTabs() {
			model.selectPane(tab.id)
			tester.Frame()
			img := tester.Image()
			if px := img.RGBAAt(20, 450); px.A != 0 {
				t.Errorf("dark=%v pane=%s sidebar pixel alpha = %d, want 0 (transparent for vibrancy)", dark, tab.id, px.A)
			}
			if px := img.RGBAAt(sidebarWidth+40, 450); px.A != 255 {
				t.Errorf("dark=%v pane=%s content pixel alpha = %d, want 255 (opaque content)", dark, tab.id, px.A)
			}
		}
	}
}
