package settings

import (
	_ "embed"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

const (
	repositoryURL = "https://github.com/hy-shine/rightclick"
)

//go:embed assets/github.svg
var githubArtwork []byte

var iconGitHub = ui.MustParseSVG(githubArtwork)

// appVersion returns the version of the app bundle, or dev for an unpackaged
// development build.
func appVersion() string {
	if version := mygo.App.Version(); version != "" {
		return version
	}
	return "dev"
}

func (model *app) aboutSection(context *ui.Context) {
	theme := context.Theme()
	ui.Column(context).Gap(10).Children(func() {
		sectionTitle(context, locales.About())
		ledger(context, ui.Column(context)).Children(func() {
			ui.Row(context).Gap(14).Padding(14, 16).AlignItems(ui.Center).Children(func() {
				mark(context, 42)
				ui.Column(context).Grow(1).MinWidth(0).Gap(3).Children(func() {
					ui.Row(context).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Text(context, locales.AppName()).FontSize(sizeSection).FontWeight(600)
						pillBadge(context, locales.VersionLine(appVersion()), theme.TextMuted, true)
					})
					ui.Text(context, locales.AboutBlurb()).
						FontSize(sizeNote).LineHeight(1.45).TextColor(theme.TextMuted)
				})
				ui.Row(context).Gap(8).AlignItems(ui.Center).Shrink(0).Children(func() {
					ui.Row(context).Tooltip(locales.OpenGitHub()).Children(func() {
						if iconButton(context.Key("about-github"), iconGitHub, locales.OpenGitHub(), false) {
							model.report(locales.Github(), model.env.openURL(repositoryURL))
						}
					})
					if quietButton(context, nil, locales.CheckUpdate()) {
						model.env.checkForUpdates()
					}
				})
			})
		})
	})
}
