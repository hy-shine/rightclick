package settings

import (
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"

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
// action that completes it. The conclusion is shown only when they all pass.
// The restart row is the one place that asks for a restart: after a folder
// change it shows the pending state, otherwise it is the manual restart.
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
			m.enableFinderExtension()
		}

	})
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, nil, ui.Color{}, locales.RestartFinder(), locales.RestartFinderHint(), func() {
			if quietButton(c, iconRestart, locales.Restart()) {
				m.confirmRestart = true
			}
		})
	})
}

func diagnosticRow(c *ui.Context, title, note, action string, statusIcon *ui.SVG, statusColor ui.Color) bool {
	clicked := false
	settingRow(c, statusIcon, statusColor, title, note, func() {
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

// enableFinderExtension registers the Finder extension and turns it on. When the
// system still reports it off, it opens System Settings, where the user switches it
// on. The work runs off the main thread, and the state is checked again afterwards.
func (m *app) enableFinderExtension() {
	win, enable, check := m.win, m.env.enableExtension, m.env.finderEnabled
	m.finder = finderChecking
	m.finderSeq++
	seq := m.finderSeq
	go func() {
		err := enable()
		on := check()
		win.Update(func() {
			if seq != m.finderSeq {
				return
			}
			m.report(locales.ExtensionEnableFailed(), err)
			m.finder = finderOff
			if on {
				m.finder = finderOn
			} else if err == nil {
				m.openFinderSettings()
			}
		})
	}()
}

// openFinderSettings opens the System Settings pane for extensions.
func (m *app) openFinderSettings() {
	m.report(locales.OpenSettingsFailed(), m.env.openFinderSettings())
}

// restartFinder restarts Finder off the main thread.
func (m *app) restartFinder() {
	win, restart := m.win, m.env.restartFinder
	go func() {
		err := restart()
		win.Update(func() {
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
		settingRow(c, nil, ui.Color{}, locales.LanguageSetting(), locales.LanguageHint(), func() {
			locationSelect(c, &m.cfg.Language, "language", locales.LanguageSetting(), config.Languages, languageLabels(), true, m.applyLanguage)
		})
		settingRow(c, nil, ui.Color{}, locales.LaunchAtLogin(), locales.LaunchAtLoginHint(), func() {
			ui.Switch(c, &m.openAtLogin).Label(locales.LaunchAtLogin()).OnChange(m.applyOpenAtLogin)
		})
		m.updateRow(c)
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
	return []string{locales.LanguageSystem(), locales.LanguageChinese(), locales.LanguageTraditionalChinese(), locales.LanguageEnglish()}
}

// applyLanguage saves the language the picker chose and shows the windows in it.
// The Finder menu reads the same setting from config.json.
func (m *app) applyLanguage() {
	m.saveConfig()
	m.env.setLocale(m.cfg.Language)
	m.env.setWindowTitle(locales.AppName())
	refreshAppMenuLabels()
}

// applyOpenAtLogin sets the login item, then shows what the system reports,
// so that a failed change does not leave the switch in the wrong place.
func (m *app) applyOpenAtLogin() {
	err := m.env.setOpenAtLogin(m.openAtLogin)
	m.openAtLogin = m.env.openAtLogin()
	m.report(locales.SetLoginFailed(), err)
}
