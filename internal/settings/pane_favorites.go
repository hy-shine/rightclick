package settings

import (
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

const (
	favoriteHeaderHeight  = 32
	favoriteShortcutWidth = 112
	favoriteEnabledWidth  = 68
)

// favoriteRows lists the favorite folders: the targets of Copy To and Move To, and the
// folders that can be added to the menu. Adding is in the toolbar of the folder pane.
func (m *app) favoriteRows(c *ui.Context) {
	if len(m.cfg.FavoriteFolders) == 0 {
		emptyState(c, iconDir, locales.NoFavoriteFolders(), locales.AddFavoriteHint(), func() {
			if primaryButton(c, iconPlus, locales.AddFolder()) {
				m.pickFavorite()
			}
		})
	} else {
		removeID := ""
		m.favoriteList.Key = func(index int) any { return m.cfg.FavoriteFolders[index].ID }
		m.favoriteList.Reorder = func(rows []int, target int) {
			m.cfg.FavoriteFolders = moveRows(m.cfg.FavoriteFolders, rows, target)
			m.saveConfig()
		}
		ledger(c, ui.Column(c)).Children(func() {
			ui.Row(c).Height(favoriteHeaderHeight).Gap(12).Padding(0, 14).AlignItems(ui.Center).Children(func() {
				ui.Text(c, locales.ContextFolder()).FontSize(sizeBody).FontWeight(600).TextColor(c.Theme().TextMuted).Grow(1)
				ui.Text(c, locales.ColEnabled()).FontSize(sizeBody).FontWeight(600).TextColor(c.Theme().TextMuted).Width(favoriteEnabledWidth).TextAlign(ui.Center)
				ui.Text(c, locales.QuickAccess()).FontSize(sizeBody).FontWeight(600).TextColor(c.Theme().TextMuted).Width(favoriteShortcutWidth).TextAlign(ui.Center)
				ui.Text(c, locales.ColActionContext()).FontSize(sizeBody).FontWeight(600).TextColor(c.Theme().TextMuted).Width(76).TextAlign(ui.Center)
			})
			ui.List(c, &m.favoriteList, len(m.cfg.FavoriteFolders), func(index int) {
				folder := &m.cfg.FavoriteFolders[index]
				itemRow(c, true, func() {
					glyphTile(c, iconDir)
				}, folder.Label, tildePath(folder.Path), func() {
					ui.Row(c).Width(favoriteEnabledWidth).Justify(ui.Center).Children(func() {
						ui.Switch(c.Key("favorite-enabled:"+folder.ID), &folder.Enabled).Label(locales.EnableLabel(folder.Label)).OnChange(m.saveConfig)
					})
					ui.Row(c).Width(favoriteShortcutWidth).Justify(ui.Center).Children(func() {
						id := "folder:" + folder.ID
						enabled := m.hasShortcut(id)
						ui.Switch(c.Key("shortcut:"+id), &enabled).Label(locales.ShortcutFor(folder.Label)).OnChange(func() {
							m.setShortcut(id, enabled)
						})
					})
					ui.Row(c).Width(76).Gap(12).AlignItems(ui.Center).Children(func() {
						if iconButton(c.Key("edit-favorite:"+folder.ID), iconEdit, locales.EditMenuName(), false) {
							m.favoriteEditID, m.favoriteName, m.favoriteEditOpen = folder.ID, folder.Label, true
						}
						if iconButton(c.Key("remove-favorite:"+folder.ID), iconTrash, locales.Delete(), true) {
							removeID = folder.ID
						}
					})
				})
			}).Height(float32(len(m.cfg.FavoriteFolders))*itemRowHeight).Dividers(1, separatorColor(c))
		})
		// Removing waits until the list has been built, as in the other panes.
		if removeID != "" {
			m.removeFavorite(removeID)
		}
	}
	ui.Text(c, locales.FavoriteFoldersHint()).FontSize(sizeNote).TextColor(c.Theme().TextMuted)
}

func (m *app) addFavorite(path string) error {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(path) || !info.IsDir() {
		return errors.New(locales.FavoriteUnavailable())
	}
	if slices.ContainsFunc(m.cfg.FavoriteFolders, func(folder config.FavoriteFolder) bool { return folder.Path == path }) {
		return nil
	}
	m.cfg.FavoriteFolders = append(m.cfg.FavoriteFolders, config.FavoriteFolder{
		ID: rand.Text(), Path: path, Label: folderName(path), Enabled: true,
	})
	if !m.hasShortcut("favorite-folders") {
		m.cfg.MenuItems = append(m.cfg.MenuItems, config.MenuItem{ID: "favorite-folders", Enabled: true, Location: config.FileTypeSubmenu})
	}
	m.saveConfig()
	return nil
}

func (m *app) pickFavorite() {
	window, open := m.win, m.env.openDialog
	go func() {
		paths, err := open(mygo.OpenDialogOptions{Parent: window, Title: locales.AddFavorite(), Directory: true})
		window.Update(func() {
			if err != nil {
				m.report(locales.AddFailed(), err)
				return
			}
			for _, path := range paths {
				if err := m.addFavorite(path); err != nil {
					m.report(locales.AddFailed(), err)
					return
				}
			}
		})
	}()
}

func (m *app) removeFavorite(id string) {
	m.cfg.FavoriteFolders = slices.DeleteFunc(m.cfg.FavoriteFolders, func(folder config.FavoriteFolder) bool { return folder.ID == id })
	m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == "folder:"+id })
	m.saveConfig()
}

func (m *app) favoriteDialog(c *ui.Context) {
	ui.DialogBase(c, &m.favoriteEditOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.3))
		panel.Width(360).Padding(22).Background(c.Theme().Surface).Border(1, c.Theme().Border).Radius(dialogRadius).Label(locales.EditMenuName())
		ui.Column(c).Gap(14).Children(func() {
			sectionTitle(c, locales.EditMenuName())
			ui.TextInput(c.Key("favorite-name"), &m.favoriteName).Label(locales.EditMenuName())
			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if quietButton(c, nil, locales.Cancel()) {
					m.favoriteEditOpen = false
				}
				if primaryButton(c, nil, locales.SaveChanges()) && strings.TrimSpace(m.favoriteName) != "" {
					for index := range m.cfg.FavoriteFolders {
						if m.cfg.FavoriteFolders[index].ID == m.favoriteEditID {
							m.cfg.FavoriteFolders[index].Label = strings.TrimSpace(m.favoriteName)
						}
					}
					m.saveConfig()
					m.favoriteEditOpen = false
				}
			})
		})
	})
}
