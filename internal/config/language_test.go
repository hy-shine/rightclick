package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDropsAnUnknownLanguage checks that a hand-edited language that the
// app cannot show falls back to the system choice instead of reaching SetLocale.
func TestLoadDropsAnUnknownLanguage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":3,"language":"xx-nowhere"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Language != "" {
		t.Errorf("language = %q, want the system choice", cfg.Language)
	}
}

// TestLoadKeepsTraditionalChinese checks that the Traditional Chinese choice is
// one the app can show, so a saved choice is not dropped.
func TestLoadKeepsTraditionalChinese(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":3,"language":"zh-Hant"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Language != "zh-Hant" {
		t.Errorf("language = %q, want zh-Hant", cfg.Language)
	}
}
