package settings

import "github.com/egoist/mygo/ui"

// The look is a flat work area in the system's own font. A cool gray desk holds
// one white panel. Ink marks the current choice and the switches, and highlighter
// yellow is kept for the primary button and for the Finder menu's own highlight,
// so it stays the one brand color. Only this file names colors.

// paperTheme returns the palette for the light or the dark appearance. The
// window calls it on each frame, with the appearance the system reports.
func paperTheme(dark bool) *ui.Theme {
	if dark {
		return &ui.Theme{
			Dark:           true,
			Background:     ui.Hex("#121417"),
			Surface:        ui.Hex("#1b1e23"),
			SurfaceHover:   ui.Hex("#242830"),
			SurfacePressed: ui.Hex("#2c313a"),
			Border:         ui.Hex("#353b45"),
			Text:           ui.Hex("#eceef2"),
			TextMuted:      ui.Hex("#9aa2ae"),
			Accent:         ui.Hex("#8e97a6"),
			AccentHover:    ui.Hex("#a3abb8"),
			AccentPressed:  ui.Hex("#7a8390"),
			AccentText:     ui.Hex("#121417"),
			Danger:         ui.Hex("#ff6b5e"),
			Warning:        ui.Hex("#f0a36b"),
			Success:        ui.Hex("#5cd69a"),
			Selection:      ui.RGBA(242, 194, 48, 0.35),
			Focus:          ui.RGBA(236, 238, 242, 0.6),
			Inverse:        ui.Hex("#eceef2"),
			InverseText:    ui.Hex("#15171b"),
			Scrollbar:      ui.RGBA(236, 238, 242, 0.28),
			Radius:         6,
			Spacing:        4,
			FontSize:       13,
		}
	}
	return &ui.Theme{
		Background:     ui.Hex("#e4e7ec"),
		Surface:        ui.Hex("#fbfbfa"),
		SurfaceHover:   ui.Hex("#f1f2f4"),
		SurfacePressed: ui.Hex("#e6e8ec"),
		Border:         ui.Hex("#cdd2da"),
		Text:           ui.Hex("#1c1f24"),
		TextMuted:      ui.Hex("#5a616d"),
		Accent:         ui.Hex("#2b2f36"),
		AccentHover:    ui.Hex("#3a3f48"),
		AccentPressed:  ui.Hex("#1c1f24"),
		AccentText:     ui.Hex("#fbfbfa"),
		Danger:         ui.Hex("#b42318"),
		Warning:        ui.Hex("#9a4308"),
		Success:        ui.Hex("#1e7b4f"),
		Selection:      ui.RGBA(255, 210, 63, 0.45),
		Focus:          ui.RGBA(28, 31, 36, 0.6),
		Inverse:        ui.Hex("#1c1f24"),
		InverseText:    ui.Hex("#fbfbfa"),
		Scrollbar:      ui.RGBA(28, 31, 36, 0.28),
		Radius:         6,
		Spacing:        4,
		FontSize:       13,
	}
}

// brandColors returns the highlighter yellow of the brand: the fill, the hover
// and pressed fills, and the text on them. Only the primary button and the
// Finder menu's highlighted row use it.
func brandColors(t *ui.Theme) (fill, hover, pressed, text ui.Color) {
	if t.Dark {
		return ui.Hex("#f2c230"), ui.Hex("#ffd23f"), ui.Hex("#d9a91c"), ui.Hex("#15171b")
	}
	return ui.Hex("#ffd23f"), ui.Hex("#ffc61a"), ui.Hex("#f0b400"), ui.Hex("#1c1f24")
}
