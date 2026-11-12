package settings

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"rightclick/config"
	"rightclick/files"
)

func TestRenderCreationReview(t *testing.T) {
	path := os.Getenv("RIGHTCLICK_REVIEW_IMAGE")
	if path == "" {
		t.Skip("render requested explicitly")
	}
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneTemplates)
	tester := ui.NewTester(model.view, 1140, 720)
	if os.Getenv("RIGHTCLICK_REVIEW_DIALOG") != "" {
		model.openCreationEditor(creationRow{index: 0})
		tester.Frame()
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, tester.Image()); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCreationListAndDialogShareEnabledState(t *testing.T) {
	store := testStore(t)
	model, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneTemplates)
	tester := ui.NewTester(model.view, mainWidth, mainHeight)
	if err := tester.Click("启用 Word 文档"); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.FileTypes[0].Enabled {
		t.Fatal("list switch did not save disabled state")
	}
	if err := tester.Click("编辑 Word 文档"); err != nil {
		t.Fatal(err)
	}
	if model.editCreation.Enabled {
		t.Fatal("dialog did not inherit disabled state")
	}
	if err := tester.Click("启用此文件类型"); err != nil {
		t.Fatal(err)
	}
	if !model.editCreation.Enabled {
		t.Fatalf("switch did not change draft; texts %q", tester.Texts())
	}
	if err := tester.Click("保存修改"); err != nil {
		t.Fatal(err)
	}
	saved, err = config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !saved.FileTypes[0].Enabled {
		t.Fatalf("dialog did not save enabled state; draft=%+v notice=%s open=%v", model.editCreation, model.editCreationNotice, model.editCreationOpen)
	}
	if model.editCreationOpen {
		t.Fatal("save did not close dialog")
	}
}

func TestCreationEditorCancelDoesNotChangeSavedContent(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneTemplates)
	tester := ui.NewTester(model.view, mainWidth, mainHeight)
	model.openCreationEditor(creationRow{index: 0})
	tester.Frame()
	model.editCreation.Label = "changed"
	if err := tester.Click("取消"); err != nil {
		t.Fatal(err)
	}
	if model.cfg.FileTypes[0].Label == "changed" {
		t.Fatal("cancel committed the draft")
	}
}

func TestCustomTemplateSwitchPreservesLocation(t *testing.T) {
	store := testStore(t)
	model, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.addTemplate(files.Template{Name: "meeting.md", Content: "# Meeting", Location: config.FileTypeMain})
	model.newMode = newModeTemplates
	model.selectPane(paneTemplates)
	tester := ui.NewTester(model.view, mainWidth, mainHeight)
	if err := tester.Click("启用 meeting"); err != nil {
		t.Fatal(err)
	}
	templates, err := files.LoadTemplates(filepath.Join(filepath.Dir(store.ConfigPath), "templates.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !templates[0].Disabled || templates[0].Location != config.FileTypeMain {
		t.Fatalf("disabled state lost template location: %+v", templates[0])
	}
}

func TestCreationFilteringIncludesDocumentsAndTemplates(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.newMode = newModeDocuments
	if got := len(model.filteredCreationRows()); got != 10 {
		t.Fatalf("document rows = %d, want 10", got)
	}
	model.creationSearch = "docx"
	if got := len(model.filteredCreationRows()); got != 2 {
		t.Fatalf("docx search rows = %d, want 2", got)
	}
	model.creationSearch = ""
	model.newMode = newModeCode
	if got := len(model.filteredCreationRows()); got != 14 {
		t.Fatalf("code rows = %d, want 14", got)
	}
}
