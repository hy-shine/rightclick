package settings

import (
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
	"github.com/hy-shine/rightclick/internal/update"
)

// updatePhase is where the check for a newer release stands.
type updatePhase int

const (
	updateIdle     updatePhase = iota // no check has run yet
	updateChecking                    // reading the newest release
	updateCurrent                     // this is the newest release
	updateNewer                       // a newer release exists
	updateFailed                      // no answer: offline, or no release is published
)

// updateCheck is what the general pane shows about the newest release.
type updateCheck struct {
	phase   updatePhase
	release update.Release
	seq     int // counts the checks, so that an older result is dropped
}

// checkForUpdate reads the newest release off the main thread. The check runs only
// when the user clicks the button; nothing is checked in the background.
func (m *app) checkForUpdate() {
	m.updates.phase = updateChecking
	m.updates.seq++
	seq := m.updates.seq
	win, latest := m.win, m.env.latestRelease
	go func() {
		release, err := latest()
		win.Update(func() { m.applyUpdate(seq, release, err) })
	}()
}

// applyUpdate records the result of a check. It runs on the main thread.
func (m *app) applyUpdate(seq int, release update.Release, err error) {
	if seq != m.updates.seq {
		return
	}
	m.updates.release = release
	m.updates.phase = updateFailed
	if err != nil {
		return
	}
	newer, err := update.Newer(release.Version, version)
	if err != nil {
		return
	}
	m.updates.phase = updateCurrent
	if newer {
		m.updates.phase = updateNewer
	}
}

// updateRow is the row of the general pane that checks for a newer release. A newer
// release is not installed here: its page opens in the browser, where the user
// downloads it.
func (m *app) updateRow(c *ui.Context) {
	note := locales.UpdateHint()
	switch m.updates.phase {
	case updateChecking:
		note = locales.UpdateChecking()
	case updateCurrent:
		note = locales.UpdateCurrent(version)
	case updateNewer:
		note = locales.UpdateAvailable(m.updates.release.Version)
	case updateFailed:
		note = locales.UpdateFailed()
	}
	settingRow(c, nil, ui.Color{}, locales.CheckUpdate(), note, func() {
		if m.updates.phase == updateNewer {
			if primaryButton(c, nil, locales.DownloadUpdate()) {
				m.report(locales.OpenDownloadFailed(), m.env.openURL(m.updates.release.URL))
			}
			return
		}
		if quietButton(c, nil, locales.CheckNow()) {
			m.checkForUpdate()
		}
	})
}
