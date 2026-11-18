package settings

import (
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/actionicons"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func (model *app) defaultsPane(context *ui.Context) {
	paneHeader(context, locales.PaneDefaults(), locales.DefaultsIntro())
	model.showNotice(context)

	ui.Column(context).Gap(10).Children(func() {
		sectionHeader(context, actionicons.Icon("hash"), locales.HashSettings(), "", "")
		ledger(context, ui.Column(context)).Children(func() {
			settingRow(context, nil, ui.Color{}, locales.DefaultHashAlgorithm(), locales.DefaultHashAlgorithmHint(), func() {
				locationSelect(context, &model.cfg.DefaultHashAlgorithm, "default-hash", locales.DefaultHashAlgorithm(), config.HashAlgorithms, config.HashAlgorithms, true, model.saveConfig)
			}).Padding(10, 16)
		})
	})

	ui.Column(context).Gap(10).Children(func() {
		sectionHeader(context, actionicons.Icon("zip"), locales.ZipSettings(), "", "")
		ledger(context, ui.Column(context)).Children(func() {
			settingRow(context, nil, ui.Color{}, locales.ZipExclude(".DS_Store"), locales.ZipDSStoreHint(), func() {
				ui.Switch(context, &model.cfg.ZipExclusions.DSStore).Label(locales.ZipExcludeToggle(".DS_Store")).OnChange(model.saveConfig)
			}).Padding(10, 16)
			settingRow(context, nil, ui.Color{}, locales.ZipExclude(".git"), locales.ZipGitHint(), func() {
				ui.Switch(context, &model.cfg.ZipExclusions.Git).Label(locales.ZipExcludeToggle(".git")).OnChange(model.saveConfig)
			}).Padding(10, 16)
			settingRow(context, nil, ui.Color{}, locales.ZipExclude("node_modules"), locales.ZipNodeModulesHint(), func() {
				ui.Switch(context, &model.cfg.ZipExclusions.NodeModules).Label(locales.ZipExcludeToggle("node_modules")).OnChange(model.saveConfig)
			}).Padding(10, 16)
		})
	})
}
