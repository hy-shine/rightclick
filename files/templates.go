package files

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/hy-shine/rightclick/locales"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/hy-shine/rightclick/internal/catalog"
)

// Template is a text that a new file starts with, such as a Markdown outline.
type Template struct {
	Label    string `json:"label,omitempty"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	Data     string `json:"data,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Location string `json:"location,omitempty"`
}

// DefaultTemplatesPath is the file the app keeps its templates in.
func DefaultTemplatesPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "RightClick", "templates.json"), nil
}

// LoadTemplates reads the templates at path. A missing file means there are none.
func LoadTemplates(path string) ([]Template, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var templates []Template
	if err := json.Unmarshal(data, &templates); err != nil {
		return nil, errors.New(locales.TemplateReadFailed(err.Error()))
	}
	return templates, nil
}

// SaveTemplates writes the templates to path. The previous file stays if the
// write fails.
func SaveTemplates(path string, templates []Template) error {
	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeAtomic(path, 0o644, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// TemplateFromFile makes a template from a text file, named after the file.
// Binary files are refused.
func TemplateFromFile(path string) (Template, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Template{}, err
	}
	if !info.Mode().IsRegular() {
		return Template{}, errors.New(locales.TemplateRegularFileOnly())
	}
	if info.Size() > 16<<20 {
		return Template{}, errors.New(locales.TemplateTooLarge())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Template{}, err
	}
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if catalog.IsDocument(extension) {
		if err := ValidateDocument(extension, data); err != nil {
			return Template{}, err
		}
		return Template{Name: filepath.Base(path), Data: base64.StdEncoding.EncodeToString(data)}, nil
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return Template{}, errors.New(locales.TemplateNotText())
	}
	return Template{Name: filepath.Base(path), Content: string(data)}, nil
}

func (template Template) Bytes() ([]byte, error) {
	if template.Data != "" {
		return base64.StdEncoding.DecodeString(template.Data)
	}
	return []byte(template.Content), nil
}
