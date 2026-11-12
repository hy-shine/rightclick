package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/files"
)

func TestCaptureTemplateValidatesBeforeOpeningEditor(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	if err := os.WriteFile(path, []byte("port: 8080\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var captured []files.Template
	present := func(template files.Template) error { captured = append(captured, template); return nil }
	if err := captureTemplate([]string{path}, present); err != nil {
		t.Fatal(err)
	}
	if len(captured) != 1 || captured[0].Content != "port: 8080\n" {
		t.Fatal("capture did not pass file content to the editor")
	}
	for _, paths := range [][]string{nil, {path, path}, {directory}, {filepath.Join(directory, "missing")}} {
		if err := captureTemplate(paths, present); err == nil {
			t.Fatalf("capture accepted invalid input %q", paths)
		}
	}
	if len(captured) != 1 {
		t.Fatal("invalid file reached the editor")
	}
	busy := errors.New("editor busy")
	if err := captureTemplate([]string{path}, func(files.Template) error { return busy }); !errors.Is(err, busy) {
		t.Fatalf("editor error was lost: %v", err)
	}
}

func TestCustomFileTypeCreatesConfiguredContent(t *testing.T) {
	directory := t.TempDir()
	cfg := config.Default()
	cfg.FileTypes = append(cfg.FileTypes, config.FileType{
		ID: "custom-config", Label: "YAML", Extension: "yaml", DefaultName: "config",
		Content: "enabled: true\n", Enabled: true, Location: config.FileTypeSubmenu,
	})
	if err := createFileFromType(cfg, "custom-config", []string{directory}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "config.yaml"))
	if err != nil || string(data) != "enabled: true\n" {
		t.Fatalf("custom file contents = %q, error = %v", data, err)
	}
}
