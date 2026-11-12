package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hy-shine/rightclick/internal/locales"
)

func TestProtectedFollowsLinksInFolders(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	mustMkdir(t, home)
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(base, alias); err != nil {
		t.Fatal(err)
	}
	if !Protected(filepath.Join(alias, "home"), home) {
		t.Error("an alias of the home folder was not protected")
	}
	if Protected(filepath.Join(alias, "home", "notes.txt"), home) {
		t.Error("a file inside the home folder was protected")
	}
}

func TestRemoveAllRefusesHomeThroughAlias(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	mustMkdir(t, home)
	writeFile(t, filepath.Join(home, "keep.txt"), "data")
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(base, alias); err != nil {
		t.Fatal(err)
	}
	removed, err := RemoveAll([]string{filepath.Join(alias, "home")}, home)
	if removed != 0 || err == nil {
		t.Errorf("RemoveAll = %d, %v; want 0 and a refusal", removed, err)
	}
	if !exists(filepath.Join(home, "keep.txt")) {
		t.Error("the home folder contents were deleted through an alias")
	}
}

func TestRemoveAllDeletesLinkNotItsTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	mustMkdir(t, target)
	writeFile(t, filepath.Join(target, "keep.txt"), "data")
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	removed, err := RemoveAll([]string{link}, filepath.Join(root, "home"))
	if removed != 1 || err != nil {
		t.Errorf("RemoveAll = %d, %v; want 1 and no error", removed, err)
	}
	if exists(link) {
		t.Error("the link was not removed")
	}
	if !exists(filepath.Join(target, "keep.txt")) {
		t.Error("removing the link deleted what it points to")
	}
}

func TestRemoveAllCountsItemInMissingFolderAsGone(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing", "item.txt")
	removed, err := RemoveAll([]string{missing}, "/Users/nobody-here")
	if removed != 1 || err != nil {
		t.Errorf("RemoveAll = %d, %v; want 1 and no error", removed, err)
	}
}

func TestCopyFolderIntoLinkToItselfFails(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "dir")
	sub := filepath.Join(src, "sub")
	mustMkdir(t, sub)
	dest := filepath.Join(root, "x")
	if err := os.Symlink(sub, dest); err != nil {
		t.Fatal(err)
	}
	_, err := Copy(src, dest)
	if err == nil || err.Error() != locales.FolderIntoItself() {
		t.Fatalf("Copy error = %v, want the folder-into-itself refusal", err)
	}
	if exists(filepath.Join(sub, "dir")) {
		t.Error("Copy created files before refusing")
	}
}

func TestMoveFolderIntoLinkToItselfFails(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "dir")
	sub := filepath.Join(src, "sub")
	mustMkdir(t, sub)
	dest := filepath.Join(root, "x")
	if err := os.Symlink(sub, dest); err != nil {
		t.Fatal(err)
	}
	_, err := Move(src, dest)
	if err == nil || err.Error() != locales.FolderIntoItself() {
		t.Fatalf("Move error = %v, want the folder-into-itself refusal", err)
	}
	if !exists(sub) {
		t.Error("the folder was moved despite the refusal")
	}
}

func TestUnzipRefusesToWriteThroughLink(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "outside")
	folder := filepath.Join(root, "downloads")
	mustMkdir(t, outside)
	mustMkdir(t, folder)
	if err := os.Symlink(outside, filepath.Join(folder, "link")); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(folder, "bundle.zip")
	writeZip(t, archive, []zipEntry{{name: "link/evil.txt", body: "written outside"}})
	_, _, err := Unzip(archive)
	if err == nil || err.Error() != locales.ArchiveThroughLink("link") {
		t.Fatalf("Unzip error = %v, want the through-link refusal", err)
	}
	if exists(filepath.Join(outside, "evil.txt")) {
		t.Error("Unzip wrote evil.txt outside the folder")
	}
}
