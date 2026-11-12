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
		{"wps-document", "WPS 文档", "docx"},
		{"wps-spreadsheet", "WPS 表格", "xlsx"},
		{"wps-presentation", "WPS 演示", "pptx"},
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
	if name := word.DisplayLabel(); name != "Word 文档" {
		t.Errorf("DisplayLabel() = %q, want the Chinese catalog name", name)
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

func TestInWatchFolders(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name  string
		path  string
		roots []string
		want  bool
	}{
		{"the folder itself", root, []string{root}, true},
		{"a child", filepath.Join(root, "sub", "file.txt"), []string{root}, true},
		{"a trailing separator", root + string(filepath.Separator), []string{root}, true},
		{"a sibling with the same prefix", root + "-copy", []string{root}, false},
		{"outside every folder", filepath.Join(root, "..", "outside.txt"), []string{root}, false},
		{"an empty whitelist", root, nil, false},
		{"an empty entry", root, []string{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InWatchFolders(tt.path, tt.roots); got != tt.want {
				t.Errorf("InWatchFolders(%q, %q) = %v, want %v", tt.path, tt.roots, got, tt.want)
			}
		})
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.json")
	cfg := Default()
	cfg.MenuItems[0].Enabled = false
	cfg.FileTypes[1].Location = FileTypeMain
	cfg.WatchFolders = []string{"/tmp/only"}
	cfg.Terminal = "iTerm"
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
	if len(got.WatchFolders) != 1 || got.WatchFolders[0] != "/tmp/only" || got.Terminal != "iTerm" {
		t.Fatalf("round trip lost settings: %+v", got)
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
		{"id":"python","label":"Python 3","extension":"py","defaultName":"脚本","content":"","location":"main"}
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
