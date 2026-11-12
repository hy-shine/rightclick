package settings

import (
	"os/exec"
	"strings"

	"github.com/egoist/mygo"

	"github.com/hy-shine/rightclick/internal/finder"
	"github.com/hy-shine/rightclick/internal/update"
)

// finderExtensionID is the bundle identifier of the Finder Sync extension
// (extension/RightClickFinderSync/Info.plist).
const finderExtensionID = finder.ExtensionID

// finderSettingsURL opens the System Settings pane for extensions. Not
// verified on this machine.
const finderSettingsURL = "x-apple.systempreferences:com.apple.ExtensionsPreferences?FinderSync"

// env holds the calls to the system. The window uses the real ones; tests
// replace them.
type env struct {
	openAtLogin        func() bool
	setOpenAtLogin     func(on bool) error
	finderEnabled      func() bool  // may block; called off the main thread
	enableExtension    func() error // registers and turns on the extension; may block; called off the main thread
	openFinderSettings func() error
	restartFinder      func() error                                        // may block; called off the main thread
	openDialog         func(opts mygo.OpenDialogOptions) ([]string, error) // blocks until dismissed
	setLocale          func(tag string)                                    // "" follows the system
	setWindowTitle     func(title string)                                  // renames the main window when it is open
	latestRelease      func() (update.Release, error)                      // reads GitHub; called off the main thread
	openURL            func(url string) error
}

// defaultEnv returns the env that calls mygo and the macOS tools.
func defaultEnv() env {
	return env{
		openAtLogin:     mygo.App.OpenAtLogin,
		setOpenAtLogin:  mygo.App.SetOpenAtLogin,
		finderEnabled:   finderExtensionEnabled,
		enableExtension: finder.Enable,
		openFinderSettings: func() error {
			return mygo.Shell.OpenExternal(finderSettingsURL)
		},
		restartFinder:  finder.Restart,
		openDialog:     mygo.Dialog.Open,
		setLocale:      mygo.App.SetLocale,
		setWindowTitle: setMainWindowTitle,
		latestRelease:  update.Latest,
		openURL:        mygo.Shell.OpenExternal,
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
	return pluginkitEnabled(string(out))
}

// pluginkitEnabled reads the listing pluginkit prints: the first non-space character of
// a line that starts with "+" means the plug-in is on.
func pluginkitEnabled(listing string) bool {
	return strings.HasPrefix(strings.TrimLeft(listing, " \t\r\n"), "+")
}
