package settings

import (
	"github.com/hy-shine/rightclick/internal/locales"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
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

// sidebar builds the pane list with a subtle accent for the selected pane.
func (m *app) sidebar(c *ui.Context) {
	// The window has no title bar of its own: the sidebar runs up to the top edge
	// and leaves room for the window controls, and the strip above the sign
	// moves the window.
	bar := c.TitleBar()
	ui.Column(c).Width(sidebarWidth).Padding(0, 10, 12).Children(func() {
		ui.Row(c).Height(max(bar.Height, 22)).DragWindow()
		ui.Row(c).Gap(8).Padding(6, 10).Margin(0, 0, 20).AlignItems(ui.Center).DragWindow().Children(func() {
			mark(c, 22)
			ui.Text(c, locales.AppName()).FontSize(14).FontWeight(600)
		})
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			for index := range paneTabs() {
				m.navRow(c, index)
			}
		})
	})
}

// navRow builds the row of the pane at index i. Selection uses a tinted surface
// and stronger text. When a setup
// prerequisite is still missing, the diagnostics row carries a warning mark.
func (m *app) navRow(c *ui.Context, i int) {
	t := c.Theme()
	p := paneTabs()[i]
	selected := i == m.tab
	label, glyphColor := t.Text, t.TextMuted
	bg := ui.Transparent
	if selected {
		bg = t.Background.Mix(t.Accent, 0.16)
		glyphColor = t.Accent.Mix(t.Text, 0.35)
		if c.Preferences().HighContrast {
			bg = t.Background.Mix(t.Accent, 0.26)
			glyphColor = t.Text
		}
	}
	step := m.setupStep()
	alert := p.id == paneDiagnostics && step.blocked()
	btn := ui.ButtonBase(c.Key("nav:"+p.id)).Gap(9).Padding(8, 10).AlignItems(ui.Center).Justify(ui.Start).Radius(t.Radius).Background(bg).
		Role(ui.RoleTab).Checked(selected).Label(p.label).
		OnClick(func() { m.selectPane(p.id) }).
		OnShortcut(0, ui.KeyDown, func() { m.selectPane(paneTabs()[min(i+1, len(paneTabs())-1)].id) }).
		OnShortcut(0, ui.KeyUp, func() { m.selectPane(paneTabs()[max(i-1, 0)].id) })
	if !selected {
		tint(btn, bg, t.SurfaceHover, t.SurfacePressed)
	}
	btn.Children(func() {
		glyph(c, paneIcons[p.id], 16, glyphColor)
		name := ui.Text(c, p.label).FontSize(sizeBody).TextColor(label).Grow(1)
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
