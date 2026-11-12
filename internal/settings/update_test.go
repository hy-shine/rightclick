package settings

import (
	"errors"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
	"github.com/hy-shine/rightclick/internal/update"
)

// TestApplyUpdateSetsThePhaseFromTheRelease checks that a result says whether the
// newest release is newer than this app, the same as, or unreachable.
func TestApplyUpdateSetsThePhaseFromTheRelease(t *testing.T) {
	m := readyApp(t)
	m.updates.seq = 1
	m.applyUpdate(1, update.Release{Version: "0.2.0", URL: "https://example.com/r"}, nil)
	if m.updates.phase != updateNewer || m.updates.release.URL != "https://example.com/r" {
		t.Errorf("a newer release gave phase %d, release %+v", m.updates.phase, m.updates.release)
	}
	m.applyUpdate(1, update.Release{Version: version}, nil)
	if m.updates.phase != updateCurrent {
		t.Errorf("the same release gave phase %d, want current", m.updates.phase)
	}
	m.applyUpdate(1, update.Release{}, errors.New("offline"))
	if m.updates.phase != updateFailed {
		t.Errorf("an unreachable check gave phase %d, want failed", m.updates.phase)
	}
	m.applyUpdate(1, update.Release{Version: "0.2.0-beta"}, nil)
	if m.updates.phase != updateFailed {
		t.Errorf("a pre-release tag gave phase %d, want failed", m.updates.phase)
	}
}

// TestApplyUpdateDropsAnOlderResult keeps a slow answer from overwriting the result of
// a later check.
func TestApplyUpdateDropsAnOlderResult(t *testing.T) {
	m := readyApp(t)
	m.updates.seq = 2
	m.updates.phase = updateChecking
	m.applyUpdate(1, update.Release{Version: "0.2.0"}, nil)
	if m.updates.phase != updateChecking {
		t.Errorf("an older result changed the phase to %d", m.updates.phase)
	}
}

// TestGeneralPaneChecksOnlyWhenAskedAndOffersTheDownload checks that choosing the
// general pane makes no request, and that a newer release shows its download button.
func TestGeneralPaneChecksOnlyWhenAskedAndOffersTheDownload(t *testing.T) {
	m := readyApp(t)
	m.selectPane(paneGeneral)
	if m.updates.phase != updateIdle {
		t.Fatalf("choosing the general pane started a check: phase %d", m.updates.phase)
	}
	screen := ui.NewTester(m.view, mainWidth, mainHeight)
	if _, ok := screen.Find(locales.CheckUpdate()); !ok {
		t.Fatal("the check-for-updates row is missing from the general pane")
	}
	if _, ok := screen.Find(locales.UpdateHint()); !ok {
		t.Error("the idle row does not say when it checks")
	}

	m.updates.phase = updateNewer
	m.updates.release = update.Release{Version: "0.2.0", URL: "https://example.com/r"}
	screen = ui.NewTester(m.view, mainWidth, mainHeight)
	if _, ok := screen.Find(locales.UpdateAvailable("0.2.0")); !ok {
		t.Error("a newer release is not named in the row")
	}
	if _, ok := screen.Find(locales.DownloadUpdate()); !ok {
		t.Error("a newer release has no download button")
	}
}
