package actionicons

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestEveryNameParses(t *testing.T) {
	for id := range Names {
		if Icon(id) == nil {
			t.Errorf("action %q has no parseable icon", id)
		}
	}
}

func TestEveryFileIsMapped(t *testing.T) {
	entries, err := fs.ReadDir(files, "icons")
	if err != nil {
		t.Fatal(err)
	}
	mapped := map[string]bool{}
	for _, name := range Names {
		mapped[name+".svg"] = true
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".svg" {
			continue
		}
		if !mapped[entry.Name()] {
			t.Errorf("icon %q is shipped but no action uses it", entry.Name())
		}
	}
}

// TestNamesMatchFinderExtension keeps the Go and Swift maps in step, the same
// way actions_handlers_test.go keeps the action lists in step.
func TestNamesMatchFinderExtension(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "extension", "RightClickFinderSync", "ActionIcons.swift"))
	if err != nil {
		t.Fatal(err)
	}
	block := regexp.MustCompile(`(?s)static let names: \[String: String\] = \[(.*?)\n    \]`).FindSubmatch(source)
	if block == nil {
		t.Fatal("ActionIcons.names not found")
	}
	inSwift := map[string]string{}
	for _, match := range regexp.MustCompile(`"([a-z-]+)": "([a-z-]+)"`).FindAllSubmatch(block[1], -1) {
		inSwift[string(match[1])] = string(match[2])
	}
	if len(inSwift) != len(Names) {
		t.Fatalf("Swift maps %d actions, Go maps %d", len(inSwift), len(Names))
	}
	for id, name := range Names {
		if inSwift[id] != name {
			t.Errorf("action %q: Go icon %q, Swift icon %q", id, name, inSwift[id])
		}
	}
}
