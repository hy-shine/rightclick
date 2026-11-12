package settings

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
)

// paneIcons are the icons of the sidebar rows, by pane ID.
var paneIcons = map[string]*ui.SVG{
	paneMenu:        iconMenu,
	paneTemplates:   iconNew,
	paneFolders:     iconDir,
	paneOpeners:     iconOpen,
	paneDiagnostics: iconShield,
	paneGeneral:     iconGeneral,
}

// sidebarWidth is the width of the pane list, in points.
const sidebarWidth = 176

// sidebar builds the pane list on the desk, left of the sheet of paper. The
// chosen pane is the one raised as a white panel. Each row shows the state its
// pane controls, so the list says what is set up without opening a pane.
func (m *app) sidebar(c *ui.Context) {
	// The window has no title bar of its own: the sidebar runs up to the top edge
	// and leaves room for the window controls, and the strip above the sign
	// moves the window.
	bar := c.TitleBar()
	ui.Column(c).Width(sidebarWidth).Padding(0, 10, 12).Children(func() {
		ui.Row(c).Height(max(bar.Height, 22)).DragWindow()
		ui.Row(c).Gap(9).Padding(0, 10).Margin(0, 0, 22).AlignItems(ui.Center).DragWindow().Children(func() {
			mark(c, 28)
			ui.Text(c, "右键助手").FontSize(17).FontWeight(700)
		})
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			for index := range paneTabs {
				m.navRow(c, index)
			}
		})
	})
}

// navRow builds the row of the pane at index i. The chosen pane's row is a white
// panel on the desk, with its name in bold: the list marks a choice without color.
func (m *app) navRow(c *ui.Context, i int) {
	t := c.Theme()
	p := paneTabs[i]
	selected := i == m.tab
	label, glyphColor := t.Text, t.TextMuted
	bg := ui.Transparent
	if selected {
		bg = t.SurfaceHover
		label, glyphColor = t.Text, t.Text
	}
	ui.ButtonBase(c.Key("nav:"+p.id)).Gap(10).Padding(11, 10).AlignItems(ui.Center).Justify(ui.Start).Radius(t.Radius+2).Background(bg).
		FocusRing(false).Role(ui.RoleTab).Checked(selected).Label(p.label).
		OnClick(func() { m.selectPane(p.id) }).
		OnShortcut(0, ui.KeyDown, func() { m.selectPane(paneTabs[min(i+1, len(paneTabs)-1)].id) }).
		OnShortcut(0, ui.KeyUp, func() { m.selectPane(paneTabs[max(i-1, 0)].id) }).Children(func() {
		glyph(c, paneIcons[p.id], 17, glyphColor)
		name := ui.Text(c, p.label).FontSize(sizeBody).TextColor(label)
		if selected {
			name.Bold()
		}
	})
}

// navSummary is the line under a pane's name in the sidebar.
func (m *app) navSummary(id string) string {
	switch id {
	case paneMenu:
		// A missing step means the menu does not show in Finder, so the count
		// would claim something that is not true yet.
		if step := m.setupStep(); step.blocked() {
			return step.summary()
		}
		return fmt.Sprintf("已启用 %d 项", enabledCount(m.cfg.MenuItems))
	case paneTemplates:
		return fmt.Sprintf("%d 种类型 · %d 个模板", len(m.cfg.FileTypes), len(m.templates))
	case paneFolders:
		if len(m.cfg.WatchFolders) == 0 {
			return "未设置文件夹范围"
		}
		return fmt.Sprintf("%d 个文件夹", len(m.cfg.WatchFolders))
	case paneOpeners:
		return "终端：" + m.cfg.Terminal
	case paneDiagnostics:
		return m.setupStep().summary()
	case paneGeneral:
		if m.openAtLogin {
			return "开机时启动"
		}
		return "开机时不启动"
	}
	return ""
}

// enabledCount counts the menu entries that are switched on.
func enabledCount(items []config.MenuItem) int {
	n := 0
	for _, item := range items {
		if item.Enabled {
			n++
		}
	}
	return n
}
