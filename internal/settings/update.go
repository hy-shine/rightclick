package settings

import (
	"log"
	"os/exec"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

// latestReleaseURL is the newest release of RightClick on GitHub. The app downloads
// nothing: "Check for Updates" only opens this page, and the user installs from there.
const latestReleaseURL = "https://github.com/hy-shine/rightclick/releases/latest"

// updateRow is the row of the general pane that opens the newest release page.
func (m *app) updateRow(c *ui.Context) {
	settingRow(c, nil, ui.Color{}, locales.CheckUpdate(), locales.UpdateHint(), func() {
		if quietButton(c, nil, locales.CheckNow()) {
			m.env.checkForUpdates()
		}
	})
}

// openLatestRelease opens the newest release page in the default browser.
func openLatestRelease() {
	if err := exec.Command("/usr/bin/open", latestReleaseURL).Run(); err != nil {
		log.Printf("update: open %s: %v", latestReleaseURL, err)
	}
}
