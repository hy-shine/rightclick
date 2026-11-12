package settings

import (
	"slices"
	"testing"

	"rightclick/config"
)

// previewHas reports whether the preview shows label in place ctx.
func previewHas(m *app, ctx int, label string) bool {
	for _, item := range m.previewTopItems(ctx) {
		if item.label == label {
			return true
		}
	}
	for _, g := range m.previewGroups(ctx) {
		if slices.ContainsFunc(g, func(it previewItem) bool { return it.label == label }) {
			return true
		}
	}
	return false
}

func TestPreviewFollowsContext(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		ctx   int
		label string
		want  bool
	}{
		{ctxBlank, "新建文件", true},
		{ctxBlank, "拷贝路径", false},
		{ctxFolder, "在终端打开", true},
		{ctxFolder, "压缩为 ZIP", false}, // zip needs files, not only a folder
		{ctxFile, "计算哈希", true},
		{ctxFile, "比较两个文件的哈希", false},
		{ctxFile, "解压到此处", true},
		{ctxTwoFiles, "比较两个文件的哈希", true},
		{ctxTwoFiles, "计算哈希", false},
	}
	for _, tc := range cases {
		if got := previewHas(m, tc.ctx, tc.label); got != tc.want {
			t.Errorf("place %d, %q shown = %v, want %v", tc.ctx, tc.label, got, tc.want)
		}
	}
}

func TestPreviewHidesDisabledEntries(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	if !previewHas(m, ctxBlank, "新建文件") {
		t.Fatal("新建文件 should show on blank space before it is switched off")
	}
	m.cfg.MenuItems[0].Enabled = false // new-file is the first entry of config.Default
	if previewHas(m, ctxBlank, "新建文件") {
		t.Error("新建文件 still shows after it is switched off")
	}
}

func TestPreviewKeepsFileTypesInsideNewFile(t *testing.T) {
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	for i := range m.cfg.FileTypes {
		if m.cfg.FileTypes[i].ID == "json" {
			m.cfg.FileTypes[i].Location = config.FileTypeMain
		}
	}
	if previewHas(m, ctxBlank, "新建 JSON") {
		t.Error("JSON leaked to the top level of the preview")
	}
	if !previewHas(m, ctxBlank, "新建文件") {
		t.Error("New File is missing from the preview")
	}
	for i := range m.cfg.FileTypes {
		m.cfg.FileTypes[i].Enabled = false
	}
	if previewHas(m, ctxBlank, "新建文件") {
		t.Error("新建文件 still shows when every file type is hidden")
	}
}
