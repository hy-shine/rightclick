package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// TestLocationDropdownShowsTheMainMenuFirst opens the dropdown and checks that the
// Main Menu row sits above the RightClick row on screen.
func TestLocationDropdownShowsTheMainMenuFirst(t *testing.T) {
	value := config.FileTypeSubmenu
	tester := ui.NewTester(func(c *ui.Context) {
		locationSelect(c, &value, "test", "Location", locationValues(), locationLabels(), true, func() {})
	}, 240, 200)
	if err := tester.Click(locales.LocInApp()); err != nil {
		t.Fatal(err)
	}
	lastMain, lastInApp := -1, -1
	for i, text := range tester.Texts() {
		switch text {
		case locales.LocFinderTopShort():
			lastMain = i
		case locales.LocInApp():
			lastInApp = i
		}
	}
	if lastMain < 0 || lastInApp < 0 {
		t.Fatalf("both choices are not shown: %q", tester.Texts())
	}
	if lastMain > lastInApp {
		t.Errorf("Main Menu is listed below RightClick: %q", tester.Texts())
	}
}
