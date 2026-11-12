package settings

import (
	"testing"

	"rightclick/config"
	"rightclick/locales"
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

// TestPreviewLeavesOutTheEditorWhenNoEditorIsListed checks that the preview follows
// the Finder menu: the editor entry is shown only when an editor is listed. Both
// the top level and the submenu are checked.
func TestPreviewLeavesOutTheEditorWhenNoEditorIsListed(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	label := locales.MenuOpenEditor()

	setEditorLocation := func(location string) {
		for i := range m.cfg.MenuItems {
			if m.cfg.MenuItems[i].ID == "open-editor" {
				m.cfg.MenuItems[i].Location = location
			}
		}
	}

	setEditorLocation(config.FileTypeMain)
	m.cfg.Editors = []string{"Cursor"}
	if !previewHas(m, ctxBlank, label) {
		t.Error("top level: the editor entry is missing although an editor is listed")
	}
	m.cfg.Editors = []string{}
	if previewHas(m, ctxBlank, label) {
		t.Error("top level: the editor entry is shown although no editor is listed")
	}

	setEditorLocation(config.FileTypeSubmenu)
	m.cfg.Editors = []string{"Cursor"}
	if !submenuHas(m, ctxBlank, label) {
		t.Error("submenu: the editor entry is missing although an editor is listed")
	}
	m.cfg.Editors = []string{}
	if submenuHas(m, ctxBlank, label) {
		t.Error("submenu: the editor entry is shown although no editor is listed")
	}
}
