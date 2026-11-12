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
// this app lit, on a light blue square. It matches resources/icon.png.
var appMark = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 824 824">
<defs><linearGradient id="blue" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#a9cdfb"/><stop offset="1" stop-color="#5b93ee"/></linearGradient></defs>
<rect width="824" height="824" rx="190" fill="url(#blue)"/>
<g transform="translate(-100 -100)">
<rect x="262" y="232" width="500" height="560" rx="44" fill="#ffffff"/>
<rect x="326" y="309" width="200" height="22" rx="11" fill="#cbd5e1"/>
<rect x="326" y="405" width="268" height="22" rx="11" fill="#cbd5e1"/>
<rect x="290" y="470" width="444" height="84" rx="24" fill="#1d4ed8"/>
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

// itemRowHeight is the height of every row in a list of items: a program, a
// favorite folder. It holds two lines, so a row with one
// line sits at the same pitch as a row with two.
const itemRowHeight = 48

// itemGripSlot is the width of the drag handle. Only a list that can be
// reordered has one, so the other lists do not start with a blank strip.
const itemGripSlot = 10

// itemIconSize is the side of the icon in a list row. macOS shows app icons in lists
// at this size, so the artwork itself fills the slot without a tile around it.
const itemIconSize = 32

// glyphTile draws a row's icon when the row has no artwork of its own, such as a
// folder or an app with no icon: a glyph on a soft tile of the same size as an app icon.
func glyphTile(c *ui.Context, s *ui.SVG) {
	t := c.Theme()
	ui.Row(c).Size(itemIconSize, itemIconSize).Radius(8).Background(t.Background).Shrink(0).
		AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		glyph(c, s, 16, t.TextMuted)
	})
}

// itemRow is one row of a list of items. From the left: the drag handle, when grip
// is true, the icon slot (icon fills it), the name with an optional note, then the
// trailing controls. Every list of items builds its rows with it.
func itemRow(c *ui.Context, grip bool, icon func(), title, note string, trailing func()) {
	t := c.Theme()
	ui.Row(c).Height(itemRowHeight).Gap(12).Padding(0, 14).AlignItems(ui.Center).Children(func() {
		if grip {
			ui.Row(c).Size(itemGripSlot, 14).Shrink(0).AlignItems(ui.Center).Children(func() {
				glyph(c, iconGrip, 10, t.TextMuted)
			})
		}
		ui.Row(c).Size(itemIconSize, itemIconSize).Shrink(0).AlignItems(ui.Center).Justify(ui.Center).Children(icon)
		ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
			ui.Text(c, title).FontSize(sizeBody).SingleLine()
			if note != "" {
				ui.Text(c, note).FontSize(sizeNote).TextColor(t.TextMuted).SingleLine()
			}
		})
		trailing()
	})
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
	return actionButton(c, primaryKind, s, label, ui.Start)
}

// primaryButtonCentered is a primaryButton that centers itself in a row, for
// actions that sit beside a heading instead of at the start of a pane. Without
// it, a centered quiet button next to a start-aligned primary one leaves the
// pair on different baselines.
func primaryButtonCentered(c *ui.Context, s *ui.SVG, label string) bool {
	return actionButton(c, primaryKind, s, label, ui.Center)
}

// quietButton is a button with a border and no fill, for actions that are not
// the next step. It reports a click.
func quietButton(c *ui.Context, s *ui.SVG, label string) bool {
	return actionButton(c, quietKind, s, label, ui.Start)
}

// quietButtonCentered is a quietButton that centers itself in a row, for
// actions that sit beside a heading instead of at the start of a pane.
func quietButtonCentered(c *ui.Context, s *ui.SVG, label string) bool {
	return actionButton(c, quietKind, s, label, ui.Center)
}

// buttonKind picks the look of an action button. primaryKind is the filled one of
// a view; quietKind has a border and no fill.
type buttonKind int

const (
	primaryKind buttonKind = iota
	quietKind
)

// buttonLook holds what tells the two action buttons apart. The rest of their
// construction is shared, so both keep the same height, radius and icon size.
type buttonLook struct {
	fill, hover, pressed ui.Color
	text, icon           ui.Color
	padX                 float32
	weight               int
	bordered             bool
}

// actionButton draws the button behind primaryButton and quietButton. align says
// where it sits in its row. It reports a click.
func actionButton(c *ui.Context, kind buttonKind, s *ui.SVG, label string, align ui.Align) bool {
	t := c.Theme()
	look := buttonLook{fill: t.Surface, hover: t.SurfaceHover, pressed: t.SurfacePressed,
		text: t.Text, icon: t.TextMuted, padX: 12, bordered: true}
	if kind == primaryKind {
		look = buttonLook{fill: t.Accent, hover: t.AccentHover, pressed: t.AccentPressed,
			text: t.AccentText, icon: t.AccentText, padX: 14, weight: 600}
	}
	b := ui.ButtonBase(c).Gap(6).Padding(7, look.padX).Height(controlHeight).Radius(t.Radius).AlignSelf(align).Label(label)
	if look.bordered {
		b.Border(1, separatorColor(c))
	}
	b.Children(func() {
		if s != nil {
			glyph(c, s, 14, look.icon)
		}
		text := ui.Text(c, label).FontSize(sizeBody).TextColor(look.text)
		if look.weight != 0 {
			text.FontWeight(look.weight)
		}
	})
	tint(b, look.fill, look.hover, look.pressed)
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
	// The hit area is 32 by 26 points, the target the edit button had, so every
	// icon-only button is the same size.
	b := ui.ButtonBase(c).Size(32, 26).Radius(t.Radius).Label(label).Center()
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

// settingRow is one row of a ledger: a title and a note on the left, with the
// control that sets it on the right, inside trailing. icon, when not nil, sits
// before the text in iconColor. Setting rows and status rows share this layout.
func settingRow(c *ui.Context, icon *ui.SVG, iconColor ui.Color, title, note string, trailing func()) {
	t := c.Theme()
	ui.Row(c).Gap(14).Padding(14, 16).AlignItems(ui.Center).Children(func() {
		if icon != nil {
			glyph(c, icon, 17, iconColor)
		}
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, title).FontSize(sizeBody).FontWeight(600)
			if note != "" {
				ui.Text(c, note).FontSize(sizeNote).LineHeight(1.45).TextColor(t.TextMuted)
			}
		})
		trailing()
	})
}
