package settings

import (
	"rightclick/locales"
	"slices"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
)

// setupStep is the first thing still missing before the Finder menu can show,
// in the order the user should fix them. stepReady means it can show.
type setupStep int

const (
	stepChecking  setupStep = iota // the Finder extension state is not known yet
	stepExtension                  // the Finder extension is off
	stepFolder                     // the folder whitelist is empty
	stepItem                       // no menu entry is switched on
	stepRestart                    // a folder changed since Finder last started
	stepReady                      // the menu can show in Finder
)

// setupStep returns the first missing step. Finder shows the menu only when all
// of them hold, so the sidebar, the menu page and the diagnostics all read this.
func (m *app) setupStep() setupStep {
	switch {
	case m.finder == finderChecking:
		return stepChecking
	case m.finder == finderOff:
		return stepExtension
	case len(m.cfg.WatchFolders) == 0:
		return stepFolder
	case enabledCount(m.cfg.MenuItems) == 0:
		return stepItem
	case m.foldersChanged:
		return stepRestart
	}
	return stepReady
}

// blocked reports whether a step that keeps the menu from showing in Finder is
// missing. A check still running is not a blocker.
func (s setupStep) blocked() bool {
	return s != stepChecking && s != stepReady
}

// summary is the short conclusion of a step, for the sidebar.
func (s setupStep) summary() string {
	switch s {
	case stepExtension:
		return locales.StatusNeedsExtension()
	case stepFolder:
		return locales.StatusWaitingFolders()
	case stepItem:
		return locales.StatusNoItems()
	case stepRestart:
		return locales.StatusNeedsRestart()
	case stepReady:
		return locales.StatusReady()
	}
	return locales.StatusChecking()
}

// addFolderActions offers the ways to add to the folder whitelist: the picker,
// and with offerDefaults, the Desktop, Documents and Downloads folders at once.
func (m *app) addFolderActions(c *ui.Context, offerDefaults bool) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		if primaryButton(c, iconPlus, locales.AddFolder()) {
			m.pickFolder()
		}
		if offerDefaults && quietButton(c, nil, locales.UseDefaultFolders()) {
			m.useDefaultFolders()
		}
	})
}

// useDefaultFolders adds the folders that config.Default names to the whitelist.
// Folders that are already in it stay as they are.
func (m *app) useDefaultFolders() {
	added := false
	for _, dir := range config.Default().WatchFolders {
		if !slices.Contains(m.cfg.WatchFolders, dir) {
			m.cfg.WatchFolders = append(m.cfg.WatchFolders, dir)
			added = true
		}
	}
	if added {
		m.foldersChanged = true
	}
	m.saveConfig()
}
