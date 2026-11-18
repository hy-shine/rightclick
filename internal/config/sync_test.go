package config

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// TestMenuDefaultsMatchFinderExtension verifies that Go's default menu IDs,
// default-off flags, default locations, and default file types stay aligned
// with the Swift FinderSync extension.
func TestMenuDefaultsMatchFinderExtension(t *testing.T) {
	menuConfigPath := filepath.Join("..", "..", "extension", "RightClickFinderSync", "MenuConfig.swift")
	menuConfigSrc, err := os.ReadFile(menuConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	// 1. defaultMenuIDs == MenuActions.defaultOrder (exact order).
	groupsMatch := regexp.MustCompile(`(?s)static let groups: \[\[String\]\] = \[(.*?)\n    \]`).FindSubmatch(menuConfigSrc)
	if groupsMatch == nil {
		t.Fatal("MenuActions.groups not found in MenuConfig.swift")
	}
	var swiftOrder []string
	for _, m := range regexp.MustCompile(`"([a-z-]+)"`).FindAllSubmatch(groupsMatch[1], -1) {
		swiftOrder = append(swiftOrder, string(m[1]))
	}
	if !slices.Equal(defaultMenuIDs, swiftOrder) {
		t.Errorf("defaultMenuIDs = %q, Swift MenuActions.groups flattened = %q", defaultMenuIDs, swiftOrder)
	}

	// 2. defaultOffMenuIDs == MenuActions.defaultOff (exact set).
	offMatch := regexp.MustCompile(`(?s)static let defaultOff: Set<String> = \[(.*?)\]`).FindSubmatch(menuConfigSrc)
	if offMatch == nil {
		t.Fatal("MenuActions.defaultOff not found in MenuConfig.swift")
	}
	swiftOff := map[string]bool{}
	for _, m := range regexp.MustCompile(`"([a-z-]+)"`).FindAllSubmatch(offMatch[1], -1) {
		swiftOff[string(m[1])] = true
	}
	if len(swiftOff) != len(defaultOffMenuIDs) {
		t.Errorf("defaultOff count mismatch: Go has %d (%v), Swift has %d (%v)",
			len(defaultOffMenuIDs), defaultOffMenuIDs, len(swiftOff), swiftOff)
	}
	for id := range defaultOffMenuIDs {
		if !swiftOff[id] {
			t.Errorf("Go defaultOffMenuIDs has %q, missing from Swift MenuActions.defaultOff", id)
		}
	}
	for id := range swiftOff {
		if !defaultOffMenuIDs[id] {
			t.Errorf("Swift MenuActions.defaultOff has %q, missing from Go defaultOffMenuIDs", id)
		}
	}

	// 3. DefaultMenuLocation matches Swift's fallback main-menu ID list.
	extPath := filepath.Join("..", "..", "extension", "RightClickFinderSync", "FinderSyncExtension.swift")
	extSrc, err := os.ReadFile(extPath)
	if err != nil {
		t.Fatal(err)
	}
	locMatch := regexp.MustCompile(`\(\[(.*?)\]\.contains\(id\) \? FileTypeLocation\.main : FileTypeLocation\.submenu\)`).FindSubmatch(extSrc)
	if locMatch == nil {
		t.Fatal("fallback main-menu location list not found in FinderSyncExtension.swift")
	}
	swiftMain := map[string]bool{}
	for _, m := range regexp.MustCompile(`"([a-z-]+)"`).FindAllSubmatch(locMatch[1], -1) {
		swiftMain[string(m[1])] = true
	}
	for _, id := range defaultMenuIDs {
		wantMain := DefaultMenuLocation(id) == FileTypeMain
		if swiftMain[id] != wantMain {
			t.Errorf("default location for %q: Go main=%v, Swift main=%v", id, wantMain, swiftMain[id])
		}
	}

	// 4. defaultFileTypes IDs and extensions match Swift FileTypes.defaults in exact order.
	ftMatch := regexp.MustCompile(`(?s)static let defaults: \[FileTypeSetting\] = \[(.*?)\n    \]`).FindSubmatch(menuConfigSrc)
	if ftMatch == nil {
		t.Fatal("FileTypes.defaults not found in MenuConfig.swift")
	}
	entryRe := regexp.MustCompile(`FileTypeSetting\(id: "([^"]+)", label: "[^"]*", ext: "([^"]+)"`)
	swiftFileTypes := entryRe.FindAllStringSubmatch(string(ftMatch[1]), -1)
	if len(swiftFileTypes) != len(defaultFileTypes) {
		t.Fatalf("FileTypes.defaults count = %d, Go defaultFileTypes = %d", len(swiftFileTypes), len(defaultFileTypes))
	}
	for i, ft := range defaultFileTypes {
		if swiftFileTypes[i][1] != ft.ID || swiftFileTypes[i][2] != ft.Extension {
			t.Errorf("file type[%d]: Go=(%s, %s), Swift=(%s, %s)",
				i, ft.ID, ft.Extension, swiftFileTypes[i][1], swiftFileTypes[i][2])
		}
	}

	// 5. HTML default content is locale-neutral.
	for _, ft := range defaultFileTypes {
		if ft.ID == "html" && strings.Contains(ft.Content, `lang="zh-CN"`) {
			t.Error("default HTML template still hardcodes lang=\"zh-CN\"")
		}
	}
}
