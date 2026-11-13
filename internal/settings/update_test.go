package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

// TestUpdateButtonChecksForUpdatesWhenAsked checks that choosing the general pane checks
// nothing, and that the check button starts one check.
func TestUpdateButtonChecksForUpdatesWhenAsked(t *testing.T) {
	opened := 0
	e := testEnv()
	e.checkForUpdates = func() { opened++ }
	m, err := newApp(testStore(t), e)
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneGeneral)
	if opened != 0 {
		t.Fatalf("choosing the general pane checked for updates %d times", opened)
	}
	screen := ui.NewTester(m.view, mainWidth, mainHeight)
	if _, ok := screen.Find(locales.CheckUpdate()); !ok {
		t.Fatal("the check-for-updates button is missing from the general pane")
	}
	if err := screen.Click(locales.CheckUpdate()); err != nil {
		t.Fatal(err)
	}
	if opened != 1 {
		t.Errorf("the check button checked for updates %d times, want 1", opened)
	}
}
