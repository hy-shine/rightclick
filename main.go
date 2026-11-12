// Command rightclick is the right-click helper. Its main window configures the
// Finder right-click menu; the Finder Sync extension (extension/) opens
// rightclick:// URLs, and this app does the work.
package main

import (
	"log"
	"os"

	"github.com/egoist/mygo"
	"rightclick/settings"
)

func main() {
	// Register before Run. OnOpenURL also delivers the URL the app was launched
	// with (docs/app.md, "Deep links").
	mygo.App.OnOpenURL(routeURL)
	// Closing the main window keeps the app running, so the Finder actions stay
	// available. Without a listener, mygo quits the app when its last window
	// closes.
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
			if cfg, err := loadConfig(); err != nil {
				log.Printf("settings: %v", err)
			} else {
				mygo.App.SetLocale(cfg.Language)
			}
		}
		settings.Open()
		// pluginkit blocks, so the check runs off the main thread. Its result
		// goes back to the main thread, which shows the diagnostics tab when the
		// extension is off.
		go func() {
			enabled := settings.ExtensionEnabled()
			mygo.RunOnMain(func() {
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
