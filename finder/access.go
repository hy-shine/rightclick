package finder

import (
	"os"
	"path/filepath"
)

// FullDiskAccess reports whether the app can read locations that macOS guards
// with the Full Disk Access switch. The check reads one file and nothing else, so
// it stays cheap enough to run at start-up.
func FullDiskAccess() bool {
	return canRead(protectedPath())
}

// protectedPath is a location that only an app with Full Disk Access can read.
// The TCC database exists on every macOS version that has the switch.
func protectedPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support", "com.apple.TCC", "TCC.db")
}

// canRead reports whether path can be opened for reading.
func canRead(path string) bool {
	if path == "" {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}
