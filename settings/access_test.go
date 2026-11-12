package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/files"
)

func TestFoldersPaneShowsAccessState(t *testing.T) {
	granted := true
	opened := ""
	e := testEnv()
	e.folderAccess = func(string) files.Access {
		if granted {
			return files.AccessOK
		}
		return files.AccessDenied
	}
	e.openPrivacy = func(anchor string) error {
		opened = anchor
		return nil
	}
	m, err := newApp(testStore(t), e)
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneFolders)
	tt := ui.NewTester(m.view, mainWidth, 900)
	if !slices.Contains(tt.Texts(), "Authorized") {
		t.Fatalf("an authorized folder is not marked: %q", tt.Texts())
	}
	if slices.Contains(tt.Texts(), "Needs Access") {
		t.Error("an authorized folder is marked as needing access")
	}

	granted = false
	if err := tt.Click("Check Again"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !slices.Contains(tt.Texts(), "Needs Access") {
		t.Fatalf("a refused folder is not marked: %q", tt.Texts())
	}
	if !slices.Contains(tt.Texts(), "Some folders are not authorized") {
		t.Error("the pane does not offer the way to grant access")
	}
	if err := tt.Click("Grant Access"); err != nil {
		t.Fatal(err)
	}
	if opened != privacyFilesAndFolders {
		t.Errorf("opened %q, want %q", opened, privacyFilesAndFolders)
	}
}

// TestFoldersPaneSeparatesMissingFromRefused keeps a folder that is gone from
// asking for a permission it does not need.
func TestFoldersPaneSeparatesMissingFromRefused(t *testing.T) {
	e := testEnv()
	e.folderAccess = func(string) files.Access { return files.AccessMissing }
	m, err := newApp(testStore(t), e)
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneFolders)
	screen := ui.NewTester(m.view, mainWidth, 900)
	if !slices.Contains(screen.Texts(), "Path Unavailable") {
		t.Fatalf("a missing folder is not marked: %q", screen.Texts())
	}
	if slices.Contains(screen.Texts(), "Some folders are not authorized") {
		t.Error("a missing folder asks for access it cannot grant")
	}
}

func TestDiagnosticsShowsFullDiskAccess(t *testing.T) {
	tests := []struct {
		name    string
		granted bool
		note    string
		action  bool
	}{
		{
			name:    "missing",
			granted: false,
			note:    "Some actions need access to protected locations.",
			action:  true,
		},
		{
			name:    "granted",
			granted: true,
			note:    "RightClick can read protected locations.",
			action:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opened := ""
			e := testEnv()
			e.fullDiskAccess = func() bool { return tt.granted }
			e.openPrivacy = func(anchor string) error {
				opened = anchor
				return nil
			}
			m, err := newApp(testStore(t), e)
			if err != nil {
				t.Fatal(err)
			}
			m.selectPane(paneDiagnostics)
			screen := ui.NewTester(m.view, mainWidth, 1200)
			if !slices.Contains(screen.Texts(), "Full Disk Access") {
				t.Fatalf("the pane has no Full Disk Access step: %q", screen.Texts())
			}
			if !slices.Contains(screen.Texts(), tt.note) {
				t.Fatalf("note %q is missing: %q", tt.note, screen.Texts())
			}
			if has := slices.Contains(screen.Texts(), "Grant Access"); has != tt.action {
				t.Fatalf("action shown = %v, want %v", has, tt.action)
			}
			if !tt.action {
				return
			}
			if err := screen.Click("Grant Access"); err != nil {
				t.Fatal(err)
			}
			if opened != privacyAllFiles {
				t.Errorf("opened %q, want %q", opened, privacyAllFiles)
			}
		})
	}
}
