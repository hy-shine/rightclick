package settings

import (
	"os"
	"path/filepath"
	"testing"

	"rightclick/config"
	"rightclick/files"
)

// makeReadOnly stops the files of store from being written. The folder is opened
// again when the test ends, so that the temporary folder can be removed.
func makeReadOnly(t *testing.T, store Store) {
	t.Helper()
	dir := filepath.Dir(store.ConfigPath)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(dir, 0o755); err != nil {
			t.Error(err)
		}
	})
}

// TestFailedConfigSaveKeepsTheSavedValues checks that a change which cannot be
// written is undone in the window, so that the window matches the file.
func TestFailedConfigSaveKeepsTheSavedValues(t *testing.T) {
	store := testStore(t)
	cfg := config.Default()
	cfg.Terminal = "Ghostty"
	if err := config.Save(store.ConfigPath, cfg); err != nil {
		t.Fatal(err)
	}
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	makeReadOnly(t, store)

	m.cfg.Terminal = "iTerm"
	m.saveConfig()

	if m.cfg.Terminal != "Ghostty" {
		t.Errorf("window shows terminal %q, want the saved %q", m.cfg.Terminal, "Ghostty")
	}
	if m.notice == "" {
		t.Error("the failed save is not reported below the pane")
	}
	saved, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Terminal != "Ghostty" {
		t.Errorf("file holds terminal %q, want %q", saved.Terminal, "Ghostty")
	}
}

// TestFailedTemplateSaveKeepsTheSavedTemplates checks that an added template that
// cannot be written is removed again, and that the selection stays in range.
func TestFailedTemplateSaveKeepsTheSavedTemplates(t *testing.T) {
	store := testStore(t)
	saved := []files.Template{{Name: "meeting.md", Content: "# 会议\n", Location: config.FileTypeSubmenu}}
	if err := files.SaveTemplates(store.TemplatesPath, saved); err != nil {
		t.Fatal(err)
	}
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	makeReadOnly(t, store)

	m.templates = append(m.templates, files.Template{Name: "draft.md", Content: "", Location: config.FileTypeSubmenu})
	m.tplSel = len(m.templates) - 1
	m.saveTemplates()

	if len(m.templates) != 1 || m.templates[0].Name != "meeting.md" {
		t.Errorf("window shows %d templates, want the saved one", len(m.templates))
	}
	if m.tplSel != 0 {
		t.Errorf("selection = %d, want 0 after the added template is removed", m.tplSel)
	}
	if m.notice == "" {
		t.Error("the failed save is not reported below the pane")
	}
}
