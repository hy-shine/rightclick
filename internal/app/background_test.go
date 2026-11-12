package app

import (
	"strings"
	"testing"
)

// TestAppleScriptReceivesArgumentsAsText checks that the text reaches a script
// unchanged, with a leading dash and non-ASCII characters. The clipboard, the
// notification and the dialogs of the background process all depend on it.
func TestAppleScriptReceivesArgumentsAsText(t *testing.T) {
	const join = `on run argv
	return (item 1 of argv) & "|" & (item 2 of argv)
end run`
	out, err := runAppleScript(join, "-abc", "中文 名.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(out), "-abc|中文 名.txt"; got != want {
		t.Errorf("script returned %q, want %q", got, want)
	}
}

// TestAppleScriptErrorCarriesTheScriptMessage checks that a failing script reports
// its own message, not only the exit status.
func TestAppleScriptErrorCarriesTheScriptMessage(t *testing.T) {
	_, err := runAppleScript(`error "the chosen name is wrong"`)
	if err == nil || !strings.Contains(err.Error(), "the chosen name is wrong") {
		t.Errorf("error = %v, want the script's own message", err)
	}
}
