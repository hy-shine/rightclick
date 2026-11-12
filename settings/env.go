package settings

import (
	"os/exec"
	"strings"

	"github.com/egoist/mygo"

	"github.com/hy-shine/rightclick/files"
	"github.com/hy-shine/rightclick/finder"
)

// finderExtensionID is the bundle identifier of the Finder Sync extension
// (extension/RightClickFinderSync/Info.plist).
const finderExtensionID = "com.example.rightclick.FinderSync"

// finderSettingsURL opens the System Settings pane for extensions. Not
// verified on this machine.
const finderSettingsURL = "x-apple.systempreferences:com.apple.ExtensionsPreferences?FinderSync"

// Anchors of the Privacy panes in System Settings. Not verified on this machine.
const (
	privacyFilesAndFolders = "Privacy_FilesAndFolders"
	privacyAllFiles        = "Privacy_AllFiles"
)

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
	folderAccess       func(dir string) files.Access                       // may block on a network volume
	fullDiskAccess     func() bool
	openPrivacy        func(anchor string) error
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
		restartFinder:  finder.Restart,
		openDialog:     mygo.Dialog.Open,
		setLocale:      mygo.App.SetLocale,
		folderAccess:   files.FolderAccess,
		fullDiskAccess: finder.FullDiskAccess,
		openPrivacy: func(anchor string) error {
			return mygo.Shell.OpenExternal("x-apple.systempreferences:com.apple.preference.security?" + anchor)
		},
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
