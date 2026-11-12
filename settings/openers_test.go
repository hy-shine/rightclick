package settings

import (
	"slices"
	"testing"
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
