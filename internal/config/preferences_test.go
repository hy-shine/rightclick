package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferenceDefaultsAndInvalidValues(t *testing.T) {
	for _, sample := range []struct {
		name       string
		json       string
		appearance string
		algorithm  string
	}{
		{"old configuration", `{"version":3,"zipExclusions":{"dsStore":false,"git":true},"copyBaseFolder":"/tmp/base"}`, "", "SHA-256"},
		{"light appearance", `{"version":3,"appearance":"light","defaultHashAlgorithm":"SHA-1"}`, "light", "SHA-1"},
		{"dark appearance", `{"version":3,"appearance":"dark","defaultHashAlgorithm":"MD5"}`, "dark", "MD5"},
		{"system appearance", `{"version":3,"appearance":"","defaultHashAlgorithm":"SHA-256"}`, "", "SHA-256"},
		{"unknown preferences", `{"version":3,"appearance":"night","defaultHashAlgorithm":"SHA-512"}`, "", "SHA-256"},
		{"empty algorithm", `{"version":3,"defaultHashAlgorithm":""}`, "", "SHA-256"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(sample.json), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Appearance != sample.appearance || cfg.DefaultHashAlgorithm != sample.algorithm {
				t.Fatalf("appearance/algorithm = %q/%q, want %q/%q", cfg.Appearance, cfg.DefaultHashAlgorithm, sample.appearance, sample.algorithm)
			}
			if sample.name == "old configuration" && (cfg.CopyBaseFolder != "/tmp/base" || cfg.ZipExclusions != (ZipExclusions{Git: true})) {
				t.Fatalf("existing copy/archive settings changed: %+v", cfg)
			}
		})
	}
	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Appearance != "" || cfg.DefaultHashAlgorithm != "SHA-256" {
		t.Fatalf("missing configuration defaults = %q/%q", cfg.Appearance, cfg.DefaultHashAlgorithm)
	}
	defaults := Default()
	if defaults.Appearance != "" || defaults.DefaultHashAlgorithm != "SHA-256" {
		t.Fatalf("new install defaults = %q/%q", defaults.Appearance, defaults.DefaultHashAlgorithm)
	}
}

func TestPreferencesRoundTripKeepsCopyAndArchiveSettings(t *testing.T) {
	for _, appearance := range Appearances {
		for _, algorithm := range HashAlgorithms {
			t.Run(appearance+"/"+algorithm, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "config.json")
				cfg := Default()
				cfg.Appearance = appearance
				cfg.DefaultHashAlgorithm = algorithm
				cfg.CopyBaseFolder = t.TempDir()
				cfg.ZipExclusions = ZipExclusions{Git: true, NodeModules: true}
				cfg.Language = "en"
				cfg.MenuItems[0].Enabled = false
				if err := Save(path, cfg); err != nil {
					t.Fatal(err)
				}
				saved, err := Load(path)
				if err != nil {
					t.Fatal(err)
				}
				if saved.Appearance != appearance || saved.DefaultHashAlgorithm != algorithm {
					t.Fatalf("saved appearance/algorithm = %q/%q, want %q/%q", saved.Appearance, saved.DefaultHashAlgorithm, appearance, algorithm)
				}
				if saved.CopyBaseFolder != cfg.CopyBaseFolder || saved.ZipExclusions != cfg.ZipExclusions || saved.Language != cfg.Language || saved.MenuItems[0].Enabled {
					t.Fatalf("saving preferences changed other settings: %+v", saved)
				}
			})
		}
	}
}
