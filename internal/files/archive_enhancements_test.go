package files

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestZipSelectedFolders(test *testing.T) {
	root := test.TempDir()
	first := filepath.Join(root, "project")
	second := filepath.Join(root, "photos")
	file := filepath.Join(root, "notes.txt")
	writeFile(test, filepath.Join(first, "nested", "readme.md"), "notes")
	writeFile(test, filepath.Join(second, "photo.txt"), "photo")
	writeFile(test, file, "note")
	for _, scenario := range []struct {
		name  string
		paths []string
		want  []string
	}{
		{name: "one folder", paths: []string{first}, want: []string{"project/nested/readme.md"}},
		{name: "several folders", paths: []string{first, second}, want: []string{"photos/photo.txt", "project/nested/readme.md"}},
		{name: "mixed", paths: []string{first, file}, want: []string{"notes.txt", "project/nested/readme.md"}},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			archive, err := Zip(scenario.paths, test.TempDir(), "bundle.zip")
			if err != nil {
				test.Fatal(err)
			}
			reader, err := zip.OpenReader(archive)
			if err != nil {
				test.Fatal(err)
			}
			defer reader.Close()
			var names []string
			for _, entry := range reader.File {
				if !entry.FileInfo().IsDir() {
					names = append(names, entry.Name)
				}
			}
			slices.Sort(names)
			if !slices.Equal(names, scenario.want) {
				test.Fatalf("archive files = %q, want %q", names, scenario.want)
			}
		})
	}
}

func TestUnzipToFolderKeepsExistingNames(test *testing.T) {
	root := test.TempDir()
	archive := filepath.Join(root, "bundle.v1.ZIP")
	writeZip(test, archive, []zipEntry{{name: "project/readme.md", body: "hello"}, {name: "empty/", mode: os.ModeDir | 0o755}})
	writeFile(test, filepath.Join(root, "bundle.v1"), "keep this file")
	mustMkdir(test, filepath.Join(root, "bundle.v1 2"))
	writeFile(test, filepath.Join(root, "bundle.v1 2", "existing.md"), "keep this folder")
	folder, written, skipped, err := UnzipToFolder(archive)
	if err != nil || written != 1 || skipped != 0 || folder != filepath.Join(root, "bundle.v1 3") {
		test.Fatalf("UnzipToFolder = %q, %d, %d, %v", folder, written, skipped, err)
	}
	if got := readFile(test, filepath.Join(folder, "project", "readme.md")); got != "hello" {
		test.Fatalf("extracted content = %q", got)
	}
	if got := readFile(test, filepath.Join(root, "bundle.v1")); got != "keep this file" {
		test.Fatal("existing file was overwritten")
	}
	if got := readFile(test, filepath.Join(root, "bundle.v1 2", "existing.md")); got != "keep this folder" {
		test.Fatal("existing folder was merged")
	}
	if info, err := os.Stat(filepath.Join(folder, "empty")); err != nil || !info.IsDir() {
		test.Fatal("empty directory missing")
	}
}

func TestUnzipToFolderChecksLimitsBeforeCreatingFolder(test *testing.T) {
	previous := maxUnzipBytes
	test.Cleanup(func() { maxUnzipBytes = previous })
	maxUnzipBytes = 3
	root := test.TempDir()
	archive := filepath.Join(root, "bundle.zip")
	writeZip(test, archive, []zipEntry{{name: "large.txt", body: "too large"}})
	folder, written, _, err := UnzipToFolder(archive)
	if err == nil || folder != "" || written != 0 || exists(filepath.Join(root, "bundle")) {
		test.Fatalf("refused archive left a folder: %q, %d, %v", folder, written, err)
	}
}

func TestUnzipToFolderPartialErrorPreservesCompletedFiles(test *testing.T) {
	root := test.TempDir()
	archive := filepath.Join(root, "bundle.zip")
	writeZip(test, archive, []zipEntry{{name: "good.txt", body: "keep"}, {name: "../escape.txt", body: "refuse"}})
	folder, written, _, err := UnzipToFolder(archive)
	if err == nil || written != 1 || folder != filepath.Join(root, "bundle") {
		test.Fatalf("partial extraction = %q, %d, %v", folder, written, err)
	}
	if got := readFile(test, filepath.Join(folder, "good.txt")); got != "keep" || exists(filepath.Join(root, "escape.txt")) {
		test.Fatal("partial extraction lost content or escaped the output folder")
	}
}

func TestUnzipToFolderSkipsLinksAndKeepsQuarantine(test *testing.T) {
	root := test.TempDir()
	archive := filepath.Join(root, "bundle.zip")
	writeZip(test, archive, []zipEntry{{name: "link", body: "../../outside", mode: os.ModeSymlink | 0o777}, {name: "nested/notes.md", body: "keep"}})
	if err := writeQuarantine(archive, testQuarantine); err != nil {
		test.Fatal(err)
	}
	folder, written, skipped, err := UnzipToFolder(archive)
	if err != nil || written != 1 || skipped != 1 || exists(filepath.Join(folder, "link")) {
		test.Fatalf("link extraction = %q, %d, %d, %v", folder, written, skipped, err)
	}
	for _, path := range []string{folder, filepath.Join(folder, "nested"), filepath.Join(folder, "nested", "notes.md")} {
		value, present, err := readQuarantine(path)
		if err != nil || !present || !bytes.Equal(value, testQuarantine) {
			test.Fatalf("quarantine not preserved for %q: %q, %v", path, value, err)
		}
	}
}

func TestZipExclusionsApplyOnlyBelowSelections(test *testing.T) {
	root := test.TempDir()
	project := filepath.Join(root, "project")
	for _, name := range []string{".DS_Store", ".git/config", "node_modules/package/index.js", "nested/.DS_Store", "nested/.git/config", "nested/node_modules/file.js", "nested/notes.md"} {
		writeFile(test, filepath.Join(project, name), "keep or exclude")
	}
	if err := os.Mkdir(filepath.Join(project, "empty"), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.Symlink("nested/notes.md", filepath.Join(project, "notes-link")); err != nil {
		test.Fatal(err)
	}
	for _, scenario := range []struct {
		name       string
		paths      []string
		exclusions []string
		absent     []string
		present    []string
	}{
		{name: "default cleanup", paths: []string{project}, exclusions: []string{".DS_Store"}, absent: []string{"project/.DS_Store", "project/nested/.DS_Store"}, present: []string{"project/.git/config", "project/node_modules/package/index.js"}},
		{name: "git", paths: []string{project}, exclusions: []string{".git"}, absent: []string{"project/.git/", "project/nested/.git/"}, present: []string{"project/.DS_Store", "project/node_modules/package/index.js"}},
		{name: "dependencies", paths: []string{project}, exclusions: []string{"node_modules"}, absent: []string{"project/node_modules/", "project/nested/node_modules/"}, present: []string{"project/.git/config", "project/.DS_Store"}},
		{name: "all", paths: []string{project}, exclusions: []string{".DS_Store", ".git", "node_modules"}, absent: []string{"project/.DS_Store", "project/.git/", "project/node_modules/", "project/nested/.DS_Store", "project/nested/.git/", "project/nested/node_modules/"}},
		{name: "explicit selection", paths: []string{project, filepath.Join(project, ".DS_Store"), filepath.Join(project, ".git")}, exclusions: []string{".DS_Store", ".git"}, absent: []string{"project/.DS_Store", "project/.git/"}, present: []string{".DS_Store", ".git/config"}},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			archive, err := Zip(scenario.paths, test.TempDir(), "bundle.zip", scenario.exclusions...)
			if err != nil {
				test.Fatal(err)
			}
			reader, err := zip.OpenReader(archive)
			if err != nil {
				test.Fatal(err)
			}
			defer reader.Close()
			var names []string
			for _, entry := range reader.File {
				names = append(names, entry.Name)
			}
			for _, name := range scenario.absent {
				if slices.Contains(names, name) {
					test.Errorf("excluded entry remains: %s", name)
				}
			}
			for _, name := range append(scenario.present, "project/empty/", "project/notes-link", "project/nested/notes.md") {
				if !slices.Contains(names, name) {
					test.Errorf("expected entry absent: %s", name)
				}
			}
		})
	}
}
