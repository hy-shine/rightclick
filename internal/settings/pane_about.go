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
	ui.Column(context).Gap(12).Margin(8, 0, 0).Children(func() {
		sectionTitle(context, locales.About())
		ui.Row(context).Gap(12).AlignItems(ui.Center).Children(func() {
			mark(context, 40)
			ui.Column(context).Grow(1).Gap(2).Children(func() {
				ui.Text(context, locales.AppName()).FontSize(sizeSection).FontWeight(600)
				ui.Text(context, locales.VersionLine(appVersion())).FontSize(sizeNote).TextColor(theme.TextMuted)
			})
			ui.Row(context).Gap(8).AlignItems(ui.Center).Children(func() {
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
		ui.Text(context, locales.AboutBlurb()).
			FontSize(sizeNote).LineHeight(1.5).TextColor(theme.TextMuted)
	})
}
