package settings

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// openersPane sets the terminal and the list of editors.
func (m *app) openersPane(c *ui.Context) {
	t := c.Theme()
	paneHeader(c, "打开方式", "菜单里的「在终端打开」和「用…打开」使用这里的设置。")
	sectionTitle(c, "终端")
	ledger(c, ui.Column(c)).Children(func() {
		settingRow(c, "默认终端", "「在终端打开」用它打开文件夹。", func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Select(c.Key("terminal"), &m.cfg.Terminal, terminalChoices(m.cfg.Terminal)).
					Label("默认终端").Width(150).OnChange(m.saveConfig)
				if quietButton(c, iconOpen, "选择应用…") {
					m.pickTerminal()
				}
			})
		})
	})
	sectionTitle(c, "编辑器")
	remove := -1
	if len(m.cfg.Editors) == 0 {
		emptyState(c, iconCode, "还没有编辑器", "在下面输入应用的名称，如 Cursor。")
	} else {
		ledger(c, ui.Column(c)).Children(func() {
			for i, name := range m.cfg.Editors {
				ui.Row(c).Gap(12).Padding(0, 8, 0, 16).Height(44).AlignItems(ui.Center).Children(func() {
					glyph(c, iconCode, 17, t.TextMuted)
					ui.Text(c, name).FontSize(sizeBody).Grow(1)
					if i == 0 {
						ui.Text(c, "默认").FontSize(sizeCaption).FontWeight(600).
							TextColor(t.TextMuted).Padding(2, 6).
							Background(t.SurfaceHover).Radius(t.Radius)
					}
					if ghostButton(c, "移除", true) {
						remove = i
					}
				})
			}
		})
	}
	// Removing happens after the rows are built, as in the folders pane.
	if remove >= 0 {
		m.cfg.Editors = slices.Delete(m.cfg.Editors, remove, remove+1)
		m.saveConfig()
	}
	ui.Row(c).Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c.Key("new-editor"), &m.newEditor).
			Label("编辑器应用名称").Placeholder("应用名称，如 Cursor").Grow(1).OnSubmit(m.addEditor)
		if primaryButton(c, iconPlus, "添加") {
			m.addEditor()
		}
		if quietButton(c, iconOpen, "选择应用…") {
			m.pickEditor()
		}
	})
	ui.Text(c, "“用默认编辑器打开”使用列表第一项；未列出的应用也可以通过选择已安装的 .app 添加。").
		FontSize(sizeCaption).LineHeight(1.45).TextColor(t.TextMuted)
}

// addEditor adds the application named in the input to the editors and clears the input.
func (m *app) addEditor() {
	name := strings.TrimSpace(m.newEditor)
	m.newEditor = ""
	if name == "" || slices.Contains(m.cfg.Editors, name) {
		return
	}
	m.cfg.Editors = append(m.cfg.Editors, name)
	m.saveConfig()
}

// pickTerminal lets the user choose any installed application for the Open in
// Terminal action. The app name is what open(1) receives with -a.
func (m *app) pickTerminal() {
	m.pickApplication("选择终端应用", func(name string) {
		m.cfg.Terminal = name
		m.saveConfig()
	})
}

// pickEditor lets the user add any installed application as an editor.
func (m *app) pickEditor() {
	m.pickApplication("选择编辑器应用", func(name string) {
		m.newEditor = ""
		if slices.Contains(m.cfg.Editors, name) {
			return
		}
		m.cfg.Editors = append(m.cfg.Editors, name)
		m.saveConfig()
	})
}

// pickApplication opens an application package picker and publishes the chosen
// display name on the UI thread.
func (m *app) pickApplication(title string, apply func(string)) {
	win, open := m.win, m.env.openDialog
	go func() {
		paths, err := open(mygo.OpenDialogOptions{
			Parent:  win,
			Title:   title,
			Message: "选择一个已安装的应用",
			Filters: []mygo.FileFilter{{
				Name:       "应用程序",
				Extensions: []string{"app"},
			}},
		})
		if err == nil && len(paths) == 0 {
			return
		}
		var name string
		if err == nil {
			name, err = applicationName(paths[0])
		}
		win.Update(func() {
			if err != nil {
				m.report("选择应用失败", err)
				return
			}
			apply(name)
		})
	}()
}

// applicationName returns the name open -a uses for a selected .app package.
func applicationName(path string) (string, error) {
	base := filepath.Base(path)
	if !strings.EqualFold(filepath.Ext(base), ".app") {
		return "", fmt.Errorf("不是应用程序包：%s", base)
	}
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("应用程序名称为空：%s", path)
	}
	return name, nil
}

// terminalChoices returns the terminals to offer. The current setting is
// added when it is not one of them, so that the select shows it.
func terminalChoices(current string) []string {
	choices := []string{
		"Terminal",
		"iTerm",
		"Warp",
		"Ghostty",
		"kitty",
		"Alacritty",
		"WezTerm",
		"Hyper",
		"Tabby",
	}
	if current != "" && !slices.Contains(choices, current) {
		choices = append(choices, current)
	}
	return choices
}
