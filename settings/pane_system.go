package settings

import (
	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/locales"

	"github.com/egoist/mygo/ui"
)

// finderState is what the diagnostics pane shows about the Finder extension.
type finderState int

const (
	finderChecking finderState = iota
	finderOn
	finderOff
)

// diagnosticsPane lists the steps that the Finder menu needs, each with the
// action that completes it, and the Full Disk Access state after them. The
// conclusion is shown only when the four menu steps pass.
func (m *app) diagnosticsPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, locales.PaneDiagnostics(), locales.DiagnosticsIntro())
	if m.setupStep() == stepReady {
		banner(c, t.Success, iconCheck, locales.ReadyInFinder(), locales.ReadyInFinderBody(), "")
	}
	ledger(c, ui.Column(c)).Children(func() {
		extensionTitle := locales.BannerExtensionOff()
		extension, extensionAction := locales.ExtensionOpenHint(), locales.GoEnable()
		extensionIcon, extensionTone := iconAlert, t.Warning
		switch m.finder {
		case finderChecking:
			extensionTitle, extension, extensionAction = locales.FinderExtension(), locales.CheckingItem(), ""
			extensionIcon, extensionTone = iconInfo, t.TextMuted
		case finderOn:
			extensionTitle = locales.CheckExtension()
			extension, extensionAction = locales.ExtensionEnabledHint(), ""
			extensionIcon, extensionTone = iconCheck, t.Success
		}
		if diagnosticRow(c, extensionTitle, extension, extensionAction, extensionIcon, extensionTone) {
			m.openFinderSettings()
		}

		folders, foldersAction := locales.FolderScopeHint(), locales.AddEllipsis()
		if n := len(m.cfg.WatchFolders); n > 0 {
			folders, foldersAction = locales.FoldersAdded(n), ""
		}
		if checkRow(c, locales.CheckFolderScope(), len(m.cfg.WatchFolders) > 0, folders, foldersAction) {
			m.pickFolder()
		}

		items, itemsAction := locales.AtLeastOneItem(), locales.GoEdit()
		if n := enabledCount(m.cfg.MenuItems); n > 0 {
			items, itemsAction = locales.ItemsEnabled(n), ""
		}
		if checkRow(c, locales.CheckItemsEnabled(), enabledCount(m.cfg.MenuItems) > 0, items, itemsAction) {
			m.selectPane(paneMenu)
		}

		restart, restartAction := locales.RestartNeededBody(), locales.Restart()
		if !m.foldersChanged {
			restart, restartAction = locales.FolderScopeApplied(), ""
		}
		if checkRow(c, locales.CheckFinderReloaded(), !m.foldersChanged, restart, restartAction) {
			m.confirmRestart = true
		}

		// Full Disk Access is not one of the steps that keep the menu from
		// showing, so it does not change the conclusion above. Actions that read
		// protected locations need it.
		diskNote, diskAction := locales.DiskAccessGranted(), ""
		if !m.diskAccess {
			diskNote, diskAction = locales.DiskAccessMissing(), locales.GrantAccess()
		}
		if checkRow(c, locales.CheckDiskAccess(), m.diskAccess, diskNote, diskAction) {
			m.openPrivacySettings(privacyAllFiles)
		}
	})
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, locales.RestartFinder(), locales.RestartFinderHint(), func() {
			if quietButton(c, iconRestart, locales.Restart()) {
				m.confirmRestart = true
			}
		})
	})
}

// checkRow is one step of the diagnostics checklist. A passed step shows a check
// mark. A missing step shows a warning and, when action is not empty, a button
// with that label, which checkRow reports as a click.
func checkRow(c *ui.Context, title string, pass bool, note, action string) bool {
	t := c.Theme()
	statusIcon, statusColor := iconAlert, t.Warning
	if pass {
		statusIcon, statusColor = iconCheck, t.Success
	}
	return diagnosticRow(c, title, note, action, statusIcon, statusColor)
}

func diagnosticRow(c *ui.Context, title, note, action string, statusIcon *ui.SVG, statusColor ui.Color) bool {
	t := c.Theme()
	clicked := false
	ui.Row(c).Gap(14).Padding(14, 16).AlignItems(ui.Center).Children(func() {
		glyph(c, statusIcon, 17, statusColor)
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, title).FontSize(sizeBody).FontWeight(600)
			if note != "" {
				ui.Text(c, note).FontSize(sizeNote).LineHeight(1.45).TextColor(t.TextMuted)
			}
		})
		if action != "" {
			clicked = quietButton(c, nil, action)
		}
	})
	return clicked
}

// checkFinder asks the system whether the Finder extension is on. The check
// runs off the main thread, and its result comes back through the window's
// Update. A result that a later check has overtaken is dropped.
func (m *app) checkFinder() {
	m.finder = finderChecking
	m.finderSeq++
	seq := m.finderSeq
	win, check := m.win, m.env.finderEnabled
	go func() {
		on := check()
		win.Update(func() {
			if seq != m.finderSeq {
				return
			}
			m.finder = finderOff
			if on {
				m.finder = finderOn
			}
		})
	}()
}

// openFinderSettings opens the System Settings pane for extensions.
func (m *app) openFinderSettings() {
	m.report(locales.OpenSettingsFailed(), m.env.openFinderSettings())
}

// openPrivacySettings opens one of the Privacy panes in System Settings.
func (m *app) openPrivacySettings(anchor string) {
	m.report(locales.OpenSettingsFailed(), m.env.openPrivacy(anchor))
}

// restartFinder restarts Finder off the main thread. When it works, Finder has
// read the watched folders again, so the folders no longer count as changed.
func (m *app) restartFinder() {
	win, restart := m.win, m.env.restartFinder
	go func() {
		err := restart()
		win.Update(func() {
			if err == nil {
				m.foldersChanged = false
			}
			m.report(locales.RestartFinderFailed(), err)
		})
	}()
}

// generalPane holds the switch for starting the app at login, and below it the
// app's name and version, so that the sidebar needs no separate About pane.
func (m *app) generalPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, locales.PaneGeneral(), "")
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, locales.LanguageSetting(), locales.LanguageHint(), func() {
			locationSelect(c, &m.cfg.Language, "language", locales.LanguageSetting(), config.Languages, languageLabels(), true, m.applyLanguage)
		})
		settingRow(c, locales.LaunchAtLogin(), locales.LaunchAtLoginHint(), func() {
			ui.Switch(c, &m.openAtLogin).Label(locales.LaunchAtLogin()).OnChange(m.applyOpenAtLogin)
		})
	})
	sectionTitle(c, locales.About())
	ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
		mark(c, 56)
		ui.Column(c).Gap(4).Children(func() {
			ui.Text(c, locales.AppName()).FontSize(22).FontWeight(700)
			ui.Text(c, locales.VersionLine(version)).FontSize(sizeBody).TextColor(t.TextMuted)
		})
	})
	ui.Text(c, locales.AboutBlurb()).
		FontSize(sizeBody).LineHeight(1.6).TextColor(t.TextMuted)
}

// languageLabels returns the picker text for each tag in config.Languages, in the
// same order. The picker shows these labels and stores the tag.
func languageLabels() []string {
	return []string{locales.LanguageSystem(), locales.LanguageChinese(), locales.LanguageEnglish()}
}

// applyLanguage saves the language the picker chose and shows the windows in it.
// The Finder menu reads the same setting from config.json.
func (m *app) applyLanguage() {
	m.saveConfig()
	m.env.setLocale(m.cfg.Language)
}

// applyOpenAtLogin sets the login item, then shows what the system reports,
// so that a failed change does not leave the switch in the wrong place.
func (m *app) applyOpenAtLogin() {
	err := m.env.setOpenAtLogin(m.openAtLogin)
	m.openAtLogin = m.env.openAtLogin()
	m.report(locales.SetLoginFailed(), err)
}
