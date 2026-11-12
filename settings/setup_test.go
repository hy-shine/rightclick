package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/config"
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
	for i, p := range paneTabs() {
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
	if got, want := m.navSummary(paneMenu), "The extension must be turned on"; got != want {
		t.Errorf("menu line with the extension off = %q, want %q", got, want)
	}
	if got, want := m.navSummary(paneDiagnostics), "The extension must be turned on"; got != want {
		t.Errorf("diagnostics line with the extension off = %q, want %q", got, want)
	}
	m.finder = finderOn
	m.cfg.WatchFolders = nil
	if got, want := m.navSummary(paneMenu), "Waiting for a folder scope"; got != want {
		t.Errorf("menu line with no folder = %q, want %q", got, want)
	}
}

// TestMenuPageOmitsSetupWarnings checks that setup and permission blockers stay
// out of the menu editor page and are signaled in diagnostics instead.
func TestMenuPageOmitsSetupWarnings(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*app)
	}{
		{"extension off", func(m *app) { m.finder = finderOff }},
		{"no folder", func(m *app) { m.cfg.WatchFolders = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := readyApp(t)
			tc.mutate(m)
			texts := ui.NewTester(m.view, mainWidth, mainHeight).Texts()
			for _, unwanted := range []string{"The Finder extension is not on yet", "The extension must be turned on", "No folder scope is set", "Waiting for a folder scope"} {
				if slices.Contains(texts, unwanted) {
					t.Errorf("menu page still shows setup warning %q; texts %q", unwanted, texts)
				}
			}
		})
	}
}

// TestDiagnosticsSaysReadyOnlyWhenEveryStepPasses uses one window per app, as the
// app does: a second tester on the same app would show its lists twice.
func TestDiagnosticsSaysReadyOnlyWhenEveryStepPasses(t *testing.T) {
	m := readyApp(t)
	selectDiagnostics(m)
	if !slices.Contains(ui.NewTester(m.view, mainWidth, mainHeight).Texts(), "Available in Finder") {
		t.Error("all four steps pass, but the diagnostics pane does not say the menu is usable")
	}

	changed := readyApp(t)
	changed.foldersChanged = true
	selectDiagnostics(changed)
	if slices.Contains(ui.NewTester(changed.view, mainWidth, mainHeight).Texts(), "Available in Finder") {
		t.Error("a folder changed and Finder has not restarted, but the pane says the menu is usable")
	}
}

// TestSidebarMarksDiagnosticsWhenBlocked checks that the sidebar diagnostics
// item displays an alert indicator only while a setup prerequisite is missing.
func TestSidebarMarksDiagnosticsWhenBlocked(t *testing.T) {
	blocked := readyApp(t)
	blocked.finder = finderOff
	tt := ui.NewTester(blocked.view, mainWidth, mainHeight)
	if _, ok := tt.Find("Permissions The extension must be turned on"); !ok {
		t.Fatal("sidebar diagnostics row does not show warning mark when extension is off")
	}

	ready := readyApp(t)
	ttReady := ui.NewTester(ready.view, mainWidth, mainHeight)
	if _, ok := ttReady.Find("Permissions The extension must be turned on"); ok {
		t.Fatal("sidebar diagnostics row still shows warning mark when all steps pass")
	}
}

// TestRestartFromDiagnosticsAsksFirst checks that the restart button on the
// diagnostics pane opens the confirmation dialog, which lives at the window level.
func TestRestartFromDiagnosticsAsksFirst(t *testing.T) {
	m := readyApp(t)
	m.foldersChanged = true
	selectDiagnostics(m)
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	if err := tt.Click("Restart"); err != nil {
		t.Fatal(err)
	}
	if !m.confirmRestart {
		t.Fatal("the restart button did not ask for confirmation")
	}
	if !slices.Contains(tt.Texts(), "Restart Finder?") {
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

func TestMenuPageOffersFreeOrdering(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	texts := ui.NewTester(m.view, minWidth, minHeight).Texts()
	// Open With also names the openers pane, so it cannot tell a group header
	// from the sidebar; File Tools names only the old group.
	for _, label := range []string{"Paths and Move", "File Tools"} {
		if slices.Contains(texts, label) {
			t.Errorf("fixed category still splits the sortable list: %q", label)
		}
	}
}

// TestMenuPreviewOpensAndClosesOnDemand checks that preview stays a temporary
// overlay and follows the selected Finder context while it is open.
func TestMenuPreviewOpensAndClosesOnDemand(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	if slices.Contains(tt.Texts(), "Menu Preview") {
		t.Fatal("menu preview is visible before it is opened")
	}
	if err := tt.Click("Preview"); err != nil {
		t.Fatal(err)
	}
	if !m.menuPreviewOpen || !slices.Contains(tt.Texts(), "Menu Preview") {
		t.Fatal("menu preview did not open")
	}
	if err := tt.Click("Blank Space"); err != nil {
		t.Fatal(err)
	}
	if m.previewCtx != ctxBlank {
		t.Fatalf("preview context = %d, want blank", m.previewCtx)
	}
	if err := tt.Click("Close Preview"); err != nil {
		t.Fatal(err)
	}
	if m.menuPreviewOpen || slices.Contains(tt.Texts(), "Menu Preview") {
		t.Fatal("menu preview did not close")
	}
}
