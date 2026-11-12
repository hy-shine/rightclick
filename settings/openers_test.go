package settings

import (
	"slices"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/locales"
)

// TestProgramSearchFiltersTheList checks that the search box keeps only the
// matching programs, and says so when none match.
func TestProgramSearchFiltersTheList(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	m.cfg.Programs = []string{"Visual Studio Code", "Cursor", "Terminal"}
	m.selectPane(paneOpeners)
	m.programSearch = "CUR"
	texts := ui.NewTester(m.view, minWidth, 900).Texts()
	if !slices.Contains(texts, "Cursor") || slices.Contains(texts, "Visual Studio Code") {
		t.Errorf("search for CUR shows %q, want Cursor only", texts)
	}
	m.programSearch = "no-such-app"
	texts = ui.NewTester(m.view, minWidth, 900).Texts()
	if !slices.Contains(texts, locales.NoProgramMatches()) {
		t.Errorf("empty search state is missing: %q", texts)
	}
}

// TestChooseAppPanelDoesNotGreyOutBundles checks the picker options: an "app"
// extension filter would grey out every application bundle in the panel.
func TestChooseAppPanelDoesNotGreyOutBundles(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	opts := make(chan mygo.OpenDialogOptions, 1)
	m.env.openDialog = func(o mygo.OpenDialogOptions) ([]string, error) {
		opts <- o
		return nil, nil
	}
	m.pickEditor()
	o := <-opts
	if o.DefaultPath != "/Applications" {
		t.Errorf("the panel opens at %q, want /Applications", o.DefaultPath)
	}
	for _, f := range o.Filters {
		if slices.Contains(f.Extensions, "app") {
			t.Errorf("the panel filters by the app extension, which greys out bundles: %+v", f)
		}
	}
}

func TestApplicationName(t *testing.T) {
	tests := []struct {
		path    string
		want    string
		wantErr bool
	}{
		{path: "/Applications/Visual Studio Code.app", want: "Visual Studio Code"},
		{path: "/Applications/Warp.app", want: "Warp"},
		{path: "/Applications/NotAnApp", wantErr: true},
	}
	for _, tt := range tests {
		got, err := applicationName(tt.path)
		if (err != nil) != tt.wantErr {
			t.Fatalf("applicationName(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("applicationName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestEditorRowsNameTheirCheckBox keeps the second list without a header from
// carrying an unexplained box: the label says what it adds, and the row still
// fits the smallest window.
func TestEditorRowsNameTheirCheckBox(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneOpeners)
	screen := ui.NewTester(model.view, minWidth, 900)
	if !slices.Contains(screen.Texts(), locales.QuickAccess()) {
		t.Errorf("the editor rows have no label for the check box: %q", screen.Texts())
	}
	remove, ok := screen.Find(locales.Remove())
	if !ok {
		t.Fatal("the remove button is missing")
	}
	if remove.X+remove.W > minWidth {
		t.Errorf("the remove button ends at %.1f pt, past the window width %d", remove.X+remove.W, minWidth)
	}
}
