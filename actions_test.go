package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rightclick/config"
	"rightclick/files"
)

func TestFileTypeFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileType config.FileType
		want     string
		wantErr  bool
	}{
		{
			name:     "normal type",
			fileType: config.FileType{DefaultName: "未命名", Extension: "txt"},
			want:     "未命名.txt",
		},
		{
			name:     "extension dot",
			fileType: config.FileType{DefaultName: "脚本", Extension: ".py"},
			want:     "脚本.py",
		},
		{
			name:     "name already has extension",
			fileType: config.FileType{DefaultName: "notes.md", Extension: "md"},
			want:     "notes.md",
		},
		{
			name:     "empty base uses default",
			fileType: config.FileType{DefaultName: "", Extension: "go"},
			want:     "未命名.go",
		},
		{
			name:     "path in base",
			fileType: config.FileType{DefaultName: "../x", Extension: "txt"},
			wantErr:  true,
		},
		{
			name:     "path in extension",
			fileType: config.FileType{DefaultName: "x", Extension: "a/b"},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fileTypeFileName(tt.fileType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("fileTypeFileName(%+v) error = %v, wantErr %v", tt.fileType, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("fileTypeFileName(%+v) = %q, want %q", tt.fileType, got, tt.want)
			}
		})
	}
}

func TestCreateFileFromTypeWritesContentAndNumbersNames(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	for i := range cfg.FileTypes {
		if cfg.FileTypes[i].ID == "json" {
			cfg.FileTypes[i].Content = "{\"ok\": true}\n"
		}
	}
	for range 2 {
		if err := createFileFromType(cfg, "json", []string{dir}); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "未命名.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "{\"ok\": true}\n"; got != want {
		t.Errorf("created content = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dir, "未命名 2.json")); err != nil {
		t.Errorf("second file was not numbered: %v", err)
	}
}

func TestCreateFileFromTypeRejectsHiddenType(t *testing.T) {
	cfg := config.Default()
	for index := range cfg.FileTypes {
		if cfg.FileTypes[index].ID == "text" {
			cfg.FileTypes[index].Enabled = false
		}
	}
	if err := createFileFromType(cfg, "text", []string{t.TempDir()}); err == nil {
		t.Fatal("hidden file type was created")
	}
}

func TestFindTemplate(t *testing.T) {
	templates := []files.Template{
		{Name: "周报.md", Content: "# 周报"},
		{Name: "会议.txt", Content: "议题"},
	}
	got, err := findTemplate(templates, "会议.txt")
	if err != nil {
		t.Fatalf("findTemplate(会议.txt) error = %v", err)
	}
	if got.Content != "议题" {
		t.Fatalf("findTemplate(会议.txt) content = %q, want %q", got.Content, "议题")
	}
	if _, err := findTemplate(templates, "不存在.md"); err == nil {
		t.Fatal("findTemplate with a missing name returned no error")
	}
	if _, err := findTemplate([]files.Template{{Name: "../x.md"}}, "../x.md"); err == nil {
		t.Fatal("findTemplate accepted a template name with a folder part")
	}
	if _, err := findTemplate(nil, ""); err == nil {
		t.Fatal("findTemplate with an empty name returned no error")
	}
}

func TestArchiveName(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  string
	}{
		{name: "one file", paths: []string{"/Users/a/报告.pdf"}, want: "报告.zip"},
		{name: "one folder", paths: []string{"/Users/a/photos"}, want: "photos.zip"},
		{name: "several items", paths: []string{"/Users/a/x.txt", "/Users/a/y"}, want: "归档.zip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := archiveName(tt.paths); got != tt.want {
				t.Fatalf("archiveName(%q) = %q, want %q", tt.paths, got, tt.want)
			}
		})
	}
}

func TestHashAlgorithm(t *testing.T) {
	tests := []struct {
		name    string
		want    files.Algorithm
		wantErr bool
	}{
		{name: "", want: files.SHA256},
		{name: "SHA-256", want: files.SHA256},
		{name: "SHA-1", want: files.SHA1},
		{name: "MD5", want: files.MD5},
		{name: "sha256", wantErr: true},
	}
	for _, tt := range tests {
		got, err := hashAlgorithm(tt.name)
		if (err != nil) != tt.wantErr {
			t.Fatalf("hashAlgorithm(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		if got != tt.want {
			t.Fatalf("hashAlgorithm(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestCopyTextJoin(t *testing.T) {
	paths := []string{"/Users/a/x.txt", "/Users/a/b c"}
	if got, want := joinLines(paths), "/Users/a/x.txt\n/Users/a/b c"; got != want {
		t.Fatalf("copy-path text = %q, want %q", got, want)
	}
	if got, want := joinLines(baseNames(paths)), "x.txt\nb c"; got != want {
		t.Fatalf("copy-name text = %q, want %q", got, want)
	}
	if got := baseNames(paths); !reflect.DeepEqual(got, []string{"x.txt", "b c"}) {
		t.Fatalf("baseNames = %q", got)
	}
}

func TestRequireOne(t *testing.T) {
	if _, err := requireOne(nil); err == nil {
		t.Fatal("requireOne(nil) returned no error")
	}
	if _, err := requireOne([]string{"/a", "/b"}); err == nil {
		t.Fatal("requireOne with two paths returned no error")
	}
	got, err := requireOne([]string{"/a"})
	if err != nil || got != "/a" {
		t.Fatalf("requireOne(/a) = %q, %v", got, err)
	}
}

func TestTransferOp(t *testing.T) {
	for _, action := range []string{"copy-to", "move-to"} {
		if _, err := transferOp(action); err != nil {
			t.Fatalf("transferOp(%q) error = %v", action, err)
		}
	}
	if _, err := transferOp("zip"); err == nil {
		t.Fatal("transferOp(zip) returned no error")
	}
}

func TestCompareSummary(t *testing.T) {
	title, body := compareSummary(files.SHA256, false, "/a/x.txt", "h1", "/b/y.txt", "h2")
	if title != "SHA-256 比较：不同" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{"x.txt", "h1", "y.txt", "h2"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body %q does not contain %q", body, want)
		}
	}
	title, _ = compareSummary(files.MD5, true, "/a/x", "h", "/b/y", "h")
	if title != "MD5 比较：相同" {
		t.Fatalf("title = %q", title)
	}
}
