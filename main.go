// Command rightclick is the right-click helper. Its main window configures the
// Finder right-click menu; the Finder Sync extension (extension/) opens
// rightclick:// URLs, and this app does the work.
package main

import (
	"log"

	"github.com/egoist/mygo"
	"rightclick/settings"
)

func main() {
	// Register before Run. OnOpenURL also delivers the URL the app was launched
	// with (docs/app.md, "Deep links").
	mygo.App.OnOpenURL(routeURL)
	// Closing the main window keeps the app running, so the Finder actions and
	// the tray stay available. Without a listener, mygo quits the app when its
	// last window closes.
	mygo.App.OnWindowAllClosed(func() {})
	// Clicking the Dock icon with no window open shows the main window again.
	mygo.App.OnActivate(func(hasVisibleWindows bool) {
		if !hasVisibleWindows {
			settings.Open()
		}
	})

	mygo.App.WhenReady(func() {
		tray, err := mygo.NewTray(mygo.TrayOptions{
			ID:      "main",
			Title:   "右",
			ToolTip: "右键助手",
			Menu:    trayMenu("扩展状态：检查中…", false),
		})
		if err != nil {
			// The main window does not depend on the tray, so the app keeps going.
			log.Printf("菜单栏图标：%v", err)
		}
		settings.Open()
		// pluginkit blocks, so the check runs off the main thread. Its result
		// goes back to the main thread, which updates the tray and shows the
		// diagnostics tab when the extension is off.
		go func() {
			enabled := settings.ExtensionEnabled()
			mygo.RunOnMain(func() {
				if tray != nil {
					tray.SetMenu(trayMenu(extensionStatusLabel(enabled), !enabled))
				}
				if !enabled {
					settings.OpenDiagnostics()
				}
			})
		}()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// extensionStatusLabel is the tray line that shows the Finder extension state.
func extensionStatusLabel(enabled bool) string {
	if enabled {
		return "扩展状态：已启用"
	}
	return "扩展状态：未启用"
}

// trayMenu builds the tray menu. The first line only shows status. When the
// extension is off, a line below it opens the diagnostics tab.
func trayMenu(status string, extensionOff bool) *mygo.Menu {
	items := []*mygo.MenuItem{
		{Label: status, Disabled: true},
	}
	if extensionOff {
		items = append(items, &mygo.MenuItem{
			Label: "启用 Finder 扩展…",
			Click: func(*mygo.MenuItem, *mygo.Window) { settings.OpenDiagnostics() },
		})
	}
	items = append(items,
		mygo.Separator(),
		&mygo.MenuItem{Label: "打开主窗口", Click: func(*mygo.MenuItem, *mygo.Window) { settings.Open() }},
		mygo.Separator(),
		&mygo.MenuItem{Label: "退出", Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	)
	return mygo.NewMenu(items)
}
