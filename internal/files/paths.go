package files

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/hy-shine/rightclick/internal/locales"
)

func TerminalPaths(paths []string) string {
	quoted := make([]string, 0, len(paths))
	for _, path := range paths {
		quoted = append(quoted, "'"+strings.ReplaceAll(path, "'", "'\\''")+"'")
	}
	return strings.Join(quoted, " ")
}

func RelativePaths(paths []string, base string) (string, error) {
	relative, err := relativePaths(paths, base)
	if err != nil {
		return "", err
	}
	return strings.Join(relative, "\n"), nil
}

func MarkdownFileLinks(paths []string, base string) (string, error) {
	relative, err := relativePaths(paths, base)
	if err != nil {
		return "", err
	}
	labels := strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "`", "\\`", "*", "\\*", "_", "\\_", "<", "&lt;", ">", "&gt;", "&", "&amp;", "\n", "&#10;", "\r", "&#13;")
	destinations := strings.NewReplacer("(", "%28", ")", "%29", ":", "%3A", "&", "%26")
	links := make([]string, 0, len(paths))
	for index, path := range paths {
		target := &url.URL{Path: relative[index]}
		links = append(links, "["+labels.Replace(filepath.Base(path))+"]("+destinations.Replace(target.EscapedPath())+")")
	}
	return strings.Join(links, "\n"), nil
}

func relativePaths(paths []string, base string) ([]string, error) {
	if !filepath.IsAbs(base) {
		return nil, errors.New(locales.ErrPathNotAbsolute(base))
	}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return nil, errors.New(locales.ErrPathNotAbsolute(path))
		}
		relative, err := filepath.Rel(base, path)
		if err != nil {
			return nil, err
		}
		result = append(result, filepath.ToSlash(relative))
	}
	return result, nil
}
