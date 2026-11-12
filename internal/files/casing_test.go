package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hy-shine/rightclick/internal/locales"
)

// skipIfCaseSensitive skips a test that needs a volume where a capitalised spelling
// names the same folder. APFS is case-insensitive by default, but not every volume is.
// It returns the capitalised spelling of path.
func skipIfCaseSensitive(t *testing.T, path string) string {
	t.Helper()
	upper := strings.ToUpper(path)
	if _, err := os.Stat(upper); err != nil {
		t.Skip("the volume is case-sensitive, so the capitalised spelling is another folder")
	}
	return upper
}

func TestProtectedIgnoresCaseOnCaseInsensitiveVolume(t *testing.T) {
	home := t.TempDir()
	upper := skipIfCaseSensitive(t, home)
	if !Protected(upper, home) {
		t.Errorf("Protected(%q) = false, want true: it is the home folder spelled in capitals", upper)
	}
	if Protected(filepath.Join(upper, "notes.txt"), home) {
		t.Error("a file inside the home folder was protected")
	}
}

func TestProtectedIgnoresCaseForVolumesFolder(t *testing.T) {
	home := t.TempDir()
	skipIfCaseSensitive(t, "/Volumes")
	if !Protected("/VOLUMES/Data", home) {
		t.Error("a volume's top folder spelled in capitals was not protected")
	}
}

func TestRemoveAllRefusesHomeSpelledInCapitals(t *testing.T) {
	home := t.TempDir()
	upper := skipIfCaseSensitive(t, home)
	writeFile(t, filepath.Join(home, "keep.txt"), "data")
	removed, err := RemoveAll([]string{upper}, home)
	if removed != 0 || err == nil {
		t.Errorf("RemoveAll = %d, %v; want 0 and a refusal", removed, err)
	}
	if !exists(filepath.Join(home, "keep.txt")) {
		t.Error("the home folder was deleted through a capitalised path")
	}
}

func TestCopyFolderIntoItselfIgnoresCase(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "dir")
	sub := filepath.Join(src, "sub")
	mustMkdir(t, sub)
	upper := skipIfCaseSensitive(t, sub)
	_, err := Copy(src, upper)
	if err == nil || err.Error() != locales.FolderIntoItself() {
		t.Fatalf("Copy error = %v, want the folder-into-itself refusal", err)
	}
	if exists(filepath.Join(sub, "dir")) {
		t.Error("Copy created a folder inside itself before refusing")
	}
}

func TestMoveFolderIntoItselfIgnoresCase(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "dir")
	sub := filepath.Join(src, "sub")
	mustMkdir(t, sub)
	upper := skipIfCaseSensitive(t, sub)
	_, err := Move(src, upper)
	if err == nil || err.Error() != locales.FolderIntoItself() {
		t.Fatalf("Move error = %v, want the folder-into-itself refusal", err)
	}
	if !exists(sub) {
		t.Error("the folder was moved despite the refusal")
	}
}

func TestMoveIntoSameFolderIgnoresCase(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	writeFile(t, file, "data")
	upper := skipIfCaseSensitive(t, dir)
	_, err := Move(file, upper)
	if err == nil || err.Error() != locales.AlreadyInFolder() {
		t.Fatalf("Move error = %v, want the already-in-folder refusal", err)
	}
	if !exists(file) {
		t.Error("the file was renamed despite the refusal")
	}
}
