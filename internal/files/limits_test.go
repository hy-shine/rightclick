package files

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestUnzipRefusesTooManyItems(t *testing.T) {
	old := maxUnzipEntries
	maxUnzipEntries = 2
	t.Cleanup(func() { maxUnzipEntries = old })
	dir := t.TempDir()
	archive := filepath.Join(dir, "many.zip")
	writeZip(t, archive, []zipEntry{
		{name: "a.txt", body: "a"},
		{name: "b.txt", body: "b"},
		{name: "c.txt", body: "c"},
	})
	_, _, err := Unzip(archive)
	if err == nil || err.Error() != locales.ArchiveTooManyItems() {
		t.Fatalf("Unzip error = %v, want the too-many-items message", err)
	}
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if exists(filepath.Join(dir, name)) {
			t.Errorf("%s was written although the archive was refused", name)
		}
	}
}

func TestUnzipRefusesTotalOverBudget(t *testing.T) {
	old := maxUnzipBytes
	maxUnzipBytes = 30
	t.Cleanup(func() { maxUnzipBytes = old })
	dir := t.TempDir()
	archive := filepath.Join(dir, "big.zip")
	writeZip(t, archive, []zipEntry{
		{name: "a.txt", body: strings.Repeat("a", 20)},
		{name: "b.txt", body: strings.Repeat("b", 20)},
	})
	_, _, err := Unzip(archive)
	if err == nil || err.Error() != locales.ArchiveTooLarge() {
		t.Fatalf("Unzip error = %v, want the too-large message", err)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if exists(filepath.Join(dir, name)) {
			t.Errorf("%s was written although the archive was refused", name)
		}
	}
}

func TestUnzipRefusesHighCompressionRatio(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bomb.zip")
	writeZip(t, archive, []zipEntry{{name: "zeros.bin", body: strings.Repeat("\x00", 10<<20)}})
	_, _, err := Unzip(archive)
	if err == nil || err.Error() != locales.ArchiveRatioTooHigh() {
		t.Fatalf("Unzip error = %v, want the ratio message", err)
	}
	if exists(filepath.Join(dir, "zeros.bin")) {
		t.Error("zeros.bin was written although the archive was refused")
	}
}

func TestValidateDocumentRefusesTotalOverBudget(t *testing.T) {
	var fileType config.FileType
	for _, ft := range config.Default().FileTypes {
		if ft.Extension == "docx" && ft.Template != "" {
			fileType = ft
			break
		}
	}
	if fileType.Extension != "docx" {
		t.Fatal("no docx template in the defaults")
	}
	content, err := FileTypeContent(fileType)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDocument("docx", content); err != nil {
		t.Fatalf("the default docx should validate: %v", err)
	}
	old := maxDocumentBytes
	maxDocumentBytes = 100
	t.Cleanup(func() { maxDocumentBytes = old })
	err = ValidateDocument("docx", content)
	if err == nil || err.Error() != locales.DocumentTooLarge() {
		t.Errorf("ValidateDocument error = %v, want the too-large message", err)
	}
}
