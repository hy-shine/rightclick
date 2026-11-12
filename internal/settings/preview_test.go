package settings

import (
	"slices"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
)

// previewHas reports whether the preview shows label in place ctx.
func previewHas(m *app, ctx int, label string) bool {
	for _, item := range m.previewTopItems(ctx) {
		if item.label == label {
			return true
		}
	}
	for _, g := range m.previewGroups(ctx) {
		if slices.ContainsFunc(g, func(it previewItem) bool { return it.label == label }) {
			return true
		}
	}
	return false
}

func TestPreviewFollowsContext(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	// Open with Terminal starts switched off, so this test turns it on.
	for i := range m.cfg.MenuItems {
		if m.cfg.MenuItems[i].ID == "open-terminal" {
			m.cfg.MenuItems[i].Enabled = true
		}
	}
	cases := []struct {
		ctx   int
		label string
		want  bool
	}{
		{ctxBlank, "New File", true},
		{ctxBlank, "Copy Path", true},
		{ctxFolder, "Open with Terminal", true},
		{ctxFolder, "Compress to ZIP", false}, // zip needs files, not only a folder
		{ctxFile, "Compute Hash", true},
		{ctxFile, "Compare File Hashes", false},
		{ctxFile, "Extract Here", true},
		{ctxTwoFiles, "Compare File Hashes", true},
		{ctxTwoFiles, "Compute Hash", false},
	}
	for _, tc := range cases {
		if got := previewHas(m, tc.ctx, tc.label); got != tc.want {
			t.Errorf("place %d, %q shown = %v, want %v", tc.ctx, tc.label, got, tc.want)
		}
	}
}

func TestPreviewHidesDisabledEntries(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	if !previewHas(m, ctxBlank, "New File") {
		t.Fatal("New File should show on blank space before it is switched off")
	}
	m.cfg.MenuItems[0].Enabled = false // new-file is the first entry of config.Default
	if previewHas(m, ctxBlank, "New File") {
		t.Error("New File still shows after it is switched off")
	}
}

func TestPreviewKeepsFileTypesInsideNewFile(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	for i := range m.cfg.FileTypes {
		if m.cfg.FileTypes[i].ID == "json" {
			m.cfg.FileTypes[i].Location = config.FileTypeMain
		}
	}
	if previewHas(m, ctxBlank, "New JSON") {
		t.Error("JSON leaked to the top level of the preview")
	}
	if !previewHas(m, ctxBlank, "New File") {
		t.Error("New File is missing from the preview")
	}
	for i := range m.cfg.FileTypes {
		m.cfg.FileTypes[i].Enabled = false
	}
	if previewHas(m, ctxBlank, "New File") {
		t.Error("New File still shows when every file type is hidden")
	}
}
