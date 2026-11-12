package files

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"rightclick/config"
)

func TestOfficeTemplatesCreateAndRoundTripUnchanged(t *testing.T) {
	for _, fileType := range config.Default().FileTypes {
		if fileType.Template == "" {
			continue
		}
		t.Run(fileType.ID, func(t *testing.T) {
			content, err := FileTypeContent(fileType)
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateDocument(fileType.Extension, content); err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			path, err := CreateFile(directory, "blank."+fileType.Extension, content)
			if err != nil {
				t.Fatal(err)
			}
			template, err := TemplateFromFile(path)
			if err != nil {
				t.Fatal(err)
			}
			store := filepath.Join(directory, "templates.json")
			if err := SaveTemplates(store, []Template{template}); err != nil {
				t.Fatal(err)
			}
			loaded, err := LoadTemplates(store)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := loaded[0].Bytes()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(content, restored) {
				t.Fatal("import/save/load changed document bytes")
			}
			second, err := CreateFile(directory, "blank."+fileType.Extension, restored)
			if err != nil {
				t.Fatal(err)
			}
			if second == path {
				t.Fatal("creation overwrote an existing file")
			}
		})
	}
}

func TestDocumentImportRejectsFakeOfficeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fake.docx")
	if err := os.WriteFile(path, []byte("not an office document"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := TemplateFromFile(path); err == nil {
		t.Fatal("accepted invalid document")
	}
}
