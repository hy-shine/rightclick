package settings

import (
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

// updateRow is the row of the general pane that checks for updates in the app.
func (m *app) updateRow(c *ui.Context) {
	settingRow(c, nil, ui.Color{}, locales.CheckUpdate(), locales.UpdateHint(), func() {
		if quietButton(c, nil, locales.CheckNow()) {
			m.env.checkForUpdates()
		}
	})
}
