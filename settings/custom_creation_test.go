package settings

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/files"
	"github.com/hy-shine/rightclick/internal/catalog"
	"github.com/hy-shine/rightclick/locales"
)

func TestCustomFileTypeCreateCancelEditAndDelete(t *testing.T) {
	for _, language := range []string{"zh-Hans", "en"} {
		t.Run(language, func(t *testing.T) {
			useLocale(t, language)
			model := readyApp(t)
			model.selectPane(paneTemplates)
			tester := ui.NewTester(model.view, minWidth, minHeight)
			originalCount := len(model.cfg.FileTypes)
			if err := tester.Click(locales.AddFileType()); err != nil {
				t.Fatal(err)
			}
			if err := tester.Click(locales.Cancel()); err != nil {
				t.Fatal(err)
			}
			if len(model.cfg.FileTypes) != originalCount {
				t.Fatal("cancel created a file type")
			}
			if err := tester.Click(locales.AddFileType()); err != nil {
				t.Fatal(err)
			}
			model.editCreation.Label = "YAML config"
			model.editCreation.Extension = ".yaml"
			model.editCreation.DefaultName = "config"
			model.editCreation.Content = "enabled: true\n"
			tester.Frame()
			if err := tester.Click(locales.SaveChanges()); err != nil {
				t.Fatal(err)
			}
			saved, err := config.Load(model.store.ConfigPath)
			if err != nil {
				t.Fatal(err)
			}
			custom := saved.FileTypes[len(saved.FileTypes)-1]
			if len(saved.FileTypes) != originalCount+1 || custom.Extension != "yaml" || custom.Content != "enabled: true\n" {
				t.Fatalf("custom type was not persisted: %+v", custom)
			}
			model.setShortcut("type:"+custom.ID, true)
			model.openCreationEditor(creationRow{index: len(model.cfg.FileTypes) - 1})
			model.editCreation.Content = "enabled: false\n"
			model.saveCreationEditor()
			if len(model.cfg.FileTypes) != originalCount+1 || model.cfg.FileTypes[len(model.cfg.FileTypes)-1].Content != "enabled: false\n" {
				t.Fatal("editing appended a duplicate or lost content")
			}
			model.openCreationEditor(creationRow{index: len(model.cfg.FileTypes) - 1})
			tester.Frame()
			if err := tester.Click(locales.DeleteFileType()); err != nil {
				t.Fatal(err)
			}
			if !model.confirmDeleteType || len(model.cfg.FileTypes) != originalCount+1 {
				t.Fatal("custom type deleted without confirmation")
			}
			if err := tester.Click(locales.Delete()); err != nil {
				t.Fatal(err)
			}
			saved, err = config.Load(model.store.ConfigPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(saved.FileTypes) != originalCount || model.hasShortcut("type:"+custom.ID) {
				t.Fatalf("deleted type or shortcut survived: count=%d original=%d shortcut=%v dialog=%v confirm=%v notice=%q", len(saved.FileTypes), originalCount, model.hasShortcut("type:"+custom.ID), model.editCreationOpen, model.confirmDeleteType, model.notice)
			}
		})
	}
}

func TestCustomFileTypeRejectsInvalidExtensionsAndCreatesRealDocuments(t *testing.T) {
	model := readyApp(t)
	count := len(model.cfg.FileTypes)
	for _, extension := range []string{"", "../json", "a/b", "a\\b", "a\x00b", "a..b", "a b", "."} {
		model.newFileType()
		model.editCreation.Label = "Config"
		model.editCreation.Extension = extension
		model.saveCreationEditor()
		if !model.editCreationOpen || model.editCreationNotice == "" || len(model.cfg.FileTypes) != count {
			t.Fatalf("invalid extension %q was accepted", extension)
		}
	}
	model.newFileType()
	model.editCreation.Label = "Report"
	model.editCreation.Extension = "docx"
	model.saveCreationEditor()
	if model.editCreationOpen {
		t.Fatal(model.editCreationNotice)
	}
	data, err := files.FileTypeContent(model.cfg.FileTypes[len(model.cfg.FileTypes)-1])
	if err != nil {
		t.Fatal(err)
	}
	if err := files.ValidateDocument("docx", data); err != nil {
		t.Fatal(err)
	}
}

func TestImportedTemplateWaitsForSaveAndAvoidsOverwrite(t *testing.T) {
	model := readyApp(t)
	model.addTemplate(files.Template{Name: "meeting.md", Content: "original"})
	if err := model.beginTemplate(files.Template{Name: "meeting.md", Content: "new"}); err != nil {
		t.Fatal(err)
	}
	if model.editCreation.DefaultName != "meeting 2" {
		t.Fatalf("collision name = %q", model.editCreation.DefaultName)
	}
	if err := model.beginTemplate(files.Template{Name: "another.md"}); err == nil {
		t.Fatal("incoming capture overwrote the active draft")
	}
	saved, err := files.LoadTemplates(model.store.TemplatesPath)
	if err != nil || len(saved) != 1 || saved[0].Content != "original" {
		t.Fatal("opening an import changed saved templates")
	}
	tester := ui.NewTester(model.view, minWidth, minHeight)
	if err := tester.Click(locales.Cancel()); err != nil {
		t.Fatal(err)
	}
	if err := model.beginTemplate(files.Template{Name: "meeting.md", Content: "new"}); err != nil {
		t.Fatal(err)
	}
	model.saveCreationEditor()
	saved, err = files.LoadTemplates(model.store.TemplatesPath)
	if err != nil || len(saved) != 2 || saved[1].Name != "meeting 2.md" || saved[0].Content != "original" {
		t.Fatalf("import did not preserve both templates: %+v, %v", saved, err)
	}
	model.openCreationEditor(creationRow{index: 0, template: true})
	model.duplicateTemplate()
	model.saveCreationEditor()
	if len(model.templates) != 3 || model.templates[2].Name != "meeting 3.md" || model.templates[2].Content != "original" {
		t.Fatal("duplicate failed to preserve content or avoid collision")
	}
}

func TestCaptureOfficeAndExtensionlessTemplate(t *testing.T) {
	model := readyApp(t)
	for _, name := range []string{"report.docx", "LICENSE", ".env"} {
		path := filepath.Join(t.TempDir(), name)
		data := []byte("sample text")
		if name == "report.docx" {
			var err error
			data, err = catalog.Document("docx")
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		template, err := files.TemplateFromFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := model.beginTemplate(template); err != nil {
			t.Fatal(err)
		}
		model.saveCreationEditor()
		if model.editCreationOpen {
			t.Fatal(model.editCreationNotice)
		}
		saved := model.templates[len(model.templates)-1]
		content, err := saved.Bytes()
		if err != nil || saved.Name != name || !slices.Equal(content, data) {
			t.Fatalf("%s lost its name or contents", name)
		}
	}
}
