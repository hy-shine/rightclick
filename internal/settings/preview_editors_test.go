package settings

import (
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// submenuHas reports whether a submenu group of the preview shows label.
func submenuHas(m *app, ctx int, label string) bool {
	for _, group := range m.previewGroups(ctx) {
		for _, item := range group {
			if item.label == label {
				return true
			}
		}
	}
	return false
}

// TestPreviewShowsTheDefaultOpenEntry checks that the default open entry follows
// its location setting; it no longer depends on a configured program.
func TestPreviewShowsTheDefaultOpenEntry(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	label := locales.MenuOpenEditor()
	m.cfg.Programs = []string{}
	setEditorLocation := func(location string) {
		for i := range m.cfg.MenuItems {
			if m.cfg.MenuItems[i].ID == "open-editor" {
				m.cfg.MenuItems[i].Location = location
			}
		}
	}
	setEditorLocation(config.FileTypeMain)
	if !previewHas(m, ctxBlank, label) {
		t.Error("top level: the default open entry is missing")
	}
	setEditorLocation(config.FileTypeSubmenu)
	if !submenuHas(m, ctxBlank, label) {
		t.Error("submenu: the default open entry is missing")
	}
}
