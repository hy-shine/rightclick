package settings

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo"

	"rightclick/config"
	"rightclick/files"
)

type creationRow struct {
	key      string
	index    int
	template bool
}

func (m *app) filteredCreationRows() []creationRow {
	search := strings.ToLower(strings.TrimSpace(m.creationSearch))
	var rows []creationRow
	if m.newMode != newModeTemplates {
		for index, fileType := range m.cfg.FileTypes {
			code := isCodeExtension(fileType.Extension)
			if m.newMode == newModeDocuments && code || m.newMode == newModeCode && !code {
				continue
			}
			if !strings.Contains(strings.ToLower(fileType.Label+" "+fileType.Extension), search) {
				continue
			}
			rows = append(rows, creationRow{key: "type:" + fileType.ID, index: index})
		}
	}
	if m.newMode == newModeFileTypes || m.newMode == newModeTemplates {
		for index, template := range m.templates {
			if strings.Contains(strings.ToLower(template.Name), search) {
				rows = append(rows, creationRow{key: "template:" + template.Name, index: index, template: true})
			}
		}
	}
	return rows
}

func isCodeExtension(extension string) bool {
	return slices.Contains([]string{"json", "html", "css", "js", "swift", "py", "go", "rs", "java", "c", "cpp", "sql", "ts", "yaml", "yml", "sh"}, extension)
}

func (m *app) creationType(row creationRow) config.FileType {
	if !row.template {
		return m.cfg.FileTypes[row.index]
	}
	template := m.templates[row.index]
	extension := strings.TrimPrefix(filepath.Ext(template.Name), ".")
	kind := ""
	if template.Data != "" {
		kind = extension
	}
	location := template.Location
	if location != config.FileTypeMain {
		location = config.FileTypeSubmenu
	}
	label := template.Label
	if label == "" {
		label = strings.TrimSuffix(template.Name, filepath.Ext(template.Name))
	}
	return config.FileType{ID: row.key, Label: label, Extension: extension,
		DefaultName: strings.TrimSuffix(template.Name, filepath.Ext(template.Name)), Content: template.Content, Enabled: !template.Disabled,
		Location: location, Template: kind, TemplateData: template.Data, TemplateName: template.Name}
}

func (m *app) openCreationEditor(row creationRow) {
	m.editCreation = m.creationType(row)
	m.editCreationIndex, m.editCreationTemplate = row.index, row.template
	m.editCreationNotice, m.editCreationOpen = "", true
}

func (m *app) saveCreationEditor() {
	draft := m.editCreation
	draft.Label, draft.DefaultName = strings.TrimSpace(draft.Label), strings.TrimSpace(draft.DefaultName)
	if draft.Label == "" || draft.DefaultName == "" || strings.ContainsAny(draft.DefaultName, "/\\\x00") || draft.DefaultName == "." || draft.DefaultName == ".." {
		m.editCreationNotice = "请输入有效名称，文件名中不能包含路径分隔符。"
		return
	}
	if m.editCreationTemplate {
		template := m.templates[m.editCreationIndex]
		template.Name = draft.DefaultName + "." + draft.Extension
		template.Label = draft.Label
		for index, existing := range m.templates {
			if index != m.editCreationIndex && existing.Name == template.Name {
				m.editCreationNotice = "模板名称已存在"
				return
			}
		}
		template.Content, template.Data, template.Disabled, template.Location = draft.Content, draft.TemplateData, !draft.Enabled, draft.Location
		updated := slices.Clone(m.templates)
		updated[m.editCreationIndex] = template
		if err := files.SaveTemplates(m.store.TemplatesPath, updated); err != nil {
			m.editCreationNotice = err.Error()
			return
		}
		m.templates = updated
	} else {
		updated := slices.Clone(m.cfg.FileTypes)
		updated[m.editCreationIndex] = draft
		cfg := m.cfg
		cfg.FileTypes = updated
		if err := config.Save(m.store.ConfigPath, cfg); err != nil {
			m.editCreationNotice = err.Error()
			return
		}
		m.cfg = cfg
	}
	m.editCreationOpen = false
}

func (m *app) replaceCreationTemplate() {
	win, open, extension, index, custom := m.win, m.env.openDialog, m.editCreation.Extension, m.editCreationIndex, m.editCreationTemplate
	go func() {
		paths, err := open(mygo.OpenDialogOptions{Parent: win, Title: "更换文档模板", Filters: []mygo.FileFilter{{Name: "文档", Extensions: []string{extension}}}})
		if err == nil && len(paths) == 0 {
			return
		}
		var template files.Template
		if err == nil {
			template, err = files.TemplateFromFile(paths[0])
			if err == nil && strings.TrimPrefix(filepath.Ext(template.Name), ".") != extension {
				err = fmt.Errorf("请选择相同格式的模板")
			}
		}
		win.Update(func() {
			if !m.editCreationOpen || m.editCreationIndex != index || m.editCreationTemplate != custom {
				return
			}
			if err != nil {
				m.editCreationNotice = err.Error()
				return
			}
			data, decodeErr := template.Bytes()
			if decodeErr != nil {
				m.editCreationNotice = decodeErr.Error()
				return
			}
			m.editCreation.TemplateData, m.editCreation.TemplateName = base64.StdEncoding.EncodeToString(data), template.Name
		})
	}()
}

func (m *app) reorderCreationRows(rows []int, to int) {
	if len(rows) == 0 {
		return
	}
	custom := m.creationRows[rows[0]].template
	var positions, moving []int
	for index, row := range m.creationRows {
		if row.template == custom {
			positions = append(positions, row.index)
			if slices.Contains(rows, index) {
				moving = append(moving, len(positions)-1)
			}
		}
	}
	target := 0
	for index, row := range m.creationRows {
		if index < to && row.template == custom {
			target++
		}
	}
	if custom {
		entries := make([]files.Template, len(positions))
		for index, position := range positions {
			entries[index] = m.templates[position]
		}
		entries = moveRows(entries, moving, target)
		for index, position := range positions {
			m.templates[position] = entries[index]
		}
		m.saveTemplates()
	} else {
		entries := make([]config.FileType, len(positions))
		for index, position := range positions {
			entries[index] = m.cfg.FileTypes[position]
		}
		entries = moveRows(entries, moving, target)
		for index, position := range positions {
			m.cfg.FileTypes[position] = entries[index]
		}
		m.saveConfig()
	}
}

func (m *app) reorderFileTypes(rows []int, to int) {
	selectedID := m.cfg.FileTypes[m.fileTypeSel].ID
	m.cfg.FileTypes = moveRows(m.cfg.FileTypes, rows, to)
	m.fileTypeSel = slices.IndexFunc(m.cfg.FileTypes, func(fileType config.FileType) bool { return fileType.ID == selectedID })
	m.saveConfig()
}

func (m *app) fileTypesAt(location string) []config.FileType {
	var types []config.FileType
	for _, fileType := range m.cfg.FileTypes {
		if fileType.Enabled && fileType.Location == location {
			types = append(types, fileType)
		}
	}
	return types
}

func (m *app) pickTemplate() {
	win, open := m.win, m.env.openDialog
	go func() {
		paths, err := open(mygo.OpenDialogOptions{Parent: win, Title: "添加文件模板", Message: "支持文本、Word、Excel、PowerPoint 和 RTF 文档。"})
		if err == nil && len(paths) == 0 {
			return
		}
		var template files.Template
		if err == nil {
			template, err = files.TemplateFromFile(paths[0])
		}
		win.Update(func() {
			if err != nil {
				m.report("添加失败", err)
				return
			}
			m.addTemplate(template)
			m.newMode = newModeTemplates
		})
	}()
}

func (m *app) addTemplate(template files.Template) {
	if slices.ContainsFunc(m.templates, func(existing files.Template) bool { return existing.Name == template.Name }) {
		m.notice = "同名模板已存在，请在编辑中更换模板。"
		return
	}
	m.templates = append(m.templates, template)
	m.tplSel = len(m.templates) - 1
	m.saveTemplates()
}

func (m *app) deleteTemplate() {
	if m.tplSel < 0 || m.tplSel >= len(m.templates) {
		return
	}
	m.templates = slices.Delete(m.templates, m.tplSel, m.tplSel+1)
	m.tplSel = min(m.tplSel, len(m.templates)-1)
	m.saveTemplates()
}
