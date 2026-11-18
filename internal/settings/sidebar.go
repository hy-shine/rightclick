package settings

import (
	"github.com/hy-shine/rightclick/internal/actionicons"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"

	"github.com/egoist/mygo/ui"
)

// paneIcons are the icons of the sidebar rows, by pane ID.
var paneIcons = map[string]*ui.SVG{
	paneMenu:      iconMenu,
	paneTemplates: iconNew,
	paneFolders:   iconDir,
	paneOpeners:   iconOpen,
	paneDefaults:  actionicons.Icon("menu"),
	paneGeneral:   iconGeneral,
}

// sidebarWidth is the width of the pane list, in points.
const sidebarWidth = 188

// sidebar builds the pane list with a translucent vibrancy material when available.
func (m *app) sidebar(c *ui.Context) {
	t := c.Theme()
	bar := c.TitleBar()
	panel := ui.Column(c).Width(sidebarWidth).Padding(0, 10, 12)
	if !c.Vibrancy() {
		panel.Background(t.Background)
	}
	panel.Children(func() {
		ui.Row(c).Height(max(bar.Height, 22)).DragWindow()
		ui.Row(c).Gap(9).Padding(6, 10).Margin(0, 0, 12).AlignItems(ui.Center).DragWindow().Children(func() {
			mark(c, 24)
			ui.Column(c).Gap(1).Grow(1).MinWidth(0).Children(func() {
				ui.Text(c, locales.AppName()).FontSize(14).FontWeight(700).SingleLine()
			})
		})
		ui.Divider(c).Background(separatorColor(c).Alpha(0.65)).Margin(0, 4, 10)
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			for index := range len(paneTabs()) - 1 {
				m.navRow(c, index)
			}
		})
		ui.Divider(c).Background(separatorColor(c).Alpha(0.65)).Margin(8, 4, 8)
		m.navRow(c, len(paneTabs())-1)
	})
}

// navRow builds the row of the pane at index i. Selection uses a lifted card or tinted accent
// pill. When a setup prerequisite is still missing, General carries a warning indicator.
func (m *app) navRow(c *ui.Context, i int) {
	t := c.Theme()
	p := paneTabs()[i]
	selected := i == m.tab
	label, glyphColor := t.Text, t.TextMuted
	bg := ui.Transparent
	if selected {
		bg = t.Background.Mix(t.Accent, 0.16)
		glyphColor = t.Accent.Mix(t.Text, 0.25)
		if c.Vibrancy() {
			bg = t.Accent.Alpha(0.16)
		}
		if c.Preferences().HighContrast {
			bg = t.Background.Mix(t.Accent, 0.26)
			glyphColor = t.Text
		}
	}
	step := m.setupStep()
	alert := p.id == paneGeneral && step.blocked()
	btn := ui.ButtonBase(c.Key("nav:"+p.id)).Gap(9).Padding(7, 10).AlignItems(ui.Center).Justify(ui.Start).Radius(t.Radius+1).Background(bg).
		Role(ui.RoleTab).Checked(selected).Label(p.label).
		OnClick(func() { m.selectPane(p.id) }).
		OnShortcut(0, ui.KeyDown, func() { m.selectPane(paneTabs()[min(i+1, len(paneTabs())-1)].id) }).
		OnShortcut(0, ui.KeyUp, func() { m.selectPane(paneTabs()[max(i-1, 0)].id) })
	if !selected {
		tint(btn, bg, t.SurfaceHover.Alpha(0.75), t.SurfacePressed.Alpha(0.85))
	}
	btn.Children(func() {
		glyph(c, paneIcons[p.id], 16, glyphColor)
		name := ui.Text(c, p.label).FontSize(sizeBody).TextColor(label).Grow(1).SingleLine()
		if selected {
			name.FontWeight(600)
		}
		if alert {
			alertColor := t.Warning
			if selected {
				alertColor = label
			}
			ui.Icon(c, iconAlert).Size(14, 14).TextColor(alertColor).Label(p.label + " " + step.summary())
		}
	})
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
