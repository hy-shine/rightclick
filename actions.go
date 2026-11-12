package main

import (
	"errors"
	"fmt"
	"github.com/hy-shine/rightclick/locales"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/hy-shine/rightclick/config"
	"github.com/hy-shine/rightclick/files"
	"github.com/hy-shine/rightclick/finder"
	"github.com/hy-shine/rightclick/settings"
)

// routeURL handles a rightclick:// URL sent by the Finder extension. The action
// runs on a goroutine of its own, so that slow file work does not hold up the
// main thread that delivers URLs. The goroutine ends when the action returns.
func routeURL(raw string) {
	routeURLWith(raw, func(run func()) { go run() })
}

// routeURLWith parses the URL and checks it against the settings at once, then
// hands the action to start, which decides when it runs. A failure before that
// point, or in the action, is shown as a notification; an ordinary file action
// that succeeds shows nothing.
func routeURLWith(raw string, start func(run func())) {
	action, params, paths, err := parseURL(raw)
	if err != nil {
		routeFailure(raw, err)
		return
	}
	cfg, err := loadConfig()
	if err != nil {
		routeFailure(raw, err)
		return
	}
	if err := requireInWatchFolders(paths, cfg.WatchFolders); err != nil {
		routeFailure(raw, err)
		return
	}
	start(func() {
		if err := runAction(action, params, paths, cfg); err != nil {
			routeFailure(raw, err)
		}
	})
}

// routeFailure logs the URL and error before notifying, so a missing notification
// still leaves a record in Console for diagnosis.
func routeFailure(raw string, err error) {
	log.Printf("rightclick %s: %v", raw, err)
	notifyFailure(err)
}

// requireInWatchFolders keeps every action inside the same folder whitelist
// that controls where the Finder menu appears.
func requireInWatchFolders(paths, folders []string) error {
	for _, path := range paths {
		if !config.InWatchFolders(path, folders) {
			return errors.New(locales.OutsideFolders(path))
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
		"copy-basename": func(_ url.Values, paths []string, _ config.Config) error {
			return copyText(joinLines(stemNames(paths)))
		},
		"copy-extension": func(_ url.Values, paths []string, _ config.Config) error {
			extensions := extensionNames(paths)
			if len(extensions) == 0 {
				return errors.New(locales.NoExtensions())
			}
			return copyText(joinLines(extensions))
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
		"add-template": func(_ url.Values, paths []string, _ config.Config) error {
			return captureTemplate(paths, func(template files.Template) error {
				result := make(chan error, 1)
				mygo.RunOnMain(func() { result <- settings.OpenTemplateDraft(template) })
				return <-result
			})
		},
		"copy-to": transferAction("copy-to"),
		"move-to": transferAction("move-to"),
		"open-favorite": func(params url.Values, _ []string, cfg config.Config) error {
			path, err := cfg.FavoritePath(params.Get("favorite"))
			if err != nil {
				return err
			}
			return files.Open(path)
		},
		"open-terminal": openTerminal,
		"open-editor":   openEditor,
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
		"delete-permanently": func(_ url.Values, paths []string, _ config.Config) error {
			return deletePermanently(paths)
		},
	}
}

// runAction performs one action and returns its error.
func runAction(action string, params url.Values, paths []string, cfg config.Config) error {
	handler, ok := actionHandlers()[action]
	if !ok {
		return errors.New(locales.ErrUnknownAction(action))
	}
	return handler(params, paths, cfg)
}

func captureTemplate(paths []string, present func(files.Template) error) error {
	path, err := requireOne(paths)
	if err != nil {
		return err
	}
	template, err := files.TemplateFromFile(path)
	if err != nil {
		return err
	}
	return present(template)
}

// createFolder makes a folder in the folder the action was started in.
func createFolder(paths []string) error {
	dir, err := requireOne(paths)
	if err != nil {
		return err
	}
	_, err = files.CreateFolder(dir, locales.DefaultNewFolder())
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
		return errors.New(locales.FileTypeNotFound(id))
	}
	if !fileType.Enabled || fileType.Location == config.FileTypeHidden {
		return errors.New(locales.FileTypeHidden(fileType.DisplayLabel()))
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
			return errors.New(locales.CreatedButOpenFailed(files.Name(path), err.Error()))
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
	base := strings.TrimSpace(fileType.DisplayDefaultName())
	ext := strings.TrimPrefix(strings.TrimSpace(fileType.Extension), ".")
	if base == "" || strings.ContainsAny(base, `/\`) {
		return "", errors.New(locales.InvalidDefaultName(base))
	}
	if ext == "" || strings.ContainsAny(ext, `/\`) {
		return "", errors.New(locales.InvalidExtension(fileType.Extension))
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
		return errors.New(locales.TemplateDisabled(tpl.Name))
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
			return errors.New(locales.CreatedButOpenFailed(files.Name(created), err.Error()))
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
			return files.Template{}, errors.New(locales.TemplateNameInvalid(t.Name))
		}
		return t, nil
	}
	return files.Template{}, errors.New(locales.TemplateNotFound(name))
}

// transferOp returns the file operation behind copy-to or move-to.
func transferOp(action string) (func(src, dir string) (string, error), error) {
	switch action {
	case "copy-to":
		return files.Copy, nil
	case "move-to":
		return files.Move, nil
	}
	return nil, errors.New(locales.ErrUnknownAction(action))
}

// transferAction returns the handler for copy-to or move-to.
func transferAction(action string) actionHandler {
	return func(params url.Values, paths []string, cfg config.Config) error {
		if params.Has("favorite") {
			destination, err := cfg.FavoritePath(params.Get("favorite"))
			if err != nil {
				return err
			}
			operation, err := transferOp(action)
			if err != nil {
				return err
			}
			return transferTo(operation, paths, destination)
		}
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
	title := locales.CopyToTitle()
	if action == "move-to" {
		title = locales.MoveToTitle()
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
			errs = append(errs, fmt.Errorf("%s: %w", files.Name(p), err))
		}
	}
	return errors.Join(errs...)
}

// chooseFolder shows a folder picker and returns the chosen folder, or "" when the
// user cancels.
func chooseFolder(title string) (string, error) {
	dirs, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Title:             title,
		Message:           locales.ChooseDestination(),
		ButtonLabel:       locales.ChooseButton(),
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

// openTerminal opens each folder in Terminal.
func openTerminal(params url.Values, paths []string, _ config.Config) error {
	app := strings.TrimSpace(params.Get("app"))
	if app == "" {
		app = "Terminal"
	}
	return openEach(app, foldersOf(paths))
}

// openEditor opens the selection with its default application. A menu entry may
// still name a program, which keeps older URLs working.
func openEditor(params url.Values, paths []string, cfg config.Config) error {
	if app := strings.TrimSpace(params.Get("app")); app != "" {
		return openWithApp(params, paths, cfg)
	}
	return openDefaultEach(paths)
}

// openWithApp opens the selection in the named app. Terminal programs get the
// folder that holds the selection, because they cannot open a file directly.
func openWithApp(params url.Values, paths []string, _ config.Config) error {
	app := params.Get("app")
	if app == "" {
		return errors.New(locales.ErrURLNoApp())
	}
	if config.IsTerminalApp(app) {
		paths = foldersOf(paths)
	}
	return openEach(app, paths)
}

// foldersOf returns the folder for each path: the path itself when it is a
// folder, otherwise its parent.
func foldersOf(paths []string) []string {
	folders := make([]string, 0, len(paths))
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			folders = append(folders, path)
			continue
		}
		folders = append(folders, filepath.Dir(path))
	}
	return folders
}

// openWith and openDefault are the system open calls. Tests replace them.
var openWith = files.OpenWith
var openDefault = files.Open

// openEach opens each path with app. Every path is tried; the errors of the
// failed ones are returned together.
func openEach(app string, paths []string) error {
	return openEachWith(openWith, app, paths)
}

// openDefaultEach opens each path with its default application. Every path is
// tried; the errors of the failed ones are returned together.
func openDefaultEach(paths []string) error {
	var errs []error
	for _, p := range paths {
		if err := openDefault(p); err != nil {
			errs = append(errs, errors.New(locales.OpenFailed(files.Name(p), err.Error())))
		}
	}
	return errors.Join(errs...)
}

// openEachWith calls open for each path. It takes the open call as a parameter so
// that tests can replace the system call.
func openEachWith(open func(app, path string) error, app string, paths []string) error {
	var errs []error
	for _, p := range paths {
		if err := open(app, p); err != nil {
			errs = append(errs, errors.New(locales.OpenWithFailed(files.Name(p), app, err.Error())))
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
	return locales.DefaultArchiveName()
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
			errs = append(errs, fmt.Errorf("%s: %w", files.Name(p), err))
		}
	}
	return written, errors.Join(errs...)
}

// unzipAction runs unzipItems and shows how many files it wrote. A failure
// reports the count too, because the archives that succeeded stay extracted.
func unzipAction(_ url.Values, paths []string, _ config.Config) error {
	written, err := unzipItems(paths)
	if err != nil {
		return errors.New(locales.UnzipPartial(written, err.Error()))
	}
	notify(locales.NotifyUnzipped(), locales.UnzipWritten(written))
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
		notify(locales.NotifyHiddenFiles(), locales.NotifyHidden())
	} else {
		notify(locales.NotifyHiddenFiles(), locales.NotifyShown())
	}
	return nil
}

// deletePermanently deletes the items for good once the user confirms. The dialog
// lists the names and says the deletion cannot be undone. Cancel is the default
// answer, so Return does nothing. A protected location is refused before the dialog.
func deletePermanently(paths []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	for _, p := range paths {
		if files.Protected(p, home) {
			return errors.New(locales.DeleteRefused(p))
		}
	}
	answer, err := mygo.Dialog.Message(mygo.MessageOptions{
		Type:          mygo.MessageWarning,
		Message:       locales.DeletePermanentlyQuestion(len(paths)),
		Detail:        deleteDetail(paths),
		Buttons:       []string{locales.DeletePermanently(), locales.Cancel()},
		DefaultButton: 1,
		CancelButton:  1,
	})
	if err != nil {
		return err
	}
	if answer.Button != 0 {
		return nil
	}
	removed, err := files.RemoveAll(paths, home)
	if removed > 0 {
		notify(locales.DeletePermanently(), locales.DeletedItems(removed))
	}
	return err
}

// deleteDetailLimit is how many names the delete dialog lists before it says how
// many more there are.
const deleteDetailLimit = 5

// deleteDetail returns the text under the delete dialog's question: the names of
// the items, then a note that the deletion cannot be undone.
func deleteDetail(paths []string) string {
	var lines []string
	for i, p := range paths {
		if i == deleteDetailLimit {
			lines = append(lines, locales.DeleteMoreItems(len(paths)-deleteDetailLimit))
			break
		}
		lines = append(lines, files.Name(p))
	}
	lines = append(lines, "", locales.DeletePermanentlyDetail())
	return strings.Join(lines, "\n")
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
	notify(locales.HashCopied(string(algo)), files.Name(path)+"\n"+digest)
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
		return "", "", errors.New(locales.CompareNeedsTwo())
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
	verdict := locales.VerdictDifferent()
	if same {
		verdict = locales.VerdictSame()
	}
	title = locales.HashCompareTitle(string(algo), verdict)
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
	return "", errors.New(locales.UnsupportedHash(name))
}

// requireOne returns the only path. Actions that work on one folder or file use it.
func requireOne(paths []string) (string, error) {
	if len(paths) != 1 {
		return "", errors.New(locales.NeedOneItem())
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

// stemNames returns the name of each path without its extension. For a folder or
// a name that starts with a dot, the name has no extension and stays as it is.
func stemNames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, p := range paths {
		names = append(names, files.BaseName(p))
	}
	return names
}

// extensionNames returns the extension of each path without its dot. A path with
// no extension adds no line, so the clipboard does not gain blank entries.
func extensionNames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, p := range paths {
		if ext := files.Ext(p); ext != "" {
			names = append(names, ext)
		}
	}
	return names
}

// copyText replaces the clipboard with text.
func copyText(text string) error {
	return mygo.Clipboard.Write(transfer.TextData(text))
}

// notifyFailure shows an error in a notification.
func notifyFailure(err error) {
	notify(locales.NotifyFailed(), err.Error())
}

// notify shows a notification. A failed notification is only logged, because the
// action has already run and its result should not be hidden by a notification error.
func notify(title, body string) {
	err := mygo.NewNotification(mygo.NotificationOptions{Title: title, Body: body}).Show()
	if err != nil {
		log.Printf("notification %q failed: %v", title, err)
	}
}
