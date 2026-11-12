package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/hy-shine/rightclick/locales"
)

func (cfg Config) FavoritePath(id string) (string, error) {
	for _, folder := range cfg.FavoriteFolders {
		if folder.ID != id {
			continue
		}
		if !folder.Enabled || !filepath.IsAbs(folder.Path) {
			return "", errors.New(locales.FavoriteUnavailable())
		}
		info, err := os.Stat(folder.Path)
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return "", errors.New(locales.FavoriteUnavailable())
		}
		return folder.Path, nil
	}
	return "", errors.New(locales.FavoriteUnavailable())
}
