package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// extensionBundle is the Finder Sync extension inside the app bundle, in Contents/PlugIns.
const extensionBundle = "RightClickFinderSync.appex"

// executablePath returns the path of the running program. Tests replace it.
var executablePath = os.Executable

// ExtensionPath returns the Finder extension inside the app that runs this program.
// A program run outside an app bundle has none, and that is an error.
func ExtensionPath() (string, error) {
	exe, err := executablePath()
	if err != nil {
		return "", err
	}
	// The program is Contents/MacOS/RightClick, and the extension is Contents/PlugIns.
	path := filepath.Clean(filepath.Join(filepath.Dir(exe), "..", "PlugIns", extensionBundle))
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("finder extension: %w", err)
	}
	return path, nil
}

// Register tells the system where the Finder extension is, so that it can be turned
// on. It does nothing when the system already has the extension at that path, so an
// extension the user switched off stays off. The app calls it at every start, which
// keeps a rebuilt or moved app working.
func Register() error {
	path, err := ExtensionPath()
	if err != nil {
		return err
	}
	known, err := registeredAt(path)
	if err != nil {
		return err
	}
	if known {
		return nil
	}
	out, err := runCommand("pluginkit", "-a", path)
	if err != nil {
		return commandError("pluginkit -a", err, out)
	}
	return nil
}

// Enable registers the extension if it is missing, then turns it on with
// pluginkit -e use. macOS may still ask the user to confirm in System Settings, so
// the caller checks the state afterwards.
func Enable() error {
	if err := Register(); err != nil {
		return err
	}
	out, err := runCommand("pluginkit", "-e", "use", "-i", ExtensionID)
	if err != nil {
		return commandError("pluginkit -e use", err, out)
	}
	return nil
}

// registeredAt reports whether the system has the extension registered at path.
// pluginkit -mv ends each plug-in's line with its path, after a tab.
func registeredAt(path string) (bool, error) {
	out, err := runCommand("pluginkit", "-mv", "-p", "com.apple.FinderSync", "-i", ExtensionID)
	if err != nil {
		return false, commandError("pluginkit -mv", err, out)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ExtensionID) && strings.HasSuffix(strings.TrimRight(line, " \t\r"), "\t"+path) {
			return true, nil
		}
	}
	return false, nil
}

// commandError adds the command's output to its error, when there is any.
func commandError(command string, err error, out []byte) error {
	text := strings.TrimSpace(string(out))
	if text == "" {
		return fmt.Errorf("%s: %w", command, err)
	}
	return fmt.Errorf("%s: %w: %s", command, err, text)
}
