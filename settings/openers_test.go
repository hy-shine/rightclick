package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
)

func TestTerminalChoicesIncludeCommonAppsAndCurrent(t *testing.T) {
	choices := terminalChoices("CustomTerm")
	for _, want := range []string{
		"Terminal", "iTerm", "Warp", "Ghostty", "kitty",
		"Alacritty", "WezTerm", "Hyper", "Tabby", "CustomTerm",
	} {
		if !slices.Contains(choices, want) {
			t.Errorf("terminal choices do not contain %q: %q", want, choices)
		}
	}
}

func TestApplicationName(t *testing.T) {
	tests := []struct {
		path    string
		want    string
		wantErr bool
	}{
		{path: "/Applications/Visual Studio Code.app", want: "Visual Studio Code"},
		{path: "/Applications/Warp.app", want: "Warp"},
		{path: "/Applications/NotAnApp", wantErr: true},
	}
	for _, tt := range tests {
		got, err := applicationName(tt.path)
		if (err != nil) != tt.wantErr {
			t.Fatalf("applicationName(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("applicationName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestSetDefaultEditorMovesItToFront(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneOpeners)
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	if err := tt.Click("设为默认"); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Editors) < 2 || saved.Editors[0] != "Cursor" {
		t.Fatalf("editors = %q, want Cursor first", saved.Editors)
	}
}
