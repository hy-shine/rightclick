package settings

import (
	"github.com/hy-shine/rightclick/files"
	"github.com/hy-shine/rightclick/locales"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// foldersPane edits the single folder whitelist that the Finder extension and
// every action share.
func (m *app) foldersPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, locales.PaneFolders(), locales.FolderScopeIntro())
	if m.folderAccess == nil {
		// A pane built without paneChanged still shows the real access state.
		m.refreshFolderAccess()
	}
	if m.foldersChanged && banner(c, t.Warning, iconRestart, locales.FolderScopePending(), locales.FolderScopePendingBody(), locales.Restart()) {
		m.confirmRestart = true
	}
	if m.foldersNeedAccess() && banner(c, t.Warning, iconAlert, locales.FolderAccessTitle(), locales.FolderAccessBody(), locales.GrantAccess()) {
		m.openPrivacySettings(privacyFilesAndFolders)
	}
	remove := -1
	if len(m.cfg.WatchFolders) == 0 {
		emptyState(c, iconDir, locales.NoFolderScopeTitle(), locales.NoFolderScopeBody())
	} else {
		ledger(c, ui.Column(c)).Children(func() {
			for i, dir := range m.cfg.WatchFolders {
				ui.Row(c).Gap(12).Padding(12, 14).AlignItems(ui.Center).Children(func() {
					ui.Row(c).Size(30, 30).Radius(8).Background(t.Background).
						AlignItems(ui.Center).Justify(ui.Center).Children(func() {
						glyph(c, iconDir, 16, t.TextMuted)
					})
					ui.Column(c).Grow(1).Gap(2).Children(func() {
						ui.Text(c, folderName(dir)).FontSize(sizeBody).FontWeight(600).SingleLine()
						ui.Text(c, tildePath(dir)).FontSize(sizeNote).TextColor(t.TextMuted).SingleLine()
					})
					folderAccessBadge(c, m.folderAccessOf(dir))
					if iconButton(c, iconTrash, locales.Remove(), true) {
						remove = i
					}
				})
			}
		})
	}
	// Removing happens after the rows are built, so that the loop does not
	// index a list that is changing.
	if remove >= 0 {
		m.cfg.WatchFolders = slices.Delete(m.cfg.WatchFolders, remove, remove+1)
		m.foldersChanged = true
		m.saveConfig()
		m.refreshFolderAccess()
	}
	m.addFolderActions(c, len(m.cfg.WatchFolders) == 0)
	ui.Row(c).Gap(8).Children(func() {
		if quietButton(c, iconRestart, locales.CheckAgain()) {
			m.refreshFolderAccess()
		}
	})
}

// refreshFolderAccess reads whether each watched folder can be listed. macOS
// answers with a permission error when the user refused access, which is what
// the badge shows. The read runs on the main thread when the pane is shown: the
// watched folders are local in practice, and a network volume can make the pane
// wait.
func (m *app) refreshFolderAccess() {
	access := make(map[string]files.Access, len(m.cfg.WatchFolders))
	for _, dir := range m.cfg.WatchFolders {
		access[dir] = m.env.folderAccess(dir)
	}
	m.folderAccess = access
}

// folderAccessOf reports what the last check found for dir.
func (m *app) folderAccessOf(dir string) files.Access {
	return m.folderAccess[dir]
}

// foldersNeedAccess reports whether the user refused access to any watched
// folder. A missing folder gets its own badge and no grant prompt.
func (m *app) foldersNeedAccess() bool {
	for _, dir := range m.cfg.WatchFolders {
		if m.folderAccess[dir] == files.AccessDenied {
			return true
		}
	}
	return false
}

// folderAccessBadge shows whether the app may read the folder. It replaces the
// plain "Includes subfolders" note, as ui.md section 3.3 describes.
func folderAccessBadge(c *ui.Context, access files.Access) {
	t := c.Theme()
	s, tone, label := iconAlert, t.Warning, locales.FolderNeedsAccess()
	switch access {
	case files.AccessOK:
		s, tone, label = iconCheck, t.Success, locales.FolderGranted()
	case files.AccessMissing:
		label = locales.FolderMissing()
	}
	ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
		glyph(c, s, 14, tone)
		ui.Text(c, label).FontSize(sizeCaption).FontWeight(600).TextColor(tone)
	})
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

// pickFolder asks for a folder and adds it to the watched folders. The dialog
// runs off the main thread.
func (m *app) pickFolder() {
	win, open := m.win, m.env.openDialog
	go func() {
		paths, err := open(mygo.OpenDialogOptions{Parent: win, Title: locales.ChooseFolder(), Directory: true})
		if err == nil && len(paths) == 0 {
			return // canceled
		}
		win.Update(func() {
			if err != nil {
				m.report(locales.AddFailed(), err)
				return
			}
			for _, p := range paths {
				if !slices.Contains(m.cfg.WatchFolders, p) {
					m.cfg.WatchFolders = append(m.cfg.WatchFolders, p)
					m.foldersChanged = true
				}
			}
			m.saveConfig()
			m.refreshFolderAccess()
		})
	}()
}
