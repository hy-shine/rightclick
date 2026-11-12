// Package config holds the settings that the app and its Finder extension share.
// The app writes the file; the extension reads it when it starts.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Version is the file format version that Save writes.
const Version = 3

// A file type is shown in one of these places in the Finder menu.
const (
	FileTypeHidden  = "hidden"
	FileTypeSubmenu = "submenu"
	FileTypeMain    = "main"
)

// MenuItem is one entry of the Finder menu. Its ID is also the action name in
// rightclick:// URLs.
type MenuItem struct {
	ID       string `json:"id"`
	Enabled  bool   `json:"enabled"`
	Location string `json:"location"`
}

// FileType is one file the New File submenu can create.
type FileType struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Extension    string `json:"extension"`
	DefaultName  string `json:"defaultName"`
	Content      string `json:"content"`
	Location     string `json:"location"`
	Enabled      bool   `json:"enabled"`
	Template     string `json:"template,omitempty"`
	TemplateData string `json:"templateData,omitempty"`
	TemplateName string `json:"templateName,omitempty"`
}

// Config is the saved settings.
type Config struct {
	Version   int        `json:"version"`
	MenuItems []MenuItem `json:"menuItems"`
	FileTypes []FileType `json:"fileTypes"`
	// WatchFolders is the single folder whitelist. The Finder extension shows
	// the menu only inside these folders, and URL actions are refused outside
	// them.
	WatchFolders    []string `json:"watchFolders"`
	Terminal        string   `json:"terminal"`
	Editors         []string `json:"editors"`
	OpenAfterCreate bool     `json:"openAfterCreate"`
}

// defaultMenuIDs lists the menu entries in their default order, as in ui.md section 2.
var defaultMenuIDs = []string{
	"new-file", "new-folder", "new-template",
	"copy-path", "copy-name", "copy-to", "move-to",
	"open-terminal", "open-editor", "open-with",
	"zip", "unzip", "hash", "compare-hash",
	"toggle-hidden",
}

// defaultFileTypes lists the text files the New File submenu creates. Binary
// document formats need real templates and are deliberately not included.
var defaultFileTypes = []FileType{
	{ID: "word", Label: "Word 文档", Extension: "docx", DefaultName: "未命名", Template: "docx", Location: FileTypeSubmenu},
	{ID: "excel", Label: "Excel 工作簿", Extension: "xlsx", DefaultName: "未命名", Template: "xlsx", Location: FileTypeSubmenu},
	{ID: "powerpoint", Label: "PowerPoint 演示", Extension: "pptx", DefaultName: "未命名", Template: "pptx", Location: FileTypeSubmenu},
	{ID: "wps-document", Label: "WPS 文档", Extension: "docx", DefaultName: "未命名", Template: "docx", Location: FileTypeSubmenu},
	{ID: "wps-spreadsheet", Label: "WPS 表格", Extension: "xlsx", DefaultName: "未命名", Template: "xlsx", Location: FileTypeSubmenu},
	{ID: "wps-presentation", Label: "WPS 演示", Extension: "pptx", DefaultName: "未命名", Template: "pptx", Location: FileTypeSubmenu},
	{ID: "rtf", Label: "富文本文档", Extension: "rtf", DefaultName: "未命名", Template: "rtf", Location: FileTypeSubmenu},
	{
		ID: "text", Label: "文本文件", Extension: "txt", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "markdown", Label: "Markdown", Extension: "md", DefaultName: "未命名",
		Content: "# 标题\n\n", Location: FileTypeSubmenu,
	},
	{
		ID: "json", Label: "JSON", Extension: "json", DefaultName: "未命名",
		Content: "{}\n", Location: FileTypeSubmenu,
	},
	{
		ID: "csv", Label: "CSV", Extension: "csv", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "html", Label: "HTML", Extension: "html", DefaultName: "未命名",
		Content: "<!doctype html>\n<html lang=\"zh-CN\">\n<head>\n" +
			"  <meta charset=\"utf-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n" +
			"  <title>未命名</title>\n</head>\n<body>\n\n</body>\n</html>\n",
		Location: FileTypeSubmenu,
	},
	{
		ID: "css", Label: "CSS", Extension: "css", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "javascript", Label: "JavaScript", Extension: "js", DefaultName: "未命名",
		Content: "\"use strict\";\n", Location: FileTypeSubmenu,
	},
	{
		ID: "swift", Label: "Swift", Extension: "swift", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "python", Label: "Python", Extension: "py", DefaultName: "未命名",
		Content: "#!/usr/bin/env python3\n", Location: FileTypeSubmenu,
	},
	{
		ID: "go", Label: "Go", Extension: "go", DefaultName: "未命名",
		Content: "package main\n\n", Location: FileTypeSubmenu,
	},
	{
		ID: "rust", Label: "Rust", Extension: "rs", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "java", Label: "Java", Extension: "java", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "c", Label: "C", Extension: "c", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "cpp", Label: "C++", Extension: "cpp", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "sql", Label: "SQL", Extension: "sql", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "typescript", Label: "TypeScript", Extension: "ts", DefaultName: "未命名",
		Location: FileTypeSubmenu,
	},
	{
		ID: "shell", Label: "Shell", Extension: "sh", DefaultName: "未命名",
		Content: "#!/usr/bin/env bash\n", Location: FileTypeSubmenu,
	},
}

// Default returns the settings a new install starts with.
func Default() Config {
	items := make([]MenuItem, 0, len(defaultMenuIDs))
	for _, id := range defaultMenuIDs {
		items = append(items, MenuItem{ID: id, Enabled: true, Location: DefaultMenuLocation(id)})
	}
	var folders []string
	if home, err := os.UserHomeDir(); err == nil {
		folders = []string{
			filepath.Join(home, "Desktop"),
			filepath.Join(home, "Documents"),
			filepath.Join(home, "Downloads"),
		}
	}
	return Config{
		Version:      Version,
		MenuItems:    items,
		FileTypes:    DefaultFileTypes(),
		WatchFolders: folders,
		Terminal:     "Terminal",
		Editors:      []string{"Visual Studio Code", "Cursor"},
	}
}

// DefaultPath is the file the app keeps its settings in.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "RightClick", "config.json"), nil
}

// InWatchFolders reports whether path is one of folders or inside one. This is
// the single folder whitelist shared by the Finder menu and every URL action.
func InWatchFolders(path string, folders []string) bool {
	path = filepath.Clean(path)
	for _, folder := range folders {
		if folder == "" {
			continue
		}
		folder = filepath.Clean(folder)
		if path == folder {
			return true
		}
		rel, err := filepath.Rel(folder, path)
		if err != nil {
			continue
		}
		if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// Load reads the settings at path. A missing file means the defaults. Menu
// entries that the file does not mention yet are appended, so that a new
// version of the app can add entries without losing the user's order.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("设置文件无法读取：%w", err)
	}
	cfg.normalize()
	cfg.addMissingMenuItems()
	return cfg, nil
}

// normalize fills newer file type settings and migrates the two creation
// entries used by format version 1.
func (cfg *Config) normalize() {
	hadFileTypes := len(cfg.FileTypes) > 0
	legacyMenu := slices.Clone(cfg.MenuItems)

	cfg.FileTypes = normalizeFileTypes(cfg.FileTypes, cfg.Version < Version)
	cfg.addMissingFileTypes()
	if !hadFileTypes {
		migrateLegacyFileTypeState(cfg.FileTypes, legacyMenu)
	}

	var migrated []MenuItem
	insertNewFile := -1
	legacyEnabled := false
	legacySeen := false
	for _, item := range cfg.MenuItems {
		switch item.ID {
		case "new-text", "new-markdown":
			legacySeen = true
			if insertNewFile < 0 {
				insertNewFile = len(migrated)
			}
			legacyEnabled = legacyEnabled || item.Enabled
		case "new-file":
			insertNewFile = -1
			migrated = append(migrated, item)
		default:
			migrated = append(migrated, item)
		}
	}
	if legacySeen && !slices.ContainsFunc(migrated, func(item MenuItem) bool { return item.ID == "new-file" }) {
		item := MenuItem{ID: "new-file", Enabled: true}
		if legacySeen {
			item.Enabled = legacyEnabled
		}
		migrated = slices.Insert(migrated, min(insertNewFile, len(migrated)), item)
	}
	cfg.MenuItems = migrated
	for index := range cfg.MenuItems {
		if cfg.MenuItems[index].Location != FileTypeMain && cfg.MenuItems[index].Location != FileTypeSubmenu {
			cfg.MenuItems[index].Location = DefaultMenuLocation(cfg.MenuItems[index].ID)
		}
	}
	cfg.Version = Version
}

// normalizeFileTypes keeps valid unique entries and canonicalizes each
// location.
func normalizeFileTypes(types []FileType, legacy bool) []FileType {
	seen := make(map[string]bool, len(types))
	out := make([]FileType, 0, len(types))
	for _, fileType := range types {
		fileType.ID = strings.TrimSpace(fileType.ID)
		if fileType.ID == "" || seen[fileType.ID] {
			continue
		}
		seen[fileType.ID] = true
		if legacy {
			fileType.Enabled = fileType.Location != FileTypeHidden
		}
		if fileType.Location == FileTypeHidden {
			fileType.Enabled = false
			fileType.Location = FileTypeSubmenu
		}
		switch fileType.Location {
		case FileTypeMain, FileTypeSubmenu:
		default:
			fileType.Location = FileTypeSubmenu
		}
		out = append(out, fileType)
	}
	return out
}

// addMissingFileTypes appends defaults introduced in later versions without
// moving entries the user already arranged.
func (cfg *Config) addMissingFileTypes() {
	have := make(map[string]bool, len(cfg.FileTypes))
	for _, fileType := range cfg.FileTypes {
		have[fileType.ID] = true
	}
	for _, fileType := range DefaultFileTypes() {
		if !have[fileType.ID] {
			cfg.FileTypes = append(cfg.FileTypes, fileType)
		}
	}
}

// migrateLegacyFileTypeState carries the old New Text and Markdown switches
// into the new per-type locations.
func migrateLegacyFileTypeState(types []FileType, items []MenuItem) {
	enabled := make(map[string]bool, len(items))
	present := make(map[string]bool, len(items))
	for _, item := range items {
		enabled[item.ID] = item.Enabled
		present[item.ID] = true
	}
	for i := range types {
		legacyID := ""
		switch types[i].ID {
		case "text":
			legacyID = "new-text"
		case "markdown":
			legacyID = "new-markdown"
		}
		if present[legacyID] && !enabled[legacyID] {
			types[i].Enabled = false
		}
	}
}

// addMissingMenuItems appends default menu entries that cfg does not have.
func (cfg *Config) addMissingMenuItems() {
	have := map[string]bool{}
	for _, it := range cfg.MenuItems {
		have[it.ID] = true
	}
	for _, id := range defaultMenuIDs {
		if !have[id] {
			cfg.MenuItems = append(cfg.MenuItems, MenuItem{ID: id, Enabled: true, Location: DefaultMenuLocation(id)})
		}
	}
}

func DefaultFileTypes() []FileType {
	types := slices.Clone(defaultFileTypes)
	for index := range types {
		types[index].Enabled = true
	}
	return types
}

func DefaultMenuLocation(id string) string {
	switch id {
	case "new-file", "copy-path", "open-terminal":
		return FileTypeMain
	}
	return FileTypeSubmenu
}

// Save writes cfg to path. The previous file stays if the write fails.
func Save(path string, cfg Config) error {
	cfg.Version = Version
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config.*.part")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
