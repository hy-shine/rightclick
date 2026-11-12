package settings

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/hy-shine/rightclick/internal/locales"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/egoist/mygo"

	"errors"
	"github.com/hy-shine/rightclick/internal/catalog"
	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/files"
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
	return templateFileType(m.templates[row.index])
}

func templateFileType(template files.Template) config.FileType {
	extension := strings.TrimPrefix(filepath.Ext(template.Name), ".")
	base := strings.TrimSuffix(template.Name, filepath.Ext(template.Name))
	if base == "" {
		base, extension = template.Name, ""
	}
	kind := ""
	if template.Data != "" {
		kind = extension
	}
	label := template.Label
	if label == "" {
		label = base
	}
	return config.FileType{ID: "template:" + template.Name, Label: label, Extension: extension,
		DefaultName: base, Content: template.Content, Enabled: !template.Disabled,
		Location: config.FileTypeSubmenu, Template: kind, TemplateData: template.Data, TemplateName: template.Name}
}

func (m *app) openCreationEditor(row creationRow) {
	m.editCreation = m.creationType(row)
	m.editCreationNew = false
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
	if !m.editCreationTemplate {
		draft.Extension = strings.TrimPrefix(strings.TrimSpace(draft.Extension), ".")
		if !validFileExtension(draft.Extension) {
			m.editCreationNotice = locales.InvalidExtension(draft.Extension)
			return
		}
		if catalog.IsDocument(draft.Extension) && draft.Template == "" {
			draft.Template = draft.Extension
		}
		if _, err := files.FileTypeContent(draft); err != nil {
			m.editCreationNotice = err.Error()
			return
		}
	}
	if m.editCreationTemplate {
		var template files.Template
		if !m.editCreationNew {
			template = m.templates[m.editCreationIndex]
		}
		previousName := template.Name
		template.Name = draft.DefaultName
		if draft.Extension != "" {
			template.Name += "." + draft.Extension
		}
		template.Label = draft.Label
		for index, existing := range m.templates {
			if (m.editCreationNew || index != m.editCreationIndex) && strings.EqualFold(existing.Name, template.Name) {
				m.editCreationNotice = locales.TemplateNameExists()
				return
			}
		}
		template.Content, template.Data, template.Disabled, template.Location = draft.Content, draft.TemplateData, !draft.Enabled, config.FileTypeSubmenu
		updated := slices.Clone(m.templates)
		if m.editCreationNew {
			updated = append(updated, template)
		} else {
			updated[m.editCreationIndex] = template
		}
		if err := files.SaveTemplates(m.store.TemplatesPath, updated); err != nil {
			m.editCreationNotice = err.Error()
			return
		}
		m.templates = updated
		if !m.editCreationNew && previousName != template.Name {
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
		if m.editCreationNew {
			updated = append(updated, draft.Stored())
		} else {
			updated[m.editCreationIndex] = draft.Stored()
		}
		cfg := m.cfg
		cfg.FileTypes = updated
		if err := config.Save(m.store.ConfigPath, cfg); err != nil {
			m.editCreationNotice = err.Error()
			return
		}
		m.cfg = cfg
	}
	if m.editCreationNew {
		m.creationSearch = ""
		m.newMode = newModeFileTypes
		if m.editCreationTemplate {
			m.newMode = newModeTemplates
			m.creationSel = len(m.templates) - 1
		} else {
			m.creationSel = len(m.cfg.FileTypes) - 1
		}
	}
	m.editCreationOpen = false
}

func validFileExtension(extension string) bool {
	if extension == "" || strings.HasSuffix(extension, ".") || strings.Contains(extension, "..") {
		return false
	}
	for _, char := range extension {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '-' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func (m *app) newFileType() {
	m.editCreation = config.FileType{
		ID: "custom-" + rand.Text(), Extension: "txt", DefaultName: locales.DefaultUntitled(),
		Enabled: true, Location: config.FileTypeSubmenu,
	}
	m.editCreationIndex, m.editCreationTemplate, m.editCreationNew = -1, false, true
	m.editCreationNotice, m.editCreationOpen = "", true
}

func (m *app) beginTemplate(template files.Template) error {
	if m.editCreationOpen {
		return errors.New(locales.FinishEditingFirst())
	}
	m.editCreation = templateFileType(template)
	base := m.editCreation.DefaultName
	for suffix := 2; slices.ContainsFunc(m.templates, func(existing files.Template) bool {
		name := m.editCreation.DefaultName
		if m.editCreation.Extension != "" {
			name += "." + m.editCreation.Extension
		}
		return strings.EqualFold(existing.Name, name)
	}); suffix++ {
		m.editCreation.DefaultName = base + " " + strconv.Itoa(suffix)
	}
	m.editCreationIndex, m.editCreationTemplate, m.editCreationNew = -1, true, true
	m.editCreation.Enabled = true
	m.editCreationNotice, m.editCreationOpen = "", true
	m.selectPane(paneTemplates)
	return nil
}

func (m *app) duplicateTemplate() {
	template := m.templates[m.editCreationIndex]
	m.editCreationOpen = false
	m.report(locales.AddFailed(), m.beginTemplate(template))
}

func (m *app) deleteCustomType() {
	index := m.editCreationIndex
	if index < 0 || index >= len(m.cfg.FileTypes) || !strings.HasPrefix(m.cfg.FileTypes[index].ID, "custom-") {
		return
	}
	id := m.cfg.FileTypes[index].ID
	m.cfg.FileTypes = slices.Delete(m.cfg.FileTypes, index, index+1)
	m.cfg.MenuItems = slices.DeleteFunc(m.cfg.MenuItems, func(item config.MenuItem) bool { return item.ID == "type:"+id })
	m.creationSel = 0
	m.saveConfig()
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
			m.report(locales.AddFailed(), m.beginTemplate(template))
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
	m.templates = slices.Delete(m.templates, m.tplSel, m.tplSel+1)
	m.tplSel = min(m.tplSel, len(m.templates)-1)
	m.saveTemplates()
}
