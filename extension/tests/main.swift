import AppKit
import Foundation

// The strings come from the .lproj folders next to the extension sources.
let stringsRoot = URL(fileURLWithPath: #filePath).deletingLastPathComponent().deletingLastPathComponent()
    .appendingPathComponent("RightClickFinderSync")
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!

let configJSON = """
{"menuItems":[{"id":"new-file","enabled":true,"location":"main"}],"fileTypes":[
{"id":"text","label":"Text","extension":"txt","defaultName":"untitled","content":"","location":"submenu","enabled":false},
{"id":"json","label":"JSON","extension":"json","defaultName":"untitled","content":"{}","location":"main","enabled":true}
]}
"""
let config = try JSONDecoder().decode(AppConfig.self, from: Data(configJSON.utf8))
let target = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp", isDirectory: true))], isBlankArea: true)
let provider = FinderSyncExtension()
guard let menu = provider.buildMenu(target: target, config: config, templates: []) else { fatalError("missing menu") }
// The menu shows in every folder, so a target outside any folder still gets one.
let outsideTarget = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/Users/Shared", isDirectory: true))], isBlankArea: true)
precondition(provider.buildMenu(target: outsideTarget, config: config, templates: []) != nil, "menu is missing outside a folder list")
precondition(menu.items.contains { $0.title == "New JSON" } == false, "file type leaked to Finder root")
guard let newFile = menu.items.first(where: { $0.title == "New File" }) else { fatalError("missing root New File") }
precondition(newFile.submenu?.items.contains { $0.title == "Text" } == false, "disabled type is still visible")
precondition(newFile.submenu?.items.contains { $0.title == "JSON" } == true, "enabled type is missing from the New File submenu")
precondition(newFile.submenu?.items.contains { $0.title == "Word Document" } == true, "missing Office default")
precondition(newFile.submenu?.items.contains { $0.title == "Go" } == true, "missing code language default")
precondition(newFile.submenu?.items.contains { $0.title == "WPS Writer" } == true, "missing WPS default")
guard let jsonItem = newFile.submenu?.items.first(where: { $0.title == "JSON" }),
      let url = provider.menuURL(for: jsonItem) else { fatalError("missing action URL") }
precondition(url.host == "new-file" && url.query?.contains("type=json") == true, "wrong creation route")
let templates = [TemplateEntry(name: "meeting.md", disabled: nil, location: "main", label: "Meeting Notes")]
guard let templateMenu = provider.buildMenu(target: target, config: config, templates: templates),
      let templateNewFile = templateMenu.items.first(where: { $0.title == "New File" }) else {
    fatalError("missing New File with a template")
}
precondition(templateNewFile.submenu?.items.contains { $0.title == "Meeting Notes" } == true,
             "template is missing from the New File submenu")
// Each build renumbers the items, so the template's tag must map to the template URL in the latest menu.
guard let templateItem = templateNewFile.submenu?.items.first(where: { $0.title == "Meeting Notes" }),
      let templateURL = provider.menuURL(for: templateItem) else { fatalError("template item has no URL") }
precondition(templateURL.host == "new-template" && templateURL.query?.contains("name=meeting.md") == true,
             "template item routed to the wrong action")
precondition(templateMenu.items.contains { $0.title == "New Meeting Notes" } == false,
             "template leaked to Finder root")
let fileTarget = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp/document.txt"))], isBlankArea: false)
let fileMenu = provider.buildMenu(target: fileTarget, config: config, templates: [])
precondition(fileMenu?.items.contains { $0.title == "Open with Terminal" || $0.title == "New File" } == false, "folder actions visible on a file")

// The name without its extension is listed for a file, and still works for a name with no extension.
precondition(allTitles(fileMenu).contains("Copy Name (Without Extension)"), "copy name without extension is missing")
let namedWithoutExtension = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp/README"))], isBlankArea: false)
precondition(namedWithoutExtension.allows("copy-basename"), "copy name without extension is hidden for a name with no extension")

// A saved entry for the removed Open With submenu must not bring it back.
let staleConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"open-with","enabled":true,"location":"submenu"}]}"#.utf8))
guard let staleMenu = provider.buildMenu(target: fileTarget, config: staleConfig, templates: []) else {
    fatalError("missing menu for a stale config")
}
precondition(!allTitles(staleMenu).contains("Open With…"), "a removed action still builds a menu row")

// A saved entry for the removed Copy Extension must not bring it back.
let staleExtensionConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"copy-extension","enabled":true,"location":"submenu"}]}"#.utf8))
guard let staleExtensionMenu = provider.buildMenu(target: fileTarget, config: staleExtensionConfig, templates: []) else {
    fatalError("missing menu for a stale config")
}
precondition(!allTitles(staleExtensionMenu).contains("Copy Extension") && !allTitles(staleExtensionMenu).contains("copy-extension"), "a removed action still builds a menu row")

// Action icons come from the Lucide set. Warm them from the source folder and
// check that every action and every menu row carries one.
ActionIcons.warm(from: stringsRoot.deletingLastPathComponent().deletingLastPathComponent()
    .appendingPathComponent("internal/actionicons/icons"))
for id in MenuActions.defaultOrder {
    precondition(ActionIcons.image(for: id) != nil, "no icon for action \(id)")
}
func menuItems(_ menu: NSMenu?) -> [NSMenuItem] {
    (menu?.items ?? []).flatMap { [$0] + menuItems($0.submenu) }
}
guard let iconMenu = provider.buildMenu(target: fileTarget, config: config, templates: []) else {
    fatalError("missing menu with icons")
}
for item in menuItems(iconMenu) where !item.isSeparatorItem {
    precondition(item.image != nil, "menu item \(item.title) has no icon")
}
print("Finder menu verified: New File submenu placement, disabled types, Office defaults, templates, URL routing, copy variants and file context")

// The same menu in English, read from en.lproj.
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
guard let englishMenu = provider.buildMenu(target: target, config: config, templates: []),
      let englishNewFile = englishMenu.items.first(where: { $0.title == "New File" }) else {
    fatalError("missing English New File")
}
precondition(englishNewFile.submenu?.items.contains { $0.title == "Word Document" } == true, "built-in type not in English")
precondition(englishNewFile.submenu?.items.contains { $0.title == "JSON" } == true, "stored label lost in English")

// Every language has the same keys, and every key has text in each language.
let english = (NSDictionary(contentsOf: stringsRoot.appendingPathComponent("en.lproj/Localizable.strings")) as? [String: String]) ?? [:]
let chinese = (NSDictionary(contentsOf: stringsRoot.appendingPathComponent("zh-Hans.lproj/Localizable.strings")) as? [String: String]) ?? [:]
let traditional = (NSDictionary(contentsOf: stringsRoot.appendingPathComponent("zh-Hant.lproj/Localizable.strings")) as? [String: String]) ?? [:]
precondition(!english.isEmpty && Set(english.keys) == Set(chinese.keys) && Set(english.keys) == Set(traditional.keys), "en, zh-Hans and zh-Hant strings have different keys")
for (name, bundlePath) in [("en", "en.lproj"), ("zh-Hans", "zh-Hans.lproj"), ("zh-Hant", "zh-Hant.lproj")] {
    Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent(bundlePath))!
    for key in english.keys {
        precondition(Localized.text(key) != key, "no \(name) text for \(key)")
    }
}
print("Finder menu strings verified in English and Chinese")

let shortcutsConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"editors":["Cursor"],"favoriteFolders":[{"id":"archive","path":"/private/tmp","label":"Archive","enabled":true}],"menuItems":[{"id":"app:Cursor","enabled":true,"location":"main","label":"Work Editor"},{"id":"type:json","enabled":true,"location":"main"},{"id":"folder:archive","enabled":true,"location":"main"},{"id":"new-file","enabled":true,"location":"main"},{"id":"new-folder","enabled":true,"location":"submenu"},{"id":"template:meeting.md","enabled":true,"location":"main"},{"id":"favorite-folders","enabled":true,"location":"submenu"},{"id":"hash","enabled":true,"location":"submenu"},{"id":"copy-name","enabled":true,"location":"submenu"}]}"#.utf8))
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
guard let shortcutsMenu = provider.buildMenu(target: target, config: shortcutsConfig, templates: templates) else { fatalError("missing shortcuts") }
precondition(Array(shortcutsMenu.items.prefix(3)).map(\.title) == ["RightClick", "Work Editor", "Archive"], "shortcut order or naming is wrong")
precondition(!shortcutsMenu.items.contains(where: { $0.title.hasPrefix("New ") && $0.submenu == nil }), "a file type or template reached the main menu: \(shortcutsMenu.items.map(\.title))")
precondition(shortcutsMenu.items.first(where: { $0.title == "New File" })?.submenu?.items.contains(where: { $0.title == "New Folder" }) == true, "New Folder is not inside New File")
guard let archiveItem = shortcutsMenu.items.first(where: { $0.title == "Archive" }),
      let archiveURL = provider.menuURL(for: archiveItem) else { fatalError("favorite route missing") }
precondition(archiveURL.host == "open-favorite" && archiveURL.query?.contains("favorite=archive") == true, "favorite route must use its saved ID")
guard let shortcutsFileMenu = provider.buildMenu(target: fileTarget, config: shortcutsConfig, templates: templates),
      let appSubmenu = shortcutsFileMenu.items.first(where: { $0.title == "RightClick" })?.submenu else { fatalError("missing file menu") }
// The RightClick group sits above the rows of the menu, so it is the first item.
precondition(shortcutsFileMenu.items.first?.title == "RightClick", "RightClick group is not the first Finder menu item")
let orderedTitles = appSubmenu.items.map(\.title)
precondition(orderedTitles.firstIndex(of: "Compute Hash")! < orderedTitles.firstIndex(of: "Copy Name")!, "menu regrouped a custom order")
guard let copyMenu = appSubmenu.items.first(where: { $0.title == "Copy To…" })?.submenu,
      let copyFavorite = copyMenu.items.first(where: { $0.title == "Archive" }),
      let copyFavoriteURL = provider.menuURL(for: copyFavorite) else { fatalError("copy favorite missing") }
precondition(copyFavoriteURL.host == "copy-to" && copyFavoriteURL.query?.contains("favorite=archive") == true, "copy favorite routes incorrectly")
precondition(copyMenu.items.contains(where: { $0.title == "Choose Another Folder…" }), "folder picker fallback is missing")
precondition(!copyMenu.items.contains(where: { $0.isSeparatorItem }), "a separator splits the folder destinations")
precondition(copyMenu.items.allSatisfy { $0.image != nil }, "a folder destination has no icon")
precondition(!shortcutsFileMenu.items.contains(where: { $0.title.hasPrefix("New ") }), "creation shortcuts appear on a selected file")
print("Shortcuts, custom ordering, names and favorite destinations verified")


let captureConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"add-template","enabled":true,"location":"main"}]}"#.utf8))
guard let captureMenu = provider.buildMenu(target: fileTarget, config: captureConfig, templates: []),
      let captureItem = captureMenu.items.first(where: { $0.title == "Add as Template…" }),
      let captureURL = provider.menuURL(for: captureItem) else { fatalError("capture template action missing") }
precondition(captureURL.host == "add-template" && captureURL.query?.contains("document.txt") == true, "capture route lost source file")
precondition(!target.allows("add-template"), "capture action appears on a folder background")
let multipleFiles = MenuTarget(items: fileTarget.items + fileTarget.items, isBlankArea: false)
precondition(!multipleFiles.allows("add-template"), "capture accepts multiple files")
let captureOff = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"add-template","enabled":false,"location":"main"}]}"#.utf8))
precondition(!MenuActions.enabledIDs(from: captureOff.menuItems).contains("add-template"), "disabled capture action still appears")
let customTypeConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"fileTypes":[{"id":"custom-config","label":"Project Config","extension":"yaml","defaultName":"config","content":"port: 8080","location":"submenu","enabled":true}]}"#.utf8))
guard let customTypeMenu = provider.buildMenu(target: target, config: customTypeConfig, templates: []),
      let customTypeItem = customTypeMenu.items.first(where: { $0.title == "New File" })?.submenu?.items.first(where: { $0.title == "Project Config" }),
      let customTypeURL = provider.menuURL(for: customTypeItem) else { fatalError("custom type absent from Finder") }
precondition(customTypeURL.host == "new-file" && customTypeURL.query?.contains("type=custom-config") == true, "custom type route is wrong")
print("Template capture context and custom file type routing verified")

// The destructive action starts switched off and appears only when turned on.
func allTitles(_ menu: NSMenu?) -> [String] {
    (menu?.items ?? []).flatMap { [$0.title] + allTitles($0.submenu) }
}

/// Returns the first item with title, at any depth.
func firstItem(_ menu: NSMenu?, title: String) -> NSMenuItem? {
    for item in menu?.items ?? [] {
        if item.title == title { return item }
        if let found = firstItem(item.submenu, title: title) { return found }
    }
    return nil
}
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
precondition(!MenuActions.enabledIDs(from: nil).contains("delete-permanently"), "destructive action is on by default")
precondition(!MenuActions.enabledIDs(from: nil).contains("open-terminal"), "open with Terminal is on by default")
let deleteOn = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"delete-permanently","enabled":true,"location":"submenu"}]}"#.utf8))
let fileWithDelete = provider.buildMenu(target: fileTarget, config: deleteOn, templates: [])
precondition(allTitles(fileWithDelete).contains("Delete Permanently"), "switched-on action is missing for a file")
let folderWithDelete = provider.buildMenu(target: target, config: deleteOn, templates: [])
precondition(!allTitles(folderWithDelete).contains("Delete Permanently"), "action shown on the blank area of a folder")
print("Permanent delete is off by default and shown for selected items only")

// Copy Path and Copy Name also work on the blank area, where they copy the folder itself.
// Copy To and Move To need selected items, so they stay off the blank area.
let blankCopyConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"copy-path","enabled":true,"location":"main"},{"id":"copy-name","enabled":true,"location":"submenu"}]}"#.utf8))
precondition(target.allows("copy-path") && target.allows("copy-name") && target.allows("copy-basename"), "copy actions are hidden on the blank area")
precondition(!target.allows("copy-to") && !target.allows("move-to"), "copy-to or move-to shows on the blank area")
guard let blankCopyMenu = provider.buildMenu(target: target, config: blankCopyConfig, templates: []) else {
    fatalError("missing menu for the blank area")
}
guard let blankCopyPath = firstItem(blankCopyMenu, title: "Copy Path"),
      let blankCopyPathURL = provider.menuURL(for: blankCopyPath) else { fatalError("blank Copy Path is missing or has no URL") }
let sentPath = URLComponents(url: blankCopyPathURL, resolvingAgainstBaseURL: false)?.queryItems?.first { $0.name == "path" }?.value
precondition(blankCopyPathURL.host == "copy-path" && sentPath == "/private/tmp", "blank Copy Path does not send the folder: \(blankCopyPathURL)")
precondition(firstItem(blankCopyMenu, title: "Copy Name") != nil, "blank Copy Name is missing")
print("Copy Path and Copy Name are offered on the blank area for the folder itself")

// The language setting chooses the strings; an empty setting follows the system.
let stringsContainer = Bundle(url: stringsRoot)!
Localized.select(language: "en", from: stringsContainer)
precondition(Localized.text("finder.newFile") == "New File", "English setting ignored")
Localized.select(language: "zh-Hans", from: stringsContainer)
precondition(Localized.text("finder.newFile") == "新建文件", "Chinese setting ignored")
Localized.select(language: "zh-Hant", from: stringsContainer)
precondition(Localized.text("finder.newFile") == "新增檔案", "Traditional Chinese setting ignored")
print("Finder menu follows the app's language setting")

Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
// Children follow New File: with New File switched off, neither it nor New Folder appears anywhere.
let newFileOffConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"new-file","enabled":false,"location":"main"},{"id":"new-folder","enabled":true,"location":"submenu"}]}"#.utf8))
func deepTitles(_ menu: NSMenu?) -> [String] {
    (menu?.items ?? []).flatMap { [$0.title] + deepTitles($0.submenu) }
}
let offTitles = deepTitles(provider.buildMenu(target: target, config: newFileOffConfig, templates: []))
precondition(!offTitles.contains("New Folder") && !offTitles.contains("New File"), "children stayed visible with New File off: \(offTitles)")

// Compute Hash is one entry that sends no algorithm: the window picks SHA-256 first
// and offers the other two itself.
func findItem(_ menu: NSMenu?, titled title: String) -> NSMenuItem? {
    for item in menu?.items ?? [] {
        if item.title == title { return item }
        if let found = findItem(item.submenu, titled: title) { return found }
    }
    return nil
}
let hashConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"hash","enabled":true,"location":"submenu"}]}"#.utf8))
let hashMenu = provider.buildMenu(target: fileTarget, config: hashConfig, templates: [])
guard let hashItem = findItem(hashMenu, titled: "Compute Hash") else { fatalError("missing Compute Hash") }
precondition(hashItem.submenu == nil, "Compute Hash still opens a submenu")
precondition(hashItem.image != nil, "Compute Hash has no icon")
guard let hashURL = provider.menuURL(for: hashItem) else { fatalError("Compute Hash has no route") }
precondition(hashURL.host == "hash", "Compute Hash routes to \(hashURL.host ?? "nothing")")
precondition(hashURL.query?.contains("algo=") != true, "Compute Hash still sends an algorithm")
print("Compute Hash is one entry and sends no algorithm")

precondition(!MenuActions.enabledIDs(from: nil).contains("copy-terminal-path"), "terminal path is on by default")
let terminalPathConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"copy-terminal-path","enabled":true,"location":"submenu"}]}"#.utf8))
for selection in [target, fileTarget, multipleFiles] {
    guard let terminalMenu = provider.buildMenu(target: selection, config: terminalPathConfig, templates: []),
          let terminalItem = firstItem(terminalMenu, title: "Copy as Terminal Path"),
          let terminalURL = provider.menuURL(for: terminalItem) else { fatalError("terminal path action missing") }
    let paths = URLComponents(url: terminalURL, resolvingAgainstBaseURL: false)?.queryItems?.filter { $0.name == "path" }.compactMap(\.value)
    precondition(terminalURL.host == "copy-terminal-path" && paths == selection.items.map { $0.url.path }, "terminal path lost selection")
    precondition(terminalItem.image != nil, "terminal path has no icon")
}
print("Terminal path default, contexts, icon, and selection routing verified")

precondition(!MenuActions.enabledIDs(from: nil).contains("copy-relative-path"), "relative path is on by default")
let relativePathConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"copy-relative-path","enabled":true,"location":"submenu"}]}"#.utf8))
for selection in [target, fileTarget, multipleFiles] {
    guard let relativeMenu = provider.buildMenu(target: selection, config: relativePathConfig, templates: []),
          let relativeItem = firstItem(relativeMenu, title: "Copy Relative Path…"),
          let relativeURL = provider.menuURL(for: relativeItem) else { fatalError("relative path action missing") }
    let paths = URLComponents(url: relativeURL, resolvingAgainstBaseURL: false)?.queryItems?.filter { $0.name == "path" }.compactMap(\.value)
    precondition(relativeURL.host == "copy-relative-path" && paths == selection.items.map { $0.url.path }, "relative path lost selection")
    precondition(relativeItem.image != nil, "relative path has no icon")
}
print("Relative path default, contexts, icon, and selection routing verified")

precondition(!MenuActions.enabledIDs(from: nil).contains("copy-markdown-link"), "Markdown link is on by default")
let markdownLinkConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"copy-markdown-link","enabled":true,"location":"submenu"}]}"#.utf8))
for selection in [target, fileTarget, multipleFiles] {
    guard let markdownMenu = provider.buildMenu(target: selection, config: markdownLinkConfig, templates: []),
          let markdownItem = firstItem(markdownMenu, title: "Copy Markdown File Link…"),
          let markdownURL = provider.menuURL(for: markdownItem) else { fatalError("Markdown link action missing") }
    let paths = URLComponents(url: markdownURL, resolvingAgainstBaseURL: false)?.queryItems?.filter { $0.name == "path" }.compactMap(\.value)
    precondition(markdownURL.host == "copy-markdown-link" && paths == selection.items.map { $0.url.path }, "Markdown link lost selection")
    precondition(markdownItem.image != nil, "Markdown link has no icon")
}
print("Markdown link default, contexts, icon, and selection routing verified")

let selectedFolder = MenuTarget(items: target.items, isBlankArea: false)
precondition(selectedFolder.allows("zip"), "Compress to ZIP is missing for a selected folder")
let selectedFolders = MenuTarget(items: target.items + outsideTarget.items, isBlankArea: false)
let mixedSelection = MenuTarget(items: target.items + fileTarget.items, isBlankArea: false)
let folderZipConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"zip","enabled":true,"location":"submenu"}]}"#.utf8))
for selection in [selectedFolder, selectedFolders, mixedSelection] {
    guard let folderZipMenu = provider.buildMenu(target: selection, config: folderZipConfig, templates: []),
          let folderZipItem = firstItem(folderZipMenu, title: "Compress to ZIP"),
          let folderZipURL = provider.menuURL(for: folderZipItem) else { fatalError("folder ZIP action missing") }
    let paths = URLComponents(url: folderZipURL, resolvingAgainstBaseURL: false)?.queryItems?.filter { $0.name == "path" }.compactMap(\.value)
    precondition(folderZipURL.host == "zip" && paths == selection.items.map { $0.url.path }, "folder ZIP lost selection")
}
precondition(!target.allows("zip"), "ZIP is visible on a blank area")
print("Folder ZIP selections and action parameters verified")

let zipFileTarget = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp/bundle.ZIP"))], isBlankArea: false)
precondition(zipFileTarget.allows("unzip-to-folder"), "named-folder extraction is missing for ZIP")
precondition(!target.allows("unzip-to-folder") && !fileTarget.allows("unzip-to-folder") && !multipleFiles.allows("unzip-to-folder"), "named-folder extraction has the wrong contexts")
guard let extractionMenu = provider.buildMenu(target: zipFileTarget, config: nil, templates: []),
      let extractionItem = firstItem(extractionMenu, title: "Extract to Folder"),
      let extractionURL = provider.menuURL(for: extractionItem) else { fatalError("named-folder extraction action missing") }
precondition(extractionItem.image != nil && extractionURL.host == "unzip-to-folder", "named-folder extraction icon or route is wrong")
precondition(URLComponents(url: extractionURL, resolvingAgainstBaseURL: false)?.queryItems?.first { $0.name == "path" }?.value == "/private/tmp/bundle.ZIP", "named-folder extraction lost the archive path")
precondition(firstItem(extractionMenu, title: "Extract Here") != nil, "Extract Here was removed")
print("Both extraction menus, archive context, icons, and routes verified")

// An app's item shows that app's own icon. An app that is not installed has none.
let programConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"programs":["Finder"],"menuItems":[{"id":"app:Finder","enabled":true,"location":"main"}]}"#.utf8))
guard let programMenu = provider.buildMenu(target: fileTarget, config: programConfig, templates: []) else {
    fatalError("missing menu for the program config")
}
let finderIcon = ProgramIcons.image(for: "Finder")
precondition(finderIcon != nil, "Finder's own icon was not found")
precondition(programMenu.items.contains { $0.image === finderIcon }, "the app item does not show the app's own icon")
precondition(ProgramIcons.image(for: "No Such Program Here") == nil, "a program that is not installed got an icon")
