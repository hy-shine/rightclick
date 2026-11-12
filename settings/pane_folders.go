package settings

import (
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
	paneHeader(c, "文件夹范围", "统一的文件夹白名单：右键助手只在下列文件夹及其子文件夹中出现。")
	if m.foldersChanged && banner(c, t.Warning, iconRestart, "文件夹范围还没有生效", "重新启动访达后，新的文件夹范围才会生效。", "重新启动") {
		m.confirmRestart = true
	}
	remove := -1
	if len(m.cfg.WatchFolders) == 0 {
		emptyState(c, iconDir, "还没有设置文件夹范围", "添加文件夹后，右键助手才会出现在其中。")
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
					ui.Text(c, "含子文件夹").FontSize(sizeCaption).TextColor(t.TextMuted)
					if ghostButton(c, "移除", true) {
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
	}
	m.addFolderActions(c, len(m.cfg.WatchFolders) == 0)
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
		paths, err := open(mygo.OpenDialogOptions{Parent: win, Title: "选择文件夹", Directory: true})
		if err == nil && len(paths) == 0 {
			return // canceled
		}
		win.Update(func() {
			if err != nil {
				m.report("添加失败", err)
				return
			}
			for _, p := range paths {
				if !slices.Contains(m.cfg.WatchFolders, p) {
					m.cfg.WatchFolders = append(m.cfg.WatchFolders, p)
					m.foldersChanged = true
				}
			}
			m.saveConfig()
		})
	}()
}
