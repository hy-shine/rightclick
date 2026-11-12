package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDisabledTypeRetainsRootPlacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := Default()
	cfg.FileTypes[0].Enabled, cfg.FileTypes[0].Location = false, FileTypeMain
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.FileTypes[0].Enabled || loaded.FileTypes[0].Location != FileTypeMain {
		t.Fatalf("lost disabled state or placement: %+v", loaded.FileTypes[0])
	}
}

func TestVersionTwoHiddenTypeMigratesToDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":2,"fileTypes":[{"id":"text","label":"Text","extension":"txt","defaultName":"custom","content":"hello","location":"hidden"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	item := loaded.FileTypes[0]
	if item.Enabled || item.Location != FileTypeSubmenu || item.Content != "hello" || item.DefaultName != "custom" {
		t.Fatalf("migration lost state: %+v", item)
	}
}
