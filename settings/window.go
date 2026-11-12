package settings

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The open main window and its state. Only the main thread touches them.
var (
	openWindow *mygo.Window
	openApp    *app
)

// Size of the main window, in points. It opens at mainWidth by mainHeight and
// cannot be made smaller than minWidth by minHeight. At minWidth the menu page
// still shows its editor and its preview side by side.
const (
	mainWidth  = 980
	mainHeight = 560
	minWidth   = 940
	minHeight  = 480
)

// Open shows the main window. The first call creates it; later calls bring
// the same window to the front, and a window the user closed is created
// again. Call it on the main thread after the app is ready.
func Open() {
	if openWindow != nil && !openWindow.IsDestroyed() {
		if openWindow.IsMinimized() {
			openWindow.Restore()
		}
		openWindow.Show()
		openApp.checkFinder() // the sidebar and the menu show the state of this check
		return
	}
	store, err := DefaultStore()
	if err != nil {
		reportOpenError(err)
		return
	}
	m, err := newApp(store, defaultEnv())
	if err != nil {
		reportOpenError(err)
		return
	}
	m.win = mygo.NewWindow(mygo.WindowOptions{
		Title:     "右键助手",
		Width:     mainWidth,
		Height:    mainHeight,
		MinWidth:  minWidth,
		MinHeight: minHeight,
		// A new key, so a size saved under "main" by an earlier build does not keep the
		// window narrower than mainWidth.
		StateKey:      "main-v2",
		TitleBarStyle: mygo.TitleBarHiddenInset, // the view draws the top edge; see sidebar
		Content:       ui.View(m.view),
	})
	openWindow, openApp = m.win, m
	m.checkFinder()
}

// reportOpenError tells the user why the window did not open. Dialog.Error
// waits until the dialog is dismissed, so it runs off the main thread.
func reportOpenError(err error) {
	log.Printf("设置窗口无法打开：%v", err)
	go mygo.Dialog.Error("设置窗口无法打开", err.Error())
}

// OpenDiagnostics shows the main window on the diagnostics pane, where the
// Finder extension state and the steps to enable it are. Call it on the main
// thread, like Open.
func OpenDiagnostics() {
	Open()
	if openApp == nil {
		return // Open could not create the window and has already reported why
	}
	openApp.selectPane(paneDiagnostics)
}

// ExtensionEnabled reports whether macOS has the Finder extension switched on.
// It runs pluginkit and waits for it, so call it off the main thread.
func ExtensionEnabled() bool {
	return finderExtensionEnabled()
}
