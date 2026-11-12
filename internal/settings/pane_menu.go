package settings

import (
	"github.com/hy-shine/rightclick/internal/locales"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
)

// The places a right-click can happen. The menu preview shows the menu for one
// of them, chosen with a switch above it.
const (
	ctxBlank = iota
	ctxFolder
	ctxFile
	ctxTwoFiles
)

// previewContexts names the places in the order of the preview's switch.
func previewContexts() []string {
	return []string{locales.ContextBlank(), locales.MenuNewFolder(), locales.ContextFile(), locales.ContextTwoFiles()}
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
		"add-template":       {locales.CaptureTemplate(), locales.WhereOneFile()},
		"favorite-folders":   {locales.FavoriteFolders(), locales.WhereAnywhere()},
		"copy-path":          {locales.MenuCopyPath(), locales.WhereAnywhere()},
		"copy-name":          {locales.MenuCopyName(), locales.WhereAnywhere()},
		"copy-basename":      {locales.MenuCopyBaseName(), locales.WhereAnywhere()},
		"copy-to":            {locales.MenuCopyTo(), locales.WhereSelection()},
		"move-to":            {locales.MenuMoveTo(), locales.WhereSelection()},
		"open-terminal":      {locales.MenuOpenTerminal(), locales.WhereBlankFolder()},
		"zip":                {locales.MenuZip(), locales.WhereSelectedFiles()},
		"unzip":              {locales.MenuUnzip(), locales.WhereZipOnly()},
		"hash":               {locales.MenuHash(), locales.WhereOneFile()},
		"compare-hash":       {locales.MenuCompareHash(), locales.WhereTwoFiles()},
		"toggle-hidden":      {locales.MenuToggleHidden(), locales.WhereBlankFolder()},
		"delete-permanently": {locales.DeletePermanently(), locales.WhereSelection()},
	}
}

// gripMark is the drag handle of a menu row. Its viewBox is cropped to the three
// pairs of dots, so the dots start at the left edge of the handle's box. The
// section header lines up with that edge. gripScale is the scale the 24-unit
// handle had in its 10 pt box before, so the dots keep their size.
var gripMark = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="7.7 4.7 8.6 14.6" ` +
	`fill="none" stroke="#000" stroke-linecap="round"><path stroke-width="2.6" ` +
	`d="M9 6h.01M15 6h.01M9 12h.01M15 12h.01M9 18h.01M15 18h.01"/></svg>`))

const gripScale = 10.0 / 24

// gripWidth is the width of the dots, and so of the handle's box. A wider box
// would leave an empty gap between the dots and the entry icon.
const gripWidth = 8.6 * gripScale

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
	{"new-file", "new-folder", "new-template"},
	{"copy-path", "copy-name", "copy-basename", "copy-to", "move-to"},
	{"open-terminal"},
	{"zip", "unzip", "hash"},
	{"compare-hash"},
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
	"add-template":       {ctxFile},
	"copy-path":          {ctxBlank, ctxFolder, ctxFile, ctxTwoFiles},
	"copy-name":          {ctxBlank, ctxFolder, ctxFile, ctxTwoFiles},
	"copy-basename":      {ctxBlank, ctxFolder, ctxFile, ctxTwoFiles},
	"copy-to":            {ctxFolder, ctxFile, ctxTwoFiles},
	"move-to":            {ctxFolder, ctxFile, ctxTwoFiles},
	"open-terminal":      {ctxBlank, ctxFolder},
	"zip":                {ctxFile, ctxTwoFiles},
	"unzip":              {ctxFile},
	"hash":               {ctxFile},
	"compare-hash":       {ctxTwoFiles},
	"toggle-hidden":      {ctxBlank, ctxFolder},
	"delete-permanently": {ctxFolder, ctxFile, ctxTwoFiles},
}

// menuRowHeight is the height of one row of the menu editor, in points.
const menuRowHeight = 36

// menuWhenWidth is the width of the condition column beside an entry.
const menuWhenWidth = 150

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

// groupRows returns the indexes in m.cfg.MenuItems of the entries in group g,
// in config order.
func (m *app) groupRows(g int) []int {
	var rows []int
	for i, item := range m.cfg.MenuItems {
		if g < 0 || slices.Contains(menuGroups[g], item.ID) {
			rows = append(rows, i)
		}
	}
	return rows
}

// menuDisplayRows returns a group's rows that match the current search. The
// underlying config order stays unchanged while a search is active.
func (m *app) menuDisplayRows(g int) []int {
	rows := m.groupRows(g)
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

// reorderGroup moves the entries of group g at rows, which are positions
// within the group, to before the entry now at to, and saves. Entries of other
// groups keep their places in the config.
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
}

// previewItem is one entry of the menu preview.
type previewItem struct{ id, label string }

// previewGroups returns the entries that the preview shows in place ctx: the
// enabled entries that apply there, one slice per group that has any, in menu
// order.
func (m *app) previewGroups(ctx int) [][]previewItem {
	var items []previewItem
	for _, item := range m.cfg.MenuItems {
		if !item.Enabled || item.Location == config.FileTypeMain || !m.menuItemAvailable(item, ctx) {
			continue
		}
		if item.ID == "new-template" || item.ID == "new-folder" {
			continue
		}
		if item.ID == "new-file" && !m.hasNewFileContents() {
			continue
		}
		items = append(items, previewItem{item.ID, m.menuEntry(item).label})
	}
	if len(items) == 0 {
		return nil
	}
	return [][]previewItem{items}
}

func (m *app) previewTopItems(ctx int) []previewItem {
	var items []previewItem
	for _, item := range m.cfg.MenuItems {
		if !item.Enabled || item.Location != config.FileTypeMain || !m.menuItemAvailable(item, ctx) {
			continue
		}
		if item.ID == "new-template" || item.ID == "new-folder" {
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
		ui.Column(c).Grow(1).Gap(6).Children(func() {
			ui.Text(c, locales.PaneMenu()).FontSize(sizeTitle).FontWeight(600)
			ui.Text(c, locales.MenuIntro()).FontSize(sizeNote).LineHeight(1.5).TextColor(t.TextMuted)
		})
		ui.TextInput(c.Key("menu-search"), &m.menuSearch).
			Label(locales.SearchMenu()).
			Placeholder(locales.SearchMenuPlaceholder()).
			Width(170)
		if ghostButton(c, locales.RestoreDefaults(), false) {
			m.confirmRestore = true
		}
		if quietButtonCentered(c, iconEye, locales.Preview()) {
			m.menuPreviewOpen = true
		}
	})
	m.showNotice(c)
	m.menuEditor(c)
}

func (m *app) menuEditor(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Margin(20, 0, 8, 0).Padding(0, 8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, locales.ColActionContext()).FontSize(sizeCaption).TextColor(t.TextMuted).Grow(1)
		ui.Text(c, locales.ColLocation()).FontSize(sizeCaption).TextColor(t.TextMuted).Width(118).Shrink(0)
		ui.Text(c, locales.ColEnabled()).FontSize(sizeCaption).TextColor(t.TextMuted).Width(48).Shrink(0).TextAlign(ui.Center)
	})
	ui.Divider(c).Background(separatorColor(c))
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Children(func() {
			rows := m.menuDisplayRows(-1)
			if len(rows) == 0 {
				ui.Text(c, locales.NoMenuMatches()).FontSize(sizeNote).TextColor(t.TextMuted).Padding(16, 12)
			} else {
				m.menuGroup(c, -1, rows)
			}
		})
	})
	ui.Text(c, locales.MenuArrangeHint()).FontSize(sizeCaption).TextColor(t.TextMuted).Margin(8, 0, 0)
}

func (m *app) menuPreviewDialog(c *ui.Context) {
	ui.DialogBase(c, &m.menuPreviewOpen, func(backdrop, panel ui.Element) {
		t := c.Theme()
		_, height := c.Size()
		backdrop.Background(ui.RGBA(0, 0, 0, .3))
		panel.Width(520).MaxHeight(height-32).Padding(22).Background(t.Surface).Border(1, t.Border).Radius(10).Label(locales.MenuPreview())
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
			ui.Scroll(c).MaxHeight(height - 140).Children(func() { m.menuStage(c) })
		})
	})
}

// Sizes of the menu drawing, in points. A menu row is stageRow high. The parent
// menu has stageAbove rows over the app's row and stageBelow under it, and
// the submenu starts level with that row.
const (
	stageRow    = 24
	stageAbove  = 2
	stageBelow  = 3
	stageEdge   = 4   // padding inside a menu card
	parentWidth = 112 // the parent menu card
	subWidth    = 220 // the submenu card
	stageSubTop = stageAbove * stageRow
	stageGap    = 6 // space between the parent menu and the submenu
)

// menuStage is the pane's hero: the places to choose from, and a drawing of the
// Finder menu in the chosen place. It sits on the paper, with no mat of its own.
func (m *app) menuStage(c *ui.Context) {
	t := c.Theme()
	groups := m.previewGroups(m.previewCtx)
	n := len(m.previewTopItems(m.previewCtx))
	for _, g := range groups {
		n += len(g)
	}
	ui.Column(c).Gap(16).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			m.placeSwitch(c)
			ui.Spacer(c)
			ui.Text(c, locales.ShowCount(n)).FontSize(sizeNote).TextColor(t.TextMuted).SingleLine()
		})
		ui.Row(c).Gap(stageGap).AlignItems(ui.Start).Children(func() {
			m.parentMenu(c)
			m.submenu(c, groups)
		})
	})
}

// placeSwitch is the row of places above the drawing, in a gray track. The
// chosen place is a white chip in bold, as the chosen pane is in the sidebar.
func (m *app) placeSwitch(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(2).Padding(3).Radius(t.Radius+4).Background(t.Background).Border(1, t.Border).Role(ui.RoleRadioGroup).Label(locales.RightClickLocation()).Children(func() {
		for i, name := range previewContexts() {
			selected := i == m.previewCtx
			var label ui.Element
			b := ui.ButtonBase(c).Padding(5, 14).Radius(t.Radius).Label(name).Role(ui.RoleRadio).Checked(selected).Children(func() {
				label = ui.Text(c, name).FontSize(sizeBody).TextColor(t.TextMuted)
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

// menuCard is the look of a menu that floats: a card with a border and a
// shadow. e is the element to give it.
func menuCard(c *ui.Context, e ui.Element, shadow float32) ui.Element {
	t := c.Theme()
	return e.Padding(stageEdge).Background(t.Surface).Border(1, t.Border).Radius(8).
		Shadow(0, 2+shadow/4, 6+shadow/2, 0, ui.RGBA(0, 0, 0, 0.06+shadow/200))
}

// parentMenu draws the menu that holds the app's submenu, with its row lit. The other
// rows are bars, since what they say does not matter here.
func (m *app) parentMenu(c *ui.Context) {
	fill, text := c.Theme().Accent, c.Theme().AccentText
	t := c.Theme()
	bar := func(w float32) {
		ui.Row(c).Height(stageRow).Padding(0, 10).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Size(w, 6).Radius(3).Background(t.Border)
		})
	}
	menuCard(c, ui.Column(c).Width(parentWidth), 0).Children(func() {
		bar(52)
		bar(68)
		for _, item := range m.previewTopItems(m.previewCtx) {
			ui.Row(c).Height(stageRow).Gap(5).Padding(0, 6).AlignItems(ui.Center).Children(func() {
				m.entryMark(c, item.id, 13, t.TextMuted)
				ui.Text(c, item.label).FontSize(10).SingleLine()
			})
		}
		ui.Row(c).Height(stageRow).Gap(4).Padding(0, 6, 0, 10).AlignItems(ui.Center).Radius(5).Background(fill).Children(func() {
			ui.Text(c, locales.AppName()).FontSize(sizeNote).FontWeight(600).TextColor(text).Grow(1)
			glyph(c, iconChevron, 11, text)
		})
		bar(60)
		bar(42)
		bar(56)
	})
}

// submenu draws the app's submenu as Finder shows it in the chosen place, with
// a line between groups. A row lights up under the pointer, as a menu's rows do.
func (m *app) submenu(c *ui.Context, groups [][]previewItem) {
	fill, text := c.Theme().Accent, c.Theme().AccentText
	t := c.Theme()
	menuCard(c, ui.Column(c).Width(subWidth).Margin(stageSubTop, 0, 0, 0), 12).Children(func() {
		for g, items := range groups {
			if g > 0 {
				ui.Divider(c).Margin(4, 6)
			}
			for _, it := range items {
				var icon, label ui.Element
				row := ui.Row(c).Height(stageRow).Gap(8).Padding(0, 8).AlignItems(ui.Center).Radius(5).Children(func() {
					icon = m.entryMark(c, it.id, 15, t.TextMuted)
					label = ui.Text(c, it.label).FontSize(sizeBody).TextColor(t.Text)
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

// sharedWhen returns the condition that most of the entries in rows share, when
// two or more share one, and "" when none do. The group shows that condition once
// above its rows, so a row shows its own condition only when it differs.
func (m *app) sharedWhen(rows []int) string {
	counts := map[string]int{}
	best, most := "", 1
	for _, i := range rows {
		when := menuEntryFor(m.cfg.MenuItems[i].ID).when
		counts[when]++
		if counts[when] > most {
			best, most = when, counts[when]
		}
	}
	return best
}

// menuGroup builds one group of the menu editor: its entries, each with a switch.
// A row can be dragged only within its group.
func (m *app) menuGroup(c *ui.Context, g int, rows []int) {
	t := c.Theme()
	list := &m.menuLists[max(0, g)]
	list.Key = func(i int) any { return m.cfg.MenuItems[rows[i]].ID }
	list.Reorder = nil
	if strings.TrimSpace(m.menuSearch) == "" {
		list.Reorder = func(rows []int, to int) { m.reorderGroup(g, rows, to) }
	}
	ui.List(c, list, len(rows), func(i int) {
		item := &m.cfg.MenuItems[rows[i]]
		entry := m.menuEntry(*item)
		labelColor, iconColor := t.Text, t.TextMuted
		if !item.Enabled {
			labelColor, iconColor = t.TextMuted, t.TextMuted.Alpha(0.6)
		}
		buildRow := func(height float32) {
			row := ui.Row(c).Height(height).Padding(0, 8).AlignItems(ui.Center).Radius(5)
			row.Children(func() {
				ui.Row(c).Gap(9).Grow(1).MinWidth(0).AlignItems(ui.Center).Children(func() {
					if strings.TrimSpace(m.menuSearch) == "" {
						gripColor := ui.Transparent
						if row.Hovered() {
							gripColor = t.TextMuted.Alpha(.45)
						}
						ui.Row(c).Size(gripWidth, 14).Shrink(0).AlignItems(ui.Center).Children(func() {
							ui.Icon(c, gripMark).Size(8.6*gripScale, 14.6*gripScale).Shrink(0).TextColor(gripColor)
						})
					} else {
						ui.Row(c).Size(gripWidth, 14).Shrink(0)
					}
					m.entryMark(c, item.ID, 16, iconColor)
					ui.Text(c, entry.label).FontSize(sizeBody).TextColor(labelColor).SingleLine().Grow(1).MinWidth(0)
					// The condition keeps its own column, so every entry stays on one
					// line; a group's shared condition stays in its header.
					ui.Text(c, entry.when).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine().Width(menuWhenWidth).Shrink(0)
				})
				ui.Row(c).Width(118).Shrink(0).AlignItems(ui.Center).Children(func() {
					if item.ID == "new-template" || item.ID == "new-folder" {
						ui.Text(c, locales.MenuNewFileInside()).FontSize(sizeCaption).TextColor(t.TextMuted)
					} else {
						locationSelect(c, &item.Location, "action:"+item.ID, locales.LocationLabel(entry.label),
							locationValues(), locationLabels(), item.Enabled, m.saveConfig)
					}
				})
				ui.Row(c).Width(48).Shrink(0).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
					creationSwitch(c, &item.Enabled, locales.EnableLabel(entry.label)).OnChange(m.saveConfig)
				})
			})
			if row.Hovered() {
				row.Background(t.SurfaceHover)
			}
			if row.RowDoubleClicked() {
				m.menuNameID, m.menuName, m.menuNameOpen = item.ID, item.Label, true
			}
		}
		// The first submenu entry after the main entries gets a rule above it. The rule
		// takes one point of the row's height, so the rows keep their pitch.
		if i > 0 && item.Location != config.FileTypeMain && m.cfg.MenuItems[rows[i-1]].Location == config.FileTypeMain {
			ui.Column(c).Height(menuRowHeight).Children(func() {
				ui.Row(c).Height(1).Background(separatorColor(c))
				buildRow(menuRowHeight - 1)
			})
			return
		}
		buildRow(menuRowHeight)
	}).Height(float32(len(rows)) * menuRowHeight)
}
