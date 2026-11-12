package finder

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeBundle makes an app bundle with the Finder extension in it, and points the
// program's path at that bundle. It returns the extension's path.
func fakeBundle(t *testing.T) string {
	t.Helper()
	app := filepath.Join(t.TempDir(), "RightClick.app")
	exeDir := filepath.Join(app, "Contents", "MacOS")
	appex := filepath.Join(app, "Contents", "PlugIns", extensionBundle)
	for _, dir := range []string{exeDir, appex} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	exe := filepath.Join(exeDir, "RightClick")
	executablePath = func() (string, error) { return exe, nil }
	t.Cleanup(func() { executablePath = os.Executable })
	return appex
}

// fakePluginkit replaces the commands. listing is what pluginkit -mv prints. It
// returns the commands that ran, in order, and an error from pluginkit -e when
// enableErr is set.
func fakePluginkit(t *testing.T, listing string, enableErr error) *[][]string {
	t.Helper()
	var calls [][]string
	old := runCommand
	t.Cleanup(func() { runCommand = old })
	runCommand = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		switch {
		case len(args) > 0 && args[0] == "-mv":
			return []byte(listing), nil
		case len(args) > 0 && args[0] == "-e" && enableErr != nil:
			return []byte("denied"), enableErr
		}
		return nil, nil
	}
	return &calls
}

// listingAt is the line pluginkit -mv prints for an extension registered at path.
func listingAt(path string) string {
	return "+    " + ExtensionID + "(0.1.0)\tUUID\t2026-11-11 15:29:30 +0000\t" + path + "\n (1 plug-in)\n"
}

func TestExtensionPathIsInsideTheApp(t *testing.T) {
	appex := fakeBundle(t)
	got, err := ExtensionPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != appex {
		t.Errorf("ExtensionPath() = %q, want %q", got, appex)
	}
}

func TestExtensionPathNeedsTheBundle(t *testing.T) {
	fakeBundle(t)
	if err := os.RemoveAll(filepath.Join(filepath.Dir(mustExe(t)), "..", "PlugIns")); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtensionPath(); err == nil {
		t.Error("ExtensionPath() succeeded without the extension in the bundle")
	}
}

func mustExe(t *testing.T) string {
	t.Helper()
	exe, err := executablePath()
	if err != nil {
		t.Fatal(err)
	}
	return exe
}

func TestRegisterSkipsAPathThatIsAlreadyKnown(t *testing.T) {
	appex := fakeBundle(t)
	calls := fakePluginkit(t, listingAt(appex), nil)
	if err := Register(); err != nil {
		t.Fatal(err)
	}
	for _, c := range *calls {
		if len(c) > 1 && c[1] == "-a" {
			t.Errorf("Register ran %v although the path is already registered", c)
		}
	}
}

func TestRegisterAddsAMissingExtension(t *testing.T) {
	appex := fakeBundle(t)
	calls := fakePluginkit(t, "", nil)
	if err := Register(); err != nil {
		t.Fatal(err)
	}
	want := []string{"pluginkit", "-a", appex}
	if last := (*calls)[len(*calls)-1]; !reflect.DeepEqual(last, want) {
		t.Errorf("last command = %v, want %v", last, want)
	}
}

// A moved or rebuilt-elsewhere app has another path registered. Its path is missing,
// so the current one is registered too.
func TestRegisterAddsTheCurrentPathWhenOnlyAnotherIsKnown(t *testing.T) {
	appex := fakeBundle(t)
	calls := fakePluginkit(t, listingAt("/Applications/Old.app/Contents/PlugIns/RightClickFinderSync.appex"), nil)
	if err := Register(); err != nil {
		t.Fatal(err)
	}
	want := []string{"pluginkit", "-a", appex}
	if last := (*calls)[len(*calls)-1]; !reflect.DeepEqual(last, want) {
		t.Errorf("last command = %v, want %v", last, want)
	}
}

func TestEnableRegistersThenTurnsOn(t *testing.T) {
	appex := fakeBundle(t)
	calls := fakePluginkit(t, "", nil)
	if err := Enable(); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"pluginkit", "-mv", "-p", "com.apple.FinderSync", "-i", ExtensionID},
		{"pluginkit", "-a", appex},
		{"pluginkit", "-e", "use", "-i", ExtensionID},
	}
	if !reflect.DeepEqual(*calls, want) {
		t.Errorf("commands = %v, want %v", *calls, want)
	}
}

func TestEnableReportsTheCommandThatFailed(t *testing.T) {
	fakeBundle(t)
	fakePluginkit(t, "", errors.New("exit status 1"))
	err := Enable()
	if err == nil {
		t.Fatal("Enable() succeeded although pluginkit -e failed")
	}
	if !strings.Contains(err.Error(), "pluginkit -e use") || !strings.Contains(err.Error(), "denied") {
		t.Errorf("error %q does not name the command and its output", err)
	}
}
