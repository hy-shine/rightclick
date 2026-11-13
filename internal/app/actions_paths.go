package app

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

var copyConfigPath = config.DefaultPath

var chooseCopyBaseFolder = func(start string) (string, error) {
	folders, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Title:       locales.CopyBaseTitle(),
		Message:     locales.CopyBaseHint(),
		ButtonLabel: locales.ChooseButton(),
		DefaultPath: start,
		Directory:   true,
	})
	if err != nil || len(folders) == 0 {
		return "", err
	}
	return folders[0], nil
}

func copyPathsFromBase(paths []string, cfg config.Config, format func([]string, string) (string, error)) error {
	start := cfg.CopyBaseFolder
	if info, err := os.Stat(start); err != nil || !info.IsDir() || !filepath.IsAbs(start) {
		start = ""
	}
	base, err := chooseCopyBaseFolder(start)
	if err != nil || base == "" {
		return err
	}
	if !filepath.IsAbs(base) {
		return errors.New(locales.ErrPathNotAbsolute(base))
	}
	info, err := os.Stat(base)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New(locales.CopyBaseNotFolder(base))
	}
	text, err := format(paths, base)
	if err != nil {
		return err
	}
	if err := copyText(text); err != nil {
		return err
	}
	if err := rememberCopyBase(base); err != nil {
		return errors.New(locales.CopyBaseSaveFailed(err.Error()))
	}
	return nil
}

func rememberCopyBase(base string) error {
	path, err := copyConfigPath()
	if err != nil {
		return err
	}
	latest, err := config.Load(path)
	if err != nil {
		return err
	}
	latest.CopyBaseFolder = base
	return config.Save(path, latest)
}
