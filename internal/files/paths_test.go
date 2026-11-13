package files

import (
	"bytes"
	"os/exec"
	"slices"
	"testing"
)

func TestTerminalPathsRoundTrip(test *testing.T) {
	paths := []string{
		"/Work/My Project/report.txt",
		"/Work/it's a file.txt",
		"/Work/$(printf substitution)`printf backtick`$HOME; & (draft).txt",
		"/Work/back\\slash\nnext line.txt",
		"/Work/中文/说明.md",
	}
	for _, shell := range []string{"zsh", "bash"} {
		test.Run(shell, func(test *testing.T) {
			binary, err := exec.LookPath(shell)
			if err != nil {
				test.Skipf("%s unavailable: %v", shell, err)
			}
			for _, selection := range [][]string{paths[:1], paths} {
				command := "set -- " + TerminalPaths(selection) + "; printf '%s\\000' \"$@\""
				output, err := exec.CommandContext(test.Context(), binary, "-c", command).Output()
				if err != nil {
					test.Fatal(err)
				}
				var arguments []string
				for _, argument := range bytes.Split(bytes.TrimSuffix(output, []byte{0}), []byte{0}) {
					arguments = append(arguments, string(argument))
				}
				if !slices.Equal(arguments, selection) {
					test.Fatalf("arguments = %q, want %q", arguments, selection)
				}
			}
		})
	}
}

func TestRelativePaths(test *testing.T) {
	for _, scenario := range []struct {
		name  string
		base  string
		paths []string
		want  string
	}{
		{name: "nested file", base: "/Work/demo", paths: []string{"/Work/demo/assets/logo.png"}, want: "assets/logo.png"},
		{name: "base itself", base: "/Work/demo", paths: []string{"/Work/demo"}, want: "."},
		{name: "outside base", base: "/Work/demo", paths: []string{"/Work/shared/notes.md"}, want: "../shared/notes.md"},
		{name: "ordered selection", base: "/Work/demo", paths: []string{"/Work/demo/assets", "/Work/demo/report.txt"}, want: "assets\nreport.txt"},
		{name: "unicode", base: "/Work/项目", paths: []string{"/Work/项目/文档/说明.md"}, want: "文档/说明.md"},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			got, err := RelativePaths(scenario.paths, scenario.base)
			if err != nil || got != scenario.want {
				test.Fatalf("RelativePaths = %q, %v; want %q", got, err, scenario.want)
			}
		})
	}
	for _, scenario := range []struct {
		base  string
		paths []string
	}{
		{base: "", paths: []string{"/Work/demo/report.txt"}},
		{base: "demo", paths: []string{"/Work/demo/report.txt"}},
		{base: "/Work/demo", paths: []string{"/Work/demo/report.txt", "relative.txt"}},
	} {
		if got, err := RelativePaths(scenario.paths, scenario.base); err == nil || got != "" {
			test.Fatalf("accepted invalid paths: %q, %v", got, err)
		}
	}
}

func TestMarkdownFileLinks(test *testing.T) {
	for _, scenario := range []struct {
		name  string
		paths []string
		want  string
	}{
		{name: "nested", paths: []string{"/Work/demo/assets/logo.png"}, want: `[logo.png](assets/logo.png)`},
		{name: "punctuation", paths: []string{"/Work/demo/[draft] (v2)#?.md"}, want: `[\[draft\] (v2)#?.md](%5Bdraft%5D%20%28v2%29%23%3F.md)`},
		{name: "backslash", paths: []string{"/Work/demo/back\\slash.md"}, want: `[back\\slash.md](back%5Cslash.md)`},
		{name: "colon", paths: []string{"/Work/demo/report:final.md"}, want: `[report:final.md](report%3Afinal.md)`},
		{name: "markup", paths: []string{"/Work/demo/<a>&*_`name.md"}, want: `[&lt;a&gt;&amp;\*\_\` + "`" + `name.md](%3Ca%3E%26%2A_%60name.md)`},
		{name: "entity", paths: []string{"/Work/demo/&copy;.md"}, want: `[&amp;copy;.md](%26copy;.md)`},
		{name: "unicode", paths: []string{"/Work/demo/说明.md"}, want: `[说明.md](%E8%AF%B4%E6%98%8E.md)`},
		{name: "selection", paths: []string{"/Work/demo/one.md", "/Work/two.md"}, want: "[one.md](one.md)\n[two.md](../two.md)"},
		{name: "newline", paths: []string{"/Work/demo/line\nname.md"}, want: `[line&#10;name.md](line%0Aname.md)`},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			got, err := MarkdownFileLinks(scenario.paths, "/Work/demo")
			if err != nil || got != scenario.want {
				test.Fatalf("MarkdownFileLinks = %q, %v; want %q", got, err, scenario.want)
			}
		})
	}
}
