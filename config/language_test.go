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
