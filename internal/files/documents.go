package files

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"github.com/hy-shine/rightclick/internal/locales"
	"io"
	"strings"

	"errors"
	"github.com/hy-shine/rightclick/internal/catalog"
	"github.com/hy-shine/rightclick/internal/config"
)

func FileTypeContent(fileType config.FileType) ([]byte, error) {
	if fileType.Template == "" {
		return []byte(fileType.Content), nil
	}
	if fileType.Extension != fileType.Template {
		return nil, errors.New(locales.DocumentFormatMismatch(fileType.Extension))
	}
	if fileType.TemplateData != "" {
		data, err := base64.StdEncoding.DecodeString(fileType.TemplateData)
		if err != nil {
			return nil, errors.New(locales.TemplateDataInvalid(err.Error()))
		}
		if err := ValidateDocument(fileType.Extension, data); err != nil {
			return nil, err
		}
		return data, nil
	}
	return catalog.Document(fileType.Template)
}

// maxDocumentBytes is the most decompressed XML an Office document may hold in all of
// its parts together. Each part has its own limit, but without a total, many parts
// just under that limit could still add up to a very large document. It is a
// variable so that tests can lower it.
var maxDocumentBytes int64 = 64 << 20

// errDocumentTooLarge stops the reading of a document once its budget is spent.
var errDocumentTooLarge = errors.New("office document too large")

// budgetReader reads from r until the budget left is spent. Every part of a document
// draws on the same budget.
type budgetReader struct {
	r    io.Reader
	left *int64
}

func (b *budgetReader) Read(p []byte) (int, error) {
	if *b.left <= 0 {
		return 0, errDocumentTooLarge
	}
	if int64(len(p)) > *b.left {
		p = p[:*b.left]
	}
	n, err := b.r.Read(p)
	*b.left -= int64(n)
	return n, err
}

func ValidateDocument(extension string, data []byte) error {
	if extension == "rtf" {
		if !bytes.HasPrefix(bytes.TrimSpace(data), []byte(`{\rtf`)) {
			return errors.New(locales.NotRTF())
		}
		return nil
	}
	required := map[string][]string{
		"docx": {"[Content_Types].xml", "_rels/.rels", "word/document.xml"},
		"xlsx": {"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml"},
		"pptx": {"[Content_Types].xml", "_rels/.rels", "ppt/presentation.xml"},
	}
	parts, known := required[extension]
	if !known {
		return errors.New(locales.DocumentFormatUnsupported(extension))
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return errors.New(locales.DocumentCorrupt(err.Error()))
	}
	found := make(map[string]bool)
	budget := maxDocumentBytes
	for _, entry := range archive.File {
		if !strings.HasSuffix(entry.Name, ".xml") && !strings.HasSuffix(entry.Name, ".rels") {
			continue
		}
		if entry.UncompressedSize64 > 16<<20 {
			return errors.New(locales.DocumentPartTooLarge(entry.Name))
		}
		reader, err := entry.Open()
		if err != nil {
			return err
		}
		decoder := xml.NewDecoder(&budgetReader{r: io.LimitReader(reader, 16<<20), left: &budget})
		for {
			_, err = decoder.Token()
			if err != nil {
				break
			}
		}
		closeErr := reader.Close()
		if errors.Is(err, errDocumentTooLarge) {
			return errors.New(locales.DocumentTooLarge())
		}
		if err != io.EOF {
			return errors.New(locales.DocumentPartInvalid(entry.Name, err.Error()))
		}
		if closeErr != nil {
			return closeErr
		}
		found[entry.Name] = true
	}
	for _, part := range parts {
		if !found[part] {
			return errors.New(locales.DocumentPartMissing(part))
		}
	}
	return nil
}
