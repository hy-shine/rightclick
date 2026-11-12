package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestDefaultMenuIsEnabledInOrder(t *testing.T) {
	cfg := Default()
	if len(cfg.MenuItems) != len(defaultMenuIDs) {
		t.Fatalf("got %d menu items, want %d", len(cfg.MenuItems), len(defaultMenuIDs))
	}
	for i, id := range defaultMenuIDs {
		wantEnabled := !defaultOffMenuIDs[id]
		if cfg.MenuItems[i].ID != id || cfg.MenuItems[i].Enabled != wantEnabled {
			t.Fatalf("item %d = %+v, want enabled %t", i, cfg.MenuItems[i], wantEnabled)
		}
	}
	if len(cfg.FileTypes) != len(defaultFileTypes) {
		t.Fatalf("got %d file types, want %d", len(cfg.FileTypes), len(defaultFileTypes))
	}
	for i, want := range defaultFileTypes {
		if got := cfg.FileTypes[i]; got.ID != want.ID || got.Extension != want.Extension ||
			got.Location != FileTypeSubmenu {
			t.Errorf("file type %d = %+v, want default %+v", i, got, want)
		}
	}
}

// Open with Terminal is listed in the menu editor but starts switched off.
func TestOpenTerminalStartsOff(t *testing.T) {
	for _, item := range Default().MenuItems {
		if item.ID == "open-terminal" && item.Enabled {
			t.Fatal("open-terminal is switched on by default")
		}
	}
}

func TestDefaultFileTypesIncludeCodeLanguages(t *testing.T) {
	tests := []struct {
		id, label, extension, content string
	}{
		{"go", "Go", "go", "package main\n\n"},
		{"rust", "Rust", "rs", ""},
		{"java", "Java", "java", ""},
		{"c", "C", "c", ""},
		{"cpp", "C++", "cpp", ""},
		{"sql", "SQL", "sql", ""},
		{"typescript", "TypeScript", "ts", ""},
		{"shell", "Shell", "sh", "#!/usr/bin/env bash\n"},
	}
	cfg := Default()
	for _, tt := range tests {
		index := slices.IndexFunc(cfg.FileTypes, func(fileType FileType) bool { return fileType.ID == tt.id })
		if index < 0 {
			t.Errorf("default file types lack %q", tt.id)
			continue
		}
		got := cfg.FileTypes[index]
		want := FileType{ID: tt.id, Label: tt.label, Extension: tt.extension,
			Content: tt.content, Location: FileTypeSubmenu, Enabled: true}
		if got != want {
			t.Errorf("file type %q = %+v, want %+v", tt.id, got, want)
		}
	}
}

func TestDefaultFileTypesIncludeWPSFormats(t *testing.T) {
	tests := []struct {
		id, label, extension string
	}{
		{"wps-document", "WPS Writer", "docx"},
		{"wps-spreadsheet", "WPS Spreadsheet", "xlsx"},
		{"wps-presentation", "WPS Presentation", "pptx"},
	}
	cfg := Default()
	for _, tt := range tests {
		index := slices.IndexFunc(cfg.FileTypes, func(fileType FileType) bool { return fileType.ID == tt.id })
		if index < 0 {
			t.Errorf("default file types lack %q", tt.id)
			continue
		}
		got := cfg.FileTypes[index]
		// A built-in type stores no name; its label comes from the catalogs.
		want := FileType{ID: tt.id, Extension: tt.extension,
			Template: tt.extension, Location: FileTypeSubmenu, Enabled: true}
		if got != want {
			t.Errorf("file type %q = %+v, want %+v", tt.id, got, want)
		}
		if name := got.DisplayLabel(); name != tt.label {
			t.Errorf("file type %q DisplayLabel() = %q, want %q", tt.id, name, tt.label)
		}
	}
}

func TestBuiltinNamesFollowTheLanguage(t *testing.T) {
	word := FileType{ID: "word", Extension: "docx"}
	for _, tt := range []struct{ tag, label, defaultName string }{
		{"zh-Hans", "Word 文档", "未命名"},
		{"en", "Word Document", "Untitled"},
	} {
		useLocale(t, tt.tag)
		if got := word.DisplayLabel(); got != tt.label {
			t.Errorf("%s: DisplayLabel() = %q, want %q", tt.tag, got, tt.label)
		}
		if got := word.DisplayDefaultName(); got != tt.defaultName {
			t.Errorf("%s: DisplayDefaultName() = %q, want %q", tt.tag, got, tt.defaultName)
		}
	}
}

func TestStoredDropsNamesTheLanguageSupplies(t *testing.T) {
	useLocale(t, "en")
	fileType := FileType{ID: "word", Label: "Word Document", Extension: "docx", DefaultName: "Untitled"}
	if got := fileType.Stored(); got.Label != "" || got.DefaultName != "" {
		t.Errorf("Stored() = label %q, default %q, want both empty", got.Label, got.DefaultName)
	}
	custom := FileType{ID: "custom", Label: "My Notes", Extension: "txt", DefaultName: "notes"}
	if got := custom.Stored(); got != custom {
		t.Errorf("Stored() changed a custom type: %+v", got)
	}
}

// The fixture holds the legacy Chinese names on purpose: they are the data that
// older files stored, not UI text.
func TestLoadClearsChineseNamesFromOlderFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"version":3,"fileTypes":[{"id":"word","label":"Word 文档","extension":"docx","defaultName":"未命名","template":"docx","enabled":true}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	word := cfg.FileTypes[slices.IndexFunc(cfg.FileTypes, func(f FileType) bool { return f.ID == "word" })]
	if word.Label != "" || word.DefaultName != "" {
		t.Errorf("legacy names kept: label %q, default %q", word.Label, word.DefaultName)
	}
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(saved), "Word 文档") {
		t.Error("the saved file still has the Chinese name, so the Finder extension would show it")
	}
	if name := word.DisplayLabel(); name != "Word Document" {
		t.Errorf("DisplayLabel() = %q, want the catalog name", name)
	}
}

func TestLoadMissingFileGivesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MenuItems) != len(defaultMenuIDs) {
		t.Fatalf("got %d menu items, want the defaults", len(got.MenuItems))
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.json")
	cfg := Default()
	cfg.MenuItems[0].Enabled = false
	cfg.FileTypes[1].Location = FileTypeMain
	cfg.Programs = []string{"iTerm", "Warp"}
	cfg.OpenAfterCreate = true
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != Version {
		t.Fatalf("version %d, want %d", got.Version, Version)
	}
	if got.MenuItems[0].Enabled {
		t.Fatal("disabled menu item came back enabled")
	}
	if got.FileTypes[1].Location != FileTypeSubmenu || !got.OpenAfterCreate {
		t.Fatalf("round trip lost file type settings: %+v", got)
	}
	if !slices.Equal(got.Programs, []string{"iTerm", "Warp"}) {
		t.Fatalf("round trip lost settings: %+v", got)
	}
}

func TestLoadMergesTheLegacyProgramLists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":3,"editors":["Cursor"],"terminals":["Warp"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Cursor", "Warp"}; !slices.Equal(got.Programs, want) {
		t.Fatalf("programs = %q, want %q", got.Programs, want)
	}
}

func TestLoadSeedsProgramsFromTheLegacyTerminal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":3,"terminal":"Warp"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Warp"}; !slices.Equal(got.Programs, want) {
		t.Fatalf("programs = %q, want %q", got.Programs, want)
	}
}

func TestLoadMigratesLegacyNewEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"version":1,"menuItems":[
		{"id":"new-text","enabled":false},
		{"id":"new-markdown","enabled":true},
		{"id":"new-folder","enabled":true}
	]}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if slices.ContainsFunc(got.MenuItems, func(item MenuItem) bool {
		return item.ID == "new-text" || item.ID == "new-markdown"
	}) {
		t.Fatalf("legacy creation entries remain: %+v", got.MenuItems)
	}
	if len(got.MenuItems) == 0 || got.MenuItems[0].ID != "new-file" || !got.MenuItems[0].Enabled {
		t.Fatalf("migrated menu does not start with enabled new-file: %+v", got.MenuItems)
	}
	for _, fileType := range got.FileTypes {
		if fileType.ID == "text" && fileType.Enabled {
			t.Errorf("disabled new-text was not migrated: %+v", fileType)
		}
		if fileType.ID == "markdown" && !fileType.Enabled {
			t.Errorf("markdown was disabled: %+v", fileType)
		}
	}
}

func TestLoadAppendsMissingFileTypes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"version":2,"fileTypes":[
		{"id":"python","label":"Python 3","extension":"py","defaultName":"script","content":"","location":"main"}
	]}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.FileTypes[0].Label != "Python 3" || got.FileTypes[0].Location != FileTypeSubmenu {
		t.Errorf("the configured type changed: %+v", got.FileTypes[0])
	}
	if len(got.FileTypes) != len(defaultFileTypes) {
		t.Fatalf("owed %d file types, got %d", len(defaultFileTypes), len(got.FileTypes))
	}
}

func TestLoadAppendsMenuItemsTheFileDoesNotMention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"version":1,"menuItems":[{"id":"zip","enabled":false}]}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.MenuItems) != len(defaultMenuIDs) {
		t.Fatalf("got %d items, want %d", len(got.MenuItems), len(defaultMenuIDs))
	}
	if got.MenuItems[0].ID != "zip" || got.MenuItems[0].Enabled {
		t.Fatalf("the user's first item changed: %+v", got.MenuItems[0])
	}
}

func TestLoadRejectsBrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected an error for a broken file")
	}
}

// TestLoadDropsRemovedOpenWith keeps an older file from carrying a menu entry
// whose submenu no longer exists. App shortcut entries use the same URL action,
// so only the menu entry is dropped.
func TestLoadDropsRemovedOpenWith(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"version":3,"menuItems":[{"id":"open-with","enabled":true,"location":"submenu"},{"id":"app:Cursor","enabled":true,"location":"main"}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range cfg.MenuItems {
		if item.ID == "open-with" {
			t.Fatal("a removed menu entry survived the load")
		}
	}
	if !slices.ContainsFunc(cfg.MenuItems, func(item MenuItem) bool { return item.ID == "app:Cursor" }) {
		t.Error("the app shortcut entry was dropped with the removed entry")
	}
}

// TestGroupMenuByLocationKeepsOrderInEachPart checks that the main entries move
// ahead of the submenu entries and that each part keeps its own order.
func TestGroupMenuByLocationKeepsOrderInEachPart(t *testing.T) {
	cfg := Config{MenuItems: []MenuItem{
		{ID: "a", Location: FileTypeSubmenu}, {ID: "b", Location: FileTypeMain},
		{ID: "c", Location: FileTypeSubmenu}, {ID: "d", Location: FileTypeMain},
	}}
	cfg.GroupMenuByLocation()
	var got []string
	for _, item := range cfg.MenuItems {
		got = append(got, item.ID)
	}
	if want := []string{"b", "d", "a", "c"}; !slices.Equal(got, want) {
		t.Fatalf("order = %q, want %q", got, want)
	}
}

// TestChildEntriesStayInsideNew checks that file types, templates and New Folder
// never keep a top-level entry of their own. Only New File chooses a place.
func TestChildEntriesStayInsideNew(t *testing.T) {
	cfg := Config{MenuItems: []MenuItem{
		{ID: "new-file", Enabled: true, Location: FileTypeMain},
		{ID: "new-folder", Enabled: true, Location: FileTypeMain},
		{ID: "type:word", Enabled: true, Location: FileTypeMain},
		{ID: "template:notes.md", Enabled: true, Location: FileTypeMain},
	}}
	cfg.normalize()
	for _, item := range cfg.MenuItems {
		if strings.HasPrefix(item.ID, "type:") || strings.HasPrefix(item.ID, "template:") {
			t.Fatalf("%s still reaches the menu", item.ID)
		}
		if item.ID == "new-folder" && item.Location != FileTypeSubmenu {
			t.Fatalf("new folder has location %q, want it inside New File", item.Location)
		}
		if item.ID == "new-file" && item.Location != FileTypeMain {
			t.Fatalf("new file lost its own place: %q", item.Location)
		}
	}
}

// TestLoadDropsTheRemovedCopyExtension keeps a saved Copy Extension entry from
// coming back after the action was removed.
func TestLoadDropsTheRemovedCopyExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"version":3,"menuItems":[{"id":"copy-extension","enabled":true,"location":"submenu"},{"id":"copy-path","enabled":true,"location":"main"}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if slices.ContainsFunc(cfg.MenuItems, func(item MenuItem) bool { return item.ID == "copy-extension" }) {
		t.Error("the removed Copy Extension entry was kept")
	}
}
