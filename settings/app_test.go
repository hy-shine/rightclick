package settings

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/files"
)

// testEnv replaces every system call, so that the tests change no system state.
func testEnv() env {
	return env{
		openAtLogin:        func() bool { return false },
		setOpenAtLogin:     func(bool) error { return nil },
		finderEnabled:      func() bool { return false },
		openFinderSettings: func() error { return nil },
		restartFinder:      func() error { return nil },
		openDialog:         func(mygo.OpenDialogOptions) ([]string, error) { return nil, nil },
		setLocale:          func(string) {},
		folderAccess:       func(string) files.Access { return files.AccessOK },
		fullDiskAccess:     func() bool { return true },
		openPrivacy:        func(string) error { return nil },
	}
}

// testStore keeps the files of one test in its own temporary folder.
func testStore(t *testing.T) Store {
	dir := t.TempDir()
	return Store{
		ConfigPath:    filepath.Join(dir, "config.json"),
		TemplatesPath: filepath.Join(dir, "templates.json"),
	}
}

func TestSwitchSavesMenuItem(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, minWidth, minHeight)
	// The editor is on the first screen even in the smallest window.
	tt.Frame()
	if err := tt.Click("Enable New File"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.MenuItems[0]; got.ID != "new-file" || got.Enabled {
		t.Errorf("new-file saved as %+v, want enabled=false", got)
	}
}

func TestTabOrder(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	texts := ui.NewTester(m.view, 760, 560).Texts()
	last := -1
	for _, p := range paneTabs() {
		i := slices.Index(texts, p.label)
		if i <= last {
			t.Fatalf("tab %q at text %d, want after text %d; texts %q", p.label, i, last, texts)
		}
		last = i
	}
}

func TestClickingTabShowsItsPane(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, 760, 560)
	if err := tt.Click("General"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(tt.Texts(), "Open at Login") {
		t.Fatalf("after clicking General the general pane is not shown; texts %q", tt.Texts())
	}
}

func TestGeneralPaneShowsVersion(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneGeneral)
	texts := ui.NewTester(m.view, mainWidth, mainHeight).Texts()
	if !slices.Contains(texts, "Version "+version) {
		t.Errorf("general pane does not show the version; texts %q", texts)
	}
}

func TestSidebarHasNoAboutPane(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	texts := ui.NewTester(m.view, mainWidth, mainHeight).Texts()
	if slices.Contains(texts, "About") {
		t.Errorf("sidebar still has a About pane; texts %q", texts)
	}
}

func TestAddedTemplateIsSaved(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "meeting.md")
	if err := os.WriteFile(src, []byte("# Meeting\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tpl, err := files.TemplateFromFile(src)
	if err != nil {
		t.Fatal(err)
	}
	m.addTemplate(tpl)
	got, err := files.LoadTemplates(store.TemplatesPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []files.Template{{Name: "meeting.md", Content: "# Meeting\n", Location: config.FileTypeSubmenu}}
	if !slices.Equal(got, want) {
		t.Errorf("templates read back = %+v, want %+v", got, want)
	}
}

func TestNewPaneListsBuiltInFileTypes(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneTemplates)
	m.newMode = newModeCode
	texts := ui.NewTester(m.view, mainWidth, 900).Texts()
	for _, want := range []string{
		"JSON", "HTML",
		"CSS", "JavaScript", "Swift", "Python",
		"Go", "Rust", "Java", "C", "C++", "SQL", "TypeScript", "Shell",
	} {
		if !slices.Contains(texts, want) {
			t.Errorf("new pane does not show file type %q; texts %q", want, texts)
		}
	}
	if slices.Contains(texts, "Finder Menu") {
		t.Errorf("new pane still shows the Finder menu preview; texts %q", texts)
	}
}

func TestNewPaneSwitchesToUserTemplates(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.templates = nil
	m.tplSel = -1
	m.selectPane(paneTemplates)
	tt := ui.NewTester(m.view, mainWidth, mainHeight)
	if err := tt.Click("My Templates"); err != nil {
		t.Fatal(err)
	}
	if m.newMode != newModeTemplates {
		t.Errorf("newMode = %d, want templates", m.newMode)
	}
	if !slices.Contains(tt.Texts(), "Add Template…") {
		t.Errorf("empty user template view does not show its action; texts %q", tt.Texts())
	}
}

func TestReorderFileTypesKeepsSelection(t *testing.T) {
	store := testStore(t)
	m, err := newApp(store, testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.fileTypeSel = slices.IndexFunc(m.cfg.FileTypes, func(fileType config.FileType) bool { return fileType.ID == "markdown" })
	m.reorderFileTypes([]int{m.fileTypeSel}, 0)
	if got, want := m.cfg.FileTypes[0].ID, "markdown"; got != want {
		t.Errorf("first file type after reorder = %q, want %q", got, want)
	}
	if got, want := m.cfg.FileTypes[m.fileTypeSel].ID, "markdown"; got != want {
		t.Errorf("selected file type after reorder = %q, want %q", got, want)
	}
	cfg, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cfg.FileTypes[0].ID, "markdown"; got != want {
		t.Errorf("saved first file type = %q, want %q", got, want)
	}
}

func TestCreationPaneHasNoMenuLocationControl(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.selectPane(paneTemplates)
	texts := ui.NewTester(m.view, mainWidth, mainHeight).Texts()
	for _, unwanted := range []string{"Menu location", "Finder Top Custom", "New File submenu"} {
		if slices.Contains(texts, unwanted) {
			t.Errorf("creation pane still shows %q", unwanted)
		}
	}
}

// menuIDs returns the IDs of the menu entries, in order.
func menuIDs(items []config.MenuItem) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	return ids
}

// TestReorderStopsAtTheMainEntries checks that a submenu entry moved to the top
// of its group stops below the main entries of that group.
func TestReorderStopsAtTheMainEntries(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.reorderGroup(1, []int{3}, 0) // move copy-extension to the top of the group
	var got []string
	for _, i := range m.groupRows(1) {
		got = append(got, m.cfg.MenuItems[i].ID)
	}
	want := []string{"copy-path", "copy-extension", "copy-name", "copy-basename", "copy-to", "move-to"}
	if !slices.Equal(got, want) {
		t.Errorf("group order = %q, want %q", got, want)
	}
}

func TestSidebarCountsEnabledMenuItems(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := m.navSummary(paneMenu), "17 items enabled"; got != want {
		t.Errorf("sidebar line = %q, want %q", got, want)
	}
	m.cfg.MenuItems[0].Enabled = false
	if got, want := m.navSummary(paneMenu), "16 items enabled"; got != want {
		t.Errorf("after one entry is switched off, sidebar line = %q, want %q", got, want)
	}
}

func TestMoveRows(t *testing.T) {
	items := []string{"A", "B", "C", "D"}
	cases := []struct {
		rows []int
		to   int
		want string
	}{
		{[]int{0}, 2, "BACD"},
		{[]int{0, 1}, 4, "CDAB"},
		{[]int{3}, 0, "DABC"},
		{[]int{1}, 1, "ABCD"},
	}
	for _, c := range cases {
		if got := strings.Join(moveRows(items, c.rows, c.to), ""); got != c.want {
			t.Errorf("moveRows(%v, %d) = %s, want %s", c.rows, c.to, got, c.want)
		}
	}
}
