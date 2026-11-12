package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
)

// readyApp returns an app whose four steps all pass: the extension is on, a
// folder is watched, an entry is on, and Finder has read the folders.
func readyApp(t *testing.T) *app {
	t.Helper()
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.finder = finderOn
	return m
}

// selectDiagnostics shows the diagnostics pane. It sets the tab directly, because
// selectPane would start a Finder check, which needs a window.
func selectDiagnostics(m *app) {
	for i, p := range paneTabs {
		if p.id == paneDiagnostics {
			m.tab = i
		}
	}
}

func TestSetupStepNamesTheFirstMissingStep(t *testing.T) {
	cases := []struct {
		name    string
		finder  finderState
		folders []string
		enabled bool
		changed bool
		want    setupStep
	}{
		{"checking", finderChecking, []string{"/a"}, true, false, stepChecking},
		{"extension off", finderOff, []string{"/a"}, true, false, stepExtension},
		{"no folder", finderOn, nil, true, false, stepFolder},
		{"no entry on", finderOn, []string{"/a"}, false, false, stepItem},
		{"folder changed", finderOn, []string{"/a"}, true, true, stepRestart},
		{"all pass", finderOn, []string{"/a"}, true, false, stepReady},
	}
	for _, tc := range cases {
		m := readyApp(t)
		m.finder = tc.finder
		m.cfg.WatchFolders = tc.folders
		for i := range m.cfg.MenuItems {
			m.cfg.MenuItems[i].Enabled = tc.enabled
		}
		m.foldersChanged = tc.changed
		if got := m.setupStep(); got != tc.want {
			t.Errorf("%s: step = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestSidebarNamesTheMissingStep(t *testing.T) {
	m := readyApp(t)
	m.finder = finderOff
	if got, want := m.navSummary(paneMenu), "需要开启扩展"; got != want {
		t.Errorf("menu line with the extension off = %q, want %q", got, want)
	}
	if got, want := m.navSummary(paneDiagnostics), "需要开启扩展"; got != want {
		t.Errorf("diagnostics line with the extension off = %q, want %q", got, want)
	}
	m.finder = finderOn
	m.cfg.WatchFolders = nil
	if got, want := m.navSummary(paneMenu), "等待设置文件夹范围"; got != want {
		t.Errorf("menu line with no folder = %q, want %q", got, want)
	}
}

// TestMenuPageLeadsWithTheMissingStep checks that a first run with no folder
// shows the folder step, with both ways to add one, above the editor.
func TestMenuPageLeadsWithTheMissingStep(t *testing.T) {
	m := readyApp(t)
	m.cfg.WatchFolders = nil
	texts := ui.NewTester(m.view, mainWidth, mainHeight).Texts()
	for _, want := range []string{"还没有设置文件夹范围", "添加文件夹…", "使用默认文件夹（桌面、文稿、下载）"} {
		if !slices.Contains(texts, want) {
			t.Errorf("menu page does not show %q; texts %q", want, texts)
		}
	}
}

// TestDiagnosticsSaysReadyOnlyWhenEveryStepPasses uses one window per app, as the
// app does: a second tester on the same app would show its lists twice.
func TestDiagnosticsSaysReadyOnlyWhenEveryStepPasses(t *testing.T) {
	m := readyApp(t)
	selectDiagnostics(m)
	if !slices.Contains(ui.NewTester(m.view, mainWidth, mainHeight).Texts(), "访达中已经可用") {
		t.Error("all four steps pass, but the diagnostics pane does not say the menu is usable")
	}

	changed := readyApp(t)
	changed.foldersChanged = true
	selectDiagnostics(changed)
	if slices.Contains(ui.NewTester(changed.view, mainWidth, mainHeight).Texts(), "访达中已经可用") {
		t.Error("a folder changed and Finder has not restarted, but the pane says the menu is usable")
	}
}

// TestRestartFromMenuPageAsksFirst checks that the restart button on the menu page
// opens the confirmation dialog, which lives at the window level.
func TestRestartFromMenuPageAsksFirst(t *testing.T) {
	m := readyApp(t)
	m.foldersChanged = true
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	if err := tt.Click("重新启动"); err != nil {
		t.Fatal(err)
	}
	if !m.confirmRestart {
		t.Fatal("the restart button did not ask for confirmation")
	}
	if !slices.Contains(tt.Texts(), "重新启动访达？") {
		t.Errorf("the confirmation dialog is not shown; texts %q", tt.Texts())
	}
}

func TestUseDefaultFoldersKeepsWatchedOnes(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.cfg.WatchFolders = []string{"/tmp/only"}
	m.useDefaultFolders()
	want := append([]string{"/tmp/only"}, config.Default().WatchFolders...)
	if !m.foldersChanged {
		t.Error("adding folders did not mark them as changed for Finder")
	}
	cfg, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.WatchFolders, want) {
		t.Errorf("watch folders saved = %q, want %q", cfg.WatchFolders, want)
	}
}

func TestMenuPageUsesFunctionalSections(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	texts := ui.NewTester(m.view, 760, 560).Texts()
	for _, label := range []string{"路径与移动", "打开应用"} {
		if !slices.Contains(texts, label) {
			t.Errorf("missing functional section %q", label)
		}
	}
}

// TestPreviewSitsBesideTheEditorInAWideWindow checks that at the default window
// size the preview is on the first screen, to the right of the editor.
func TestMenuPreviewOpensAndClosesOnDemand(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	if err := tt.Click("预览"); err != nil {
		t.Fatal(err)
	}
	if !m.menuPreviewOpen || !slices.Contains(tt.Texts(), "菜单预览") {
		t.Fatal("preview did not open")
	}
	if err := tt.Click("关闭预览"); err != nil {
		t.Fatal(err)
	}
	if m.menuPreviewOpen || slices.Contains(tt.Texts(), "菜单预览") {
		t.Fatal("preview did not close")
	}
}
