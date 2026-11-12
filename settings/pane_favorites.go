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
	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/locales"
)

// favoriteRowHeight keeps the list dense: a folder needs a name and its path,
// but an archive of twenty of them should still fit on one screen.
const favoriteRowHeight = 48

func (m *app) favoritesPane(c *ui.Context) {
	paneHeader(c, locales.FavoriteFolders(), locales.FavoriteFoldersHint())
	ui.Text(c, locales.FavoriteArrangeHint()).FontSize(sizeNote).TextColor(c.Theme().TextMuted)
	if len(m.cfg.FavoriteFolders) == 0 {
		emptyState(c, iconDir, locales.NoFavoriteFolders(), locales.AddFavoriteHint())
	} else {
		removeID := ""
		m.favoriteList.Key = func(index int) any { return m.cfg.FavoriteFolders[index].ID }
		m.favoriteList.Reorder = func(rows []int, target int) {
			m.cfg.FavoriteFolders = moveRows(m.cfg.FavoriteFolders, rows, target)
			m.saveConfig()
		}
		ui.List(c, &m.favoriteList, len(m.cfg.FavoriteFolders), func(index int) {
			folder := &m.cfg.FavoriteFolders[index]
			ui.Row(c).Height(favoriteRowHeight).Gap(10).Padding(6, 12).AlignItems(ui.Center).Children(func() {
				glyph(c, iconGrip, 10, c.Theme().TextMuted)
				glyph(c, iconDir, 18, c.Theme().TextMuted)
				ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
					ui.Text(c, folder.Label).FontSize(sizeBody).SingleLine()
					ui.Text(c, tildePath(folder.Path)).FontSize(sizeNote).TextColor(c.Theme().TextMuted).SingleLine()
				})
				m.shortcutToggleLabeled(c, "folder:"+folder.ID, folder.Label, locales.QuickAccess())
				ui.Switch(c.Key("favorite-enabled:"+folder.ID), &folder.Enabled).Label(locales.EnableLabel(folder.Label)).OnChange(m.saveConfig)
				if ghostButton(c, locales.EditMenuName(), false) {
					m.favoriteEditID, m.favoriteName, m.favoriteEditOpen = folder.ID, folder.Label, true
				}
				if iconButton(c.Key("remove-favorite:"+folder.ID), iconTrash, locales.Remove(), true) {
					removeID = folder.ID
				}
			})
		}).Height(float32(len(m.cfg.FavoriteFolders))*favoriteRowHeight).Border(1, separatorColor(c)).Radius(8)
		// Removing waits until the list has been built, as in the other panes.
		if removeID != "" {
			m.removeFavorite(removeID)
		}
	}
	if primaryButton(c, iconPlus, locales.AddFavorite()) {
		m.pickFavorite()
	}
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
		panel.Width(360).Padding(22).Background(c.Theme().Surface).Radius(10).Label(locales.EditMenuName())
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
