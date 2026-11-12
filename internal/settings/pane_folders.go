package settings

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

// foldersPane is the favorite folders pane: the targets of Copy To and Move To, and
// the folders that can be added to the menu.
func (m *app) foldersPane(c *ui.Context) {
	paneHeader(c, locales.PaneFolders(), locales.FoldersIntro())
	ui.Column(c).Gap(12).Children(func() { m.folderToolbarAndList(c) })
}

// folderToolbarAndList is the toolbar and the list under it, 12 points apart. The add
// action sits at the right end of the toolbar, as in the New pane.
func (m *app) folderToolbarAndList(c *ui.Context) {
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Spacer(c)
		if primaryButtonCentered(c, iconPlus, locales.AddFolder()) {
			m.pickFavorite()
		}
	})
	m.favoriteRows(c)
}

// folderName is the last part of a folder path, for the first line of a row.
func folderName(p string) string {
	name := filepath.Base(filepath.Clean(p))
	if name == "." || name == string(filepath.Separator) {
		return tildePath(p)
	}
	return name
}

// tildePath shortens a path in the home folder to start with "~", as Finder's
// path bar and shells do.
func tildePath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(p, home+"/"); ok {
		return "~/" + rest
	}
	return p
}
