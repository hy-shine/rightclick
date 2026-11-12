package settings

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"rightclick/config"
	"rightclick/locales"
)

func (m *app) hasShortcut(id string) bool {
	return slices.ContainsFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id })
}

func (m *app) setShortcut(id string, enabled bool) {
	if enabled {
		if !m.hasShortcut(id) {
			m.cfg.MenuItems = append(m.cfg.MenuItems, config.MenuItem{ID: id, Enabled: true, Location: config.FileTypeMain})
		}
	} else {
		m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == id })
	}
	m.saveConfig()
}

func (m *app) shortcutToggle(c *ui.Context, id, label string) {
	enabled := m.hasShortcut(id)
	ui.Checkbox(c.Key("shortcut:"+id), &enabled, "").Label(locales.ShortcutFor(label)).OnClick(func() {
		m.setShortcut(id, enabled)
	})
}

func (m *app) menuEntry(item config.MenuItem) menuEntry {
	entry := menuEntryFor(item.ID)
	kind, target, shortcut := strings.Cut(item.ID, ":")
	if shortcut {
		switch kind {
		case "type":
			for _, fileType := range m.cfg.FileTypes {
				if fileType.ID == target {
					entry = menuEntry{locales.NewShortcut(fileType.DisplayLabel()), locales.WhereBlankFolder()}
				}
			}
		case "template":
			for _, template := range m.templates {
				if template.Name == target {
					label := template.Label
					if label == "" {
						label = strings.TrimSuffix(template.Name, filepath.Ext(template.Name))
					}
					entry = menuEntry{locales.NewShortcut(label), locales.WhereBlankFolder()}
				}
			}
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
	case "type":
		return (context == ctxBlank || context == ctxFolder) && slices.ContainsFunc(m.cfg.FileTypes, func(fileType config.FileType) bool {
			return fileType.ID == target && fileType.Enabled
		})
	case "template":
		if context != ctxBlank && context != ctxFolder {
			return false
		}
		for _, template := range m.templates {
			if template.Name == target {
				return !template.Disabled
			}
		}
	case "app":
		return slices.Contains(m.cfg.Editors, target)
	case "folder":
		return slices.ContainsFunc(m.cfg.FavoriteFolders, func(folder config.FavoriteFolder) bool {
			return folder.ID == target && folder.Enabled
		})
	}
	return false
}

func (m *app) menuNameDialog(c *ui.Context) {
	ui.DialogBase(c, &m.menuNameOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.3))
		panel.Width(360).Padding(22).Background(c.Theme().Surface).Radius(10).Label(locales.RenameMenuItem())
		ui.Column(c).Gap(14).Children(func() {
			sectionTitle(c, locales.RenameMenuItem())
			ui.TextInput(c.Key("menu-name"), &m.menuName).Label(locales.EditMenuName())
			ui.Text(c, locales.MenuNameHint()).FontSize(sizeNote).TextColor(c.Theme().TextMuted)
			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
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
