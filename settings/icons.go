package settings

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// The icons share one grid: 24 units, a 1.8 stroke, round caps and joins, no
// fill. They are drawn in the color of the text around them, so a row's text
// color sets its icon color too.

// icon wraps the shapes in body in the shared grid and parses them once, at
// start-up.
func icon(body string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" ` +
		`fill="none" stroke="#000" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">` +
		body + `</svg>`))
}

var (
	// Panes.
	iconMenu    = icon(`<path d="M9.5 6h10.5M9.5 12h10.5M9.5 18h10.5"/><path d="M4.5 6h.01M4.5 12h.01M4.5 18h.01"/>`)
	iconNew     = icon(`<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/><path d="M12 12v5M9.5 14.5h5"/>`)
	iconDir     = icon(`<path d="M3.5 7.5a2 2 0 0 1 2-2h3.6l2 2.2h7.4a2 2 0 0 1 2 2v7.8a2 2 0 0 1-2 2H5.5a2 2 0 0 1-2-2z"/>`)
	iconOpen    = icon(`<path d="M13.5 4H20v6.5"/><path d="M20 4l-9 9"/><path d="M18 14.5V18a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3.5"/>`)
	iconShield  = icon(`<path d="M12 3l7 3v5.2c0 4.4-2.9 7.9-7 9.8-4.1-1.9-7-5.4-7-9.8V6z"/><path d="M9 12l2.2 2.2L15.2 10"/>`)
	iconGeneral = icon(`<path d="M4 7h8M17 7h3M4 17h3M12 17h8"/><circle cx="14.5" cy="7" r="2.3"/><circle cx="9.5" cy="17" r="2.3"/>`)
	iconInfo    = icon(`<circle cx="12" cy="12" r="9"/><path d="M12 11v5.5"/><path d="M12 7.8h.01"/>`)

	// Menu entries.
	iconText       = icon(`<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/><path d="M9 13h6M9 17h4"/>`)
	iconMarkdown   = icon(`<rect x="3" y="5.5" width="18" height="13" rx="2.2"/><path d="M7 15V9l2.5 3L12 9v6"/><path d="M16.5 9v6M14.5 13l2 2 2-2"/>`)
	iconJSON       = icon(`<path d="M9 4H7.5A2.5 2.5 0 0 0 5 6.5v3A2.5 2.5 0 0 1 2.5 12 2.5 2.5 0 0 1 5 14.5v3A2.5 2.5 0 0 0 7.5 20H9"/><path d="M15 4h1.5a2.5 2.5 0 0 1 2.5 2.5v3a2.5 2.5 0 0 0 2.5 2.5 2.5 2.5 0 0 0-2.5 2.5v3a2.5 2.5 0 0 1-2.5 2.5H15"/>`)
	iconCSV        = icon(`<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M3 9h18M3 15h18M9 4v16M15 4v16"/>`)
	iconHTML       = icon(`<path d="M8.5 7.5L4 12l4.5 4.5M15.5 7.5L20 12l-4.5 4.5"/><path d="M13.5 5.5l-3 13"/>`)
	iconCSS        = icon(`<path d="M14 4l6 6-8.5 8.5-4.5 1 1-4.5z"/><path d="M12.5 5.5l6 6M5 20c-.8-2.2.2-4.5 2.5-5.5"/>`)
	iconJavaScript = icon(`<rect x="3" y="4" width="18" height="16" rx="2.2"/><path d="M13 9v6a2 2 0 0 1-4 0M18 10.5c-.7-1-3.5-1.4-3.5.5 0 2.2 4 1 4 3.2 0 2-3 2-4 1"/>`)
	iconSwift      = icon(`<path d="M5 16c4 0 7-2.5 9-7 1 2.5 2.2 4.1 4 5-3.8-1.2-8.4.2-13 2z"/><path d="M15 5c1.8 2.5 2 5.2.2 7.8"/>`)
	iconPython     = icon(`<path d="M9 4h5a3 3 0 0 1 3 3v3H9a3 3 0 0 0-3 3v1h3"/><path d="M15 20h-5a3 3 0 0 1-3-3v-3h8a3 3 0 0 0 3-3v-1h-3"/><path d="M10 7h.01M14 17h.01"/>`)
	iconDatabase   = icon(`<path d="M4.5 5.5a7.5 2.5 0 1 0 15 0a7.5 2.5 0 1 0-15 0"/><path d="M4.5 5.5v13a7.5 2.5 0 0 0 15 0v-13"/><path d="M4.5 12a7.5 2.5 0 0 0 15 0"/>`)
	iconFolderPlus = icon(`<path d="M3.5 7.5a2 2 0 0 1 2-2h3.6l2 2.2h7.4a2 2 0 0 1 2 2v7.8a2 2 0 0 1-2 2H5.5a2 2 0 0 1-2-2z"/><path d="M12 11.5v4.5M9.8 13.8h4.4"/>`)
	iconTemplate   = icon(`<rect x="4" y="4" width="16" height="16" rx="2.2"/><path d="M4 9.5h16M9.5 9.5V20"/>`)
	iconCopy       = icon(`<rect x="8.5" y="8.5" width="11.5" height="11.5" rx="2.2"/><path d="M15.5 8.5V6.5a2 2 0 0 0-2-2h-7a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h2"/>`)
	iconName       = icon(`<path d="M5 8V6h14v2"/><path d="M12 6v12M9.5 18h5"/>`)
	iconCopyTo     = icon(`<rect x="3.5" y="3.5" width="10" height="10" rx="2"/><path d="M10.5 17.5h6a2 2 0 0 0 2-2v-6"/><path d="M15.5 20.5l3-3-3-3"/>`)
	iconMoveTo     = icon(`<path d="M4 12h13M13 7l5 5-5 5"/><path d="M21 5v14"/>`)
	iconTerminal   = icon(`<rect x="3" y="4.5" width="18" height="15" rx="2.2"/><path d="M7.5 9.5l3 2.5-3 2.5"/><path d="M13 15h3.5"/>`)
	iconCode       = icon(`<path d="M8.5 7.5L4 12l4.5 4.5M15.5 7.5L20 12l-4.5 4.5"/><path d="M13.5 5.5l-3 13"/>`)
	iconZip        = icon(`<path d="M4.5 8.5h15V18a2 2 0 0 1-2 2h-11a2 2 0 0 1-2-2z"/><path d="M3.5 4.5h17v4h-17z"/><path d="M10 12.5h4"/>`)
	iconUnzip      = icon(`<path d="M4.5 8.5h15V18a2 2 0 0 1-2 2h-11a2 2 0 0 1-2-2z"/><path d="M3.5 4.5h17v4h-17z"/><path d="M12 17v-5M9.8 14l2.2-2.2 2.2 2.2"/>`)
	iconHash       = icon(`<path d="M5 9h14M5 15h14M10.3 4L8.7 20M15.3 4l-1.6 16"/>`)
	iconCompare    = icon(`<path d="M4 8.5h13M13.5 5l3.5 3.5-3.5 3.5"/><path d="M20 15.5H7M10.5 12l-3.5 3.5 3.5 3.5"/>`)
	iconEye        = icon(`<path d="M2.5 12s3.5-6.5 9.5-6.5S21.5 12 21.5 12s-3.5 6.5-9.5 6.5S2.5 12 2.5 12z"/><circle cx="12" cy="12" r="2.8"/>`)
	iconTrash      = icon(`<path d="M4 7h16"/><path d="M9.5 7V4.8h5V7"/><path d="M6.5 7l1 12.2h9l1-12.2"/><path d="M10 11v5M14 11v5"/>`)

	// Marks.
	iconChevron     = icon(`<path d="M9.5 6l6 6-6 6"/>`)
	iconChevronDown = icon(`<path d="M6 9.5l6 6 6-6"/>`)
	iconGrip        = icon(`<path d="M9 6h.01M15 6h.01M9 12h.01M15 12h.01M9 18h.01M15 18h.01" stroke-width="2.6"/>`)
	iconPlus        = icon(`<path d="M12 5v14M5 12h14"/>`)
	iconEdit        = icon(`<path d="M4 20h4l10.5-10.5-4-4L4 16z"/><path d="M12.5 7.5l4 4"/>`)
	iconCheck       = icon(`<path d="M5 12.5l4.5 4.5L19 7.5"/>`)
	iconAlert       = icon(`<path d="M12 4l9 15.5H3z"/><path d="M12 10v4.2"/><path d="M12 17h.01"/>`)
	iconRestart     = icon(`<path d="M4.5 12a7.5 7.5 0 1 0 2.4-5.5"/><path d="M4.5 4.5v4h4"/>`)
)

// entryIcons maps the ID of a menu entry to its icon.
var entryIcons = map[string]*ui.SVG{
	"new-file":           iconNew,
	"new-folder":         iconFolderPlus,
	"new-template":       iconTemplate,
	"copy-path":          iconCopy,
	"copy-name":          iconName,
	"copy-to":            iconCopyTo,
	"move-to":            iconMoveTo,
	"open-terminal":      iconTerminal,
	"open-editor":        iconCode,
	"open-with":          iconOpen,
	"zip":                iconZip,
	"unzip":              iconUnzip,
	"hash":               iconHash,
	"compare-hash":       iconCompare,
	"toggle-hidden":      iconEye,
	"delete-permanently": iconTrash,
}

// entryIcon returns the icon of the menu entry id, or the plain document
// icon for an entry it does not know.
func entryIcon(id string) *ui.SVG {
	if target, ok := strings.CutPrefix(id, "type:"); ok {
		return fileTypeIcon(target)
	}
	if id == "favorite-folders" || strings.HasPrefix(id, "folder:") {
		return iconDir
	}
	if strings.HasPrefix(id, "app:") {
		return iconOpen
	}
	if fileTypeID, ok := strings.CutPrefix(id, "file-type:"); ok {
		return fileTypeIcon(fileTypeID)
	}
	if s, ok := entryIcons[id]; ok {
		return s
	}
	return iconText
}

// fileTypeIcons map built-in file types to document or code glyphs.
var fileTypeIcons = map[string]*ui.SVG{
	"text":       iconText,
	"markdown":   iconMarkdown,
	"json":       iconJSON,
	"csv":        iconCSV,
	"html":       iconHTML,
	"css":        iconCSS,
	"javascript": iconJavaScript,
	"swift":      iconSwift,
	"python":     iconPython,
	"go":         iconCode,
	"rust":       iconCode,
	"java":       iconCode,
	"c":          iconCode,
	"cpp":        iconCode,
	"sql":        iconDatabase,
	"typescript": iconCode,
	"shell":      iconCode,
}

// fileTypeIcon returns the icon of a built-in type.
func fileTypeIcon(id string) *ui.SVG {
	if s, ok := fileTypeIcons[id]; ok {
		return s
	}
	return iconText
}

// fileTypeColors keep file types recognizable in the settings view and menu
// preview without turning the rest of the interface into a color wheel.
var fileTypeColors = map[string][2]ui.Color{
	"text":       {ui.Hex("#687386"), ui.Hex("#aab2bf")},
	"markdown":   {ui.Hex("#4f6fdd"), ui.Hex("#9fb1ff")},
	"json":       {ui.Hex("#b56b18"), ui.Hex("#e4ae69")},
	"csv":        {ui.Hex("#27834f"), ui.Hex("#7cc79a")},
	"html":       {ui.Hex("#c45133"), ui.Hex("#ee9a82")},
	"css":        {ui.Hex("#3379b7"), ui.Hex("#83b6e3")},
	"javascript": {ui.Hex("#a97900"), ui.Hex("#e1bd58")},
	"swift":      {ui.Hex("#d25f22"), ui.Hex("#f5a176")},
	"python":     {ui.Hex("#287b9b"), ui.Hex("#83c0d7")},
	"go":         {ui.Hex("#007d9c"), ui.Hex("#5fd0ea")},
	"rust":       {ui.Hex("#b7410e"), ui.Hex("#e9865a")},
	"java":       {ui.Hex("#b07219"), ui.Hex("#e5a85a")},
	"c":          {ui.Hex("#4a5a8c"), ui.Hex("#9aa8d8")},
	"cpp":        {ui.Hex("#00529b"), ui.Hex("#6fa3e0")},
	"sql":        {ui.Hex("#336791"), ui.Hex("#7fa9d4")},
	"typescript": {ui.Hex("#2b6cb0"), ui.Hex("#7fb3f0")},
	"shell":      {ui.Hex("#3f7d3a"), ui.Hex("#8fcb7f")},
}

// fileTypeColor returns the color used for a built-in type's icon.
func fileTypeColor(id string, fallback ui.Color) ui.Color {
	if colors, ok := fileTypeColors[id]; ok {
		return colors[0]
	}
	return fallback
}

// fileTypeDarkColor returns a lighter version for dark mode.
func fileTypeDarkColor(id string, fallback ui.Color) ui.Color {
	if colors, ok := fileTypeColors[id]; ok {
		return colors[1]
	}
	return fallback
}

// themedFileTypeColor chooses the light or dark variant of a file type's icon.
func themedFileTypeColor(c *ui.Context, id string, fallback ui.Color) ui.Color {
	if c.Theme().Dark {
		return fileTypeDarkColor(id, fallback)
	}
	return fileTypeColor(id, fallback)
}
