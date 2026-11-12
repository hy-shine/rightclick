package main

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hy-shine/rightclick/config"
)

func TestFavoriteDestinationCopiesAndMovesWithoutPicker(t *testing.T) {
	for _, action := range []string{"copy-to", "move-to"} {
		t.Run(action, func(t *testing.T) {
			source := filepath.Join(t.TempDir(), "notes.txt")
			if err := os.WriteFile(source, []byte("keep this"), 0o600); err != nil {
				t.Fatal(err)
			}
			destination := t.TempDir()
			cfg := config.Default()
			cfg.FavoriteFolders = []config.FavoriteFolder{{ID: "archive", Path: destination, Label: "Archive", Enabled: true}}
			params := url.Values{"favorite": {"archive"}}
			if err := runAction(action, params, []string{source}, cfg); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(destination, "notes.txt"))
			if err != nil || string(data) != "keep this" {
				t.Fatalf("destination data = %q, error = %v", data, err)
			}
			_, err = os.Stat(source)
			if action == "move-to" && !os.IsNotExist(err) || action == "copy-to" && err != nil {
				t.Fatalf("unexpected source state after %s: %v", action, err)
			}
		})
	}
}

func TestFavoriteDestinationRejectsStaleOrDisabledID(t *testing.T) {
	cfg := config.Default()
	cfg.FavoriteFolders = []config.FavoriteFolder{{ID: "disabled", Path: t.TempDir(), Enabled: false}}
	for _, id := range []string{"disabled", "missing", ""} {
		if err := runAction("copy-to", url.Values{"favorite": {id}}, []string{filepath.Join(t.TempDir(), "source")}, cfg); err == nil {
			t.Fatalf("accepted unavailable favorite %q", id)
		}
	}
}
