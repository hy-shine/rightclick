// Package actionicons holds the icons shown for the app's menu actions. The
// files come from the Lucide set (ISC, see THIRD_PARTY_LICENSES.txt) and are
// bundled into the Finder extension. The extension uses them for action glyphs;
// folder rows use the system folder artwork and the destructive action is red.
package actionicons

import (
	"embed"
	"sync"

	"github.com/egoist/mygo/ui"
)

//go:embed icons/*.svg
var files embed.FS

// Names maps a menu action ID, or a shortcut kind, to its icon file without the
// extension. It mirrors ActionIcons.names in
// extension/RightClickFinderSync/ActionIcons.swift; change both together.
var Names = map[string]string{
	"new-file":           "new-file",
	"new-folder":         "new-folder",
	"new-template":       "new-template",
	"add-template":       "add-template",
	"copy-path":          "copy-path",
	"copy-terminal-path": "open-terminal",
	"copy-relative-path": "copy-path",
	"copy-markdown-link": "copy-path",
	"copy-name":          "copy-name",
	"copy-basename":      "copy-basename",
	"copy-to":            "copy-to",
	"move-to":            "move-to",
	"open-terminal":      "open-terminal",
	"zip":                "zip",
	"unzip":              "unzip",
	"unzip-to-folder":    "unzip",
	"hash":               "hash",
	"compare-hash":       "compare-hash",
	"toggle-hidden":      "toggle-hidden",
	"delete-permanently": "delete-permanently",
	"favorite-folders":   "favorite-folders",
	// Shortcut kinds. "type" and "template" use the file type icon instead.
	"app":    "app",
	"folder": "folder",
	// The app's own menu entry: a pointer clicking, the right-click idea.
	"menu": "menu",
	// Not an action: the folder picker row uses the folder artwork.
	"choose-folder": "folder",
}

var (
	mu    sync.Mutex
	cache = map[string]*ui.SVG{}
)

// Icon returns the icon of an action, parsed once, or nil when the action has
// none. The SVGs ship with the binary, so a missing file is a build mistake.
func Icon(id string) *ui.SVG {
	name, ok := Names[id]
	if !ok {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	if parsed, ok := cache[id]; ok {
		return parsed
	}
	data, err := files.ReadFile("icons/" + name + ".svg")
	if err != nil {
		return nil
	}
	parsed := ui.MustParseSVG(data)
	cache[id] = parsed
	return parsed
}
