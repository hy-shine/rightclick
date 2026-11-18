package settings

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/actionicons"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// The places a right-click can happen. The menu preview shows the menu for one
// of them, chosen with a switch above it.
const (
	ctxBlank = iota
	ctxFolder
	ctxFile
	ctxZip
	ctxTwoFiles
)

// previewContexts names the places in the order of the preview's switch.
func previewContexts() []string {
	return []string{
		locales.ContextBlank(),
		locales.ContextFolder(),
		locales.ContextFile(),
		locales.ContextZip(),
		locales.ContextTwoFiles(),
	}
}

// menuEntry is the label of a Finder menu entry and the condition that shows it.
type menuEntry struct{ label, when string }

// menuEntries describes the entries of config.Default, by ID. The labels are
// read on each call so that they follow the current language.
func menuEntries() map[string]menuEntry {
	return map[string]menuEntry{
		"new-file":           {locales.MenuNewFile(), locales.WhereBlankFolder()},
		"new-folder":         {locales.MenuNewFolder(), locales.WhereBlankFolder()},
		"new-template":       {locales.MenuNewTemplate(), locales.WhereBlankFolder()},
		"paste-as-file":      {locales.MenuPasteAsFile(), locales.WhereBlankFolder()},
		"add-template":       {locales.CaptureTemplate(), locales.WhereOneFile()},
		"favorite-folders":   {locales.FavoriteFolders(), locales.WhereAnywhere()},
		"copy-path":          {locales.MenuCopyPath(), locales.WhereAnywhere()},
		"copy-terminal-path": {locales.MenuCopyTerminalPath(), locales.WhereAnywhere()},
		"copy-relative-path": {locales.MenuCopyRelativePath(), locales.WhereAnywhere()},
		"copy-markdown-link": {locales.MenuCopyMarkdownLink(), locales.WhereAnywhere()},
		"copy-content":       {locales.MenuCopyContent(), locales.WhereOneFile()},
		"copy-name":          {locales.MenuCopyName(), locales.WhereAnywhere()},
		"copy-basename":      {locales.MenuCopyBaseName(), locales.WhereAnywhere()},
		"copy-to":            {locales.MenuCopyTo(), locales.WhereSelection()},
		"move-to":            {locales.MenuMoveTo(), locales.WhereSelection()},
		"open-terminal":      {locales.MenuOpenTerminal(), locales.WhereBlankFolder()},
		"zip":                {locales.MenuZip(), locales.WhereSelection()},
		"unzip":              {locales.MenuUnzip(), locales.WhereZipOnly()},
		"unzip-to-folder":    {locales.MenuUnzipToFolder(), locales.WhereZipOnly()},
		"hash":               {locales.MenuHash(), locales.WhereOneFile()},
		"compare-hash":       {locales.MenuCompareHash(), locales.WhereTwoFiles()},
		"chmod-x":            {locales.MenuMakeExecutable(), locales.WhereSelectedFiles()},
		"toggle-hidden":      {locales.MenuToggleHidden(), locales.WhereBlankFolder()},
		"delete-permanently": {locales.DeletePermanently(), locales.WhereSelection()},
	}
}

// locationValues and locationLabels are the choices of the location dropdown, in the
// order they are shown: the top of the Finder menu first, the app's own submenu second.
func locationValues() []string {
	return []string{config.FileTypeMain, config.FileTypeSubmenu}
}

func locationLabels() []string {
	return []string{locales.LocFinderTopShort(), locales.LocInApp()}
}

// menuGroups lists the IDs of each group of the Finder submenu, in the order
// the submenu shows them. It mirrors MenuActions.groups in
// extension/RightClickFinderSync/MenuConfig.swift; change both together.
var menuGroups = [][]string{
	{"new-file", "new-folder", "new-template", "paste-as-file"},
	{"copy-path", "copy-name", "copy-basename", "copy-content", "copy-to", "move-to", "copy-terminal-path", "copy-relative-path", "copy-markdown-link"},
	{"open-terminal"},
	{"zip", "unzip", "unzip-to-folder", "hash"},
	{"compare-hash"},
	{"chmod-x"},
	{"toggle-hidden"},
	{"delete-permanently"},
	{"add-template"},
}

// menuContexts lists, by entry ID, the places where Finder shows the entry. It
// mirrors allows() in extension/RightClickFinderSync/FinderSyncExtension.swift;
// change both together.
var menuContexts = map[string][]int{
	"new-file":           {ctxBlank, ctxFolder},
	"new-folder":         {ctxBlank, ctxFolder},
	"new-template":       {ctxBlank, ctxFolder},
	"paste-as-file":      {ctxBlank, ctxFolder},
	"add-template":       {ctxFile},
	"copy-path":          {ctxBlank, ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"copy-terminal-path": {ctxBlank, ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"copy-relative-path": {ctxBlank, ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"copy-markdown-link": {ctxBlank, ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"copy-content":       {ctxFile},
	"copy-name":          {ctxBlank, ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"copy-basename":      {ctxBlank, ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"copy-to":            {ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"move-to":            {ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"open-terminal":      {ctxBlank, ctxFolder},
	"zip":                {ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
	"unzip":              {ctxZip},
	"unzip-to-folder":    {ctxZip},
	"hash":               {ctxFile, ctxZip},
	"compare-hash":       {ctxTwoFiles},
	"chmod-x":            {ctxFile, ctxZip, ctxTwoFiles},
	"toggle-hidden":      {ctxBlank, ctxFolder},
	"delete-permanently": {ctxFolder, ctxFile, ctxZip, ctxTwoFiles},
}

// menuVisualSection mirrors MenuActions.visualSection in MenuConfig.swift.
func menuVisualSection(id string) int {
	if strings.HasPrefix(id, "app:") {
		return 2
	}
	if strings.HasPrefix(id, "folder:") {
		return 1
	}
	switch id {
	case "new-file", "new-folder", "new-template", "paste-as-file", "add-template":
		return 0
	case "copy-path", "copy-name", "copy-basename", "copy-content", "copy-to", "move-to",
		"favorite-folders", "copy-terminal-path", "copy-relative-path", "copy-markdown-link":
		return 1
	case "open-terminal":
		return 2
	case "zip", "unzip", "unzip-to-folder", "hash", "compare-hash":
		return 3
	case "chmod-x", "toggle-hidden":
		return 4
	case "delete-permanently":
		return 5
	default:
		return 4
	}
}

const (
	menuRowHeight     = 38
	menuWhenWidth     = 114
	menuLocationWidth = 142
)

// menuEntryFor returns the entry for id. An unknown ID is shown as it is.
func menuEntryFor(id string) menuEntry {
	if e, ok := menuEntries()[id]; ok {
		return e
	}
	return menuEntry{label: id}
}

// appliesTo reports whether Finder shows the entry id in the place ctx.
func appliesTo(id string, ctx int) bool {
	return slices.Contains(menuContexts[id], ctx)
}

func isNestedNewFileItem(id string) bool {
	return id == "new-folder" || id == "new-template"
}

// groupRows returns the indexes in m.cfg.MenuItems of the entries in group g,
// in config order. Kept for group-level unit tests.
func (m *app) groupRows(g int) []int {
	var rows []int
	for i, item := range m.cfg.MenuItems {
		if g < 0 || slices.Contains(menuGroups[g], item.ID) {
			rows = append(rows, i)
		}
	}
	return rows
}

// menuDisplayRows returns a group's rows that match the current search.
func (m *app) menuDisplayRows(g int) []int {
	return m.filterMenuRows(m.groupRows(g))
}

func (m *app) filterMenuRows(rows []int) []int {
	query := strings.ToLower(strings.TrimSpace(m.menuSearch))
	if query == "" {
		return rows
	}
	return slices.DeleteFunc(rows, func(i int) bool {
		entry := m.menuEntry(m.cfg.MenuItems[i])
		text := strings.ToLower(entry.label + " " + entry.when + " " + m.cfg.MenuItems[i].ID)
		return !strings.Contains(text, query)
	})
}

func (m *app) locationRows(location string) []int {
	var rows []int
	for i, item := range m.cfg.MenuItems {
		if isNestedNewFileItem(item.ID) {
			continue
		}
		isMain := item.Location == config.FileTypeMain
		if (location == config.FileTypeMain) == isMain {
			rows = append(rows, i)
		}
	}
	return rows
}

func (m *app) nestedNewFileRows() []int {
	var rows []int
	for i, item := range m.cfg.MenuItems {
		if isNestedNewFileItem(item.ID) {
			rows = append(rows, i)
		}
	}
	return rows
}

// reorderGroup moves the entries of group g at rows to before to, and saves.
func (m *app) reorderGroup(g int, rows []int, to int) {
	at := m.groupRows(g)
	items := make([]config.MenuItem, len(at))
	for k, i := range at {
		items[k] = m.cfg.MenuItems[i]
	}
	items = moveRows(items, rows, to)
	for k, i := range at {
		m.cfg.MenuItems[i] = items[k]
	}
	m.saveConfig()
}

// reorderLocation moves entries within either the Main Menu or Submenu section cleanly.
func (m *app) reorderLocation(location string, rows []int, to int) {
	at := m.locationRows(location)
	items := make([]config.MenuItem, len(at))
	for k, i := range at {
		items[k] = m.cfg.MenuItems[i]
	}
	items = moveRows(items, rows, to)
	for k, i := range at {
		m.cfg.MenuItems[i] = items[k]
	}
	m.saveConfig()
}

// restoreDefaultMenu puts back the entries and their order from config.Default.
func (m *app) restoreDefaultMenu() {
	defaults := config.Default().MenuItems
	for _, item := range m.cfg.MenuItems {
		if strings.Contains(item.ID, ":") || item.ID == "favorite-folders" {
			defaults = append(defaults, item)
		}
	}
	m.cfg.MenuItems = defaults
	m.saveConfig()
	m.pendingToast = locales.RestoredDefaultsToast()
}

// previewItem is one entry of the menu preview.
type previewItem struct{ id, label string }

// previewGroups returns the submenu entries that the preview shows in place ctx,
// grouped by visual section so the preview shows dividers just like Finder.
func (m *app) previewGroups(ctx int) [][]previewItem {
	var groups [][]previewItem
	lastSection := -1
	for _, item := range m.cfg.MenuItems {
		if !item.Enabled || item.Location == config.FileTypeMain || !m.menuItemAvailable(item, ctx) {
			continue
		}
		if isNestedNewFileItem(item.ID) {
			continue
		}
		if item.ID == "new-file" && !m.hasNewFileContents() {
			continue
		}
		sec := menuVisualSection(item.ID)
		entry := previewItem{item.ID, m.menuEntry(item).label}
		if len(groups) == 0 || sec != lastSection {
			groups = append(groups, []previewItem{entry})
			lastSection = sec
		} else {
			groups[len(groups)-1] = append(groups[len(groups)-1], entry)
		}
	}
	return groups
}

func (m *app) previewTopItems(ctx int) []previewItem {
	var items []previewItem
	for _, item := range m.cfg.MenuItems {
		if !item.Enabled || item.Location != config.FileTypeMain || !m.menuItemAvailable(item, ctx) {
			continue
		}
		if isNestedNewFileItem(item.ID) {
			continue
		}
		if item.ID == "new-file" && !m.hasNewFileContents() {
			continue
		}
		items = append(items, previewItem{item.ID, m.menuEntry(item).label})
	}
	return items
}

func (m *app) menuPage(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, locales.PaneMenu()).FontSize(sizeTitle).FontWeight(600)
			ui.Text(c, locales.MenuIntro()).FontSize(sizeNote).LineHeight(1.5).TextColor(t.TextMuted)
		})
		searchInput(c.Key("menu-search"), &m.menuSearch, locales.SearchMenu(), locales.SearchMenuPlaceholder()).Width(175)
		if ghostButton(c, locales.RestoreDefaults(), false) {
			m.confirmRestore = true
		}
		previewKind := quietKind
		if m.menuPreviewOpen {
			previewKind = primaryKind
		}
		if actionButton(c, previewKind, iconEye, locales.Preview(), ui.Center) {
			m.menuPreviewOpen = !m.menuPreviewOpen
		}
	})
	m.showNotice(c)
	width, _ := c.Size()
	dockedPreview := m.menuPreviewOpen && width >= 960
	if dockedPreview {
		ui.Row(c).Grow(1).Margin(16, 0, 0).Gap(18).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				m.menuEditor(c, true)
			})
			m.menuPreviewPanel(c)
		})
	} else {
		ui.Column(c).Grow(1).Margin(16, 0, 0).Children(func() {
			m.menuEditor(c, false)
		})
	}
}

func (m *app) menuEditor(c *ui.Context, compact bool) {
	t := c.Theme()
	mainRows := m.filterMenuRows(m.locationRows(config.FileTypeMain))
	subRows := m.filterMenuRows(m.locationRows(config.FileTypeSubmenu))
	nestedRows := m.filterMenuRows(m.nestedNewFileRows())
	if len(mainRows)+len(subRows)+len(nestedRows) == 0 {
		ui.Column(c).Grow(1).Justify(ui.Center).Children(func() {
			emptyState(c, iconMenu, locales.NoMenuMatches(), locales.SearchEmptyHint(), func() {
				if quietButton(c, nil, locales.ClearSearch()) {
					m.menuSearch = ""
				}
			})
		})
		return
	}
	searching := strings.TrimSpace(m.menuSearch) != ""
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Gap(18).Padding(0, 2, 8, 0).Children(func() {
			// Section 1: Finder Top-Level Menu
			ui.Column(c).Gap(8).Children(func() {
				sectionHeader(c, nil, locales.MenuSectionMain(), fmt.Sprint(len(mainRows)), locales.MenuSectionMainHint())
				ledger(c, ui.Column(c)).Children(func() {
					if len(mainRows) == 0 {
						ui.Row(c).Padding(14, 16).Children(func() {
							ui.Text(c, locales.MenuSectionMainEmpty()).FontSize(sizeNote).TextColor(t.TextMuted)
						})
					} else {
						m.menuTableHeader(c, compact)
						m.menuLocationList(c, config.FileTypeMain, &m.mainMenuList, mainRows, compact, searching)
					}
				})
			})

			// Section 2: Nested items inside New File
			if len(nestedRows) > 0 {
				ui.Column(c).Gap(8).Children(func() {
					sectionHeader(c, iconNew, locales.MenuNewFileInside(), fmt.Sprint(len(nestedRows)), locales.MenuSubmenuHint())
					ledger(c, ui.Column(c)).Children(func() {
						m.menuNestedRows(c, nestedRows, compact)
					})
				})
			}

			// Section 3: RightClick Submenu
			ui.Column(c).Gap(8).Children(func() {
				sectionHeader(c, actionicons.Icon("menu"), locales.MenuSectionSub(), fmt.Sprint(len(subRows)), locales.MenuSectionSubHint())
				ledger(c, ui.Column(c)).Children(func() {
					if len(subRows) == 0 {
						ui.Row(c).Padding(14, 16).Children(func() {
							ui.Text(c, locales.MenuSectionSubEmpty()).FontSize(sizeNote).TextColor(t.TextMuted)
						})
					} else {
						m.menuTableHeader(c, compact)
						m.menuLocationList(c, config.FileTypeSubmenu, &m.subMenuList, subRows, compact, searching)
					}
				})
			})
			ui.Text(c, locales.MenuArrangeHint()).FontSize(sizeNote).TextColor(t.TextMuted)
		})
	})
}

func (m *app) menuTableHeader(c *ui.Context, compact bool) {
	t := c.Theme()
	locWidth := float32(menuLocationWidth)
	if compact {
		locWidth = 136
	}
	ui.Row(c).Height(30).Padding(0, 10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, locales.ColActionContext()).FontSize(sizeCaption).FontWeight(600).TextColor(t.TextMuted).Grow(1)
		if !compact {
			ui.Text(c, locales.ColContext()).FontSize(sizeCaption).FontWeight(600).TextColor(t.TextMuted).Width(menuWhenWidth).Shrink(0)
		}
		ui.Text(c, locales.ColLocation()).FontSize(sizeCaption).FontWeight(600).TextColor(t.TextMuted).Width(locWidth).Shrink(0)
		ui.Text(c, locales.ColEnabled()).FontSize(sizeCaption).FontWeight(600).TextColor(t.TextMuted).Width(48).Shrink(0).TextAlign(ui.Center)
		ui.Row(c).Width(32).Shrink(0)
	})
}

func (m *app) menuLocationList(c *ui.Context, location string, list *ui.ListState, rows []int, compact, searching bool) {
	list.Key = func(i int) any { return m.cfg.MenuItems[rows[i]].ID }
	list.Reorder = nil
	if !searching {
		list.Reorder = func(moved []int, to int) { m.reorderLocation(location, moved, to) }
	}
	ui.List(c, list, len(rows), func(i int) {
		m.menuActionRow(c, rows[i], compact, !searching, true)
	}).Height(float32(len(rows))*menuRowHeight).Dividers(1, separatorColor(c))
}

func (m *app) menuNestedRows(c *ui.Context, rows []int, compact bool) {
	t := c.Theme()
	for _, idx := range rows {
		item := &m.cfg.MenuItems[idx]
		entry := m.menuEntry(*item)
		labelColor, iconColor := t.Text, t.TextMuted
		if !item.Enabled {
			labelColor, iconColor = t.TextMuted, t.TextMuted.Alpha(0.6)
		}
		row := ui.Row(c).Height(menuRowHeight).Padding(0, 10).AlignItems(ui.Center)
		row.Children(func() {
			ui.Row(c).Gap(9).Grow(1).MinWidth(0).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Size(itemGripSlot, 14).Shrink(0)
				m.entryMark(c, item.ID, 16, iconColor)
				ui.Text(c, entry.label).FontSize(sizeBody).TextColor(labelColor).SingleLine().Grow(1).MinWidth(0).Tooltip(entry.label)
				if !compact {
					ui.Text(c, entry.when).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine().Width(menuWhenWidth).Shrink(0)
				}
			})
			locWidth := float32(menuLocationWidth)
			if compact {
				locWidth = 136
			}
			ui.Row(c).Width(locWidth).Shrink(0).AlignItems(ui.Center).Children(func() {
				if item.ID == "new-template" {
					if ghostButton(c.Key("goto-tpl"), locales.ManageTemplatesLink(), false) {
						m.newMode = newModeTemplates
						m.selectPane(paneTemplates)
					}
				} else {
					pillBadge(c, locales.MenuNewFileInside(), t.TextMuted, true)
				}
			})
			ui.Row(c).Width(48).Shrink(0).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
				creationSwitch(c.Key("enable:"+item.ID), &item.Enabled, locales.EnableLabel(entry.label)).OnChange(m.saveConfig)
			})
			ui.Row(c).Width(32).Shrink(0).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
				if item.ID == "new-folder" && iconButton(c.Key("rename:"+item.ID), iconEdit, locales.EditItem(entry.label), false) {
					m.menuNameID, m.menuName, m.menuNameOpen = item.ID, item.Label, true
				}
			})
		})
		if row.Hovered() {
			row.Background(t.SurfaceHover)
		}
	}
}

func (m *app) menuActionRow(c *ui.Context, cfgIndex int, compact, allowReorder, allowRename bool) {
	t := c.Theme()
	item := &m.cfg.MenuItems[cfgIndex]
	entry := m.menuEntry(*item)
	labelColor, iconColor := t.Text, t.TextMuted
	if !item.Enabled {
		labelColor, iconColor = t.TextMuted, t.TextMuted.Alpha(0.6)
	}
	row := ui.Row(c).Height(menuRowHeight).Padding(0, 10).AlignItems(ui.Center)
	row.Children(func() {
		ui.Row(c).Gap(9).Grow(1).MinWidth(0).AlignItems(ui.Center).Children(func() {
			if allowReorder {
				gripColor := t.TextMuted.Alpha(.55)
				if row.Hovered() {
					gripColor = t.TextMuted
				}
				ui.Row(c).Size(itemGripSlot, 14).Shrink(0).AlignItems(ui.Center).Children(func() {
					glyph(c, iconGrip, 12, gripColor)
				})
			} else {
				ui.Row(c).Size(itemGripSlot, 14).Shrink(0)
			}
			m.entryMark(c, item.ID, 16, iconColor)
			ui.Text(c, entry.label).FontSize(sizeBody).TextColor(labelColor).SingleLine().Grow(1).MinWidth(0).Tooltip(entry.label)
			if !compact {
				ui.Text(c, entry.when).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine().Width(menuWhenWidth).Shrink(0)
			}
		})
		locWidth := float32(menuLocationWidth)
		if compact {
			locWidth = 136
		}
		ui.Row(c).Width(locWidth).Shrink(0).AlignItems(ui.Center).Children(func() {
			locationSelect(c, &item.Location, "action:"+item.ID, locales.LocationLabel(entry.label),
				locationValues(), locationLabels(), item.Enabled, m.saveConfig).Width(locWidth - 8)
		})
		ui.Row(c).Width(48).Shrink(0).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			creationSwitch(c.Key("enable:"+item.ID), &item.Enabled, locales.EnableLabel(entry.label)).OnChange(m.saveConfig)
		})
		ui.Row(c).Width(32).Shrink(0).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
			if allowRename && iconButton(c.Key("rename:"+item.ID), iconEdit, locales.EditItem(entry.label), false) {
				m.menuNameID, m.menuName, m.menuNameOpen = item.ID, item.Label, true
			}
		})
	})
	if row.Hovered() {
		row.Background(t.SurfaceHover)
	}
	if allowRename && row.RowDoubleClicked() {
		m.menuNameID, m.menuName, m.menuNameOpen = item.ID, item.Label, true
	}
}

// menuPreviewPanel renders the docked live preview studio on wide windows.
func (m *app) menuPreviewPanel(c *ui.Context) {
	t := c.Theme()
	hint := locales.PreviewLiveHint()
	if strings.TrimSpace(m.menuSearch) != "" {
		hint = locales.PreviewSearchHint()
	}
	ui.Column(c).Width(416).Shrink(0).Padding(16).Radius(12).
		Background(t.Surface).Border(1, separatorColor(c)).Gap(12).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, locales.MenuPreview()).FontSize(sizeSection).FontWeight(600)
			ui.Spacer(c)
			if ghostButton(c, locales.ClosePreview(), false) {
				m.menuPreviewOpen = false
			}
		})
		ui.Text(c, hint).FontSize(sizeCaption).LineHeight(1.4).TextColor(t.TextMuted)
		m.previewControls(c)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(8).Radius(10).Background(t.Background.Alpha(0.7)).Children(func() {
				m.previewMenusCompact(c)
			})
		})
	})
}

func (m *app) menuPreviewDialog(c *ui.Context) {
	ui.DialogBase(c, &m.menuPreviewOpen, func(backdrop, panel ui.Element) {
		t := c.Theme()
		_, height := c.Size()
		backdrop.Background(ui.RGBA(0, 0, 0, .3))
		panel.Width(540).MaxHeight(height-32).Padding(22).Background(t.Surface).Border(1, t.Border).Radius(dialogRadius).Label(locales.MenuPreview())
		hint := locales.PreviewLiveHint()
		if strings.TrimSpace(m.menuSearch) != "" {
			hint = locales.PreviewSearchHint()
		}
		ui.Column(c).Gap(12).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, locales.MenuPreview()).FontSize(sizeSection).FontWeight(600)
				ui.Spacer(c)
				if ghostButton(c, locales.ClosePreview(), false) {
					m.menuPreviewOpen = false
				}
			})
			ui.Text(c, hint).FontSize(sizeCaption).TextColor(t.TextMuted)
			m.previewControls(c)
			ui.Scroll(c).MaxHeight(height - 196).Children(func() { m.previewMenus(c) })
		})
	})
}

// Sizes of the menu drawing, in points.
const (
	stageRow           = 24
	stageAbove         = 2
	stageBelow         = 3
	stageEdge          = 4
	parentWidth        = 176
	subWidth           = 256
	compactParentWidth = 122
	compactSubWidth    = 232
	stageGap           = 6
)

// menuStage combines the context choices and menu drawing for standalone preview tests.
func (m *app) menuStage(c *ui.Context) {
	ui.Column(c).Gap(16).Children(func() {
		m.previewControls(c)
		m.previewMenus(c)
	})
}

// previewControls stays above the scroll area so the context can always be changed.
func (m *app) previewControls(c *ui.Context) {
	t := c.Theme()
	groups := m.previewGroups(m.previewCtx)
	n := len(m.previewTopItems(m.previewCtx))
	for _, g := range groups {
		n += len(g)
	}
	ui.Column(c).Gap(8).Children(func() {
		m.placeSwitch(c)
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Spacer(c)
			ui.Text(c, locales.ShowCount(n)).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine()
		})
	})
}

func (m *app) previewMenus(c *ui.Context) {
	ui.Row(c).Gap(stageGap).AlignItems(ui.Start).Children(func() {
		m.parentMenu(c, parentWidth)
		m.submenu(c, m.previewGroups(m.previewCtx), subWidth)
	})
}

func (m *app) previewMenusCompact(c *ui.Context) {
	ui.Row(c).Gap(stageGap).AlignItems(ui.Start).Children(func() {
		m.parentMenu(c, compactParentWidth)
		m.submenu(c, m.previewGroups(m.previewCtx), compactSubWidth)
	})
}

// placeSwitch is the row of places above the drawing, in a gray track.
func (m *app) placeSwitch(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Wrap().Gap(2).Padding(3).Radius(t.Radius+4).Background(t.Background).Border(1, separatorColor(c)).Role(ui.RoleRadioGroup).Label(locales.RightClickLocation()).Children(func() {
		for i, name := range previewContexts() {
			selected := i == m.previewCtx
			var label ui.Element
			b := ui.ButtonBase(c).Padding(4, 10).Radius(t.Radius).Label(name).Role(ui.RoleRadio).Checked(selected).Children(func() {
				label = ui.Text(c, name).FontSize(sizeNote).TextColor(t.TextMuted).SingleLine()
			})
			switch {
			case selected:
				label.TextColor(t.Text).FontWeight(600)
				b.Background(t.Surface)
			case b.Hovered():
				label.TextColor(t.Text)
				tint(b, ui.Transparent, t.SurfaceHover, t.SurfacePressed)
			default:
				tint(b, ui.Transparent, t.SurfaceHover, t.SurfacePressed)
			}
			if b.Clicked() {
				m.previewCtx = i
			}
		}
	})
}

// menuCard is the look of a menu that floats: a card with a border and a shadow.
func menuCard(c *ui.Context, e ui.Element, shadow float32) ui.Element {
	t := c.Theme()
	return e.Padding(stageEdge).Background(t.Surface).Border(1, t.Border).Radius(9).
		Shadow(0, 2+shadow/4, 8+shadow/2, 0, ui.RGBA(0, 0, 0, 0.08+shadow/200))
}

// parentMenu draws the menu that holds the app's submenu, with its row lit.
func (m *app) parentMenu(c *ui.Context, width float32) {
	fill, text := c.Theme().Accent, c.Theme().AccentText
	t := c.Theme()
	bar := func(w float32) {
		ui.Row(c).Height(stageRow).Padding(0, 10).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Size(w, 6).Radius(3).Background(t.Border)
		})
	}
	menuCard(c, ui.Column(c).Width(width), 0).Children(func() {
		bar(min(52, width*0.35))
		bar(min(68, width*0.45))
		for _, item := range m.previewTopItems(m.previewCtx) {
			ui.Row(c).Height(stageRow).Gap(5).Padding(0, 6).AlignItems(ui.Center).Children(func() {
				m.entryMark(c, item.id, 13, t.TextMuted)
				ui.Text(c, item.label).FontSize(sizeNote).SingleLine().Grow(1).MinWidth(0).Tooltip(item.label)
			})
		}
		ui.Row(c).Height(stageRow).Gap(4).Padding(0, 6, 0, 10).AlignItems(ui.Center).Radius(5).Background(fill).Children(func() {
			ui.Text(c, locales.AppName()).FontSize(sizeNote).FontWeight(600).TextColor(text).Grow(1).SingleLine()
			glyph(c, iconChevron, 11, text)
		})
		bar(min(60, width*0.40))
		bar(min(42, width*0.28))
		bar(min(56, width*0.38))
	})
}

// submenu draws the app's submenu as Finder shows it in the chosen place, with
// a line between groups. A row lights up under the pointer, as a menu's rows do.
func (m *app) submenu(c *ui.Context, groups [][]previewItem, width float32) {
	fill, text := c.Theme().Accent, c.Theme().AccentText
	t := c.Theme()
	top := float32((stageAbove + len(m.previewTopItems(m.previewCtx))) * stageRow)
	menuCard(c, ui.Column(c).Width(width).Margin(top, 0, 0, 0), 12).Children(func() {
		for g, items := range groups {
			if g > 0 {
				ui.Divider(c).Margin(4, 6).Background(separatorColor(c))
			}
			for _, it := range items {
				var icon, label ui.Element
				row := ui.Row(c).Height(stageRow).Gap(8).Padding(0, 8).AlignItems(ui.Center).Radius(5).Children(func() {
					iconColor := t.TextMuted
					if it.id == "delete-permanently" {
						iconColor = t.Danger
					}
					icon = m.entryMark(c, it.id, 15, iconColor)
					label = ui.Text(c, it.label).FontSize(sizeNote).TextColor(t.Text).SingleLine().Grow(1).MinWidth(0).Tooltip(it.label)
				})
				if row.Hovered() {
					row.Background(fill)
					icon.TextColor(text)
					label.TextColor(text)
				}
			}
		}
		if len(groups) == 0 {
			ui.Text(c, locales.NoItems()).FontSize(sizeNote).TextColor(t.TextMuted).Padding(8, 8)
		}
	})
}

// menuGroup builds one group of the menu editor. Kept for group-level unit tests.
func (m *app) menuGroup(c *ui.Context, g int, rows []int) {
	list := &m.menuLists[max(0, g)]
	list.Key = func(i int) any { return m.cfg.MenuItems[rows[i]].ID }
	list.Reorder = nil
	if strings.TrimSpace(m.menuSearch) == "" {
		list.Reorder = func(moved []int, to int) { m.reorderGroup(g, moved, to) }
	}
	ui.List(c, list, len(rows), func(i int) {
		m.menuActionRow(c, rows[i], false, strings.TrimSpace(m.menuSearch) == "", true)
	}).Height(float32(len(rows)) * menuRowHeight)
}
