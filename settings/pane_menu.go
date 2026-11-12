package settings

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
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
var previewContexts = []string{"空白处", "文件夹", "文件", "两个文件"}

// menuEntry is the label of a Finder menu entry and the condition that shows it.
type menuEntry struct{ label, when string }

// menuEntries describes the entries of config.Default, by ID.
var menuEntries = map[string]menuEntry{
	"new-file":      {"新建文件", "空白处、文件夹"},
	"new-folder":    {"文件夹", "空白处、文件夹"},
	"new-template":  {"用户模板", "空白处、文件夹"},
	"copy-path":     {"拷贝路径", "选中项"},
	"copy-name":     {"拷贝文件名", "选中项"},
	"copy-to":       {"复制到…", "选中项"},
	"move-to":       {"移动到…", "选中项"},
	"open-terminal": {"在终端打开", "空白处、文件夹"},
	"open-editor":   {"用默认编辑器打开", "任何位置"},
	"open-with":     {"用…打开", "任何位置"},
	"zip":           {"压缩为 ZIP", "选中文件"},
	"unzip":         {"解压到此处", "仅 .zip 文件"},
	"hash":          {"计算哈希", "仅一个文件"},
	"compare-hash":  {"比较两个文件的哈希", "仅两个文件"},
	"toggle-hidden": {"显示/隐藏隐藏文件（⌘⇧.）", "空白处、文件夹"},
}

// menuGroups lists the IDs of each group of the Finder submenu, in the order
// the submenu shows them. It mirrors MenuActions.groups in
// extension/RightClickFinderSync/MenuConfig.swift; change both together.
var menuGroups = [][]string{
	{"new-file", "new-folder", "new-template"},
	{"copy-path", "copy-name", "copy-to", "move-to"},
	{"open-terminal", "open-editor", "open-with"},
	{"zip", "unzip", "hash"},
	{"compare-hash"},
	{"toggle-hidden"},
}

// menuContexts lists, by entry ID, the places where Finder shows the entry. It
// mirrors allows() in extension/RightClickFinderSync/FinderSyncExtension.swift;
// change both together.
var menuContexts = map[string][]int{
	"new-file":      {ctxBlank, ctxFolder},
	"new-folder":    {ctxBlank, ctxFolder},
	"new-template":  {ctxBlank, ctxFolder},
	"copy-path":     {ctxFolder, ctxFile, ctxTwoFiles},
	"copy-name":     {ctxFolder, ctxFile, ctxTwoFiles},
	"copy-to":       {ctxFolder, ctxFile, ctxTwoFiles},
	"move-to":       {ctxFolder, ctxFile, ctxTwoFiles},
	"open-terminal": {ctxBlank, ctxFolder},
	"open-editor":   {ctxBlank, ctxFolder, ctxFile, ctxTwoFiles},
	"open-with":     {ctxBlank, ctxFolder, ctxFile, ctxTwoFiles},
	"zip":           {ctxFile, ctxTwoFiles},
	"unzip":         {ctxFile},
	"hash":          {ctxFile},
	"compare-hash":  {ctxTwoFiles},
	"toggle-hidden": {ctxBlank, ctxFolder},
}

// menuRowHeight is the height of one row of the menu editor, in points.
const menuRowHeight = 42

// menuEntryFor returns the entry for id. An unknown ID is shown as it is.
func menuEntryFor(id string) menuEntry {
	if e, ok := menuEntries[id]; ok {
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
		if slices.Contains(menuGroups[g], item.ID) {
			rows = append(rows, i)
		}
	}
	return rows
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
	m.cfg.MenuItems = config.Default().MenuItems
	m.saveConfig()
}

// previewItem is one entry of the menu preview.
type previewItem struct{ id, label string }

// previewGroups returns the entries that the preview shows in place ctx: the
// enabled entries that apply there, one slice per group that has any, in menu
// order.
func (m *app) previewGroups(ctx int) [][]previewItem {
	var groups [][]previewItem
	for g := range menuGroups {
		var items []previewItem
		for _, i := range m.groupRows(g) {
			item := m.cfg.MenuItems[i]
			if !item.Enabled || item.Location == config.FileTypeMain || !appliesTo(item.ID, ctx) {
				continue
			}
			if item.ID == "new-template" {
				continue
			}
			if item.ID == "new-file" {
				if len(m.fileTypesAt(config.FileTypeSubmenu)) > 0 {
					items = append(items, previewItem{item.ID, menuEntryFor(item.ID).label})
				}
				continue
			}
			items = append(items, previewItem{item.ID, menuEntryFor(item.ID).label})
		}
		if len(items) > 0 {
			groups = append(groups, items)
		}
	}
	return groups
}

// fileTypePreviewID marks the direct file type entries in the menu drawing.
func fileTypePreviewID(id string) string {
	return "file-type:" + id
}

func (m *app) previewTopItems(ctx int) []previewItem {
	var items []previewItem
	for _, item := range m.cfg.MenuItems {
		if !item.Enabled || item.Location != config.FileTypeMain || !appliesTo(item.ID, ctx) {
			continue
		}
		if item.ID == "new-template" {
			continue
		}
		if item.ID == "new-file" && len(m.fileTypesAt(config.FileTypeSubmenu)) == 0 {
			continue
		}
		items = append(items, previewItem{item.ID, menuEntryFor(item.ID).label})
	}
	if ctx == ctxBlank || ctx == ctxFolder {
		for _, item := range m.cfg.MenuItems {
			if item.ID != "new-file" || !item.Enabled {
				continue
			}
			for _, fileType := range m.fileTypesAt(config.FileTypeMain) {
				items = append(items, previewItem{fileTypePreviewID(fileType.ID), "新建 " + fileType.Label})
			}
		}
		for _, template := range m.templates {
			if !template.Disabled && template.Location == config.FileTypeMain {
				items = append(items, previewItem{"new-template", "新建 " + template.Name})
			}
		}
	}
	return items
}

func (m *app) menuPage(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Gap(6).Children(func() {
			ui.Text(c, "右键菜单").FontSize(22).FontWeight(600)
			ui.Text(c, "常用操作放在外面，其余收进子菜单。").FontSize(sizeNote).TextColor(t.TextMuted)
		})
		if ghostButton(c, "恢复默认", false) {
			m.restoreDefaultMenu()
		}
		if quietButton(c, iconEye, "预览") {
			m.menuPreviewOpen = true
		}
	})
	m.setupCard(c)
	m.showNotice(c)
	ui.Row(c).Margin(22, 0, 8, 0).Padding(0, 8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "动作").FontSize(sizeCaption).TextColor(t.TextMuted).Grow(1)
		ui.Text(c, "显示位置").FontSize(sizeCaption).TextColor(t.TextMuted).Width(172).Shrink(0)
		ui.Text(c, "启用").FontSize(sizeCaption).TextColor(t.TextMuted).Width(56).Shrink(0).TextAlign(ui.Center)
	})
	ui.Divider(c)
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Children(func() {
			for group := range menuGroups {
				rows := m.groupRows(group)
				if len(rows) == 0 {
					continue
				}
				if group < 4 {
					if group > 0 {
						ui.Divider(c).Margin(8, 0, 0, 0)
					}
					ui.Text(c, []string{"新建", "路径与移动", "打开应用", "文件工具"}[group]).FontSize(sizeCaption).TextColor(t.TextMuted).Padding(14, 8, 5, 8)
				}
				m.menuGroup(c, group, rows)
			}
		})
	})
}

func (m *app) menuPreviewDialog(c *ui.Context) {
	ui.DialogBase(c, &m.menuPreviewOpen, func(backdrop, panel ui.Element) {
		t := c.Theme()
		_, height := c.Size()
		backdrop.Background(ui.RGBA(0, 0, 0, .3))
		panel.Width(520).MaxHeight(height-32).Padding(22).Background(t.Surface).Border(1, t.Border).Radius(10).Label("菜单预览")
		ui.Column(c).Gap(18).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "菜单预览").FontSize(sizeSection).FontWeight(600)
				ui.Spacer(c)
				if ghostButton(c, "关闭预览", false) {
					m.menuPreviewOpen = false
				}
			})
			ui.Scroll(c).MaxHeight(height - 130).Children(func() { m.menuStage(c) })
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
	stageGap    = 6   // space between the parent menu and the submenu
	stageWidth  = 360 // the preview column: the place switch and the drawing both fit
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
			ui.Text(c, fmt.Sprintf("显示 %d 项", n)).FontSize(sizeNote).TextColor(t.TextMuted).SingleLine()
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
	ui.Row(c).Gap(2).Padding(3).Radius(t.Radius+4).Background(t.Background).Border(1, t.Border).Role(ui.RoleRadioGroup).Label("右键位置").Children(func() {
		for i, name := range previewContexts {
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
	fill, _, _, text := brandColors(c.Theme())
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
				if fileTypeID, ok := strings.CutPrefix(item.id, "file-type:"); ok {
					for _, fileType := range m.cfg.FileTypes {
						if fileType.ID == fileTypeID {
							creationFileIcon(c, fileType.Extension, 15)
						}
					}
				} else {
					glyph(c, entryIcon(item.id), 13, t.TextMuted)
				}
				ui.Text(c, item.label).FontSize(10).SingleLine()
			})
		}
		ui.Row(c).Height(stageRow).Gap(4).Padding(0, 6, 0, 10).AlignItems(ui.Center).Radius(5).Background(fill).Children(func() {
			ui.Text(c, "右键助手").FontSize(sizeNote).FontWeight(600).TextColor(text).Grow(1)
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
	fill, _, _, text := brandColors(c.Theme())
	t := c.Theme()
	menuCard(c, ui.Column(c).Width(subWidth).Margin(stageSubTop, 0, 0, 0), 12).Children(func() {
		for g, items := range groups {
			if g > 0 {
				ui.Divider(c).Margin(4, 6)
			}
			for _, it := range items {
				var icon, label ui.Element
				row := ui.Row(c).Height(stageRow).Gap(8).Padding(0, 8).AlignItems(ui.Center).Radius(5).Children(func() {
					icon = ui.Icon(c, entryIcon(it.id)).Size(15, 15).
						TextColor(typePreviewColor(c, it.id, t.TextMuted))
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
			ui.Text(c, "这里没有可用的项。").FontSize(sizeNote).TextColor(t.TextMuted).Padding(8, 8)
		}
	})
}

// typePreviewColor uses the same colors as the file type list when the preview
// row is a direct file type.
func typePreviewColor(c *ui.Context, id string, fallback ui.Color) ui.Color {
	if fileTypeID, ok := strings.CutPrefix(id, "file-type:"); ok {
		return themedFileTypeColor(c, fileTypeID, fallback)
	}
	return fallback
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
	ui.List(c, &m.menuLists[g], len(rows), func(i int) {
		item := &m.cfg.MenuItems[rows[i]]
		entry := menuEntryFor(item.ID)
		labelColor, iconColor := t.Text, t.TextMuted
		if !item.Enabled {
			labelColor, iconColor = t.TextMuted, t.TextMuted.Alpha(0.6)
		}
		row := ui.Row(c).Height(menuRowHeight).Padding(0, 8).AlignItems(ui.Center).Radius(5)
		row.Children(func() {
			ui.Row(c).Gap(9).Grow(1).MinWidth(0).AlignItems(ui.Center).Children(func() {
				gripColor := ui.Transparent
				if row.Hovered() {
					gripColor = t.TextMuted.Alpha(.45)
				}
				ui.Icon(c, iconGrip).Size(10, 14).Shrink(0).TextColor(gripColor)
				ui.Icon(c, entryIcon(item.ID)).Size(17, 17).Shrink(0).TextColor(iconColor)
				ui.Text(c, entry.label).FontSize(sizeBody).TextColor(labelColor).SingleLine().MinWidth(0).Grow(1).Tooltip(entry.label + " · " + entry.when)
			})
			ui.Row(c).Width(172).Shrink(0).AlignItems(ui.Center).Children(func() {
				locationSelect(c, &item.Location, "action:"+item.ID, "菜单位置："+entry.label,
					[]string{config.FileTypeSubmenu, config.FileTypeMain}, []string{"右键助手内", "Finder 顶层"}, item.Enabled, m.saveConfig)
			})
			ui.Row(c).Width(56).Shrink(0).Justify(ui.Center).AlignItems(ui.Center).Children(func() {
				creationSwitch(c, &item.Enabled, "启用 "+entry.label).OnChange(m.saveConfig)
			})
		})
		if row.Hovered() {
			row.Background(t.SurfaceHover)
		}
	}).Height(float32(len(rows)) * menuRowHeight)
}
