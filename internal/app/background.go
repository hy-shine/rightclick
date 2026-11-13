package app

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/hy-shine/rightclick/internal/locales"
)

// RunAction runs the menu action that a rightclick:// link names, without opening a
// window. The Finder extension starts this program with --action, through the script
// that finder.InstallActionScript writes, so the app does not start for an action.
// It returns the exit status.
func RunAction(raw string) int {
	cfg, err := LoadConfig()
	if err != nil {
		return failInBackground(raw, err)
	}
	useBackgroundInterface()
	mygo.App.SetLocale(cfg.Language)
	action, params, paths, err := parseURL(raw)
	if err != nil {
		return failInBackground(raw, err)
	}
	if action == "add-template" || action == "hash" {
		// Both show a window of the app, so the link opens the app and the window
		// appears there. The background process cannot show one: it has no run loop.
		if err := exec.Command("/usr/bin/open", raw).Run(); err != nil {
			return failInBackground(raw, err)
		}
		return 0
	}
	if err := runAction(action, params, paths, cfg); err != nil {
		return failInBackground(raw, err)
	}
	return 0
}

// failInBackground logs the error and shows it in a notification, as the app does for
// a failed link.
func failInBackground(raw string, err error) int {
	log.Printf("rightclick %s: %v", raw, err)
	notifyFailure(err)
	return 1
}

// useBackgroundInterface replaces the calls that need the app's run loop. The
// background process never runs that loop, so mygo's clipboard, notification and
// dialog calls cannot finish there. AppleScript does the same jobs.
func useBackgroundInterface() {
	copyText = setClipboard
	readClipboardText = getClipboardText
	notify = notifyWithAppleScript
	chooseFolder = chooseFolderWithAppleScript
	chooseCopyBaseFolder = chooseCopyBaseWithAppleScript
	confirmDelete = confirmDeleteWithAppleScript
}

// pbpasteCmd builds the command to read the macOS clipboard. It is a variable
// so unit tests can substitute an isolated runner.
var pbpasteCmd = func() *exec.Cmd {
	cmd := exec.Command("/usr/bin/pbpaste")
	cmd.Env = append(os.Environ(), "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8")
	return cmd
}

// getClipboardText reads up to maxBytes text from the clipboard using pbpaste.
// It sets LANG and LC_ALL to en_US.UTF-8 to preserve exact UTF-8 bytes without
// appending any newlines, bounds reading with io.LimitReader, kills pbpaste on
// overflow, and returns locales.ErrClipboardTooLarge.
func getClipboardText(maxBytes int64) (string, error) {
	cmd := pbpasteCmd()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}

	lr := io.LimitReader(stdout, maxBytes+1)
	data, readErr := io.ReadAll(lr)
	if readErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", readErr
	}
	if int64(len(data)) > maxBytes {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", errors.New(locales.ErrClipboardTooLarge())
	}
	if err := cmd.Wait(); err != nil {
		return "", err
	}
	return string(data), nil
}

// setClipboard puts text on the clipboard. AppleScript keeps the text's encoding;
// pbcopy did not, and non-ASCII file names came out garbled.
func setClipboard(text string) error {
	_, err := runAppleScript(clipboardScript, text)
	return err
}

// notifyWithAppleScript shows a notification. macOS may name the sender after the
// script host, so the notice can read as Script Editor rather than RightClick.
func notifyWithAppleScript(title, body string) {
	if _, err := runAppleScript(notifyScript, title, body); err != nil {
		log.Printf("notification: %v", err)
	}
}

// chooseFolderWithAppleScript shows a folder picker and returns the chosen folder, or
// "" when the user cancels.
func chooseFolderWithAppleScript(_ string) (string, error) {
	out, err := runAppleScript(chooseFolderScript, locales.ChooseDestination())
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(out)
	if path == "" {
		return "", nil
	}
	return filepath.Clean(path), nil
}

func chooseCopyBaseWithAppleScript(start string) (string, error) {
	out, err := runAppleScript(chooseCopyBaseScript, locales.CopyBaseHint(), start)
	if err != nil {
		return "", err
	}
	path := strings.TrimSuffix(out, "\n")
	if path == "" {
		return "", nil
	}
	return filepath.Clean(path), nil
}

// confirmDeleteWithAppleScript asks before the items are deleted. Cancel is the
// default answer, and the function reports whether the user agreed.
func confirmDeleteWithAppleScript(question, detail string) (bool, error) {
	out, err := runAppleScript(confirmDeleteScript, question, detail, locales.DeletePermanently(), locales.Cancel(), appBundlePath())
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "delete", nil
}

// appBundlePath returns the .app bundle that holds the running program, or "" when
// the program runs outside a bundle. The delete alert shows the icon of that bundle.
func appBundlePath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	bundle := filepath.Clean(filepath.Join(filepath.Dir(exe), "..", ".."))
	if filepath.Ext(bundle) != ".app" {
		return ""
	}
	return bundle
}

// runAppleScript runs script with args as its argv. The text never enters the source
// of the script, so it needs no quoting.
func runAppleScript(script string, args ...string) (string, error) {
	cmd := exec.Command("/usr/bin/osascript", append([]string{"-e", script, "--"}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && len(exit.Stderr) > 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(exit.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

const clipboardScript = `on run argv
	set the clipboard to (item 1 of argv)
end run`

const notifyScript = `on run argv
	display notification (item 2 of argv) with title (item 1 of argv)
end run`

// A cancelled picker raises error -128, which becomes an empty answer.
const chooseFolderScript = `on run argv
	try
		set chosen to choose folder with prompt (item 1 of argv)
	on error number -128
		return ""
	end try
	return POSIX path of chosen
end run`

const chooseCopyBaseScript = `on run argv
	try
		if (item 2 of argv) is "" then
			set chosen to choose folder with prompt (item 1 of argv)
		else
			set startFolder to POSIX file (item 2 of argv) as alias
			set chosen to choose folder with prompt (item 1 of argv) default location startFolder
		end if
	on error number -128
		return ""
	end try
	return POSIX path of chosen
end run`

// The answer is "delete" only when the user presses the delete button. Cancel is the
// first button, so it is the default one and Return keeps the items.
//
// The alert is a real NSAlert, built through AppleScriptObjC: AppleScript's own
// dialogs cannot mark the delete button as destructive, and the red button is how
// macOS shows an action that cannot be undone. The app's own icon names the app
// that asks.
const confirmDeleteScript = `use framework "AppKit"
use scripting additions

on run argv
	set shared to current application's NSApplication's sharedApplication()
	shared's setActivationPolicy:1
	set alert to current application's NSAlert's alloc()'s init()
	alert's setMessageText:(item 1 of argv)
	alert's setInformativeText:(item 2 of argv)
	alert's setAlertStyle:2
	alert's addButtonWithTitle:(item 4 of argv)
	set destructiveButton to alert's addButtonWithTitle:(item 3 of argv)
	destructiveButton's setHasDestructiveAction:true
	if (item 5 of argv) is not "" then
		alert's setIcon:(current application's NSWorkspace's sharedWorkspace()'s iconForFile:(item 5 of argv))
	end if
	shared's activateIgnoringOtherApps:true
	set response to alert's runModal()
	if response is 1000 then return "cancel"
	return "delete"
end run`
