package main

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"rightclick/config"
	"rightclick/locales"
)

// TestRouteStartsTheActionAfterTheChecks checks that a valid URL is checked at
// once, while the action waits for the start function to run it. The URL handler
// therefore returns before slow file work finishes.
func TestRouteStartsTheActionAfterTheChecks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.WatchFolders = []string{dir}
	path, err := config.DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	raw := "rightclick://new-folder?path=" + url.QueryEscape(dir)
	var pending []func()
	routeURLWith(raw, func(run func()) { pending = append(pending, run) })

	folder := filepath.Join(dir, locales.DefaultNewFolder())
	if _, err := os.Stat(folder); err == nil {
		t.Fatal("the folder exists before the action has started")
	}
	if len(pending) != 1 {
		t.Fatalf("%d actions started, want 1", len(pending))
	}
	pending[0]()
	if _, err := os.Stat(folder); err != nil {
		t.Errorf("the action did not create the folder: %v", err)
	}
}
