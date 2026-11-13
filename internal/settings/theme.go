package settings

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func SetAppearance(mode string) {
	source := mygo.ThemeSystem
	switch mode {
	case "light":
		source = mygo.ThemeLight
	case "dark":
		source = mygo.ThemeDark
	}
	mygo.Theme.SetSource(source)
}

// The look uses neutral gray surfaces and system accents in the system's own font.
// Primary actions use solid blue; selected navigation and table rows use a
// softer blue tint. Shared interface colors are defined here.

// paperTheme returns the palette for the light or the dark appearance. The
// window calls it on each frame, with the appearance the system reports.
func paperTheme(dark bool) *ui.Theme {
	accent, hover, pressed, text := defaultBrandColors(dark)
	if dark {
		return &ui.Theme{
			Dark:           true,
			Background:     ui.Hex("#222326"),
			Surface:        ui.Hex("#2c2d30"),
			SurfaceHover:   ui.Hex("#37383c"),
			SurfacePressed: ui.Hex("#404145"),
			Border:         ui.Hex("#48494f"),
			Text:           ui.Hex("#f5f5f7"),
			TextMuted:      ui.Hex("#a6a7ad"),
			Accent:         accent,
			AccentHover:    hover,
			AccentPressed:  pressed,
			AccentText:     text,
			Danger:         ui.Hex("#ff6b5e"),
			Warning:        ui.Hex("#f0a36b"),
			Success:        ui.Hex("#5cd69a"),
			Selection:      ui.RGBA(96, 165, 250, 0.32),
			Focus:          ui.RGBA(96, 165, 250, 0.65),
			Inverse:        ui.Hex("#f5f5f7"),
			InverseText:    ui.Hex("#222326"),
			Scrollbar:      ui.RGBA(166, 167, 173, 0.4),
			Radius:         6,
			Spacing:        4,
			FontSize:       13,
		}
	}
	return &ui.Theme{
		Background:     ui.Hex("#ededf0"),
		Surface:        ui.Hex("#ffffff"),
		SurfaceHover:   ui.Hex("#f5f5f7"),
		SurfacePressed: ui.Hex("#e5e5e9"),
		Border:         ui.Hex("#d6d6dc"),
		Text:           ui.Hex("#252528"),
		TextMuted:      ui.Hex("#68686f"),
		Accent:         accent,
		AccentHover:    hover,
		AccentPressed:  pressed,
		AccentText:     text,
		Danger:         ui.Hex("#b42318"),
		Warning:        ui.Hex("#9a4308"),
		Success:        ui.Hex("#1e7b4f"),
		Selection:      ui.RGBA(37, 99, 235, 0.2),
		Focus:          ui.RGBA(37, 99, 235, 0.6),
		Inverse:        ui.Hex("#252528"),
		InverseText:    ui.Hex("#ffffff"),
		Scrollbar:      ui.RGBA(104, 104, 111, 0.4),
		Radius:         6,
		Spacing:        4,
		FontSize:       13,
	}
}

// defaultBrandColors supplies the accent when the desktop has no preference.
func defaultBrandColors(dark bool) (fill, hover, pressed, text ui.Color) {
	if dark {
		return ui.Hex("#60a5fa"), ui.Hex("#93c5fd"), ui.Hex("#3b82f6"), ui.Hex("#0f172a")
	}
	return ui.Hex("#2563eb"), ui.Hex("#1d4ed8"), ui.Hex("#1e40af"), ui.Hex("#ffffff")
}

// selectionColors returns a subtle accent fill and readable selected text.
func selectionColors(t *ui.Theme) (fill, text ui.Color) {
	return t.Surface.Mix(t.Accent, 0.12), t.Text
}

func windowTheme(context *ui.Context) *ui.Theme {
	system := context.Theme()
	theme := paperTheme(system.Dark)
	preferences := context.Preferences()
	if preferences.Accent.A > 0 {
		theme.Accent = system.Accent
		theme.AccentHover = system.AccentHover
		theme.AccentPressed = system.AccentPressed
		theme.AccentText = system.AccentText
		theme.Selection = system.Selection
		theme.Focus = system.Focus
	}
	if preferences.HighContrast {
		theme.Border = theme.Border.Mix(theme.Text, 0.45)
		theme.TextMuted = theme.TextMuted.Mix(theme.Text, 0.4)
		theme.Focus.A = 255
	}
	return theme
}

func separatorColor(context *ui.Context) ui.Color {
	border := context.Theme().Border
	if context.Preferences().HighContrast {
		return border
	}
	return border.Alpha(0.65)
}
