package settings

import (
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
)

func TestSettingsSaveKeepsBaseFolderFromAction(test *testing.T) {
	model := readyApp(test)
	latest, err := config.Load(model.store.ConfigPath)
	if err != nil {
		test.Fatal(err)
	}
	latest.CopyBaseFolder = test.TempDir()
	if err := config.Save(model.store.ConfigPath, latest); err != nil {
		test.Fatal(err)
	}
	model.cfg.OpenAfterCreate = true
	model.saveConfig()
	saved, err := config.Load(model.store.ConfigPath)
	if err != nil {
		test.Fatal(err)
	}
	if !saved.OpenAfterCreate || saved.CopyBaseFolder != latest.CopyBaseFolder {
		test.Fatalf("settings save lost the base folder: %+v", saved)
	}
}
