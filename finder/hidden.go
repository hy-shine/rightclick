// Package finder reads and changes the Finder settings that the app's menu
// entries use.
package finder

import (
	"fmt"
	"os/exec"
	"strings"
)

// hiddenKey is the Finder preference that shows hidden files. Finder reads it
// only when it starts, so a change takes effect after a restart.
const hiddenKey = "AppleShowAllFiles"

// runCommand runs a command and returns its combined output. Tests replace it.
var runCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// HiddenFilesShown reports whether Finder shows hidden files. A missing key
// means Finder hides them, which is its default.
func HiddenFilesShown() (bool, error) {
	out, err := runCommand("defaults", "read", "com.apple.finder", hiddenKey)
	text := strings.TrimSpace(string(out))
	if err != nil {
		if strings.Contains(text, "does not exist") {
			return false, nil
		}
		return false, fmt.Errorf("读取访达设置失败：%w：%s", err, text)
	}
	switch text {
	case "1":
		return true, nil
	case "0":
		return false, nil
	}
	return false, fmt.Errorf("访达的隐藏文件设置不是 0 或 1：%q", text)
}

// SetHiddenFilesShown saves whether Finder shows hidden files, then restarts
// Finder so the change applies. The system starts Finder again by itself.
func SetHiddenFilesShown(show bool) error {
	value := "FALSE"
	if show {
		value = "TRUE"
	}
	out, err := runCommand("defaults", "write", "com.apple.finder", hiddenKey, "-bool", value)
	if err != nil {
		return fmt.Errorf("写入访达设置失败：%w：%s", err, strings.TrimSpace(string(out)))
	}
	return Restart()
}

// Restart quits Finder. The system starts it again.
func Restart() error {
	out, err := runCommand("killall", "Finder")
	if err != nil {
		return fmt.Errorf("重新启动访达失败：%w：%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
