package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveTemplatesIsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "support", "templates.json")
	if err := SaveTemplates(path, []Template{}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("templates file mode = %v, want 0600", got)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("templates folder mode = %v, want 0700", got)
	}
}
