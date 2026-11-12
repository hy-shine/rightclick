package finder

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestActionScriptDirIsTheExtensionFolder(t *testing.T) {
	want := "/Users/someone/Library/Application Scripts/com.example.rightclick.FinderSync"
	if got := ActionScriptDir("/Users/someone"); got != want {
		t.Errorf("ActionScriptDir = %q, want %q", got, want)
	}
}

// TestActionScriptPassesTheActionToTheExecutable runs the installed script with a
// stand-in executable whose path has a space and a quote, and checks that the
// action reaches the executable as the argument after --action.
func TestActionScriptPassesTheActionToTheExecutable(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "My App's", "RightClick")
	if err := os.MkdirAll(filepath.Dir(app), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(app, []byte("#!/bin/sh\necho \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	scripts := filepath.Join(dir, "scripts")
	if err := InstallActionScript(scripts, app); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(scripts, ActionScriptName)
	info, err := os.Stat(script)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("script mode = %o, want 755", info.Mode().Perm())
	}

	const link = "rightclick://copy-path?path=%2Ftmp%2Fa.txt"
	out, err := exec.Command(script, link).Output()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(out), "--action "+link+"\n"; got != want {
		t.Errorf("script ran %q, want %q", got, want)
	}
}

func TestInstallActionScriptReplacesTheOldOne(t *testing.T) {
	dir := t.TempDir()
	if err := InstallActionScript(dir, "/old/RightClick"); err != nil {
		t.Fatal(err)
	}
	if err := InstallActionScript(dir, "/new/RightClick"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ActionScriptName))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(strings.Split(string(data), "\n"), "exec '/new/RightClick' --action \"$1\"") {
		t.Errorf("script does not name the new app:\n%s", data)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("the folder holds %d files, want only the script", len(entries))
	}
}
