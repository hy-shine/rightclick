package finder

import (
	"os"
	"path/filepath"
	"strings"
)

// ExtensionID is the bundle identifier of the Finder Sync extension
// (extension/RightClickFinderSync/Info.plist).
const ExtensionID = "com.example.rightclick.FinderSync"

// ActionScriptName is the file the Finder extension runs for a menu action. The
// extension names the same file in FinderSyncExtension.swift.
const ActionScriptName = "rightclick-action"

// ActionScriptDir is the Application Scripts folder of the Finder extension. macOS
// runs the scripts in it outside the extension's sandbox. The sandboxed extension
// may read the folder but not write to it, so the app writes the script.
func ActionScriptDir(home string) string {
	return filepath.Join(home, "Library", "Application Scripts", ExtensionID)
}

// InstallActionScript writes the script that runs a menu action into dir. The
// script starts executable with --action, which runs the action without opening a
// window. The app calls this at every start, so a moved app is found again.
func InstallActionScript(dir, executable string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	script := "#!/bin/sh\n" +
		"# Written by RightClick at every start. The Finder extension runs this file.\n" +
		"exec " + shellQuote(executable) + " --action \"$1\"\n"
	path := filepath.Join(dir, ActionScriptName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(script), 0o755); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// shellQuote quotes s for a POSIX shell, in single quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
