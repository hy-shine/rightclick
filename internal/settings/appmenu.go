package settings

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"

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

// CheckForUpdates opens the updater's window. The application menu calls it, on the
// main thread.
func CheckForUpdates() {
	updater.CheckForUpdates()
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
