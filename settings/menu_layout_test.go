package settings

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestMenuNamesStayOnOneLineInNarrowEditor(t *testing.T) {
	model, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tester := ui.NewTester(func(context *ui.Context) {
		context.SetTheme(paperTheme(true))
		ui.Column(context).Width(680).Children(func() {
			for group := 2; group < len(menuGroups); group++ {
				model.menuGroup(context, group, model.groupRows(group))
			}
		})
	}, 680, 800)
	for _, label := range []string{"在终端打开", "用默认编辑器打开", "压缩为 ZIP", "比较两个文件的哈希", "显示/隐藏隐藏文件（⌘⇧.）"} {
		bounds, found := tester.Find(label)
		if !found || bounds.H > 22 || bounds.W < 100 {
			t.Errorf("menu name %q is not a readable single line: found=%v bounds=%+v", label, found, bounds)
		}
	}
}
