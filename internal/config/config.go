// Package config holds the settings that the app and its Finder extension share.
// The app writes the file; the extension reads it when it starts.
package config

import (
	"encoding/json"
	"errors"
	"github.com/hy-shine/rightclick/internal/locales"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo/i18n"
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
	Label    string `json:"label,omitempty"`
}

type FavoriteFolder struct {
	ID      string `json:"id"`
	Path    string `json:"path"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

type ZipExclusions struct {
	DSStore     bool `json:"dsStore"`
	Git         bool `json:"git"`
	NodeModules bool `json:"nodeModules"`
}

func (options ZipExclusions) Names() []string {
	var names []string
	if options.DSStore {
		names = append(names, ".DS_Store")
	}
	if options.Git {
		names = append(names, ".git")
	}
	if options.NodeModules {
		names = append(names, "node_modules")
	}
	return names
}

// FileType is one file the New File submenu can create.
type FileType struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Extension    string `json:"extension"`
	DefaultName  string `json:"defaultName"`
	Content      string `json:"content"`
	Location     string `json:"location"` // always FileTypeSubmenu; kept for old files
	Enabled      bool   `json:"enabled"`
	Template     string `json:"template,omitempty"`
	TemplateData string `json:"templateData,omitempty"`
	TemplateName string `json:"templateName,omitempty"`
}

// Languages lists the choices of the language setting. "" follows the system;
// the others are catalog tags.
var Languages = []string{"", "zh-Hans", "zh-Hant", "en"}

var Appearances = []string{"", "light", "dark"}

var HashAlgorithms = []string{"SHA-256", "SHA-1", "MD5"}

// Config is the saved settings.
type Config struct {
	// Language is the language the app shows, one of Languages.
	Language             string     `json:"language,omitempty"`
	Appearance           string     `json:"appearance,omitempty"`
	DefaultHashAlgorithm string     `json:"defaultHashAlgorithm"`
	Version              int        `json:"version"`
	MenuItems            []MenuItem `json:"menuItems"`
	FileTypes            []FileType `json:"fileTypes"`
	Programs             []string   `json:"programs"`
	// Legacy lists, read once during migration and dropped on the next save.
	Terminal        string           `json:"terminal,omitempty"`
	Terminals       []string         `json:"terminals,omitempty"`
	Editors         []string         `json:"editors,omitempty"`
	OpenAfterCreate bool             `json:"openAfterCreate"`
	FavoriteFolders []FavoriteFolder `json:"favoriteFolders,omitempty"`
	CopyBaseFolder  string           `json:"copyBaseFolder,omitempty"`
	ZipExclusions   ZipExclusions    `json:"zipExclusions"`
}

// defaultMenuIDs lists the menu entries in their default order, as in ui.md section 2.
var defaultMenuIDs = []string{
	"new-file", "new-folder", "new-template",
	"copy-path", "copy-name", "copy-basename", "copy-to", "move-to",
	"copy-terminal-path", "copy-relative-path", "copy-markdown-link",
	"open-terminal",
	"zip", "unzip", "unzip-to-folder", "hash", "compare-hash",
	"toggle-hidden", "delete-permanently",
	"add-template",
}

// defaultOffMenuIDs lists the menu entries that start switched off. Deleting for
// good, Open with Terminal and the template entries stay out of the menu until the
// user turns them on.
var defaultOffMenuIDs = map[string]bool{
	"delete-permanently": true, "open-terminal": true, "new-template": true, "add-template": true,
	"copy-terminal-path": true, "copy-relative-path": true, "copy-markdown-link": true,
}

// defaultOffFileTypeIDs lists the file types that start switched off: the programming
// languages. They stay in the list, so the settings pane can turn them on, but a new
// install shows only documents in New File.
var defaultOffFileTypeIDs = map[string]bool{
	"javascript": true, "typescript": true, "swift": true, "python": true, "go": true,
	"rust": true, "java": true, "c": true, "cpp": true, "sql": true, "shell": true,
}

// defaultFileTypes lists the text files the New File submenu creates. Binary
// document formats need real templates and are deliberately not included.
var defaultFileTypes = []FileType{
	{ID: "word", Label: "", Extension: "docx", DefaultName: "", Template: "docx", Location: FileTypeSubmenu},
	{ID: "excel", Label: "", Extension: "xlsx", DefaultName: "", Template: "xlsx", Location: FileTypeSubmenu},
	{ID: "powerpoint", Label: "", Extension: "pptx", DefaultName: "", Template: "pptx", Location: FileTypeSubmenu},
	{ID: "wps-document", Label: "", Extension: "docx", DefaultName: "", Template: "docx", Location: FileTypeSubmenu},
	{ID: "wps-spreadsheet", Label: "", Extension: "xlsx", DefaultName: "", Template: "xlsx", Location: FileTypeSubmenu},
	{ID: "wps-presentation", Label: "", Extension: "pptx", DefaultName: "", Template: "pptx", Location: FileTypeSubmenu},
	{ID: "rtf", Label: "", Extension: "rtf", DefaultName: "", Template: "rtf", Location: FileTypeSubmenu},
	{
		ID: "text", Label: "", Extension: "txt", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "markdown", Label: "Markdown", Extension: "md", DefaultName: "",
		Content: "# Title\n\n", Location: FileTypeSubmenu,
	},
	{
		ID: "json", Label: "JSON", Extension: "json", DefaultName: "",
		Content: "{}\n", Location: FileTypeSubmenu,
	},
	{
		ID: "csv", Label: "CSV", Extension: "csv", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "html", Label: "HTML", Extension: "html", DefaultName: "",
		Content: "<!doctype html>\n<html lang=\"zh-CN\">\n<head>\n" +
			"  <meta charset=\"utf-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n" +
			"  <title>Untitled</title>\n</head>\n<body>\n\n</body>\n</html>\n",
		Location: FileTypeSubmenu,
	},
	{
		ID: "css", Label: "CSS", Extension: "css", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "javascript", Label: "JavaScript", Extension: "js", DefaultName: "",
		Content: "\"use strict\";\n", Location: FileTypeSubmenu,
	},
	{
		ID: "swift", Label: "Swift", Extension: "swift", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "python", Label: "Python", Extension: "py", DefaultName: "",
		Content: "#!/usr/bin/env python3\n", Location: FileTypeSubmenu,
	},
	{
		ID: "go", Label: "Go", Extension: "go", DefaultName: "",
		Content: "package main\n\n", Location: FileTypeSubmenu,
	},
	{
		ID: "rust", Label: "Rust", Extension: "rs", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "java", Label: "Java", Extension: "java", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "c", Label: "C", Extension: "c", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "cpp", Label: "C++", Extension: "cpp", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "sql", Label: "SQL", Extension: "sql", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "typescript", Label: "TypeScript", Extension: "ts", DefaultName: "",
		Location: FileTypeSubmenu,
	},
	{
		ID: "shell", Label: "Shell", Extension: "sh", DefaultName: "",
		Content: "#!/usr/bin/env bash\n", Location: FileTypeSubmenu,
	},
}

// Default returns the settings a new install starts with.
func Default() Config {
	items := make([]MenuItem, 0, len(defaultMenuIDs))
	for _, id := range defaultMenuIDs {
		items = append(items, MenuItem{ID: id, Enabled: !defaultOffMenuIDs[id], Location: DefaultMenuLocation(id)})
	}
	return Config{
		Version:              Version,
		MenuItems:            items,
		FileTypes:            DefaultFileTypes(),
		ZipExclusions:        ZipExclusions{DSStore: true},
		DefaultHashAlgorithm: "SHA-256",
		// No app is added for the user: the Open With list is theirs to fill.
		Programs: []string{},
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

// Load reads the settings at path. A missing file means the defaults. Menu
// entries that the file does not mention yet are appended, so that a new
// version of the app can add entries without losing the user's order.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		cfg := Default()
		cfg.GroupMenuByLocation()
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	cfg := Config{ZipExclusions: ZipExclusions{DSStore: true}, DefaultHashAlgorithm: "SHA-256"}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, errors.New(locales.ConfigReadFailed(err.Error()))
	}
	legacy := hasLegacyNames(cfg.FileTypes)
	cfg.normalize()
	cfg.addMissingMenuItems()
	if legacy {
		// The Finder extension reads this file directly, so the cleared names are saved
		// here. A failed save is not fatal: the names are cleared in memory and the
		// next load tries again.
		_ = Save(path, cfg)
	}
	return cfg, nil
}

// normalize fills newer file type settings and migrates the two creation
// entries used by format version 1.
func (cfg *Config) normalize() {
	cfg.normalizePrograms()
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
		// File types, templates and New Folder live inside New File. An older file that
		// still lists them as their own menu entries drops those entries.
		if strings.HasPrefix(item.ID, "type:") || strings.HasPrefix(item.ID, "template:") {
			continue
		}
		if item.ID == "new-folder" || item.ID == "new-template" {
			item.Location = FileTypeSubmenu
		}
		switch item.ID {
		case "new-text", "new-markdown":
			legacySeen = true
			if insertNewFile < 0 {
				insertNewFile = len(migrated)
			}
			legacyEnabled = legacyEnabled || item.Enabled
		case "copy-extension":
			// Copy Extension is gone. Its menu entry is dropped on load.
			continue
		case "open-with":
			// The Open With submenu is gone; app shortcut entries still use the
			// URL action, so only the menu entry is dropped.
			continue
		case "open-editor":
			// Open with Default App is gone; Finder already opens a double-click
			// with the default app, so the entry is dropped on load.
			continue
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
	if !slices.Contains(Languages, cfg.Language) {
		cfg.Language = ""
	}
	if !slices.Contains(Appearances, cfg.Appearance) {
		cfg.Appearance = ""
	}
	if !slices.Contains(HashAlgorithms, cfg.DefaultHashAlgorithm) {
		cfg.DefaultHashAlgorithm = "SHA-256"
	}
	cfg.Version = Version
	cfg.GroupMenuByLocation()
}

// GroupMenuByLocation puts the main-menu entries before the submenu entries and
// keeps the order inside each part. Finder reads each part in order, so it shows
// the same menu afterwards. The settings pane depends on the split: it draws a
// rule between the parts, and a drag cannot move an entry across it.
func (cfg *Config) GroupMenuByLocation() {
	grouped := make([]MenuItem, 0, len(cfg.MenuItems))
	for _, item := range cfg.MenuItems {
		if item.Location == FileTypeMain {
			grouped = append(grouped, item)
		}
	}
	for _, item := range cfg.MenuItems {
		if item.Location != FileTypeMain {
			grouped = append(grouped, item)
		}
	}
	cfg.MenuItems = grouped
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
		fileType = migrateLegacyNames(fileType)
		if legacy {
			fileType.Enabled = fileType.Location != FileTypeHidden
		}
		if fileType.Location == FileTypeHidden {
			fileType.Enabled = false
		}
		// File types always live inside the New File submenu. The field stays
		// for older config files that still carry a location.
		fileType.Location = FileTypeSubmenu
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
			cfg.MenuItems = append(cfg.MenuItems, MenuItem{ID: id, Enabled: !defaultOffMenuIDs[id], Location: DefaultMenuLocation(id)})
		}
	}
}

func DefaultFileTypes() []FileType {
	types := slices.Clone(defaultFileTypes)
	for index := range types {
		types[index].Enabled = !defaultOffFileTypeIDs[types[index].ID]
	}
	return types
}

// DisplayLabel returns the name that menus and lists show for f. A built-in type
// stores no label, so its name follows the language the app shows.
func (f FileType) DisplayLabel() string {
	if f.Label != "" {
		return f.Label
	}
	if name := builtinLabel(locales.In(i18n.Locale()), f.ID); name != "" {
		return name
	}
	return f.Extension
}

// DisplayDefaultName returns the file name, without extension, that a new file of
// type f starts with.
func (f FileType) DisplayDefaultName() string {
	if f.DefaultName != "" {
		return f.DefaultName
	}
	return locales.In(i18n.Locale()).DefaultUntitled()
}

// Stored returns f ready to save. A name that equals the current language's
// built-in or default name is stored empty, so the saved file keeps following the
// language. Any other name is kept as typed.
func (f FileType) Stored() FileType {
	msgs := locales.In(i18n.Locale())
	if f.Label == builtinLabel(msgs, f.ID) {
		f.Label = ""
	}
	if f.DefaultName == msgs.DefaultUntitled() {
		f.DefaultName = ""
	}
	return f
}

// builtinLabel returns the name of a built-in file type in msgs' language, or ""
// for a type that has no built-in name.
func builtinLabel(msgs locales.Messages, id string) string {
	switch id {
	case "word":
		return msgs.FileTypeWord()
	case "excel":
		return msgs.FileTypeExcel()
	case "powerpoint":
		return msgs.FileTypePowerPoint()
	case "wps-document":
		return msgs.FileTypeWPSDocument()
	case "wps-spreadsheet":
		return msgs.FileTypeWPSSpreadsheet()
	case "wps-presentation":
		return msgs.FileTypeWPSPresentation()
	case "rtf":
		return msgs.FileTypeRTF()
	case "text":
		return msgs.FileTypeText()
	}
	return ""
}

// hasLegacyNames reports whether migrateLegacyNames would change one of types.
func hasLegacyNames(types []FileType) bool {
	for _, f := range types {
		if migrateLegacyNames(f) != f {
			return true
		}
	}
	return false
}

// migrateLegacyNames clears the names that format version 3 and earlier wrote into
// a file type: the Chinese built-in names and locales.DefaultUntitled() as the default name. Those
// were the Chinese catalog text, so they would stay Chinese after the language
// changes. An empty name means the current language's name is used.
func migrateLegacyNames(f FileType) FileType {
	zh := locales.In("zh-Hans")
	if f.Label != "" && f.Label == builtinLabel(zh, f.ID) {
		f.Label = ""
	}
	if f.DefaultName == zh.DefaultUntitled() {
		f.DefaultName = ""
	}
	return f
}

func DefaultMenuLocation(id string) string {
	switch id {
	case "new-file", "copy-path", "open-terminal":
		return FileTypeMain
	}
	return FileTypeSubmenu
}

// normalizePrograms fills the program list from the old separate editor and
// terminal lists. An empty but present list is a deliberate choice and stays
// empty.
func (cfg *Config) normalizePrograms() {
	if cfg.Programs == nil {
		legacy := append(slices.Clone(cfg.Editors), cfg.Terminals...)
		if len(legacy) == 0 && cfg.Terminal != "" {
			legacy = []string{cfg.Terminal}
		}
		cfg.Programs = legacy
	}
	programs := make([]string, 0, len(cfg.Programs))
	seen := map[string]bool{}
	for _, name := range cfg.Programs {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		programs = append(programs, name)
	}
	cfg.Programs = programs
}

// terminalApps are the programs that open a folder instead of a file. The list
// is a heuristic: any other app gets the selection itself.
var terminalApps = map[string]bool{
	"Terminal": true, "iTerm": true, "iTerm2": true, "Warp": true,
	"Ghostty": true, "kitty": true, "Alacritty": true, "WezTerm": true,
	"Hyper": true, "Tabby": true,
}

// IsTerminalApp reports whether app opens a folder rather than a file.
func IsTerminalApp(app string) bool { return terminalApps[app] }

// Save writes cfg to path. The previous file stays if the write fails.
func Save(path string, cfg Config) error {
	cfg.Version = Version
	cfg.normalizePrograms()
	for index := range cfg.FileTypes {
		cfg.FileTypes[index].Location = FileTypeSubmenu
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
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
	// The settings name the user's folders and programs, so only the owner may read them.
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
