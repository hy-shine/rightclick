package settings

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// finderState is what the diagnostics pane shows about the Finder extension.
type finderState int

const (
	finderChecking finderState = iota
	finderOn
	finderOff
)

// diagnosticsPane lists the four steps that the Finder menu needs, each with the
// action that completes it. The conclusion is shown only when all four pass.
func (m *app) diagnosticsPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, "权限与诊断", "右键菜单要四步都通过，访达里才会显示。")
	if m.setupStep() == stepReady {
		banner(c, t.Success, iconCheck, "访达中已经可用", "在文件夹范围内右键，就能看到右键助手的菜单。", "")
	}
	ledger(c, ui.Column(c)).Children(func() {
		extension, extensionAction := "在系统设置的扩展里打开右键助手。", "去启用"
		switch m.finder {
		case finderChecking:
			extension, extensionAction = "正在检查…", ""
		case finderOn:
			extension, extensionAction = "已在系统设置里启用。", ""
		}
		if checkRow(c, "Finder 扩展已启用", m.finder == finderOn, extension, extensionAction) {
			m.openFinderSettings()
		}

		folders, foldersAction := "访达只在文件夹范围内显示菜单。", "添加…"
		if n := len(m.cfg.WatchFolders); n > 0 {
			folders, foldersAction = fmt.Sprintf("已添加 %d 个文件夹。", n), ""
		}
		if checkRow(c, "已设置文件夹范围", len(m.cfg.WatchFolders) > 0, folders, foldersAction) {
			m.pickFolder()
		}

		items, itemsAction := "至少打开一项菜单。", "去编辑"
		if n := enabledCount(m.cfg.MenuItems); n > 0 {
			items, itemsAction = fmt.Sprintf("已启用 %d 项。", n), ""
		}
		if checkRow(c, "有启用的菜单项", enabledCount(m.cfg.MenuItems) > 0, items, itemsAction) {
			m.selectPane(paneMenu)
		}

		restart, restartAction := "文件夹范围变更后，要重新启动访达才会读取。", "重新启动"
		if !m.foldersChanged {
			restart, restartAction = "文件夹范围已生效。", ""
		}
		if checkRow(c, "访达已读取最新的文件夹范围", !m.foldersChanged, restart, restartAction) {
			m.confirmRestart = true
		}
	})
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, "重新启动访达", "关闭访达，然后由系统重新打开。菜单没有出现时可以试试。", func() {
			if quietButton(c, iconRestart, "重新启动") {
				m.confirmRestart = true
			}
		})
	})
}

// checkRow is one step of the diagnostics checklist. A passed step shows a check
// mark. A missing step shows a warning and, when action is not empty, a button
// with that label, which checkRow reports as a click.
func checkRow(c *ui.Context, title string, pass bool, note, action string) bool {
	t := c.Theme()
	clicked := false
	ui.Row(c).Gap(14).Padding(14, 16).AlignItems(ui.Center).Children(func() {
		if pass {
			glyph(c, iconCheck, 17, t.Success)
		} else {
			glyph(c, iconAlert, 17, t.Warning)
		}
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, title).FontSize(sizeBody).FontWeight(600)
			ui.Text(c, note).FontSize(sizeNote).LineHeight(1.45).TextColor(t.TextMuted)
		})
		if action != "" {
			clicked = quietButton(c, nil, action)
		}
	})
	return clicked
}

// checkFinder asks the system whether the Finder extension is on. The check
// runs off the main thread, and its result comes back through the window's
// Update. A result that a later check has overtaken is dropped.
func (m *app) checkFinder() {
	m.finder = finderChecking
	m.finderSeq++
	seq := m.finderSeq
	win, check := m.win, m.env.finderEnabled
	go func() {
		on := check()
		win.Update(func() {
			if seq != m.finderSeq {
				return
			}
			m.finder = finderOff
			if on {
				m.finder = finderOn
			}
		})
	}()
}

// openFinderSettings opens the System Settings pane for extensions.
func (m *app) openFinderSettings() {
	m.report("打开系统设置失败", m.env.openFinderSettings())
}

// restartFinder restarts Finder off the main thread. When it works, Finder has
// read the watched folders again, so the folders no longer count as changed.
func (m *app) restartFinder() {
	win, restart := m.win, m.env.restartFinder
	go func() {
		err := restart()
		win.Update(func() {
			if err == nil {
				m.foldersChanged = false
			}
			m.report("重新启动访达失败", err)
		})
	}()
}

// generalPane holds the switch for starting the app at login, and below it the
// app's name and version, so that the sidebar needs no separate About pane.
func (m *app) generalPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, "通用", "")
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, "开机启动", "登录 macOS 后自动启动右键助手。", func() {
			ui.Switch(c, &m.openAtLogin).Label("开机启动").OnChange(m.applyOpenAtLogin)
		})
	})
	sectionTitle(c, "关于")
	ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
		mark(c, 56)
		ui.Column(c).Gap(4).Children(func() {
			ui.Text(c, "右键助手").FontSize(22).FontWeight(700)
			ui.Text(c, "版本 "+version).FontSize(sizeBody).TextColor(t.TextMuted)
		})
	})
	ui.Text(c, "给访达的右键菜单加上新建、拷贝路径、压缩和哈希等常用动作。").
		FontSize(sizeBody).LineHeight(1.6).TextColor(t.TextMuted)
}

// applyOpenAtLogin sets the login item, then shows what the system reports,
// so that a failed change does not leave the switch in the wrong place.
func (m *app) applyOpenAtLogin() {
	err := m.env.setOpenAtLogin(m.openAtLogin)
	m.openAtLogin = m.env.openAtLogin()
	m.report("设置开机启动失败", err)
}
