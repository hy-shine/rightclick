package settings

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
)

// TestSettingsMenuGroupsMatchFinderExtension ensures pane_menu.go stays in lockstep
// with MenuConfig.swift for group structure and visual sections.
func TestSettingsMenuGroupsMatchFinderExtension(t *testing.T) {
	path := filepath.Join("..", "..", "extension", "RightClickFinderSync", "MenuConfig.swift")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	groupsMatch := regexp.MustCompile(`(?s)static let groups: \[\[String\]\] = \[(.*?)\n    \]`).FindSubmatch(src)
	if groupsMatch == nil {
		t.Fatal("MenuActions.groups not found in MenuConfig.swift")
	}
	var swiftIDs []string
	for _, m := range regexp.MustCompile(`"([a-z-]+)"`).FindAllSubmatch(groupsMatch[1], -1) {
		swiftIDs = append(swiftIDs, string(m[1]))
	}
	var goIDs []string
	for _, g := range menuGroups {
		goIDs = append(goIDs, g...)
	}
	if !slices.Equal(goIDs, swiftIDs) {
		t.Errorf("pane_menu.menuGroups = %q, Swift MenuActions.groups = %q", goIDs, swiftIDs)
	}
	for _, item := range config.Default().MenuItems {
		if _, ok := menuEntries()[item.ID]; !ok {
			t.Errorf("menuEntries missing %q", item.ID)
		}
		if _, ok := menuContexts[item.ID]; !ok {
			t.Errorf("menuContexts missing %q", item.ID)
		}
		if sec := menuVisualSection(item.ID); sec < 0 || sec > 5 {
			t.Errorf("menuVisualSection(%q) = %d out of range", item.ID, sec)
		}
	}
}
