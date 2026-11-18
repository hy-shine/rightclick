package settings

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// openersPane lists the programs whose menu items open the selection.
func (m *app) openersPane(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			paneHeader(c, locales.PaneOpeners(), locales.OpenersIntro())
		})
		if len(m.cfg.Programs) > 0 {
			if primaryButtonCentered(c, iconPlus, locales.ChooseApp()) {
				m.pickEditor()
			}
		}
	})
	m.showNotice(c)
	if len(m.cfg.Programs) > 0 {
		searchInput(c.Key("program-search"), &m.programSearch, locales.SearchPrograms(), locales.SearchProgramsPlaceholder()).Width(220)
	}
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
		emptyState(c, iconOpen, locales.NoEditors(), locales.NoEditorsBody(), func() {
			if primaryButton(c, iconPlus, locales.ChooseApp()) {
				m.pickEditor()
			}
		})
	case len(shown) == 0:
		emptyState(c, iconOpen, locales.NoProgramMatches(), locales.SearchEmptyHint(), func() {
			if quietButton(c, nil, locales.ClearSearch()) {
				m.programSearch = ""
			}
		})
	default:
		m.programList.Key = func(index int) any { return m.cfg.Programs[shown[index]] }
		m.programList.Reorder = nil
		if query == "" {
			m.programList.Reorder = func(rows []int, target int) {
				m.cfg.Programs = moveRows(m.cfg.Programs, rows, target)
				m.saveConfig()
			}
		}
		ledger(c, ui.Column(c)).Children(func() {
			ui.Row(c).Height(favoriteHeaderHeight).Gap(12).Padding(0, 14).AlignItems(ui.Center).Children(func() {
				ui.Text(c, locales.ColApp()).FontSize(sizeBody).FontWeight(600).TextColor(t.TextMuted).Grow(1)
				ui.Text(c, locales.QuickAccess()).FontSize(sizeBody).FontWeight(600).TextColor(t.TextMuted).Width(favoriteShortcutWidth).TextAlign(ui.Center)
				ui.Text(c, locales.ColActionContext()).FontSize(sizeBody).FontWeight(600).TextColor(t.TextMuted).Width(76).TextAlign(ui.Center)
			})
			ui.List(c, &m.programList, len(shown), func(row int) {
				i := shown[row]
				name := m.cfg.Programs[i]
				id := "app:" + name
				note := m.programLocationNote(id)
				itemRow(c, query == "", func() {
					if icon := m.editorIcon(name); icon != nil {
						ui.Image(c, icon).Size(itemIconSize, itemIconSize)
					} else {
						glyphTile(c, iconCode)
					}
				}, name, note, func() {
					ui.Row(c).Width(favoriteShortcutWidth).Justify(ui.Center).Children(func() {
						enabled := m.hasShortcut(id)
						ui.Switch(c.Key("shortcut:"+id), &enabled).Label(locales.ShortcutFor(name)).OnChange(func() {
							m.setShortcut(id, enabled)
						})
					})
					ui.Row(c).Width(76).Gap(12).AlignItems(ui.Center).Justify(ui.End).Children(func() {
						if m.hasShortcut(id) {
							currentLabel := ""
							if idx := slices.IndexFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id }); idx >= 0 {
								currentLabel = m.cfg.MenuItems[idx].Label
							}
							if iconButton(c.Key("rename-app:"+name), iconEdit, locales.EditItem(name), false) {
								m.menuNameID, m.menuName, m.menuNameOpen = id, currentLabel, true
							}
						}
						if iconButton(c.Key("remove-app:"+name), iconTrash, locales.Delete(), true) {
							remove = i
						}
					})
				})
			}).Height(float32(len(shown))*itemRowHeight).Dividers(1, separatorColor(c))
		})
	}
	if remove >= 0 {
		id := "app:" + m.cfg.Programs[remove]
		m.cfg.Programs = slices.Delete(m.cfg.Programs, remove, remove+1)
		m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id })
		m.saveConfig()
	}
	ui.Text(c, locales.EditorsHint()).
		FontSize(sizeNote).LineHeight(1.45).TextColor(t.TextMuted)
}

func (m *app) programLocationNote(id string) string {
	idx := slices.IndexFunc(m.cfg.MenuItems, func(item config.MenuItem) bool {
		return item.ID == id && item.Enabled
	})
	if idx < 0 {
		return ""
	}
	item := m.cfg.MenuItems[idx]
	loc := locales.LocInApp()
	if item.Location == config.FileTypeMain {
		loc = locales.LocFinderTopShort()
	}
	if item.Label != "" {
		return item.Label + " · " + loc
	}
	return loc
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
		m.ensureShortcutInSubmenu("app:" + name)
		m.saveConfig()
		m.pendingToast = locales.AppAddedToast(name)
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
