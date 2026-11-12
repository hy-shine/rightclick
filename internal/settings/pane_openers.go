package settings

import (
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
	"path/filepath"
	"slices"
	"strings"

	"errors"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// openersPane lists the programs whose menu items open the selection. The search
// box above the list filters the programs as the user types.
func (m *app) openersPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, locales.PaneOpeners(), locales.OpenersIntro())
	ui.Row(c).Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c.Key("program-search"), &m.programSearch).
			Label(locales.SearchPrograms()).Placeholder(locales.SearchProgramsPlaceholder()).Grow(1)
		if quietButton(c, iconOpen, locales.ChooseApp()) {
			m.pickEditor()
		}
	})
	// shown keeps each row's index in cfg.Programs, so removing a filtered row
	// still removes the right program.
	query := strings.ToLower(strings.TrimSpace(m.programSearch))
	var shown []int
	for i, name := range m.cfg.Programs {
		if query == "" || strings.Contains(strings.ToLower(name), query) {
			shown = append(shown, i)
		}
	}
	remove := -1
	switch {
	case len(m.cfg.Programs) == 0:
		emptyState(c, iconCode, locales.NoEditors(), locales.NoEditorsBody())
	case len(shown) == 0:
		ui.Text(c, locales.NoProgramMatches()).FontSize(sizeNote).TextColor(t.TextMuted).Padding(16, 12)
	default:
		ledger(c, ui.Column(c)).Children(func() {
			for _, i := range shown {
				name := m.cfg.Programs[i]
				itemRow(c, false, func() {
					if icon := m.editorIcon(name); icon != nil {
						ui.Image(c, icon).Size(itemIconSize, itemIconSize)
					} else {
						glyphTile(c, iconCode)
					}
				}, name, "", func() {
					m.shortcutToggleLabeled(c, "app:"+name, name, locales.QuickAccess())
					if iconButton(c, iconTrash, locales.Delete(), true) {
						remove = i
					}
				})
			}
		})
	}
	// Removing happens after the rows are built, as in the folders pane.
	if remove >= 0 {
		id := "app:" + m.cfg.Programs[remove]
		m.cfg.Programs = slices.Delete(m.cfg.Programs, remove, remove+1)
		m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id })
		m.saveConfig()
	}
	ui.Text(c, locales.EditorsHint()).
		FontSize(sizeCaption).LineHeight(1.45).TextColor(t.TextMuted)
}

// pickEditor lets the user add any installed application as a program.
func (m *app) pickEditor() {
	m.pickApplication(locales.ChooseEditorTitle(), func(name string) {
		if slices.Contains(m.cfg.Programs, name) {
			m.notice = locales.ProgramAlreadyAdded(name)
			return
		}
		m.notice = ""
		m.cfg.Programs = append(m.cfg.Programs, name)
		m.saveConfig()
	})
}

// pickApplication opens an application package picker and publishes the chosen
// display name on the UI thread.
func (m *app) pickApplication(title string, apply func(string)) {
	win, open := m.win, m.env.openDialog
	go func() {
		// No type filter: the "app" extension maps to com.apple.application-file, and a
		// bundle is com.apple.application-bundle, so macOS greys out every app. The
		// chosen path is still checked by applicationName.
		paths, err := open(mygo.OpenDialogOptions{
			Parent:      win,
			Title:       title,
			Message:     locales.ChooseInstalledApp(),
			DefaultPath: "/Applications",
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
