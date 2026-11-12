package settings

import (
	"github.com/hy-shine/rightclick/internal/locales"
)

// setupStep is the first thing still missing before the Finder menu can show,
// in the order the user should fix them. stepReady means it can show.
type setupStep int

const (
	stepChecking  setupStep = iota // the Finder extension state is not known yet
	stepExtension                  // the Finder extension is off
	stepItem                       // no menu entry is switched on
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
	case enabledCount(m.cfg.MenuItems) == 0:
		return stepItem
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
	case stepItem:
		return locales.StatusNoItems()
	case stepReady:
		return locales.StatusReady()
	}
	return locales.StatusChecking()
}
