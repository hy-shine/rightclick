// Package settings is the main window of the app. Each change is saved at
// once to the files that the app and the Finder extension read.
package settings

import (
	"errors"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/hy-shine/rightclick/internal/locales"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/files"
)

// version is the app version that the general pane shows.
const version = "0.1.0"

// Pane IDs. Each pane is one section of the main window, chosen in the sidebar.
const (
	paneMenu        = "menu"
	paneTemplates   = "templates"
	paneFolders     = "folders"
	paneOpeners     = "openers"
	paneDiagnostics = "diagnostics"
	paneGeneral     = "general"
)

// paneTab is one pane of the main window: its ID and its name in the sidebar.
type paneTab struct{ id, label string }

// paneTabs lists the panes in the order of the sidebar. The labels are read on
// each call so that they follow the current language.
func paneTabs() []paneTab {
	return []paneTab{
		{paneMenu, locales.PaneMenu()},
		{paneTemplates, locales.PaneNew()},
		{paneFolders, locales.PaneFolders()},
		{paneOpeners, locales.PaneOpeners()},
		{paneDiagnostics, locales.PaneDiagnostics()},
		{paneGeneral, locales.PaneGeneral()},
	}
}

// Store is where the settings window reads and writes its data.
type Store struct {
	ConfigPath    string
	TemplatesPath string
}

// DefaultStore returns the store at the default paths of the config and
// template files.
func DefaultStore() (Store, error) {
	configPath, err := config.DefaultPath()
	if err != nil {
		return Store{}, err
	}
	templatesPath, err := files.DefaultTemplatesPath()
	if err != nil {
		return Store{}, err
	}
	return Store{ConfigPath: configPath, TemplatesPath: templatesPath}, nil
}

// app is the state of one settings window. Its methods run on the main
// thread. Goroutines started for dialogs and system checks publish their
// results with win.Update, which also runs them on the main thread.
type app struct {
	store Store
	env   env
	win   *mygo.Window // set once NewWindow returns

	cfg       config.Config
	templates []files.Template

	tab        int    // index of the chosen pane in paneTabs; the sidebar list selects it
	previewCtx int    // index into previewContexts, where the menu preview is shown
	notice     string // error of the last action, shown below the pane
	newMode    int    // index of the Files or Templates view in the New pane

	navList          ui.ListState
	menuLists        []ui.ListState // one per group of the Finder menu, in menuGroups order
	menuSearch       string
	menuPreviewOpen  bool
	confirmRestore   bool
	menuNameOpen     bool
	menuNameID       string
	menuName         string
	favoriteList     ui.ListState
	favoriteEditID   string
	favoriteName     string
	favoriteEditOpen bool

	fileTypeList         ui.ListState
	fileTypeSel          int
	creationSel          int
	creationRows         []creationRow
	creationSearch       string
	editCreationOpen     bool
	editCreationTemplate bool
	editCreationNew      bool
	confirmDeleteType    bool
	editCreationIndex    int
	editCreation         config.FileType
	editCreationNotice   string

	tplList          ui.ListState
	tplSel           int // index of the selected template, -1 for none
	confirmDeleteTpl bool

	programSearch string
	editorIcons   map[string]*ui.Bitmap // icon of each editor app, read once; nil when it has none

	finder         finderState
	finderSeq      int // counts the checks, so that an older result is dropped
	confirmRestart bool

	openAtLogin bool
}

// newApp loads the settings in store. The system calls come from e.
func newApp(store Store, e env) (*app, error) {
	cfg, err := config.Load(store.ConfigPath)
	if err != nil {
		return nil, err
	}
	templates, err := files.LoadTemplates(store.TemplatesPath)
	if err != nil {
		return nil, err
	}
	m := &app{
		store:       store,
		env:         e,
		cfg:         cfg,
		templates:   templates,
		tab:         0, // paneMenu
		previewCtx:  ctxFile,
		fileTypeSel: 0,
		tplSel:      -1,
		openAtLogin: e.openAtLogin(),
	}
	if len(templates) > 0 {
		m.tplSel = 0
	}
	m.navList.Selected = &m.tab
	m.fileTypeList.Selected = &m.creationSel
	m.fileTypeList.Key = func(i int) any { return m.creationRows[i].key }
	m.fileTypeList.Reorder = m.reorderCreationRows
	m.tplList.Selected = &m.tplSel
	m.menuLists = make([]ui.ListState, len(menuGroups))
	for g := range menuGroups {
		list := &m.menuLists[g]
		list.Key = func(i int) any { return m.cfg.MenuItems[m.groupRows(g)[i]].ID }
		list.Reorder = func(rows []int, to int) { m.reorderGroup(g, rows, to) }
	}
	if err := config.Save(store.ConfigPath, m.cfg); err != nil {
		return nil, err
	}
	return m, nil
}

// view builds the sidebar and continuous content surface using the window theme.
func (m *app) view(c *ui.Context) {
	c.SetTheme(windowTheme(c))
	t := c.Theme()
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		m.sidebar(c)
		ui.Column(c).Width(1).Background(separatorColor(c))
		ui.Column(c).Grow(1).Background(t.Surface.Mix(t.Background, 0.3)).Children(func() {
			ui.Row(c).Height(12).DragWindow() // the top edge of the window, which moves it
			ui.Column(c).Grow(1).Padding(16, 28, 24).Children(func() {
				// The menu pane scrolls its editor and its preview on their own. Every other
				// pane scrolls as a whole, and its content is at most contentWidth wide, so
				// long lines stay readable in a wider window.
				if m.currentPane() == paneMenu {
					ui.Column(c).Grow(1).Children(func() { m.menuPage(c) })
					return
				}
				if m.currentPane() == paneTemplates {
					ui.Column(c).Grow(1).Children(func() { m.newPane(c) })
					m.showNotice(c)
					return
				}
				ui.Scroll(c).Fill().Children(func() {
					ui.Column(c).Gap(20).MaxWidth(m.contentWidth()).Children(func() {
						m.paneContent(c)
						m.showNotice(c)
					})
				})
			})
		})
	})
	m.creationDialog(c)
	m.favoriteDialog(c)
	m.menuNameDialog(c)
	m.menuPreviewDialog(c)
	if ui.AlertDialog(c, &m.confirmRestore, locales.RestoreDefaultsQuestion(), locales.RestoreDefaultsBody(), locales.Cancel(), locales.ConfirmRestore()) == 1 {
		m.restoreDefaultMenu()
	}
	// The restart dialog belongs to the window, not to one pane: the diagnostics
	// pane and the folders pane can ask for a restart too.
	if ui.AlertDialog(c, &m.confirmRestart, locales.RestartFinderTitle(), locales.RestartFinderBody(), locales.Cancel(), locales.Restart()) == 1 {
		m.restartFinder()
	}
}

// contentWidth is the widest the content of the chosen pane may be. The templates
// pane takes more room, so its two lists can sit side by side.
func (m *app) contentWidth() float32 {
	if m.currentPane() == paneTemplates {
		return 800
	}
	return 680
}

// currentPane returns the ID of the chosen pane.
func (m *app) currentPane() string {
	if m.tab < 0 || m.tab >= len(paneTabs()) {
		return paneMenu
	}
	return paneTabs()[m.tab].id
}

// selectPane shows the pane with the given ID.
func (m *app) selectPane(id string) {
	for i, p := range paneTabs() {
		if p.id == id {
			m.tab = i
		}
	}
	m.paneChanged()
}

// paneContent builds the chosen pane. The menu pane is not here: menuPage lays it out.
func (m *app) paneContent(c *ui.Context) {
	switch m.currentPane() {
	case paneTemplates:
		m.newPane(c)
	case paneFolders:
		m.foldersPane(c)
	case paneOpeners:
		m.openersPane(c)
	case paneDiagnostics:
		m.diagnosticsPane(c)
	case paneGeneral:
		m.generalPane(c)
	}
}

// paneChanged runs when a pane is chosen. The diagnostics pane checks the Finder
// extension each time it is shown.
func (m *app) paneChanged() {
	if m.currentPane() == paneDiagnostics {
		m.checkFinder()
	}
}

// saveConfig writes the settings file and reports a failure below the pane. The
// controls change m.cfg before they save. A failed write loads the file back, so
// the window shows the values that the Finder extension and the actions read.
func (m *app) saveConfig() {
	m.cfg.GroupMenuByLocation()
	err := config.Save(m.store.ConfigPath, m.cfg)
	if err != nil {
		saved, loadErr := config.Load(m.store.ConfigPath)
		if loadErr != nil {
			err = errors.Join(err, loadErr)
		} else {
			m.cfg = saved
		}
	}
	m.report(locales.SaveFailed(), err)
}

// saveTemplates writes the templates file and reports a failure below the pane.
// A failed write loads the file back, as saveConfig does.
func (m *app) saveTemplates() {
	for i := range m.templates {
		m.templates[i].Location = config.FileTypeSubmenu
	}
	err := files.SaveTemplates(m.store.TemplatesPath, m.templates)
	if err != nil {
		saved, loadErr := files.LoadTemplates(m.store.TemplatesPath)
		if loadErr != nil {
			err = errors.Join(err, loadErr)
		} else {
			m.templates = saved
			m.tplSel = min(m.tplSel, len(m.templates)-1)
		}
	}
	m.report(locales.SaveFailed(), err)
}

// report shows what went wrong below the pane, or clears the notice when err is nil.
func (m *app) report(what string, err error) {
	m.notice = ""
	if err != nil {
		m.notice = locales.NoticeFailure(what, err.Error())
	}
}

// showNotice shows the error of the last action in the danger color, when there is one.
func (m *app) showNotice(c *ui.Context) {
	if m.notice != "" {
		ui.Text(c, m.notice).FontSize(sizeNote).TextColor(c.Theme().Danger)
	}
}

// moveRows moves the items at rows so that they go before the item now at
// to, where len(items) means the end. It returns the new order and leaves
// items as it is.
func moveRows[T any](items []T, rows []int, to int) []T {
	moving := make(map[int]bool, len(rows))
	for _, r := range rows {
		if r >= 0 && r < len(items) {
			moving[r] = true
		}
	}
	var moved, kept []T
	insert := to
	for i, item := range items {
		if moving[i] {
			moved = append(moved, item)
			if i < to {
				insert--
			}
		} else {
			kept = append(kept, item)
		}
	}
	insert = max(0, min(insert, len(kept)))
	out := make([]T, 0, len(items))
	out = append(out, kept[:insert]...)
	out = append(out, moved...)
	return append(out, kept[insert:]...)
}
