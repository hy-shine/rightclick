package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"rightclick/config"
	"rightclick/files"
	"rightclick/finder"
)

// routeURL handles a rightclick:// URL sent by the Finder extension. It parses the
// URL, loads the settings, and runs the action. A failure is shown as a
// notification; an ordinary file action that succeeds shows nothing.
func routeURL(raw string) {
	action, params, paths, err := parseURL(raw)
	if err != nil {
		notifyFailure(err)
		return
	}
	cfg, err := loadConfig()
	if err != nil {
		notifyFailure(err)
		return
	}
	if err := requireInWatchFolders(paths, cfg.WatchFolders); err != nil {
		notifyFailure(err)
		return
	}
	if err := runAction(action, params, paths, cfg); err != nil {
		notifyFailure(err)
	}
}

// requireInWatchFolders keeps every action inside the same folder whitelist
// that controls where the Finder menu appears.
func requireInWatchFolders(paths, folders []string) error {
	for _, path := range paths {
		if !config.InWatchFolders(path, folders) {
			return fmt.Errorf("不在文件夹范围内：%s", path)
		}
	}
	return nil
}

// loadConfig reads the saved settings, or the defaults when none are saved yet.
func loadConfig() (config.Config, error) {
	path, err := config.DefaultPath()
	if err != nil {
		return config.Config{}, err
	}
	return config.Load(path)
}

// actionHandler runs one rightclick:// action. params holds the URL's query
// values, paths the absolute paths it names, and cfg the saved settings.
type actionHandler func(params url.Values, paths []string, cfg config.Config) error

// actionHandlers maps each action name that a rightclick:// URL may name to its
// handler. It is the only list of accepted actions: parseURL checks names against
// it, and runAction dispatches through it. The Finder menu lists the same names;
// a test keeps the two lists in step.
func actionHandlers() map[string]actionHandler {
	return map[string]actionHandler{
		"copy-path": func(_ url.Values, paths []string, _ config.Config) error {
			return copyText(joinLines(paths))
		},
		"copy-name": func(_ url.Values, paths []string, _ config.Config) error {
			return copyText(joinLines(baseNames(paths)))
		},
		"new-file": func(params url.Values, paths []string, cfg config.Config) error {
			return createFileFromType(cfg, params.Get("type"), paths)
		},
		// Format version 1 links stay valid while an older Finder extension is
		// still loaded.
		"new-text": func(_ url.Values, paths []string, cfg config.Config) error {
			return createFileFromType(cfg, "text", paths)
		},
		"new-markdown": func(_ url.Values, paths []string, cfg config.Config) error {
			return createFileFromType(cfg, "markdown", paths)
		},
		"new-folder": func(_ url.Values, paths []string, _ config.Config) error {
			return createFolder(paths)
		},
		"new-template": func(params url.Values, paths []string, cfg config.Config) error {
			templatesPath, err := files.DefaultTemplatesPath()
			if err != nil {
				return err
			}
			return createFromTemplate(templatesPath, params.Get("name"), paths, cfg)
		},
		"copy-to":       transferAction("copy-to"),
		"move-to":       transferAction("move-to"),
		"open-terminal": openTerminal,
		"open-editor":   openWithApp,
		"open-with":     openWithApp,
		"zip": func(_ url.Values, paths []string, _ config.Config) error {
			return zipItems(paths)
		},
		"unzip": unzipAction,
		"hash": func(params url.Values, paths []string, _ config.Config) error {
			return hashItem(params.Get("algo"), paths)
		},
		"compare-hash": func(params url.Values, paths []string, _ config.Config) error {
			return compareItems(params.Get("algo"), paths)
		},
		"toggle-hidden": func(_ url.Values, _ []string, _ config.Config) error {
			return toggleHiddenFiles()
		},
	}
}

// runAction performs one action and returns its error.
func runAction(action string, params url.Values, paths []string, cfg config.Config) error {
	handler, ok := actionHandlers()[action]
	if !ok {
		return fmt.Errorf("不认识的操作：%s", action)
	}
	return handler(params, paths, cfg)
}

// createFolder makes a folder in the folder the action was started in.
func createFolder(paths []string) error {
	dir, err := requireOne(paths)
	if err != nil {
		return err
	}
	_, err = files.CreateFolder(dir, "新建文件夹")
	return err
}

// createFileFromType makes one of the configured file types in the folder the
// action was started in. The type must still be visible in the Finder menu.
func createFileFromType(cfg config.Config, id string, paths []string) error {
	dir, err := requireOne(paths)
	if err != nil {
		return err
	}
	fileType, ok := configFileType(cfg, id)
	if !ok {
		return fmt.Errorf("找不到文件类型：%s", id)
	}
	if !fileType.Enabled || fileType.Location == config.FileTypeHidden {
		return fmt.Errorf("文件类型已隐藏：%s", fileType.Label)
	}
	name, err := fileTypeFileName(fileType)
	if err != nil {
		return err
	}
	content, err := files.FileTypeContent(fileType)
	if err != nil {
		return err
	}
	path, err := files.CreateFile(dir, name, content)
	if err != nil {
		return err
	}
	if cfg.OpenAfterCreate {
		if err := files.Open(path); err != nil {
			return fmt.Errorf("文件已创建，但无法打开 %s: %w", files.Name(path), err)
		}
	}
	return nil
}

// configFileType returns the configured file type with id.
func configFileType(cfg config.Config, id string) (config.FileType, bool) {
	for _, fileType := range cfg.FileTypes {
		if fileType.ID == id {
			return fileType, true
		}
	}
	return config.FileType{}, false
}

// fileTypeFileName returns the default file name, including its extension.
func fileTypeFileName(fileType config.FileType) (string, error) {
	base := strings.TrimSpace(fileType.DefaultName)
	if base == "" {
		base = "未命名"
	}
	ext := strings.TrimPrefix(strings.TrimSpace(fileType.Extension), ".")
	if base == "" || strings.ContainsAny(base, `/\`) {
		return "", fmt.Errorf("默认文件名无效：%q", fileType.DefaultName)
	}
	if ext == "" || strings.ContainsAny(ext, `/\`) {
		return "", fmt.Errorf("扩展名无效：%q", fileType.Extension)
	}
	if strings.EqualFold(filepath.Ext(base), "."+ext) {
		return base, nil
	}
	return base + "." + ext, nil
}

// createFromTemplate makes a file in the folder from the template called name in
// the templates file at templatesPath. The file is named after the template and
// starts with its content. When the settings ask for it, the file is opened
// afterwards. A failed open is reported after the file exists, as in
// createFileFromType.
func createFromTemplate(templatesPath, name string, paths []string, cfg config.Config) error {
	dir, err := requireOne(paths)
	if err != nil {
		return err
	}
	templates, err := files.LoadTemplates(templatesPath)
	if err != nil {
		return err
	}
	tpl, err := findTemplate(templates, name)
	if err != nil {
		return err
	}
	if tpl.Disabled {
		return fmt.Errorf("模板已停用：%s", tpl.Name)
	}
	content, err := tpl.Bytes()
	if err != nil {
		return err
	}
	created, err := files.CreateFile(dir, tpl.Name, content)
	if err != nil {
		return err
	}
	if cfg.OpenAfterCreate {
		if err := files.Open(created); err != nil {
			return fmt.Errorf("文件已创建，但无法打开 %s: %w", files.Name(created), err)
		}
	}
	return nil
}

// findTemplate returns the template whose name is name. The template name becomes
// a file name, so a name that is empty or has a folder part is refused.
func findTemplate(templates []files.Template, name string) (files.Template, error) {
	for _, t := range templates {
		if t.Name != name {
			continue
		}
		if t.Name == "" || strings.Contains(t.Name, "/") {
			return files.Template{}, fmt.Errorf("模板名不能用作文件名：%q", t.Name)
		}
		return t, nil
	}
	return files.Template{}, fmt.Errorf("找不到模板：%s", name)
}

// transferOp returns the file operation behind copy-to or move-to.
func transferOp(action string) (func(src, dir string) (string, error), error) {
	switch action {
	case "copy-to":
		return files.Copy, nil
	case "move-to":
		return files.Move, nil
	}
	return nil, fmt.Errorf("不认识的操作：%s", action)
}

// transferAction returns the handler for copy-to or move-to.
func transferAction(action string) actionHandler {
	return func(_ url.Values, paths []string, _ config.Config) error {
		return transferItems(action, paths)
	}
}

// transferItems asks for a destination folder, then copies or moves each item into
// it. A cancelled dialog does nothing.
func transferItems(action string, paths []string) error {
	op, err := transferOp(action)
	if err != nil {
		return err
	}
	title := "复制到"
	if action == "move-to" {
		title = "移动到"
	}
	dir, err := chooseFolder(title)
	if err != nil {
		return err
	}
	if dir == "" {
		return nil
	}
	return transferTo(op, paths, dir)
}

// transferTo runs op on each path with dir as the destination. Every item is
// tried; the errors of the failed ones are returned together.
func transferTo(op func(src, dir string) (string, error), paths []string, dir string) error {
	var errs []error
	for _, p := range paths {
		if _, err := op(p, dir); err != nil {
			errs = append(errs, fmt.Errorf("%s：%w", files.Name(p), err))
		}
	}
	return errors.Join(errs...)
}

// chooseFolder shows a folder picker and returns the chosen folder, or "" when the
// user cancels.
func chooseFolder(title string) (string, error) {
	dirs, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Title:             title,
		Message:           "选择目标文件夹",
		ButtonLabel:       "选择",
		Directory:         true,
		CreateDirectories: true,
	})
	if err != nil {
		return "", err
	}
	if len(dirs) == 0 {
		return "", nil
	}
	return dirs[0], nil
}

// openTerminal opens each folder with the terminal app from the settings.
func openTerminal(_ url.Values, paths []string, cfg config.Config) error {
	return openEach(cfg.Terminal, paths)
}

// openWithApp handles open-editor and open-with, which name the app in the app
// parameter.
func openWithApp(params url.Values, paths []string, _ config.Config) error {
	app := params.Get("app")
	if app == "" {
		return errors.New("网址里没有应用名")
	}
	return openEach(app, paths)
}

// openEach opens each path with app. Every path is tried; the errors of the
// failed ones are returned together.
func openEach(app string, paths []string) error {
	return openEachWith(files.OpenWith, app, paths)
}

// openEachWith calls open for each path. It takes the open call as a parameter so
// that tests can replace the system call.
func openEachWith(open func(app, path string) error, app string, paths []string) error {
	var errs []error
	for _, p := range paths {
		if err := open(app, p); err != nil {
			errs = append(errs, fmt.Errorf("无法用 %s 打开 %s：%w", app, files.Name(p), err))
		}
	}
	return errors.Join(errs...)
}

// archiveName returns the zip file name for paths: the name of the item for one
// path, or a fixed name for several.
func archiveName(paths []string) string {
	if len(paths) == 1 {
		return files.BaseName(paths[0]) + ".zip"
	}
	return "归档.zip"
}

// zipItems archives paths into the folder that holds the first one.
func zipItems(paths []string) error {
	_, err := files.Zip(paths, filepath.Dir(paths[0]), archiveName(paths))
	return err
}

// unzipItems extracts each archive next to itself. Every archive is tried. It
// returns how many files were written in total and the errors of the archives
// that failed, joined.
func unzipItems(paths []string) (int, error) {
	written := 0
	var errs []error
	for _, p := range paths {
		n, _, err := files.Unzip(p)
		written += n
		if err != nil {
			errs = append(errs, fmt.Errorf("%s：%w", files.Name(p), err))
		}
	}
	return written, errors.Join(errs...)
}

// unzipAction runs unzipItems and shows how many files it wrote. A failure
// reports the count too, because the archives that succeeded stay extracted.
func unzipAction(_ url.Values, paths []string, _ config.Config) error {
	written, err := unzipItems(paths)
	if err != nil {
		return fmt.Errorf("已写入 %d 个文件，%w", written, err)
	}
	notify("解压完成", fmt.Sprintf("写入 %d 个文件", written))
	return nil
}

// toggleHiddenFiles shows hidden files in Finder when they are hidden, and hides
// them when they are shown. Finder restarts to apply the change, and a
// notification says which state is now on.
func toggleHiddenFiles() error {
	shown, err := finder.HiddenFilesShown()
	if err != nil {
		return err
	}
	if err := finder.SetHiddenFilesShown(!shown); err != nil {
		return err
	}
	if shown {
		notify("隐藏文件", "已隐藏")
	} else {
		notify("隐藏文件", "已显示")
	}
	return nil
}

// hashItem computes the digest of the one selected file, copies it to the
// clipboard, and shows it in a notification.
func hashItem(algoName string, paths []string) error {
	algo, path, digest, err := digestOf(algoName, paths)
	if err != nil {
		return err
	}
	if err := copyText(digest); err != nil {
		return err
	}
	notify(string(algo)+" 已拷贝", files.Name(path)+"\n"+digest)
	return nil
}

// digestOf returns the algorithm, the path, and the digest of the one selected
// file.
func digestOf(algoName string, paths []string) (files.Algorithm, string, string, error) {
	algo, err := hashAlgorithm(algoName)
	if err != nil {
		return "", "", "", err
	}
	path, err := requireOne(paths)
	if err != nil {
		return "", "", "", err
	}
	digest, err := files.Hash(path, algo)
	if err != nil {
		return "", "", "", err
	}
	return algo, path, digest, nil
}

// compareItems compares the digests of the two selected files and shows the result.
func compareItems(algoName string, paths []string) error {
	title, body, err := compareFiles(algoName, paths)
	if err != nil {
		return err
	}
	notify(title, body)
	return nil
}

// compareFiles compares the digests of the two selected files and returns the
// notification title and body that report the result.
func compareFiles(algoName string, paths []string) (title, body string, err error) {
	algo, err := hashAlgorithm(algoName)
	if err != nil {
		return "", "", err
	}
	if len(paths) != 2 {
		return "", "", errors.New("比较哈希需要选中两个文件")
	}
	same, hashA, hashB, err := files.CompareHashes(paths[0], paths[1], algo)
	if err != nil {
		return "", "", err
	}
	title, body = compareSummary(algo, same, paths[0], hashA, paths[1], hashB)
	return title, body, nil
}

// compareSummary returns the notification title and body for a hash comparison.
// The title gives the verdict; the body lists both files with their digests.
func compareSummary(algo files.Algorithm, same bool, pathA, hashA, pathB, hashB string) (title, body string) {
	verdict := "不同"
	if same {
		verdict = "相同"
	}
	title = fmt.Sprintf("%s 比较：%s", algo, verdict)
	body = fmt.Sprintf("%s\n%s\n%s\n%s", files.Name(pathA), hashA, files.Name(pathB), hashB)
	return title, body
}

// hashAlgorithm maps the algo parameter to a hash function. An empty parameter
// means SHA-256.
func hashAlgorithm(name string) (files.Algorithm, error) {
	switch name {
	case "", string(files.SHA256):
		return files.SHA256, nil
	case string(files.SHA1):
		return files.SHA1, nil
	case string(files.MD5):
		return files.MD5, nil
	}
	return "", fmt.Errorf("不支持的哈希算法：%s", name)
}

// requireOne returns the only path. Actions that work on one folder or file use it.
func requireOne(paths []string) (string, error) {
	if len(paths) != 1 {
		return "", errors.New("需要且只能选中一个项目")
	}
	return paths[0], nil
}

// joinLines joins lines with newlines, the form that copy-path and copy-name put
// on the clipboard.
func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// baseNames returns the last element of each path.
func baseNames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, p := range paths {
		names = append(names, files.Name(p))
	}
	return names
}

// copyText replaces the clipboard with text.
func copyText(text string) error {
	return mygo.Clipboard.Write(transfer.TextData(text))
}

// notifyFailure shows an error in a notification.
func notifyFailure(err error) {
	notify("操作失败", err.Error())
}

// notify shows a notification. A failed notification is only logged, because the
// action has already run and its result should not be hidden by a notification error.
func notify(title, body string) {
	err := mygo.NewNotification(mygo.NotificationOptions{Title: title, Body: body}).Show()
	if err != nil {
		log.Printf("通知「%s」失败：%v", title, err)
	}
}
