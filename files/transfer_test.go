package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFileTakesFreeNameWhenTargetExists(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src", "a.txt")
	writeFile(t, src, "new")
	dst := filepath.Join(root, "dst")
	writeFile(t, filepath.Join(dst, "a.txt"), "old")

	got, err := Copy(src, dst)
	if err != nil {
		t.Fatalf("Copy(%q, %q): %v", src, dst, err)
	}
	if want := filepath.Join(dst, "a 2.txt"); got != want {
		t.Errorf("Copy = %q, want %q", got, want)
	}
	if content := readFile(t, got); content != "new" {
		t.Errorf("copied content = %q, want %q", content, "new")
	}
	if content := readFile(t, src); content != "new" {
		t.Errorf("original content = %q, want %q", content, "new")
	}
	if content := readFile(t, filepath.Join(dst, "a.txt")); content != "old" {
		t.Errorf("existing a.txt content = %q, want %q", content, "old")
	}
}

func TestCopyFolderIsRecursive(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "proj")
	writeFile(t, filepath.Join(src, "top.txt"), "top")
	writeFile(t, filepath.Join(src, "sub", "deep", "nested.txt"), "nested")
	dst := filepath.Join(root, "out")
	mustMkdir(t, dst)

	got, err := Copy(src, dst)
	if err != nil {
		t.Fatalf("Copy(%q, %q): %v", src, dst, err)
	}
	if want := filepath.Join(dst, "proj"); got != want {
		t.Errorf("Copy = %q, want %q", got, want)
	}
	if content := readFile(t, filepath.Join(got, "top.txt")); content != "top" {
		t.Errorf("top.txt content = %q, want %q", content, "top")
	}
	if content := readFile(t, filepath.Join(got, "sub", "deep", "nested.txt")); content != "nested" {
		t.Errorf("nested.txt content = %q, want %q", content, "nested")
	}
}

func TestCopyFolderIntoItselfFails(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "proj")
	writeFile(t, filepath.Join(src, "a.txt"), "a")
	sub := filepath.Join(src, "sub")
	mustMkdir(t, sub)

	tests := []struct {
		name string
		dst  string
	}{
		{"into itself", src},
		{"into its subfolder", sub},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Copy(src, tt.dst); err == nil {
				t.Errorf("Copy(%q, %q) error = nil, want error", src, tt.dst)
			}
			// Nothing should have been created inside the source folder.
			if exists(filepath.Join(tt.dst, "proj")) {
				t.Errorf("Copy left %q behind", filepath.Join(tt.dst, "proj"))
			}
		})
	}
}

func TestMoveRemovesSourceAndTakesFreeName(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src", "a.txt")
	writeFile(t, src, "moved")
	dst := filepath.Join(root, "dst")
	writeFile(t, filepath.Join(dst, "a.txt"), "old")

	got, err := Move(src, dst)
	if err != nil {
		t.Fatalf("Move(%q, %q): %v", src, dst, err)
	}
	if want := filepath.Join(dst, "a 2.txt"); got != want {
		t.Errorf("Move = %q, want %q", got, want)
	}
	if exists(src) {
		t.Errorf("source %q still exists after Move", src)
	}
	if content := readFile(t, got); content != "moved" {
		t.Errorf("moved content = %q, want %q", content, "moved")
	}
	if content := readFile(t, filepath.Join(dst, "a.txt")); content != "old" {
		t.Errorf("existing a.txt content = %q, want %q", content, "old")
	}
}

func TestMoveFolderTakesFreeNameWhenTargetExists(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src", "proj")
	writeFile(t, filepath.Join(src, "a.txt"), "moved")
	dst := filepath.Join(root, "dst")
	writeFile(t, filepath.Join(dst, "proj", "old.txt"), "old")

	got, err := Move(src, dst)
	if err != nil {
		t.Fatalf("Move(%q, %q): %v", src, dst, err)
	}
	if want := filepath.Join(dst, "proj 2"); got != want {
		t.Errorf("Move = %q, want %q", got, want)
	}
	if content := readFile(t, filepath.Join(got, "a.txt")); content != "moved" {
		t.Errorf("moved a.txt content = %q, want %q", content, "moved")
	}
	if content := readFile(t, filepath.Join(dst, "proj", "old.txt")); content != "old" {
		t.Errorf("existing proj/old.txt content = %q, want %q", content, "old")
	}
	if exists(src) {
		t.Errorf("source %q still exists after Move", src)
	}
}

func TestMoveLeavesNoPlaceholderBehind(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src", "a.txt")
	writeFile(t, src, "moved")
	dst := filepath.Join(root, "dst")
	mustMkdir(t, dst)

	if _, err := Move(src, dst); err != nil {
		t.Fatalf("Move(%q, %q): %v", src, dst, err)
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dst, err)
	}
	if len(entries) != 1 || entries[0].Name() != "a.txt" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("destination entries = %q, want only a.txt", names)
	}
}

func TestMoveIntoSameFolderIsRefused(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "a.txt")
	writeFile(t, src, "x")
	if _, err := Move(src, root); err == nil {
		t.Fatal("Move into the same folder error = nil, want error")
	}
	if content := readFile(t, src); content != "x" {
		t.Errorf("source content = %q after refused move, want %q", content, "x")
	}
	if exists(filepath.Join(root, "a 2.txt")) {
		t.Error("refused move left a numbered copy behind")
	}
}

func TestCopySymlinkKeepsLinkTarget(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "target.txt"), "target")
	link := filepath.Join(root, "link")
	if err := os.Symlink("target.txt", link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}
	dst := filepath.Join(root, "dst")
	mustMkdir(t, dst)

	got, err := Copy(link, dst)
	if err != nil {
		t.Fatalf("Copy(%q, %q): %v", link, dst, err)
	}
	info, err := os.Lstat(got)
	if err != nil {
		t.Fatalf("Lstat %s: %v", got, err)
	}
	if info.Mode()&fs.ModeSymlink == 0 {
		t.Errorf("%s is not a symbolic link, mode %v", got, info.Mode())
	}
	target, err := os.Readlink(got)
	if err != nil {
		t.Fatalf("Readlink %s: %v", got, err)
	}
	if target != "target.txt" {
		t.Errorf("link target = %q, want %q", target, "target.txt")
	}
}
