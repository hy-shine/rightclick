package app

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTransferItemsCopiesOrMovesIntoTheChosenFolder runs copy-to and move-to with the
// folder the picker returns, and checks where the item ends up.
func TestTransferItemsCopiesOrMovesIntoTheChosenFolder(t *testing.T) {
	oldChoose := chooseFolder
	t.Cleanup(func() { chooseFolder = oldChoose })
	for _, action := range []string{"copy-to", "move-to"} {
		source := filepath.Join(t.TempDir(), "note.txt")
		if err := os.WriteFile(source, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		dest := t.TempDir()
		chooseFolder = func(string) (string, error) { return dest, nil }
		if err := transferItems(action, []string{source}); err != nil {
			t.Fatalf("%s error = %v", action, err)
		}
		if _, err := os.Stat(filepath.Join(dest, "note.txt")); err != nil {
			t.Errorf("%s did not put the item in the chosen folder: %v", action, err)
		}
		_, err := os.Stat(source)
		if action == "copy-to" && err != nil {
			t.Errorf("copy-to removed the source: %v", err)
		}
		if action == "move-to" && !os.IsNotExist(err) {
			t.Errorf("move-to kept the source: %v", err)
		}
	}
}

// TestTransferItemsDoesNothingWhenTheDialogIsCancelled leaves the item where it is.
func TestTransferItemsDoesNothingWhenTheDialogIsCancelled(t *testing.T) {
	oldChoose := chooseFolder
	t.Cleanup(func() { chooseFolder = oldChoose })
	chooseFolder = func(string) (string, error) { return "", nil }
	dir := t.TempDir()
	source := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(source, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"copy-to", "move-to"} {
		if err := transferItems(action, []string{source}); err != nil {
			t.Fatalf("%s error = %v", action, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "note.txt" {
		t.Errorf("a cancelled transfer changed the folder: %d entries", len(entries))
	}
}
