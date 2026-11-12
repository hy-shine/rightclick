package settings

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"rightclick/config"
	"rightclick/internal/catalog"
	"rightclick/locales"
)

func TestPlaceSwitchChangesPreview(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, minWidth, minHeight)
	if err := tt.Click("预览"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(tt.Texts(), "显示 9 项") {
		t.Fatalf("a file shows 9 entries; texts %q", tt.Texts())
	}
	tt.Frame()
	if err := tt.Click("空白处"); err != nil {
		t.Fatal(err)
	}
	if m.previewCtx != ctxBlank {
		t.Errorf("previewCtx = %d after clicking 空白处, want %d", m.previewCtx, ctxBlank)
	}
	if !slices.Contains(tt.Texts(), "显示 6 项") {
		t.Errorf("blank space shows 6 entries; texts %q", tt.Texts())
	}
}

func TestMenuHeaderActionsShareABaseline(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneMenu)
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	search, ok := tt.Find(locales.SearchMenu())
	if !ok {
		t.Fatal("menu search field is missing")
	}
	restore, ok := tt.Find(locales.RestoreDefaults())
	if !ok {
		t.Fatal("restore defaults button is missing")
	}
	preview, ok := tt.Find(locales.Preview())
	if !ok {
		t.Fatal("preview button is missing")
	}
	searchMiddle := search.Y + search.H/2
	restoreMiddle := restore.Y + restore.H/2
	previewMiddle := preview.Y + preview.H/2
	if diff := searchMiddle - restoreMiddle; diff < -0.5 || diff > 0.5 {
		t.Errorf("search middle = %.1f, restore middle = %.1f, want them level", searchMiddle, restoreMiddle)
	}
	if diff := previewMiddle - restoreMiddle; diff < -0.5 || diff > 0.5 {
		t.Errorf("preview middle = %.1f, restore middle = %.1f, want them level", previewMiddle, restoreMiddle)
	}
}

func TestRestoreDefaultsSwitchesEntriesBackOn(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.cfg.MenuItems[0].Enabled = false
	m.saveConfig()
	tt := ui.NewTester(m.view, minWidth, minHeight)
	// The reset button is at the top of the editor, on the first screen.
	tt.Frame()
	if err := tt.Click("恢复默认"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MenuItems[0].Enabled {
		t.Fatal("restoring the menu happened before confirmation")
	}
	if !m.confirmRestore {
		t.Fatal("restore confirmation did not open")
	}
	if err := tt.Click("恢复"); err != nil {
		t.Fatal(err)
	}
	cfg, err = config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MenuItems[0].Enabled {
		t.Errorf("new-file still switched off after 恢复默认: %+v", cfg.MenuItems[0])
	}
}

func TestRemovingFolderIsSaved(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	first, n := m.cfg.WatchFolders[0], len(m.cfg.WatchFolders)
	m.selectPane(paneFolders)
	tt := ui.NewTester(m.view, 760, 560)
	if err := tt.Click("移除"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.WatchFolders) != n-1 || slices.Contains(cfg.WatchFolders, first) {
		t.Errorf("watch folders %q after removing the first of %d, want %q gone", cfg.WatchFolders, n, first)
	}
}

func TestEveryMenuEntryHasAnIcon(t *testing.T) {
	for _, item := range config.Default().MenuItems {
		if _, ok := entryIcons[item.ID]; !ok {
			t.Errorf("menu entry %q has no icon", item.ID)
		}
		if _, ok := menuEntries()[item.ID]; !ok {
			t.Errorf("menu entry %q has no label", item.ID)
		}
	}
}

func TestEveryFileTypeHasARealIconAsset(t *testing.T) {
	for _, fileType := range config.Default().FileTypes {
		if _, err := ui.ParseSVG(catalog.Icon(fileType.Extension)); err != nil {
			t.Errorf("%s icon: %v", fileType.ID, err)
		}
	}
}

func TestMenuPageFitsTheMinimumWindow(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, minWidth, minHeight)
	name, ok := tt.Find("新建文件")
	if !ok {
		t.Fatal("the first action is not on the pane")
	}
	preview, ok := tt.Find("预览")
	if !ok {
		t.Fatal("the preview button is not on the pane")
	}
	if name.Y+name.H > minHeight || preview.Y+preview.H > minHeight {
		t.Error("the first action or preview button is outside the window")
	}
	if slices.Contains(tt.Texts(), "显示 9 项") {
		t.Error("preview content is visible before the preview opens")
	}
}

func TestMenuSearchFiltersAndKeepsContextVisible(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.menuSearch = "zip"
	texts := ui.NewTester(m.view, minWidth, minHeight).Texts()
	if !slices.Contains(texts, "仅 .zip 文件") {
		t.Errorf("matching action does not show its context: %q", texts)
	}
	if rows := m.menuDisplayRows(3); len(rows) != 2 {
		t.Errorf("zip search returned %d editable rows, want 2", len(rows))
	}
	m.menuSearch = "not-an-action"
	texts = ui.NewTester(m.view, minWidth, minHeight).Texts()
	if !slices.Contains(texts, "没有匹配的动作") {
		t.Errorf("empty search state is missing: %q", texts)
	}
	if !slices.Contains(texts, "预览") {
		t.Error("empty search should keep the preview button visible")
	}
}

func TestCreationEditButtonHasAUsefulTarget(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneTemplates)
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	bounds, ok := tt.Find("编辑 Word 文档")
	if !ok {
		t.Fatal("edit button is missing")
	}
	if bounds.W < 30 || bounds.H < 24 {
		t.Errorf("edit button is too small: %+v", bounds)
	}
}

// TestPreviewCountFitsItsColumn checks that the entry count beside the place switch
// stays inside the preview column, in the smallest window.
func TestPreviewCountFitsItsColumn(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, minWidth, minHeight)
	if err := tt.Click("预览"); err != nil {
		t.Fatal(err)
	}
	count, ok := tt.Find("显示 9 项")
	if !ok {
		t.Fatal("the entry count is not on the pane")
	}
	if count.X+count.W > minWidth || count.Y+count.H > minHeight {
		t.Errorf("preview count is outside the window: %+v", count)
	}
}

// TestEmptyTemplatesShowOnlyTheEmptyState checks that the templates pane with no
// templates shows the empty state and its add button, but no list or editor.
func TestEmptyTemplatesShowOnlyTheEmptyState(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.templates = nil
	m.tplSel = -1
	m.newMode = newModeTemplates
	m.selectPane(paneTemplates)
	tt := ui.NewTester(m.view, 760, 560)
	texts := tt.Texts()
	if !slices.Contains(texts, "没有找到相关类型") || !slices.Contains(texts, "添加模板…") {
		t.Errorf("empty templates pane shows %q, want the empty state and 添加模板…", texts)
	}
	if slices.Contains(texts, "名称") {
		t.Errorf("empty templates pane still shows the template editor: %q", texts)
	}
}

// TestSidebarNamesPanesByTheirPurpose checks the two pane names that say what
// their pane is for.
func TestSidebarNamesPanesByTheirPurpose(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	texts := ui.NewTester(m.view, 760, 560).Texts()
	for _, want := range []string{"新建", "文件夹范围"} {
		if !slices.Contains(texts, want) {
			t.Errorf("sidebar does not show %q; texts %q", want, texts)
		}
	}
}

func TestTildePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home folder")
	}
	cases := []struct{ in, want string }{
		{home, "~"},
		{filepath.Join(home, "Desktop"), "~/Desktop"},
		{home + "x/Desktop", home + "x/Desktop"}, // a sibling that shares the prefix
		{"/Volumes/Data", "/Volumes/Data"},
	}
	for _, tc := range cases {
		if got := tildePath(tc.in); got != tc.want {
			t.Errorf("tildePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
