package settings

import (
	"github.com/hy-shine/rightclick/locales"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/catalog"
)

const (
	newModeFileTypes = iota
	newModeDocuments
	newModeCode
	newModeTemplates
)

func (m *app) newPane(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() { paneHeader(c, locales.PaneNew(), locales.NewIntro()) })
		if quietButtonCentered(c, nil, locales.AddTemplate()) {
			m.pickTemplate()
		}
		if primaryButtonCentered(c, iconPlus, locales.AddFileType()) {
			m.newFileType()
		}
	})
	ui.Row(c).Gap(12).Margin(20, 0, 12, 0).AlignItems(ui.Center).Children(func() {
		ui.Segmented(c.Key("creation-filter"), &m.newMode, locales.FilterAll(), locales.FilterDocuments(), locales.FilterCode(), locales.FilterMyTemplates()).Label(locales.FileCategory())
		ui.Spacer(c)
		ui.TextInput(c.Key("creation-search"), &m.creationSearch).Label(locales.SearchFileTypes()).Placeholder(locales.SearchPlaceholder()).Width(180)
	})
	m.creationRows = m.filteredCreationRows()
	columns := []ui.TableColumn{
		{Title: locales.ColFileType(), ID: "type", Width: 240, Fixed: true},
		{Title: locales.ColFormat(), ID: "extension", Width: 80, Fixed: true},
		{Title: locales.ColContent(), ID: "source", MinWidth: 140, Fixed: true},
		{Title: locales.ColEnabled(), ID: "enabled", Width: 68, Align: ui.Center, Fixed: true},
		{Title: locales.QuickAccess(), ID: "shortcut", Width: 68, Align: ui.Center, Fixed: true},
		{Title: "", ID: "edit", Width: 52, Fixed: true},
	}
	tableTheme := *t
	tableTheme.Accent, tableTheme.AccentText = selectionColors(t)
	c.SetTheme(&tableTheme)
	table := ui.Table(c.Key("creation-table"), &m.fileTypeList, columns, len(m.creationRows), func(row, column int) {
		c.SetTheme(t)
		m.creationCell(c, m.creationRows[row], columns[column].ID, row == m.creationSel)
		c.SetTheme(&tableTheme)
	}).Grow(1).Border(1, separatorColor(c)).Radius(8).Label(locales.FileTypesList())
	c.SetTheme(t)
	if table.Submitted() && m.creationSel >= 0 && m.creationSel < len(m.creationRows) {
		m.openCreationEditor(m.creationRows[m.creationSel])
	}
	if len(m.creationRows) == 0 {
		ui.Text(c, locales.NoMatches()).FontSize(sizeNote).TextColor(t.TextMuted).Padding(12, 0)
	}
	ui.Row(c).Margin(10, 0, 16, 0).Children(func() {
		ui.Text(c, locales.TypeCount(len(m.creationRows))).FontSize(sizeCaption).TextColor(t.TextMuted)
		ui.Spacer(c)
		ui.Text(c, locales.CreationListHint()).FontSize(sizeCaption).TextColor(t.TextMuted)
	})
	ui.Divider(c)
	ui.Row(c).Margin(14, 0, 0, 0).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Gap(3).Children(func() {
			ui.Text(c, locales.OpenAfterCreate()).FontSize(sizeNote).FontWeight(600)
			ui.Text(c, locales.OpenAfterCreateHint()).FontSize(sizeCaption).TextColor(t.TextMuted)
		})
		ui.Spacer(c)
		ui.Switch(c.Key("open-after-create"), &m.cfg.OpenAfterCreate).Label(locales.OpenAfterCreate()).OnChange(m.saveConfig)
	})
}

func (m *app) creationCell(c *ui.Context, row creationRow, column string, selected bool) {
	fileType := m.creationType(row)
	t := c.Theme()
	secondary := t.TextMuted
	if selected {
		secondary = secondary.Mix(t.Text, 0.3)
	}
	switch column {
	case "type":
		ui.Row(c).Height(20).Gap(9).AlignItems(ui.Center).Children(func() {
			glyph(c, iconGrip, 11, t.TextMuted.Alpha(.45))
			creationFileIcon(c, fileType.Extension, 20)
			color := t.Text
			if !fileType.Enabled {
				color = secondary
			}
			ui.Text(c, fileType.Label).FontSize(sizeBody).TextColor(color).SingleLine()
		})
	case "extension":
		ui.Text(c, "."+fileType.Extension).FontSize(sizeNote).TextColor(secondary).SingleLine()
	case "source":
		source := locales.BlankFile()
		if fileType.Template != "" {
			source = locales.BlankDocument()
		}
		if fileType.Content != "" {
			source = locales.CustomContent()
		}
		if fileType.TemplateName != "" {
			source = fileType.TemplateName
		}
		ui.Text(c, source).FontSize(sizeNote).TextColor(secondary).SingleLine()
	case "enabled":
		if row.template {
			enabled := !m.templates[row.index].Disabled
			creationSwitch(c.Key("enabled:"+row.key), &enabled, locales.EnableLabel(fileType.Label)).OnClick(func() { m.templates[row.index].Disabled = !enabled; m.saveTemplates() })
		} else {
			creationSwitch(c.Key("enabled:"+row.key), &m.cfg.FileTypes[row.index].Enabled, locales.EnableLabel(fileType.Label)).OnChange(m.saveConfig)
		}
	case "shortcut":
		m.shortcutToggle(c, row.key, fileType.Label)
	case "edit":
		edit := ui.ButtonBase(c.Key("edit:"+row.key)).Size(32, 24).Radius(t.Radius).
			Label(locales.EditItem(fileType.Label)).Center()
		edit.Children(func() { glyph(c, iconEdit, 15, secondary) })
		if edit.Hovered() {
			edit.Background(t.SurfaceHover)
		}
		edit.OnClick(func() { m.openCreationEditor(row) })
	}
}

func creationSwitch(c *ui.Context, enabled *bool, label string) ui.Element {
	return ui.Switch(c, enabled).Label(label)
}

func locationSelect(c *ui.Context, value *string, key, label string, values, labels []string, enabled bool, save func()) {
	t := c.Theme()
	selected := slices.Index(values, *value)
	if selected < 0 {
		selected = 0
	}
	parts := ui.SelectBase(c.Key("location:"+key), value)
	parts.Trigger.MinWidth(100).Padding(6, 10).Gap(10).AlignItems(ui.Center).Radius(t.Radius).Background(t.Surface).Border(1, t.Border).Label(label).Value(labels[selected]).TextColor(t.Text).Disabled(!enabled).OnChange(save).Children(func() {
		ui.Text(c, labels[selected]).FontSize(sizeBody).TextColor(t.Text).SingleLine()
		glyph(c, iconChevronDown, 9, t.TextMuted)
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
		title := m.editCreation.Label
		if m.editCreationNew {
			title = locales.AddFileType()
			if m.editCreationTemplate {
				title = locales.AddFileTemplate()
			}
		}
		panel.Width(440).MaxHeight(height-32).Padding(22).Background(t.Surface).Border(1, t.Border).Radius(12).Label(locales.EditFileType())
		ui.Column(c).Gap(14).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				creationFileIcon(c, m.editCreation.Extension, 32)
				ui.Column(c).Gap(3).Grow(1).Children(func() {
					ui.Text(c, title).FontSize(sizeSection).FontWeight(600)
					ui.Text(c, "."+m.editCreation.Extension).FontSize(sizeCaption).TextColor(t.TextMuted)
				})
				if ghostButton(c, locales.Close(), false) {
					m.editCreationOpen = false
				}
			})
			ui.Scroll(c).MaxHeight(height - 190).Children(func() {
				ui.Column(c).Gap(9).Children(func() {
					ui.Text(c, locales.EditMenuName()).FontSize(sizeCaption).FontWeight(600)
					ui.TextInput(c.Key("creation-name"), &m.editCreation.Label).Label(locales.EditMenuName())
					if !m.editCreationTemplate && strings.HasPrefix(m.editCreation.ID, "custom-") && m.editCreation.TemplateData == "" {
						ui.Text(c, locales.FileExtension()).FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
						ui.TextInput(c.Key("creation-extension"), &m.editCreation.Extension).Label(locales.FileExtension()).Placeholder("yaml").OnChange(func() {
							extension := strings.TrimPrefix(strings.TrimSpace(m.editCreation.Extension), ".")
							m.editCreation.Template = ""
							if catalog.IsDocument(extension) {
								m.editCreation.Template = extension
							}
						})
					}
					ui.Text(c, locales.EditDefaultName()).FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.TextInput(c.Key("creation-filename"), &m.editCreation.DefaultName).Label(locales.EditDefaultName()).Grow(1)
						ui.Text(c, "."+m.editCreation.Extension).FontSize(sizeCaption).TextColor(t.TextMuted)
					})
					if m.editCreation.Template != "" {
						ui.Text(c, locales.EditTemplate()).FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
						ui.Row(c).Gap(10).Padding(10).Border(1, t.Border).Radius(6).AlignItems(ui.Center).Children(func() {
							creationFileIcon(c, m.editCreation.Extension, 24)
							name := m.editCreation.TemplateName
							if name == "" {
								name = locales.BlankDocument() + "." + m.editCreation.Extension
							}
							ui.Text(c, name).FontSize(sizeNote).Grow(1).SingleLine()
							if quietButton(c, nil, locales.ReplaceTemplateEllipsis()) {
								m.replaceCreationTemplate()
							}
						})
						ui.Text(c, locales.KeepFormatHint()).FontSize(sizeCaption).TextColor(t.TextMuted)
					} else {
						ui.Text(c, locales.EditInitialContent()).FontSize(sizeCaption).FontWeight(600).Margin(7, 0, 0)
						ui.TextArea(c.Key("creation-content"), &m.editCreation.Content).Label(locales.EditInitialContent()).Font("Menlo").FontSize(12).Height(110)
					}
					ui.Divider(c).Margin(6, 0)
					ui.Row(c).AlignItems(ui.Center).Children(func() {
						ui.Column(c).Gap(4).Grow(1).Children(func() {
							ui.Text(c, locales.EnableTypeTitle()).FontSize(sizeNote).FontWeight(600)
							ui.Text(c, locales.EnableTypeHint()).FontSize(sizeCaption).TextColor(t.TextMuted)
						})
						creationSwitch(c.Key("creation-enabled"), &m.editCreation.Enabled, locales.EnableTypeSwitch())
					})
					if m.editCreationNotice != "" {
						ui.Text(c, m.editCreationNotice).FontSize(sizeCaption).TextColor(t.Danger)
					}
				})
			})
			ui.Row(c).Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
				if m.editCreationTemplate && !m.editCreationNew && ghostButton(c, locales.DeleteTemplate(), true) {
					m.tplSel = m.editCreationIndex
					m.confirmDeleteTpl = true
				}
				if m.editCreationTemplate && !m.editCreationNew && ghostButton(c, locales.DuplicateTemplate(), false) {
					m.duplicateTemplate()
				}
				if !m.editCreationTemplate && !m.editCreationNew && strings.HasPrefix(m.editCreation.ID, "custom-") && ghostButton(c, locales.DeleteFileType(), true) {
					m.confirmDeleteType = true
				}
				ui.Spacer(c)
				if quietButton(c, nil, locales.Cancel()) {
					m.editCreationOpen = false
				}
				if primaryButton(c, nil, locales.SaveChanges()) {
					m.saveCreationEditor()
				}
			})
		})
	})
	if ui.AlertDialog(c, &m.confirmDeleteTpl, locales.DeleteTemplateQuestion(), locales.DeleteTemplateBody(), locales.Cancel(), locales.Delete()) == 1 {
		m.deleteTemplate()
		m.editCreationOpen = false
	}
	if ui.AlertDialog(c, &m.confirmDeleteType, locales.DeleteFileTypeQuestion(), locales.DeleteFileTypeBody(), locales.Cancel(), locales.Delete()) == 1 {
		m.deleteCustomType()
	}
}
