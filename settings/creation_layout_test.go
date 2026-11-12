package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/locales"
)

// The filter track ends just after its last segment. A fixed width left an
// empty stretch of track after My Templates.
func TestCreationFilterTrackHugsItsSegments(t *testing.T) {
	const slack = 24 // points allowed between the last label and the track edge
	for _, size := range [][2]int{{minWidth, minHeight}, {mainWidth, mainHeight}} {
		model, err := newApp(testStore(t), testEnv())
		if err != nil {
			t.Fatal(err)
		}
		model.selectPane(paneTemplates)
		tester := ui.NewTester(model.view, size[0], size[1])
		tester.Frame()
		track, ok := tester.Find("File Category")
		if !ok {
			t.Fatalf("%dx%d: filter track not found", size[0], size[1])
		}
		last, ok := tester.Find("My Templates")
		if !ok {
			t.Fatalf("%dx%d: last segment not found", size[0], size[1])
		}
		if gap := track.X + track.W - (last.X + last.W); gap < 0 || gap > slack {
			t.Errorf("%dx%d: track extends %.1f pt past My Templates, want 0 to %d", size[0], size[1], gap, slack)
		}
	}
}

// TestNewPaneHeaderControlsShareABaseline checks the two header actions and the
// search field against each other: the New pane reads as one toolbar only when
// their heights, baselines and right edges match.
func TestNewPaneHeaderControlsShareABaseline(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	model.selectPane(paneTemplates)
	tester := ui.NewTester(model.view, mainWidth, mainHeight)
	tester.Frame()
	addTemplate, ok := tester.Find(locales.AddTemplate())
	if !ok {
		t.Fatal("the Add Template button is missing")
	}
	addType, ok := tester.Find(locales.AddFileType())
	if !ok {
		t.Fatal("the Add File Type button is missing")
	}
	search, ok := tester.Find(locales.SearchFileTypes())
	if !ok {
		t.Fatal("the search field is missing")
	}
	if addTemplate.H != addType.H {
		t.Errorf("button heights differ: template %.1f, type %.1f", addTemplate.H, addType.H)
	}
	templateMiddle := addTemplate.Y + addTemplate.H/2
	typeMiddle := addType.Y + addType.H/2
	if diff := templateMiddle - typeMiddle; diff > 1 || diff < -1 {
		t.Errorf("button baselines differ by %.1f pt", diff)
	}
	searchRight := search.X + search.W
	typeRight := addType.X + addType.W
	if diff := searchRight - typeRight; diff > 1 || diff < -1 {
		t.Errorf("search right edge is %.1f pt off the button edge", diff)
	}
}
