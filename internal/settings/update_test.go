package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

// TestUpdateRowOpensTheReleasePageWhenAsked checks that choosing the general pane opens
// nothing, and that the check button opens the release page once.
func TestUpdateRowOpensTheReleasePageWhenAsked(t *testing.T) {
	opened := 0
	e := testEnv()
	e.checkForUpdates = func() { opened++ }
	m, err := newApp(testStore(t), e)
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneGeneral)
	if opened != 0 {
		t.Fatalf("choosing the general pane opened the release page %d times", opened)
	}
	screen := ui.NewTester(m.view, mainWidth, mainHeight)
	if _, ok := screen.Find(locales.CheckUpdate()); !ok {
		t.Fatal("the check-for-updates row is missing from the general pane")
	}
	if err := screen.Click(locales.CheckNow()); err != nil {
		t.Fatal(err)
	}
	if opened != 1 {
		t.Errorf("the check button opened the release page %d times, want 1", opened)
	}
}
