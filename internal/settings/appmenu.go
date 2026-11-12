package settings

import (
	"github.com/egoist/mygo"

	"github.com/hy-shine/rightclick/internal/locales"
)

// checkUpdatesMenuID identifies the Check for Updates item of the application menu,
// so that its label can follow a change of language.
const checkUpdatesMenuID = "check-for-updates"

// InstallAppMenu sets the application menu. It is the default menu of macOS with one
// addition: Check for Updates, after About. The rest is the default, so Edit keeps
// Copy and Paste in text fields. Call it once the language is set.
func InstallAppMenu() {
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu, Submenu: []*mygo.MenuItem{
			{Role: mygo.RoleAbout},
			{ID: checkUpdatesMenuID, Label: locales.CheckUpdate(), Click: func(*mygo.MenuItem, *mygo.Window) {
				CheckForUpdates()
			}},
			mygo.Separator(),
			{Role: mygo.RoleServices},
			mygo.Separator(),
			{Role: mygo.RoleHide},
			{Role: mygo.RoleHideOthers},
			{Role: mygo.RoleUnhide},
			mygo.Separator(),
			{Role: mygo.RoleQuit},
		}},
		{Role: mygo.RoleFileMenu},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
	}))
}

// CheckForUpdates shows the main window on the general pane and checks for a newer
// release. The application menu calls it. Call it on the main thread, like Open.
func CheckForUpdates() {
	Open()
	if openApp == nil {
		return // Open could not create the window and has already reported why
	}
	openApp.selectPane(paneGeneral)
	openApp.checkForUpdate()
}

// refreshAppMenuLabels gives the Check for Updates item the label of the language the
// app now shows. The role items follow the language by themselves; this one does not.
func refreshAppMenuLabels() {
	menu := mygo.App.Menu()
	if menu == nil {
		return
	}
	if item := menu.ItemByID(checkUpdatesMenuID); item != nil {
		item.SetLabel(locales.CheckUpdate())
	}
}
