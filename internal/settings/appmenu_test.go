package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo"

	"github.com/hy-shine/rightclick/internal/locales"
)

// TestAppMenuChecksForUpdatesAfterAbout checks that Check for Updates comes right after
// About, and that the menus of the default menu bar, Edit among them, are still there.
func TestAppMenuChecksForUpdatesAfterAbout(t *testing.T) {
	InstallAppMenu()
	t.Cleanup(func() { mygo.App.SetMenu(nil) })

	top := mygo.App.Menu().Items()
	var roles []mygo.MenuRole
	for _, item := range top {
		roles = append(roles, item.Role)
	}
	want := []mygo.MenuRole{mygo.RoleAppMenu, mygo.RoleFileMenu, mygo.RoleEditMenu, mygo.RoleViewMenu, mygo.RoleWindowMenu}
	if !slices.Equal(roles, want) {
		t.Fatalf("menu bar roles = %q, want %q", roles, want)
	}
	app := top[0].Submenu
	if len(app) < 2 || app[0].Role != mygo.RoleAbout || app[1].ID != checkUpdatesMenuID {
		t.Fatalf("the application menu does not start with About, then Check for Updates")
	}
	if app[1].Label != locales.CheckUpdate() {
		t.Errorf("menu label = %q, want %q", app[1].Label, locales.CheckUpdate())
	}
}

// TestAppMenuLabelFollowsTheLanguage checks that the item takes the label of a language
// that is set after the menu was installed.
func TestAppMenuLabelFollowsTheLanguage(t *testing.T) {
	InstallAppMenu()
	t.Cleanup(func() { mygo.App.SetMenu(nil) })

	useLocale(t, "zh-Hans")
	refreshAppMenuLabels()
	item := mygo.App.Menu().ItemByID(checkUpdatesMenuID)
	if item == nil || item.Label != "检查更新" {
		t.Errorf("after the switch to Chinese the item reads %v, want 检查更新", item)
	}

	useLocale(t, "zh-Hant")
	refreshAppMenuLabels()
	item = mygo.App.Menu().ItemByID(checkUpdatesMenuID)
	if item == nil || item.Label != "檢查更新" {
		t.Errorf("after the switch to Traditional Chinese the item reads %v, want 檢查更新", item)
	}
}
