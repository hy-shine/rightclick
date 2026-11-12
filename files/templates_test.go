package files

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadTemplatesMissingFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "templates.json")
	got, err := LoadTemplates(path)
	if err != nil {
		t.Fatalf("LoadTemplates(%q): %v", path, err)
	}
	if got != nil {
		t.Errorf("LoadTemplates(%q) = %v, want nil", path, got)
	}
}

func TestSaveAndLoadTemplatesRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "RightClick", "templates.json")
	want := []Template{
		{Name: "Outline", Content: "# Title\n\n- item\n"},
		{Name: "Empty", Content: ""},
	}
	if err := SaveTemplates(path, want); err != nil {
		t.Fatalf("SaveTemplates(%q): %v", path, err)
	}
	got, err := LoadTemplates(path)
	if err != nil {
		t.Fatalf("LoadTemplates(%q): %v", path, err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("LoadTemplates = %+v, want %+v", got, want)
	}
}

func TestTemplateFromTextFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outline.md")
	content := "# 标题\n\n- 要点\n"
	writeFile(t, path, content)
	got, err := TemplateFromFile(path)
	if err != nil {
		t.Fatalf("TemplateFromFile(%q): %v", path, err)
	}
	want := Template{Name: "outline.md", Content: content}
	if got != want {
		t.Errorf("TemplateFromFile = %+v, want %+v", got, want)
	}
}

func TestTemplateFromBinaryFileFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bin")
	writeFile(t, path, "text\x00binary")
	if _, err := TemplateFromFile(path); err == nil {
		t.Errorf("TemplateFromFile(%q) error = nil, want error for a file with a NUL byte", path)
	}
}
