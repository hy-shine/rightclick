package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestArchiveSettingsDefaultAndRoundTrip(test *testing.T) {
	for _, source := range []string{
		`{"version":3,"menuItems":[],"fileTypes":[]}`,
		`{"version":3,"zipExclusions":{"git":true}}`,
	} {
		path := filepath.Join(test.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			test.Fatal(err)
		}
		cfg, err := Load(path)
		if err != nil {
			test.Fatal(err)
		}
		if !cfg.ZipExclusions.DSStore || cfg.ZipExclusions.NodeModules {
			test.Fatalf("missing archive options lost defaults: %+v", cfg.ZipExclusions)
		}
		cfg.CopyBaseFolder = test.TempDir()
		cfg.ZipExclusions = ZipExclusions{Git: true, NodeModules: true}
		if err := Save(path, cfg); err != nil {
			test.Fatal(err)
		}
		saved, err := Load(path)
		if err != nil {
			test.Fatal(err)
		}
		if saved.CopyBaseFolder != cfg.CopyBaseFolder || saved.ZipExclusions != cfg.ZipExclusions {
			test.Fatalf("archive settings changed: %+v", saved)
		}
		if !slices.Equal(saved.ZipExclusions.Names(), []string{".git", "node_modules"}) {
			test.Fatal("archive exclusion names are wrong")
		}
	}
	if options := Default().ZipExclusions; !options.DSStore || options.Git || options.NodeModules {
		test.Fatalf("new install defaults = %+v", options)
	}
}

func TestNewCopyActionsAreDisabledInSubmenu(test *testing.T) {
	path := filepath.Join(test.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":3,"menuItems":[{"id":"copy-path","enabled":true,"location":"main"}]}`), 0o600); err != nil {
		test.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		test.Fatal(err)
	}
	for _, cfg := range []Config{Default(), loaded} {
		for _, action := range []string{"copy-terminal-path", "copy-relative-path", "copy-markdown-link"} {
			position := slices.IndexFunc(cfg.MenuItems, func(item MenuItem) bool { return item.ID == action })
			if position < 0 || cfg.MenuItems[position].Enabled || cfg.MenuItems[position].Location != FileTypeSubmenu {
				test.Fatalf("new action %q is not disabled in the submenu", action)
			}
		}
	}
}
