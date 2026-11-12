package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"rightclick/config"
	"rightclick/files"
)

func TestCreateFromTemplateWritesContentAndRefusesDisabled(t *testing.T) {
	templatesPath := filepath.Join(t.TempDir(), "templates.json")
	if err := files.SaveTemplates(templatesPath, []files.Template{
		{Name: "周报.md", Content: "# 本周\n"},
		{Name: "停用.md", Content: "x", Disabled: true},
	}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cfg := config.Default()

	if err := createFromTemplate(templatesPath, "周报.md", []string{dir}, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "周报.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "# 本周\n"; got != want {
		t.Errorf("created content = %q, want %q", got, want)
	}
	if err := createFromTemplate(templatesPath, "停用.md", []string{dir}, cfg); err == nil {
		t.Error("disabled template was created")
	}
	if err := createFromTemplate(templatesPath, "不存在.md", []string{dir}, cfg); err == nil {
		t.Error("missing template was created")
	}
	if err := createFromTemplate(templatesPath, "周报.md", []string{dir, dir}, cfg); err == nil {
		t.Error("two folders were accepted as the destination")
	}
	if _, err := os.Stat(filepath.Join(dir, "停用.md")); err == nil {
		t.Error("disabled template left a file behind")
	}
}

func TestTransferToTriesEveryItem(t *testing.T) {
	root := t.TempDir()
	valid := filepath.Join(root, "valid.txt")
	if err := os.WriteFile(valid, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(root, "missing.txt")
	dst := t.TempDir()

	err := transferTo(files.Copy, []string{missing, valid}, dst)
	if err == nil || !strings.Contains(err.Error(), "missing.txt") {
		t.Fatalf("transferTo error = %v, want an error that names missing.txt", err)
	}
	if _, statErr := os.Stat(filepath.Join(dst, "valid.txt")); statErr != nil {
		t.Errorf("valid item was not copied after an earlier failure: %v", statErr)
	}
}

func TestZipItemsWritesArchiveBesideFirstItem(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(file, []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := zipItems([]string{file}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "notes.zip")); err != nil {
		t.Errorf("archive was not written beside the item: %v", err)
	}
}

func TestUnzipItemsKeepsGoodArchivesWhenOneFails(t *testing.T) {
	root := t.TempDir()
	item := filepath.Join(root, "src", "a.txt")
	if err := os.MkdirAll(filepath.Dir(item), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(item, []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	good, err := files.Zip([]string{item}, root, "good.zip")
	if err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(root, "bad.zip")
	if err := os.WriteFile(bad, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}

	written, err := unzipItems([]string{bad, good})
	if written != 1 {
		t.Errorf("written = %d, want 1", written)
	}
	if err == nil || !strings.Contains(err.Error(), "bad.zip") {
		t.Fatalf("unzipItems error = %v, want an error that names bad.zip", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "a.txt")); statErr != nil {
		t.Errorf("good archive was not extracted: %v", statErr)
	}
}

func TestDigestOfSelectedFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "abc.txt")
	if err := os.WriteFile(file, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}

	algo, path, digest, err := digestOf("", []string{file})
	if err != nil {
		t.Fatal(err)
	}
	if algo != files.SHA256 || path != file {
		t.Errorf("digestOf = %q, %q; want SHA-256 and %q", algo, path, file)
	}
	if want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"; digest != want {
		t.Errorf("SHA-256 digest = %s, want %s", digest, want)
	}
	if _, _, digest, err := digestOf("MD5", []string{file}); err != nil {
		t.Fatal(err)
	} else if want := "900150983cd24fb0d6963f7d28e17f72"; digest != want {
		t.Errorf("MD5 digest = %s, want %s", digest, want)
	}
	if _, _, _, err := digestOf("MD5", []string{file, file}); err == nil {
		t.Error("two files were accepted for a digest")
	}
}

func TestCompareFilesReportsVerdict(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	c := filepath.Join(dir, "c.txt")
	for path, content := range map[string]string{a: "same", b: "same", c: "other"} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	title, body, err := compareFiles("", []string{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if title != "SHA-256 比较：相同" {
		t.Errorf("title = %q, want the same verdict", title)
	}
	if !strings.Contains(body, "a.txt") || !strings.Contains(body, "b.txt") {
		t.Errorf("body %q does not name both files", body)
	}
	if title, _, err = compareFiles("", []string{a, c}); err != nil {
		t.Fatal(err)
	}
	if title != "SHA-256 比较：不同" {
		t.Errorf("title = %q, want the different verdict", title)
	}
	if _, _, err := compareFiles("", []string{a}); err == nil {
		t.Error("one file was accepted for a comparison")
	}
}

func TestOpenEachWithTriesEveryPath(t *testing.T) {
	var opened []string
	open := func(app, path string) error {
		if app != "TextEdit" {
			t.Errorf("app = %q, want TextEdit", app)
		}
		opened = append(opened, path)
		if strings.HasSuffix(path, "bad.txt") {
			return errors.New("no such app")
		}
		return nil
	}
	paths := []string{"/tmp/bad.txt", "/tmp/ok.txt"}

	err := openEachWith(open, "TextEdit", paths)
	if !reflect.DeepEqual(opened, paths) {
		t.Errorf("opened = %q, want %q", opened, paths)
	}
	if err == nil || !strings.Contains(err.Error(), "bad.txt") {
		t.Errorf("openEachWith error = %v, want an error that names bad.txt", err)
	}
}

func TestRunActionRejectsUnknownAction(t *testing.T) {
	err := runAction("bogus", nil, []string{"/tmp"}, config.Default())
	if err == nil || !strings.Contains(err.Error(), "不认识的操作") {
		t.Fatalf("runAction(bogus) error = %v, want an unknown-action error", err)
	}
}

// TestMenuActionsMatchFinderExtension keeps the action names in step with the
// Finder menu. Every name that the menu lists must be handled, and every handled
// name must be listed, except the format version 1 links that older menus used.
func TestMenuActionsMatchFinderExtension(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("extension", "RightClickFinderSync", "MenuConfig.swift"))
	if err != nil {
		t.Fatal(err)
	}
	groups := regexp.MustCompile(`(?s)static let groups: \[\[String\]\] = \[(.*?)\n    \]`).FindSubmatch(source)
	if groups == nil {
		t.Fatal("MenuActions.groups not found in MenuConfig.swift")
	}
	inMenu := map[string]bool{}
	for _, m := range regexp.MustCompile(`"([a-z-]+)"`).FindAllSubmatch(groups[1], -1) {
		inMenu[string(m[1])] = true
	}
	if len(inMenu) == 0 {
		t.Fatal("MenuActions.groups lists no actions")
	}
	handled := actionHandlers()
	for name := range inMenu {
		if _, ok := handled[name]; !ok {
			t.Errorf("Finder menu lists %q, which the app does not handle", name)
		}
	}
	legacy := map[string]bool{"new-text": true, "new-markdown": true}
	for name := range handled {
		if !legacy[name] && !inMenu[name] {
			t.Errorf("app handles %q, which the Finder menu does not list", name)
		}
	}
}
