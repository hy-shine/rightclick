package settings

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
	"rightclick/internal/catalog"
)

const (
	newModeFileTypes = iota
	newModeDocuments
	newModeCode
	newModeTemplates
)

func (m *app) newPane(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).AlignItems(ui.Center).Children(func() {
		paneHeader(c, "新建", "在当前文件夹，新建你需要的文件。")
		ui.Spacer(c)
		if primaryButton(c, iconPlus, "添加模板…") {
			m.pickTemplate()
		}
	})
	ui.Row(c).Gap(12).Margin(22, 0, 14, 0).AlignItems(ui.Center).Children(func() {
		ui.Segmented(c.Key("creation-filter"), &m.newMode, "全部", "文档", "代码", "我的模板").Label("文件分类")
		ui.Spacer(c)
		ui.TextInput(c.Key("creation-search"), &m.creationSearch).Label("搜索文件类型").Placeholder("搜索类型或格式").Width(180)
	})
	m.creationRows = m.filteredCreationRows()
	columns := []ui.TableColumn{
		{Title: "文件类型", ID: "type", MinWidth: 190, Fixed: true},
		{Title: "格式", ID: "extension", Width: 66, Fixed: true},
		{Title: "菜单位置", ID: "location", Width: 160, Fixed: true},
		{Title: "启用", ID: "enabled", Width: 56, Align: ui.Center, Fixed: true},
		{Title: "", ID: "edit", Width: 42, Fixed: true},
	}
	width, _ := c.Size()
	if width >= 1050 {
		columns = slices.Insert(columns, 2, ui.TableColumn{Title: "创建内容", ID: "source", Width: 130, Fixed: true})
	}
	tableTheme := *t
	tableTheme.Accent, tableTheme.AccentText = t.SurfaceHover, t.Text
	c.SetTheme(&tableTheme)
	table := ui.Table(c.Key("creation-table"), &m.fileTypeList, columns, len(m.creationRows), func(row, column int) {
		m.creationCell(c, m.creationRows[row], columns[column].ID)
	}).Grow(1).Border(1, t.Border).Radius(7).Label("新建文件类型")
	c.SetTheme(t)
	if table.Submitted() && m.creationSel >= 0 && m.creationSel < len(m.creationRows) {
		m.openCreationEditor(m.creationRows[m.creationSel])
	}
	if len(m.creationRows) == 0 {
		ui.Text(c, "没有找到相关类型").FontSize(sizeNote).TextColor(t.TextMuted).Padding(12, 0)
	}
	ui.Row(c).Margin(10, 0, 16, 0).Children(func() {
		ui.Text(c, fmt.Sprintf("%d 种内容", len(m.creationRows))).FontSize(sizeCaption).TextColor(t.TextMuted)
		ui.Spacer(c)
		ui.Text(c, "拖动列表调整顺序").FontSize(sizeCaption).TextColor(t.TextMuted)
	})
	ui.Divider(c)
	ui.Row(c).Margin(14, 0, 0, 0).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Gap(3).Children(func() {
			ui.Text(c, "新建后打开文件").FontSize(sizeNote).FontWeight(600)
			ui.Text(c, "使用该格式的默认应用打开。").FontSize(sizeCaption).TextColor(t.TextMuted)
		})
		ui.Spacer(c)
		ui.Switch(c.Key("open-after-create"), &m.cfg.OpenAfterCreate).Label("新建后打开文件").OnChange(m.saveConfig)
	})
}

func (m *app) creationCell(c *ui.Context, row creationRow, column string) {
	fileType := m.creationType(row)
	t := c.Theme()
	switch column {
	case "type":
		ui.Row(c).Height(28).Gap(9).AlignItems(ui.Center).Children(func() {
			glyph(c, iconGrip, 11, t.TextMuted.Alpha(.45))
			creationFileIcon(c, fileType.Extension, 23)
			color := t.Text
			if !fileType.Enabled {
				color = t.TextMuted
			}
			ui.Text(c, fileType.Label).FontSize(sizeNote).TextColor(color).SingleLine()
		})
	case "extension":
		ui.Text(c, "."+fileType.Extension).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine()
	case "source":
		source := "空白文件"
		if fileType.Template != "" {
			source = "空白文档"
		}
		if fileType.Content != "" {
			source = "自定义内容"
		}
		if fileType.TemplateName != "" {
			source = fileType.TemplateName
		}
		ui.Text(c, source).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine()
	case "location":
		if row.template {
			template := &m.templates[row.index]
			locationSelect(c, &template.Location, row.key, "菜单位置："+fileType.Label, []string{config.FileTypeSubmenu, config.FileTypeMain}, []string{"新建文件子菜单", "Finder 右键顶层"}, !template.Disabled, m.saveTemplates)
		} else {
			locationSelect(c, &m.cfg.FileTypes[row.index].Location, row.key, "显示位置："+fileType.Label, []string{config.FileTypeSubmenu, config.FileTypeMain}, []string{"新建文件子菜单", "Finder 右键顶层"}, fileType.Enabled, m.saveConfig)
		}
	case "enabled":
		if row.template {
			enabled := !m.templates[row.index].Disabled
			creationSwitch(c.Key("enabled:"+row.key), &enabled, "启用 "+fileType.Label).OnClick(func() { m.templates[row.index].Disabled = !enabled; m.saveTemplates() })
		} else {
			creationSwitch(c.Key("enabled:"+row.key), &m.cfg.FileTypes[row.index].Enabled, "启用 "+fileType.Label).OnChange(m.saveConfig)
		}
	case "edit":
		ui.ButtonBase(c.Key("edit:"+row.key)).Size(24, 24).Label("编辑 " + fileType.Label).Center().Children(func() { ui.Text(c, "…").FontSize(sizeBody).TextColor(t.TextMuted) }).OnClick(func() { m.openCreationEditor(row) })
	}
}

func creationSwitch(c *ui.Context, enabled *bool, label string) ui.Element {
	theme := c.Theme()
	accent := *theme
	accent.Accent, accent.AccentHover, accent.AccentPressed, accent.AccentText = brandColors(theme)
	c.SetTheme(&accent)
	control := ui.Switch(c, enabled).Label(label)
	c.SetTheme(theme)
	return control
}

func locationSelect(c *ui.Context, value *string, key, label string, values, labels []string, enabled bool, save func()) {
	t := c.Theme()
	selected := slices.Index(values, *value)
	if selected < 0 {
		selected = 0
	}
	parts := ui.SelectBase(c.Key("location:"+key), value)
	parts.Trigger.Width(145).Padding(5, 6).Label(label).Value(labels[selected]).TextColor(t.Text).Disabled(!enabled).OnChange(save).Children(func() {
		ui.Text(c, labels[selected]).FontSize(sizeCaption).TextColor(t.Text).SingleLine()
		ui.Spacer(c)
		glyph(c, iconChevron, 9, t.TextMuted)
	})
	parts.Popup(func(panel ui.Element) {
		panel.Padding(4).Radius(7).Background(t.Surface).Border(1, t.Border)
		for index, location := range values {
			item := parts.Item(location).Padding(7, 9).Radius(4)
			color := t.Text
			if item.Highlighted() {
				item.Background(t.Accent)
				color = t.AccentText
			}
			item.Children(func() { ui.Text(c, labels[index]).FontSize(sizeNote).TextColor(color) })
		}
	})
}

var creationIcons = make(map[string]*ui.SVG)

func creationFileIcon(c *ui.Context, extension string, size float32) {
	svg := creationIcons[extension]
	if svg == nil {
		parsed, err := ui.ParseSVG(catalog.Icon(extension))
		if err != nil {
			glyph(c, iconText, size, c.Theme().TextMuted)
			return
		}
		svg = parsed
		creationIcons[extension] = svg
	}
	if extension == "txt" || extension == "md" || extension == "rtf" {
		ui.Icon(c, svg).Size(size, size).TextColor(c.Theme().Text)
		return
	}
	ui.Image(c, svg).Size(size, size)
}

func (m *app) creationDialog(c *ui.Context) {
	ui.DialogBase(c, &m.editCreationOpen, func(backdrop, panel ui.Element) {
		t := c.Theme()
		backdrop.Background(ui.RGBA(0, 0, 0, .32))
		_, height := c.Size()
		panel.Width(440).MaxHeight(height-32).Padding(22).Background(t.Surface).Border(1, t.Border).Radius(12).Label("编辑文件类型")
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				creationFileIcon(c, m.editCreation.Extension, 32)
				ui.Column(c).Gap(3).Grow(1).Children(func() {
					ui.Text(c, m.editCreation.Label).FontSize(sizeSection).FontWeight(600)
					ui.Text(c, "."+m.editCreation.Extension).FontSize(sizeCaption).TextColor(t.TextMuted)
				})
				if ghostButton(c, "关闭", false) {
					m.editCreationOpen = false
				}
			})
			ui.Scroll(c).MaxHeight(height - 190).Children(func() {
				ui.Column(c).Gap(9).Children(func() {
					ui.Text(c, "菜单名称").FontSize(sizeCaption).FontWeight(600)
					ui.TextInput(c.Key("creation-name"), &m.editCreation.Label).Label("菜单名称")
					ui.Text(c, "默认文件名").FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.TextInput(c.Key("creation-filename"), &m.editCreation.DefaultName).Label("默认文件名").Grow(1)
						ui.Text(c, "."+m.editCreation.Extension).FontSize(sizeCaption).TextColor(t.TextMuted)
					})
					if m.editCreation.Template != "" {
						ui.Text(c, "创建模板").FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
						ui.Row(c).Gap(10).Padding(10).Border(1, t.Border).Radius(6).AlignItems(ui.Center).Children(func() {
							creationFileIcon(c, m.editCreation.Extension, 24)
							name := m.editCreation.TemplateName
							if name == "" {
								name = "空白文档." + m.editCreation.Extension
							}
							ui.Text(c, name).FontSize(sizeNote).Grow(1).SingleLine()
							if quietButton(c, nil, "更换模板…") {
								m.replaceCreationTemplate()
							}
						})
						ui.Text(c, "保留文档原始格式，内容使用对应应用编辑。").FontSize(sizeCaption).TextColor(t.TextMuted)
					} else {
						ui.Text(c, "初始内容").FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
						ui.TextArea(c.Key("creation-content"), &m.editCreation.Content).Label("初始内容").Font("Menlo").FontSize(12).Height(110)
					}
					ui.Text(c, "菜单位置").FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
					locationSelect(c, &m.editCreation.Location, "draft", "菜单位置", []string{config.FileTypeSubmenu, config.FileTypeMain}, []string{"新建文件子菜单", "Finder 右键顶层"}, true, func() {})
					ui.Divider(c).Margin(6, 0)
					ui.Row(c).AlignItems(ui.Center).Children(func() {
						ui.Column(c).Gap(4).Grow(1).Children(func() {
							ui.Text(c, "启用此类型").FontSize(sizeNote).FontWeight(600)
							ui.Text(c, "关闭后不显示在菜单中，原设置会保留。").FontSize(sizeCaption).TextColor(t.TextMuted)
						})
						creationSwitch(c.Key("creation-enabled"), &m.editCreation.Enabled, "启用此文件类型")
					})
					if m.editCreationNotice != "" {
						ui.Text(c, m.editCreationNotice).FontSize(sizeCaption).TextColor(t.Danger)
					}
				})
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if m.editCreationTemplate && ghostButton(c, "删除模板", true) {
					m.tplSel = m.editCreationIndex
					m.confirmDeleteTpl = true
				}
				ui.Spacer(c)
				if quietButton(c, nil, "取消") {
					m.editCreationOpen = false
				}
				if primaryButton(c, nil, "保存修改") {
					m.saveCreationEditor()
				}
			})
		})
	})
	if ui.AlertDialog(c, &m.confirmDeleteTpl, "删除模板？", "删除后不能恢复。", "取消", "删除") == 1 {
		m.deleteTemplate()
		m.editCreationOpen = false
	}
}
