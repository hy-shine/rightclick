package settings

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func (m *app) hasShortcut(id string) bool {
	return slices.ContainsFunc(m.cfg.MenuItems, func(item config.MenuItem) bool {
		return item.ID == id && item.Enabled
	})
}

func (m *app) setShortcut(id string, enabled bool) {
	if enabled {
		if idx := slices.IndexFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id }); idx >= 0 {
			m.cfg.MenuItems[idx].Enabled = true
		} else {
			m.cfg.MenuItems = append(m.cfg.MenuItems, config.MenuItem{ID: id, Enabled: true, Location: config.FileTypeMain})
		}
	} else {
		m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id })
	}
	m.saveConfig()
}

func (m *app) ensureShortcutInSubmenu(id string) {
	if idx := slices.IndexFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id }); idx >= 0 {
		m.cfg.MenuItems[idx].Enabled = true
		return
	}
	m.cfg.MenuItems = append(m.cfg.MenuItems, config.MenuItem{ID: id, Enabled: true, Location: config.FileTypeSubmenu})
}

// shortcutToggleLabeled shows a checkbox that adds or removes the menu shortcut for
// id. visible is the text beside the box; label names it for accessibility.
func (m *app) shortcutToggleLabeled(c *ui.Context, id, label, visible string) {
	enabled := m.hasShortcut(id)
	ui.Checkbox(c.Key("shortcut:"+id), &enabled, visible).Label(locales.ShortcutFor(label)).OnClick(func() {
		m.setShortcut(id, enabled)
	})
}

func (m *app) menuEntry(item config.MenuItem) menuEntry {
	entry := menuEntryFor(item.ID)
	kind, target, shortcut := strings.Cut(item.ID, ":")
	if shortcut {
		switch kind {
		case "app":
			entry = menuEntry{locales.OpenWithShortcut(target), locales.WhereAnywhere()}
		case "folder":
			for _, folder := range m.cfg.FavoriteFolders {
				if folder.ID == target {
					entry = menuEntry{folder.Label, locales.WhereAnywhere()}
				}
			}
		}
	}
	if item.Label != "" {
		entry.label = item.Label
	}
	return entry
}

func (m *app) menuItemAvailable(item config.MenuItem, context int) bool {
	kind, target, shortcut := strings.Cut(item.ID, ":")
	if !shortcut {
		if item.ID == "favorite-folders" {
			return slices.ContainsFunc(m.cfg.FavoriteFolders, func(folder config.FavoriteFolder) bool { return folder.Enabled })
		}
		return appliesTo(item.ID, context)
	}
	switch kind {
	case "app":
		return slices.Contains(m.cfg.Programs, target)
	case "folder":
		return slices.ContainsFunc(m.cfg.FavoriteFolders, func(folder config.FavoriteFolder) bool {
			return folder.ID == target && folder.Enabled
		})
	}
	return false
}

func (m *app) menuNameDialog(c *ui.Context) {
	ui.DialogBase(c, &m.menuNameOpen, func(backdrop, panel ui.Element) {
		t := c.Theme()
		backdrop.Background(ui.RGBA(0, 0, 0, 0.3))
		panel.Width(380).Padding(22).Background(t.Surface).Border(1, t.Border).Radius(dialogRadius).Label(locales.RenameMenuItem())
		defaultEntry := m.menuEntry(config.MenuItem{ID: m.menuNameID})
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				m.entryMark(c, m.menuNameID, 20, t.Text)
				ui.Text(c, locales.RenameActionTitle(defaultEntry.label)).FontSize(sizeSection).FontWeight(600).SingleLine().Grow(1).MinWidth(0)
			})
			ui.TextInput(c.Key("menu-name"), &m.menuName).Label(locales.EditMenuName()).Placeholder(defaultEntry.label)
			ui.Text(c, locales.MenuNameHint()).FontSize(sizeNote).TextColor(t.TextMuted)
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if strings.TrimSpace(m.menuName) != "" && ghostButton(c, locales.RestoreDefaultName(), false) {
					m.menuName = ""
				}
				ui.Spacer(c)
				if quietButton(c, nil, locales.Cancel()) {
					m.menuNameOpen = false
				}
				if primaryButton(c, nil, locales.SaveChanges()) {
					for index := range m.cfg.MenuItems {
						if m.cfg.MenuItems[index].ID == m.menuNameID {
							m.cfg.MenuItems[index].Label = strings.TrimSpace(m.menuName)
						}
					}
					m.saveConfig()
					m.menuNameOpen = false
				}
			})
		})
	})
}
