package settings

import (
	"path/filepath"
	"rightclick/config"
	"rightclick/locales"
	"slices"
	"strings"

	"errors"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// openersPane sets the terminal and the list of editors.
func (m *app) openersPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, locales.PaneOpeners(), locales.OpenersIntro())
	sectionTitle(c, locales.SectionTerminal())
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, locales.DefaultTerminal(), locales.DefaultTerminalHint(), func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Select(c.Key("terminal"), &m.cfg.Terminal, terminalChoices(m.cfg.Terminal)).
					Label(locales.DefaultTerminal()).Width(150).OnChange(m.saveConfig)
				if quietButton(c, iconOpen, locales.ChooseApp()) {
					m.pickTerminal()
				}
			})
		})
	})
	sectionTitle(c, locales.SectionEditor())
	ui.Text(c, locales.AppShortcutHint()).FontSize(sizeNote).TextColor(t.TextMuted)
	remove := -1
	setDefault := -1
	if len(m.cfg.Editors) == 0 {
		emptyState(c, iconCode, locales.NoEditors(), locales.NoEditorsBody())
	} else {
		ledger(c, ui.Column(c)).Children(func() {
			for i, name := range m.cfg.Editors {
				ui.Row(c).Gap(12).Padding(0, 8, 0, 16).Height(44).AlignItems(ui.Center).Children(func() {
					if icon := m.editorIcon(name); icon != nil {
						ui.Image(c, icon).Size(17, 17)
					} else {
						glyph(c, iconCode, 17, t.TextMuted)
					}
					ui.Text(c, name).FontSize(sizeBody).Grow(1)
					m.shortcutToggle(c, "app:"+name, name)
					if i == 0 {
						ui.Text(c, locales.DefaultBadge()).FontSize(sizeCaption).FontWeight(600).
							TextColor(t.TextMuted).Padding(2, 6).
							Background(t.SurfaceHover).Radius(t.Radius)
					} else if ghostButton(c, locales.SetAsDefault(), false) {
						setDefault = i
					}
					if ghostButton(c, locales.Remove(), true) {
						remove = i
					}
				})
			}
		})
	}
	// Removing happens after the rows are built, as in the folders pane.
	if remove >= 0 {
		id := "app:" + m.cfg.Editors[remove]
		m.cfg.Editors = slices.Delete(m.cfg.Editors, remove, remove+1)
		m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id })
		m.saveConfig()
	} else if setDefault >= 0 {
		m.setDefaultEditor(setDefault)
	}
	ui.Row(c).Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c.Key("new-editor"), &m.newEditor).
			Label(locales.EditorAppName()).Placeholder(locales.EditorAppPlaceholder()).Grow(1).OnSubmit(m.addEditor)
		if primaryButton(c, iconPlus, locales.Add()) {
			m.addEditor()
		}
		if quietButton(c, iconOpen, locales.ChooseApp()) {
			m.pickEditor()
		}
	})
	ui.Text(c, locales.EditorsHint()).
		FontSize(sizeCaption).LineHeight(1.45).TextColor(t.TextMuted)
}

// setDefaultEditor moves the editor at index to the front, where the default
// editor action reads it.
func (m *app) setDefaultEditor(index int) {
	if index <= 0 || index >= len(m.cfg.Editors) {
		return
	}
	name := m.cfg.Editors[index]
	editors := slices.Delete(slices.Clone(m.cfg.Editors), index, index+1)
	m.cfg.Editors = slices.Insert(editors, 0, name)
	m.saveConfig()
}

// addEditor adds the application named in the input to the editors and clears the input.
func (m *app) addEditor() {
	name := strings.TrimSpace(m.newEditor)
	m.newEditor = ""
	if name == "" || slices.Contains(m.cfg.Editors, name) {
		return
	}
	m.cfg.Editors = append(m.cfg.Editors, name)
	m.saveConfig()
}

// pickTerminal lets the user choose any installed application for the Open in
// Terminal action. The app name is what open(1) receives with -a.
func (m *app) pickTerminal() {
	m.pickApplication(locales.ChooseTerminalTitle(), func(name string) {
		m.cfg.Terminal = name
		m.saveConfig()
	})
}

// pickEditor lets the user add any installed application as an editor.
func (m *app) pickEditor() {
	m.pickApplication(locales.ChooseEditorTitle(), func(name string) {
		m.newEditor = ""
		if slices.Contains(m.cfg.Editors, name) {
			return
		}
		m.cfg.Editors = append(m.cfg.Editors, name)
		m.saveConfig()
	})
}

// pickApplication opens an application package picker and publishes the chosen
// display name on the UI thread.
func (m *app) pickApplication(title string, apply func(string)) {
	win, open := m.win, m.env.openDialog
	go func() {
		paths, err := open(mygo.OpenDialogOptions{
			Parent:  win,
			Title:   title,
			Message: locales.ChooseInstalledApp(),
			Filters: []mygo.FileFilter{{
				Name:       locales.ApplicationsFilter(),
				Extensions: []string{"app"},
			}},
		})
		if err == nil && len(paths) == 0 {
			return
		}
		var name string
		if err == nil {
			name, err = applicationName(paths[0])
		}
		win.Update(func() {
			if err != nil {
				m.report(locales.ChooseAppFailed(), err)
				return
			}
			apply(name)
		})
	}()
}

// applicationName returns the name open -a uses for a selected .app package.
func applicationName(path string) (string, error) {
	base := filepath.Base(path)
	if !strings.EqualFold(filepath.Ext(base), ".app") {
		return "", errors.New(locales.NotAppBundle(base))
	}
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if strings.TrimSpace(name) == "" {
		return "", errors.New(locales.AppNameEmpty(path))
	}
	return name, nil
}

// terminalChoices returns the terminals to offer. The current setting is
// added when it is not one of them, so that the select shows it.
func terminalChoices(current string) []string {
	choices := []string{
		"Terminal",
		"iTerm",
		"Warp",
		"Ghostty",
		"kitty",
		"Alacritty",
		"WezTerm",
		"Hyper",
		"Tabby",
	}
	if current != "" && !slices.Contains(choices, current) {
		choices = append(choices, current)
	}
	return choices
}
