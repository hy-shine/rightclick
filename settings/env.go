package settings

import (
	"os/exec"
	"strings"

	"github.com/egoist/mygo"

	"rightclick/finder"
)

// finderExtensionID is the bundle identifier of the Finder Sync extension
// (extension/RightClickFinderSync/Info.plist).
const finderExtensionID = "com.example.rightclick.FinderSync"

// finderSettingsURL opens the System Settings pane for extensions. Not
// verified on this machine.
const finderSettingsURL = "x-apple.systempreferences:com.apple.ExtensionsPreferences?FinderSync"

// env holds the calls to the system. The window uses the real ones; tests
// replace them.
type env struct {
	openAtLogin        func() bool
	setOpenAtLogin     func(on bool) error
	finderEnabled      func() bool // may block; called off the main thread
	openFinderSettings func() error
	restartFinder      func() error                                        // may block; called off the main thread
	openDialog         func(opts mygo.OpenDialogOptions) ([]string, error) // blocks until dismissed
	setLocale          func(tag string)                                    // "" follows the system
}

// defaultEnv returns the env that calls mygo and the macOS tools.
func defaultEnv() env {
	return env{
		openAtLogin:    mygo.App.OpenAtLogin,
		setOpenAtLogin: mygo.App.SetOpenAtLogin,
		finderEnabled:  finderExtensionEnabled,
		openFinderSettings: func() error {
			return mygo.Shell.OpenExternal(finderSettingsURL)
		},
		restartFinder: finder.Restart,
		openDialog:    mygo.Dialog.Open,
		setLocale:     mygo.App.SetLocale,
	}
}

// finderExtensionEnabled reports whether macOS has the Finder extension on.
// pluginkit prints a line that starts with "+" for an enabled plug-in, and
// nothing when it does not find the plug-in.
func finderExtensionEnabled() bool {
	out, err := exec.Command("pluginkit", "-m", "-p", "com.apple.FinderSync", "-i", finderExtensionID).Output()
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimLeft(string(out), " \t\r\n"), "+")
}
