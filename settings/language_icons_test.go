package settings

import (
	"bytes"
	"testing"

	"github.com/egoist/mygo/ui"

	"rightclick/internal/catalog"
)

func TestCodeLanguageIconsParse(t *testing.T) {
	fallback := catalog.Icon("txt")
	for _, extension := range []string{"go", "rs", "java", "c", "cpp", "sql", "ts", "sh"} {
		data := catalog.Icon(extension)
		if bytes.Equal(data, fallback) {
			t.Errorf("%s has no icon of its own", extension)
			continue
		}
		if _, err := ui.ParseSVG(data); err != nil {
			t.Errorf("%s icon does not parse: %v", extension, err)
		}
	}
}
