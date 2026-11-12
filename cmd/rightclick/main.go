// Command rightclick is the right-click helper. Its main window configures the
// Finder right-click menu. The Finder Sync extension (extension/) runs each action
// by starting this program with --action, so the action runs with no window.
package main

import (
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater/native"
	"github.com/hy-shine/rightclick/internal/app"
	"github.com/hy-shine/rightclick/internal/finder"
	"github.com/hy-shine/rightclick/internal/settings"
)

// launchedForURL records that a rightclick:// link started this process. Such a
// start runs in the background and must not show the main window.
var launchedForURL atomic.Bool

// windowOpenDelay is how long start-up waits for a launch link before it shows
// the window. macOS delivers the link just after the app finishes launching.
const windowOpenDelay = 300 * time.Millisecond

func main() {
	// The Finder extension runs an action by starting this program with --action.
	// The action runs here, and no window or Dock icon appears.
	if len(os.Args) == 3 && os.Args[1] == "--action" {
		os.Exit(app.RunAction(os.Args[2]))
	}
	// Update in the app. Finder action launches return above, so they do not
	// start the updater.
	mygo.Use(native.Plugin)
	// Register before Run. OnOpenURL also delivers the URL the app was launched
	// with (docs/app.md, "Deep links").
	mygo.App.OnOpenURL(func(raw string) {
		launchedForURL.Store(true)
		app.RouteURL(raw)
	})
	// Closing the main window keeps the app running. Without a listener, mygo quits
	// the app when its last window closes.
	mygo.App.OnWindowAllClosed(func() {})
	// Clicking the Dock icon with no window open shows the main window again.
	mygo.App.OnActivate(func(hasVisibleWindows bool) {
		if !hasVisibleWindows {
			settings.Open()
		}
	})

	mygo.App.WhenReady(func() {
		// The saved language applies before any text is read. MYGO_LOCALE, set by
		// mygo dev -lang, takes precedence for one run.
		if os.Getenv("MYGO_LOCALE") == "" {
			if cfg, err := app.LoadConfig(); err != nil {
				log.Printf("settings: %v", err)
			} else {
				mygo.App.SetLocale(cfg.Language)
			}
		}
		// After the language is set, so that the menu's own labels are in it.
		settings.InstallAppMenu()
		installActionScript()
		openedAtLogin := mygo.App.WasOpenedAtLogin()
		if shouldOpenWindow(openedAtLogin, launchedForURL.Load()) {
			time.AfterFunc(windowOpenDelay, func() {
				if shouldOpenWindow(openedAtLogin, launchedForURL.Load()) {
					mygo.RunOnMain(settings.Open)
				}
			})
		}
		// pluginkit blocks, so the check runs off the main thread. Its result
		// goes back to the main thread, which shows the diagnostics tab when the
		// extension is off. The app registers its extension first, so that a
		// rebuilt or moved app is known to the system without a manual step.
		go func() {
			if err := finder.Register(); err != nil {
				log.Printf("finder extension: %v", err)
			}
			enabled := settings.ExtensionEnabled()
			mygo.RunOnMain(func() {
				if !enabled && shouldOpenWindow(openedAtLogin, launchedForURL.Load()) {
					settings.OpenDiagnostics()
				}
			})
		}()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// shouldOpenWindow reports whether start-up shows the main window. A start at
// login and a start from a rightclick:// link stay in the background.
func shouldOpenWindow(openedAtLogin, launchedForURL bool) bool {
	return !openedAtLogin && !launchedForURL
}

// installActionScript points the script that the Finder extension runs at this app.
// The extension cannot write its own Application Scripts folder, so the app writes
// it at every start, and a moved app is found again.
func installActionScript() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("finder action script: %v", err)
		return
	}
	executable, err := os.Executable()
	if err != nil {
		log.Printf("finder action script: %v", err)
		return
	}
	if err := finder.InstallActionScript(finder.ActionScriptDir(home), executable); err != nil {
		log.Printf("finder action script: %v", err)
	}
}
