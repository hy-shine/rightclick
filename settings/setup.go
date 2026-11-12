package settings

import (
	"slices"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
)

// setupStep is the first thing still missing before the Finder menu can show,
// in the order the user should fix them. stepReady means it can show.
type setupStep int

const (
	stepChecking  setupStep = iota // the Finder extension state is not known yet
	stepExtension                  // the Finder extension is off
	stepFolder                     // the folder whitelist is empty
	stepItem                       // no menu entry is switched on
	stepRestart                    // a folder changed since Finder last started
	stepReady                      // the menu can show in Finder
)

// setupStep returns the first missing step. Finder shows the menu only when all
// of them hold, so the sidebar, the menu page and the diagnostics all read this.
func (m *app) setupStep() setupStep {
	switch {
	case m.finder == finderChecking:
		return stepChecking
	case m.finder == finderOff:
		return stepExtension
	case len(m.cfg.WatchFolders) == 0:
		return stepFolder
	case enabledCount(m.cfg.MenuItems) == 0:
		return stepItem
	case m.foldersChanged:
		return stepRestart
	}
	return stepReady
}

// blocked reports whether a step that keeps the menu from showing in Finder is
// missing. A check still running is not a blocker.
func (s setupStep) blocked() bool {
	return s != stepChecking && s != stepReady
}

// summary is the short conclusion of a step, for the sidebar.
func (s setupStep) summary() string {
	switch s {
	case stepExtension:
		return "需要开启扩展"
	case stepFolder:
		return "等待设置文件夹范围"
	case stepItem:
		return "没有启用的项"
	case stepRestart:
		return "需要重新启动访达"
	case stepReady:
		return "可以使用"
	}
	return "检查中…"
}

// setupCard names the one step that is still missing, with the action that
// completes it, so the first thing on the menu page is what to do next. It shows
// nothing when the menu can show in Finder, or while the extension is checked.
func (m *app) setupCard(c *ui.Context) {
	t := c.Theme()
	switch m.setupStep() {
	case stepExtension:
		if banner(c, t.Warning, iconAlert, "Finder 扩展还没有启用", "启用后，访达的右键菜单里才会出现这些项。", "去启用") {
			m.openFinderSettings()
		}
	case stepFolder:
		ui.Column(c).Gap(10).Children(func() {
			banner(c, t.Warning, iconAlert, "还没有设置文件夹范围", "右键助手只会在你添加的文件夹及其子文件夹里出现。", "")
			m.addFolderActions(c, true)
		})
	case stepItem:
		if banner(c, t.Warning, iconAlert, "没有启用任何菜单项", "打开至少一项，访达的右键菜单才会有内容。", "恢复默认") {
			m.restoreDefaultMenu()
		}
	case stepRestart:
		if banner(c, t.Warning, iconRestart, "需要重新启动访达", "文件夹范围变更后，要重新启动访达才会读取。", "重新启动") {
			m.confirmRestart = true
		}
	}
}

// addFolderActions offers the ways to add to the folder whitelist: the picker,
// and with offerDefaults, the Desktop, Documents and Downloads folders at once.
func (m *app) addFolderActions(c *ui.Context, offerDefaults bool) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		if primaryButton(c, iconPlus, "添加文件夹…") {
			m.pickFolder()
		}
		if offerDefaults && quietButton(c, nil, "使用默认文件夹（桌面、文稿、下载）") {
			m.useDefaultFolders()
		}
	})
}

// useDefaultFolders adds the folders that config.Default names to the whitelist.
// Folders that are already in it stay as they are.
func (m *app) useDefaultFolders() {
	added := false
	for _, dir := range config.Default().WatchFolders {
		if !slices.Contains(m.cfg.WatchFolders, dir) {
			m.cfg.WatchFolders = append(m.cfg.WatchFolders, dir)
			added = true
		}
	}
	if added {
		m.foldersChanged = true
	}
	m.saveConfig()
}
