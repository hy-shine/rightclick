package settings

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Type scale, in points: the pane title, a section heading, body text, the
// line under a title, and the smallest note. Spacing runs 4, 8, 12, 16, 24, 32.
const (
	sizeTitle   = 22
	sizeSection = 16
	sizeBody    = 13
	sizeNote    = 12
	sizeCaption = 11
)

// fade is how long a control takes to change color when the pointer reaches it.
const fade = 120 * time.Millisecond

// glyph draws the icon s, size points square, in color.
func glyph(c *ui.Context, s *ui.SVG, size float32, color ui.Color) {
	ui.Icon(c, s).Size(size, size).TextColor(color)
}

// appMark is the app's icon: a menu card, with the row of
// this app lit, on a blue square. It matches resources/icon.png.
var appMark = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 824 824">
<defs><linearGradient id="blue" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#60a5fa"/><stop offset="1" stop-color="#2563eb"/></linearGradient></defs>
<rect width="824" height="824" rx="190" fill="url(#blue)"/>
<g transform="translate(-100 -100)">
<rect x="262" y="232" width="500" height="560" rx="44" fill="#ffffff"/>
<rect x="326" y="309" width="200" height="22" rx="11" fill="#cbd5e1"/>
<rect x="326" y="405" width="268" height="22" rx="11" fill="#cbd5e1"/>
<rect x="290" y="470" width="444" height="84" rx="24" fill="#2563eb"/>
<rect x="326" y="501" width="212" height="22" rx="11" fill="#ffffff"/>
<path d="M676 490l26 22-26 22" fill="none" stroke="#ffffff" stroke-width="13" stroke-linecap="round" stroke-linejoin="round"/>
<rect x="326" y="597" width="236" height="22" rx="11" fill="#cbd5e1"/>
<rect x="326" y="693" width="172" height="22" rx="11" fill="#cbd5e1"/>
</g></svg>`))

// mark shows the app's icon, size points square. It heads the sidebar and the
// general pane.
func mark(c *ui.Context, size float32) {
	ui.Image(c, appMark).Size(size, size)
}

// paneHeader writes the title of a pane in bold and, when desc is not empty,
// the line under it.
func paneHeader(c *ui.Context, title, desc string) {
	t := c.Theme()
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, title).FontSize(sizeTitle).FontWeight(600)
		if desc != "" {
			ui.Text(c, desc).FontSize(sizeNote).LineHeight(1.5).TextColor(t.TextMuted)
		}
	})
}

// sectionTitle writes the heading of a part of a pane, in bold.
func sectionTitle(c *ui.Context, s string) {
	ui.Text(c, s).FontSize(sizeSection).FontWeight(600)
}

// ledger gives the rows of e one shared border, with a line between each pair
// of rows. Rows sit on the paper and are set apart by lines, not boxed one by
// one.
func ledger(c *ui.Context, e ui.Element) ui.Element {
	t := c.Theme()
	return e.Background(t.Surface).Border(1, separatorColor(c)).Radius(t.Radius+2).Dividers(1, separatorColor(c))
}

// tint colors the background of e by what the pointer does to it, fading
// between the colors. rest is the color when the pointer is elsewhere.
func tint(e ui.Element, rest, hover, pressed ui.Color) {
	target := rest
	switch {
	case e.Pressed():
		target = pressed
	case e.Hovered():
		target = hover
	}
	e.Background(e.AnimateColor("bg", target, ui.Animation{Duration: fade}))
}

// controlHeight is the height every button control uses, so actions that sit
// side by side share one baseline.
const controlHeight = 30

// primaryButton is the one filled button of a view: the action to take next. It
// takes the brand blue, with icon s before the label when s is not nil. It
// reports a click.
func primaryButton(c *ui.Context, s *ui.SVG, label string) bool {
	return primaryButtonAligned(c, s, label, ui.Start)
}

// primaryButtonCentered is a primaryButton that centers itself in a row, for
// actions that sit beside a heading instead of at the start of a pane. Without
// it, a centered quiet button next to a start-aligned primary one leaves the
// pair on different baselines.
func primaryButtonCentered(c *ui.Context, s *ui.SVG, label string) bool {
	return primaryButtonAligned(c, s, label, ui.Center)
}

func primaryButtonAligned(c *ui.Context, s *ui.SVG, label string, align ui.Align) bool {
	t := c.Theme()
	fill, hover, pressed, text := t.Accent, t.AccentHover, t.AccentPressed, t.AccentText
	b := ui.ButtonBase(c).Gap(6).Padding(7, 14).Height(controlHeight).Radius(t.Radius).AlignSelf(align).Label(label).Children(func() {
		if s != nil {
			glyph(c, s, 14, text)
		}
		ui.Text(c, label).FontSize(sizeBody).FontWeight(600).TextColor(text)
	})
	tint(b, fill, hover, pressed)
	return b.Clicked()
}

// quietButton is a button with a border and no fill, for actions that are not
// the next step. It reports a click.
func quietButton(c *ui.Context, s *ui.SVG, label string) bool {
	return quietButtonAligned(c, s, label, ui.Start)
}

// quietButtonCentered is a quietButton that centers itself in a row, for
// actions that sit beside a heading instead of at the start of a pane.
func quietButtonCentered(c *ui.Context, s *ui.SVG, label string) bool {
	return quietButtonAligned(c, s, label, ui.Center)
}

func quietButtonAligned(c *ui.Context, s *ui.SVG, label string, align ui.Align) bool {
	t := c.Theme()
	b := ui.ButtonBase(c).Gap(6).Padding(7, 12).Height(controlHeight).Radius(t.Radius).AlignSelf(align).
		Border(1, separatorColor(c)).Label(label).Children(func() {
		if s != nil {
			glyph(c, s, 14, t.TextMuted)
		}
		ui.Text(c, label).FontSize(sizeBody).TextColor(t.Text)
	})
	tint(b, t.Surface, t.SurfaceHover, t.SurfacePressed)
	return b.Clicked()
}

// ghostButton is text with no border, for actions that sit in every row and
// would shout if boxed. It shows its color only while the pointer is over it.
// danger sets that color to the danger color instead of the text color. It
// reports a click.
func ghostButton(c *ui.Context, label string, danger bool) bool {
	t := c.Theme()
	hover := t.Text
	if danger {
		hover = t.Danger
	}
	b := ui.ButtonBase(c).Padding(5, 10).Radius(t.Radius).Label(label)
	var text ui.Element
	b.Children(func() {
		text = ui.Text(c, label).FontSize(sizeNote).TextColor(t.TextMuted)
	})
	if b.Hovered() {
		text.TextColor(hover)
	}
	tint(b, ui.Transparent, t.SurfaceHover, t.SurfacePressed)
	return b.Clicked()
}

// iconButton is a ghostButton that shows only its icon. The label stays as the
// accessible name. The icon turns danger-colored on hover, as the text of a
// ghostButton does.
func iconButton(c *ui.Context, s *ui.SVG, label string, danger bool) bool {
	t := c.Theme()
	hover := t.Text
	if danger {
		hover = t.Danger
	}
	b := ui.ButtonBase(c).Padding(5, 6).Radius(t.Radius).Label(label)
	var mark ui.Element
	b.Children(func() {
		mark = ui.Icon(c, s).Size(15, 15).TextColor(t.TextMuted)
	})
	if b.Hovered() {
		mark.TextColor(hover)
	}
	tint(b, ui.Transparent, t.SurfaceHover, t.SurfacePressed)
	return b.Clicked()
}

// banner states one fact about the app, as a tinted band with an icon: tone is
// the color of the fact, s its icon, title the fact and note what follows from
// it. When action is not empty the band ends in a filled button, and banner
// reports a click on it.
func banner(c *ui.Context, tone ui.Color, s *ui.SVG, title, note, action string) bool {
	clicked := false
	ui.Row(c).Gap(12).Padding(12, 14).AlignItems(ui.Center).Radius(10).
		Background(tone.Alpha(0.1)).Border(1, tone.Alpha(0.3)).Children(func() {
		ui.Row(c).Size(32, 32).Radius(16).Background(tone.Alpha(0.16)).
			AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			glyph(c, s, 17, tone)
		})
		ui.Column(c).Grow(1).Gap(2).Children(func() {
			ui.Text(c, title).FontSize(sizeBody).FontWeight(600)
			if note != "" {
				ui.Text(c, note).FontSize(sizeNote).TextColor(c.Theme().TextMuted)
			}
		})
		if action != "" {
			clicked = primaryButton(c, nil, action)
		}
	})
	return clicked
}

// emptyState fills the place of a list that has nothing in it: a dashed outline
// with an icon, what is missing, and what to do about it. The caller puts the
// action under it.
func emptyState(c *ui.Context, s *ui.SVG, title, note string) {
	t := c.Theme()
	ui.Column(c).Gap(6).Padding(28, 20).AlignItems(ui.Center).Radius(t.Radius+6).
		Border(1, t.Border).BorderStyle(ui.BorderDashed).Children(func() {
		ui.Row(c).Size(40, 40).Radius(20).Background(t.Background).Margin(0, 0, 6).
			AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			glyph(c, s, 19, t.TextMuted)
		})
		ui.Text(c, title).FontSize(sizeBody).FontWeight(600)
		ui.Text(c, note).FontSize(sizeNote).LineHeight(1.5).TextColor(t.TextMuted).TextAlign(ui.Center).MaxWidth(320)
	})
}

// settingRow is one row of a ledger that names a setting and says what it does.
// The caller adds the control that sets it, inside trailing.
func settingRow(c *ui.Context, title, note string, trailing func()) {
	t := c.Theme()
	ui.Row(c).Gap(16).Padding(14, 16).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, title).FontSize(sizeBody).FontWeight(600)
			ui.Text(c, note).FontSize(sizeNote).LineHeight(1.45).TextColor(t.TextMuted)
		})
		trailing()
	})
}
