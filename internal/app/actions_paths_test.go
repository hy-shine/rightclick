package app

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/files"
)

func TestTerminalPathActionCopiesSelectedPaths(test *testing.T) {
	previous := copyText
	test.Cleanup(func() { copyText = previous })
	var copied string
	copyText = func(text string) error {
		copied = text
		return nil
	}
	paths := []string{"/Work/My Project/report.txt", "/Work/it's a file.txt"}
	params := url.Values{"path": paths}
	action, parsed, selection, err := parseURL("rightclick://copy-terminal-path?" + params.Encode())
	if err != nil {
		test.Fatal(err)
	}
	if err := runAction(action, parsed, selection, config.Default()); err != nil {
		test.Fatal(err)
	}
	if want := files.TerminalPaths(paths); copied != want {
		test.Fatalf("clipboard = %q, want %q", copied, want)
	}
}

func TestRelativeCopyRemembersBaseAndKeepsCurrentSettings(test *testing.T) {
	settingsPath := filepath.Join(test.TempDir(), "config.json")
	base := test.TempDir()
	initial := config.Default()
	initial.CopyBaseFolder = base
	if err := config.Save(settingsPath, initial); err != nil {
		test.Fatal(err)
	}
	previousPicker, previousPath, previousCopy := chooseCopyBaseFolder, copyConfigPath, copyText
	test.Cleanup(func() { chooseCopyBaseFolder, copyConfigPath, copyText = previousPicker, previousPath, previousCopy })
	copyConfigPath = func() (string, error) { return settingsPath, nil }
	var starts []string
	chooseCopyBaseFolder = func(start string) (string, error) {
		starts = append(starts, start)
		latest, err := config.Load(settingsPath)
		if err != nil {
			return "", err
		}
		latest.Language = "zh-Hant"
		latest.Programs = []string{"TextEdit"}
		if err := config.Save(settingsPath, latest); err != nil {
			return "", err
		}
		return base, nil
	}
	var copied string
	copyText = func(text string) error { copied = text; return nil }
	paths := []string{filepath.Join(base, "notes.md"), filepath.Join(base, "assets")}
	for range 2 {
		loaded, err := config.Load(settingsPath)
		if err != nil {
			test.Fatal(err)
		}
		if err := runAction("copy-relative-path", nil, paths, loaded); err != nil {
			test.Fatal(err)
		}
	}
	if !slices.Equal(starts, []string{base, base}) || copied != "notes.md\nassets" {
		test.Fatalf("picker starts = %q, clipboard = %q", starts, copied)
	}
	saved, err := config.Load(settingsPath)
	if err != nil {
		test.Fatal(err)
	}
	if saved.CopyBaseFolder != base || saved.Language != "zh-Hant" || !slices.Equal(saved.Programs, []string{"TextEdit"}) {
		test.Fatalf("latest settings lost: %+v", saved)
	}
}

func TestRelativeCopyCancellationAndErrorsKeepClipboard(test *testing.T) {
	settingsPath := filepath.Join(test.TempDir(), "config.json")
	if err := config.Save(settingsPath, config.Default()); err != nil {
		test.Fatal(err)
	}
	previousPicker, previousPath, previousCopy := chooseCopyBaseFolder, copyConfigPath, copyText
	test.Cleanup(func() { chooseCopyBaseFolder, copyConfigPath, copyText = previousPicker, previousPath, previousCopy })
	copyConfigPath = func() (string, error) { return settingsPath, nil }
	before, err := os.ReadFile(settingsPath)
	if err != nil {
		test.Fatal(err)
	}
	for _, scenario := range []struct {
		name string
		base string
		err  error
	}{
		{name: "cancelled"},
		{name: "picker error", err: errors.New("picker failed")},
		{name: "missing folder", base: filepath.Join(test.TempDir(), "missing")},
		{name: "relative folder", base: "relative"},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			chooseCopyBaseFolder = func(string) (string, error) { return scenario.base, scenario.err }
			copyText = func(string) error { test.Fatal("clipboard changed"); return nil }
			err := runAction("copy-relative-path", nil, []string{"/Work/file.md"}, config.Default())
			if scenario.name == "cancelled" && err != nil || scenario.name != "cancelled" && err == nil {
				test.Fatalf("error = %v", err)
			}
			after, readErr := os.ReadFile(settingsPath)
			if readErr != nil || string(after) != string(before) {
				test.Fatal("settings changed after cancellation or failure")
			}
		})
	}
}

func TestMarkdownCopyUsesTheRememberedRelativeBase(test *testing.T) {
	settingsPath := filepath.Join(test.TempDir(), "config.json")
	base := test.TempDir()
	cfg := config.Default()
	cfg.CopyBaseFolder = base
	if err := config.Save(settingsPath, cfg); err != nil {
		test.Fatal(err)
	}
	previousPicker, previousPath, previousCopy := chooseCopyBaseFolder, copyConfigPath, copyText
	test.Cleanup(func() { chooseCopyBaseFolder, copyConfigPath, copyText = previousPicker, previousPath, previousCopy })
	copyConfigPath = func() (string, error) { return settingsPath, nil }
	var copied string
	copyText = func(text string) error { copied = text; return nil }
	chooseCopyBaseFolder = func(start string) (string, error) {
		if start != base {
			test.Fatalf("picker start = %q, want %q", start, base)
		}
		return base, nil
	}
	if err := runAction("copy-markdown-link", nil, []string{filepath.Join(base, "[draft] notes.md")}, cfg); err != nil {
		test.Fatal(err)
	}
	if want := `[\[draft\] notes.md](%5Bdraft%5D%20notes.md)`; copied != want {
		test.Fatalf("clipboard = %q, want %q", copied, want)
	}
	chooseCopyBaseFolder = func(string) (string, error) { return "", nil }
	copyText = func(string) error { test.Fatal("cancelled picker changed clipboard"); return nil }
	if err := runAction("copy-markdown-link", nil, []string{filepath.Join(base, "notes.md")}, cfg); err != nil {
		test.Fatal(err)
	}
}

func TestRelativeCopyFormatsEveryPathBeforeClipboardWrite(test *testing.T) {
	previousPicker, previousCopy := chooseCopyBaseFolder, copyText
	test.Cleanup(func() { chooseCopyBaseFolder, copyText = previousPicker, previousCopy })
	base := test.TempDir()
	chooseCopyBaseFolder = func(string) (string, error) { return base, nil }
	copyText = func(string) error { test.Fatal("invalid selection changed clipboard"); return nil }
	if err := runAction("copy-relative-path", nil, []string{filepath.Join(base, "notes.md"), "relative.md"}, config.Default()); err == nil {
		test.Fatal("invalid path accepted")
	}
}
