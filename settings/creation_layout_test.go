package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The filter track ends just after its last segment. A fixed width left an
// empty stretch of track after 我的模板.
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
		track, ok := tester.Find("文件分类")
		if !ok {
			t.Fatalf("%dx%d: filter track not found", size[0], size[1])
		}
		last, ok := tester.Find("我的模板")
		if !ok {
			t.Fatalf("%dx%d: last segment not found", size[0], size[1])
		}
		if gap := track.X + track.W - (last.X + last.W); gap < 0 || gap > slack {
			t.Errorf("%dx%d: track extends %.1f pt past 我的模板, want 0 to %d", size[0], size[1], gap, slack)
		}
	}
}
