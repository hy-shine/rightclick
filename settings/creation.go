package settings

import (
	"encoding/base64"
	"path/filepath"
	"rightclick/locales"
	"slices"
	"strings"

	"github.com/egoist/mygo"

	"errors"
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
			if !strings.Contains(strings.ToLower(fileType.DisplayLabel()+" "+fileType.Extension), search) {
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
		fileType := m.cfg.FileTypes[row.index]
		fileType.Label, fileType.DefaultName = fileType.DisplayLabel(), fileType.DisplayDefaultName()
		return fileType
	}
	template := m.templates[row.index]
	extension := strings.TrimPrefix(filepath.Ext(template.Name), ".")
	kind := ""
	if template.Data != "" {
		kind = extension
	}
	label := template.Label
	if label == "" {
		label = strings.TrimSuffix(template.Name, filepath.Ext(template.Name))
	}
	return config.FileType{ID: row.key, Label: label, Extension: extension,
		DefaultName: strings.TrimSuffix(template.Name, filepath.Ext(template.Name)), Content: template.Content, Enabled: !template.Disabled,
		Location: config.FileTypeSubmenu, Template: kind, TemplateData: template.Data, TemplateName: template.Name}
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
		m.editCreationNotice = locales.InvalidName()
		return
	}
	if m.editCreationTemplate {
		template := m.templates[m.editCreationIndex]
		previousName := template.Name
		template.Name = draft.DefaultName + "." + draft.Extension
		template.Label = draft.Label
		for index, existing := range m.templates {
			if index != m.editCreationIndex && existing.Name == template.Name {
				m.editCreationNotice = locales.TemplateNameExists()
				return
			}
		}
		template.Content, template.Data, template.Disabled, template.Location = draft.Content, draft.TemplateData, !draft.Enabled, config.FileTypeSubmenu
		updated := slices.Clone(m.templates)
		updated[m.editCreationIndex] = template
		if err := files.SaveTemplates(m.store.TemplatesPath, updated); err != nil {
			m.editCreationNotice = err.Error()
			return
		}
		m.templates = updated
		if previousName != template.Name {
			for index := range m.cfg.MenuItems {
				if m.cfg.MenuItems[index].ID == "template:"+previousName {
					m.cfg.MenuItems[index].ID = "template:" + template.Name
				}
			}
			m.saveConfig()
		}
	} else {
		draft.Location = config.FileTypeSubmenu
		updated := slices.Clone(m.cfg.FileTypes)
		updated[m.editCreationIndex] = draft.Stored()
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
		paths, err := open(mygo.OpenDialogOptions{Parent: win, Title: locales.ReplaceTemplate(), Filters: []mygo.FileFilter{{Name: locales.FilterDocuments(), Extensions: []string{extension}}}})
		if err == nil && len(paths) == 0 {
			return
		}
		var template files.Template
		if err == nil {
			template, err = files.TemplateFromFile(paths[0])
			if err == nil && strings.TrimPrefix(filepath.Ext(template.Name), ".") != extension {
				err = errors.New(locales.ChooseSameFormat())
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

// hasNewFileContents reports whether the New File submenu has anything to show.
func (m *app) hasNewFileContents() bool {
	if slices.ContainsFunc(m.cfg.FileTypes, func(fileType config.FileType) bool {
		return fileType.Enabled
	}) {
		return true
	}
	if !slices.ContainsFunc(m.cfg.MenuItems, func(item config.MenuItem) bool {
		return item.ID == "new-template" && item.Enabled
	}) {
		return false
	}
	return slices.ContainsFunc(m.templates, func(template files.Template) bool {
		return !template.Disabled
	})
}

func (m *app) pickTemplate() {
	win, open := m.win, m.env.openDialog
	go func() {
		paths, err := open(mygo.OpenDialogOptions{Parent: win, Title: locales.AddFileTemplate(), Message: locales.AddFileTemplateHint()})
		if err == nil && len(paths) == 0 {
			return
		}
		var template files.Template
		if err == nil {
			template, err = files.TemplateFromFile(paths[0])
		}
		win.Update(func() {
			if err != nil {
				m.report(locales.AddFailed(), err)
				return
			}
			m.addTemplate(template)
			m.newMode = newModeTemplates
		})
	}()
}

func (m *app) addTemplate(template files.Template) {
	if slices.ContainsFunc(m.templates, func(existing files.Template) bool { return existing.Name == template.Name }) {
		m.notice = locales.SameNameExists()
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
	name := m.templates[m.tplSel].Name
	m.templates = slices.Delete(m.templates, m.tplSel, m.tplSel+1)
	m.tplSel = min(m.tplSel, len(m.templates)-1)
	m.saveTemplates()
	if !slices.ContainsFunc(m.templates, func(template files.Template) bool { return template.Name == name }) {
		m.setShortcut("template:"+name, false)
	}
}
