package files

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"rightclick/config"
	"rightclick/internal/catalog"
)

func FileTypeContent(fileType config.FileType) ([]byte, error) {
	if fileType.Template == "" {
		return []byte(fileType.Content), nil
	}
	if fileType.Extension != fileType.Template {
		return nil, fmt.Errorf("文档格式与模板不一致：%s", fileType.Extension)
	}
	if fileType.TemplateData != "" {
		data, err := base64.StdEncoding.DecodeString(fileType.TemplateData)
		if err != nil {
			return nil, fmt.Errorf("模板数据无效: %w", err)
		}
		if err := ValidateDocument(fileType.Extension, data); err != nil {
			return nil, err
		}
		return data, nil
	}
	return catalog.Document(fileType.Template)
}

func ValidateDocument(extension string, data []byte) error {
	if extension == "rtf" {
		if !bytes.HasPrefix(bytes.TrimSpace(data), []byte(`{\rtf`)) {
			return fmt.Errorf("不是有效的 RTF 文档")
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
		return fmt.Errorf("不支持的文档格式：%s", extension)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("文档文件损坏: %w", err)
	}
	found := make(map[string]bool)
	for _, entry := range archive.File {
		if !strings.HasSuffix(entry.Name, ".xml") && !strings.HasSuffix(entry.Name, ".rels") {
			continue
		}
		if entry.UncompressedSize64 > 16<<20 {
			return fmt.Errorf("文档组件过大：%s", entry.Name)
		}
		reader, err := entry.Open()
		if err != nil {
			return err
		}
		decoder := xml.NewDecoder(io.LimitReader(reader, 16<<20))
		for {
			_, err = decoder.Token()
			if err != nil {
				break
			}
		}
		closeErr := reader.Close()
		if err != io.EOF {
			return fmt.Errorf("文档组件无效 %s: %w", entry.Name, err)
		}
		if closeErr != nil {
			return closeErr
		}
		found[entry.Name] = true
	}
	for _, part := range parts {
		if !found[part] {
			return fmt.Errorf("文档缺少组件：%s", part)
		}
	}
	return nil
}
