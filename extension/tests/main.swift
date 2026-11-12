import AppKit
import Foundation

// The strings come from the .lproj folders next to the extension sources.
let stringsRoot = URL(fileURLWithPath: #filePath).deletingLastPathComponent().deletingLastPathComponent()
    .appendingPathComponent("RightClickFinderSync")
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("zh-Hans.lproj"))!

let configJSON = """
{"watchFolders":["/private/tmp"],"menuItems":[{"id":"new-file","enabled":true,"location":"main"}],"fileTypes":[
{"id":"text","label":"Text","extension":"txt","defaultName":"untitled","content":"","location":"submenu","enabled":false},
{"id":"json","label":"JSON","extension":"json","defaultName":"untitled","content":"{}","location":"main","enabled":true}
]}
"""
let config = try JSONDecoder().decode(AppConfig.self, from: Data(configJSON.utf8))
let target = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp", isDirectory: true))], isBlankArea: true)
precondition(config.contains(target), "whitelisted folder was rejected")
let outsideTarget = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/Users/Shared", isDirectory: true))], isBlankArea: true)
precondition(!config.contains(outsideTarget), "folder outside the whitelist was accepted")
let fileInside = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp/document.txt"))], isBlankArea: false)
precondition(config.contains(fileInside), "file inside the whitelist was rejected")
let provider = FinderSyncExtension()
guard let menu = provider.buildMenu(target: target, config: config, templates: []) else { fatalError("missing menu") }
precondition(menu.items.contains { $0.title == "新建 JSON" } == false, "file type leaked to Finder root")
guard let newFile = menu.items.first(where: { $0.title == "新建文件" }) else { fatalError("missing root New File") }
precondition(newFile.submenu?.items.contains { $0.title == "Text" } == false, "disabled type is still visible")
precondition(newFile.submenu?.items.contains { $0.title == "JSON" } == true, "enabled type is missing from the New File submenu")
precondition(newFile.submenu?.items.contains { $0.title == "Word 文档" } == true, "missing Office default")
precondition(newFile.submenu?.items.contains { $0.title == "Go" } == true, "missing code language default")
precondition(newFile.submenu?.items.contains { $0.title == "WPS 文档" } == true, "missing WPS default")
guard let jsonItem = newFile.submenu?.items.first(where: { $0.title == "JSON" }),
      let url = provider.menuURL(for: jsonItem) else { fatalError("missing action URL") }
precondition(url.host == "new-file" && url.query?.contains("type=json") == true, "wrong creation route")
let templates = [TemplateEntry(name: "meeting.md", disabled: nil, location: "main", label: "会议记录")]
guard let templateMenu = provider.buildMenu(target: target, config: config, templates: templates),
      let templateNewFile = templateMenu.items.first(where: { $0.title == "新建文件" }) else {
    fatalError("missing New File with a template")
}
precondition(templateNewFile.submenu?.items.contains { $0.title == "会议记录" } == true,
             "template is missing from the New File submenu")
// Each build renumbers the items, so the template's tag must map to the template URL in the latest menu.
guard let templateItem = templateNewFile.submenu?.items.first(where: { $0.title == "会议记录" }),
      let templateURL = provider.menuURL(for: templateItem) else { fatalError("template item has no URL") }
precondition(templateURL.host == "new-template" && templateURL.query?.contains("name=meeting.md") == true,
             "template item routed to the wrong action")
precondition(templateMenu.items.contains { $0.title == "新建 会议记录" } == false,
             "template leaked to Finder root")
let fileTarget = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp/document.txt"))], isBlankArea: false)
let fileMenu = provider.buildMenu(target: fileTarget, config: config, templates: [])
precondition(fileMenu?.items.contains { $0.title == "在终端打开" || $0.title == "新建文件" } == false, "folder actions visible on a file")
print("Finder menu verified: New File submenu placement, disabled types, Office defaults, templates, URL routing and file context")

// The same menu in English, read from en.lproj.
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
guard let englishMenu = provider.buildMenu(target: target, config: config, templates: []),
      let englishNewFile = englishMenu.items.first(where: { $0.title == "New File" }) else {
    fatalError("missing English New File")
}
precondition(englishNewFile.submenu?.items.contains { $0.title == "Word Document" } == true, "built-in type not in English")
precondition(englishNewFile.submenu?.items.contains { $0.title == "JSON" } == true, "stored label lost in English")

// Both languages have the same keys, and every key has text in each language.
let english = (NSDictionary(contentsOf: stringsRoot.appendingPathComponent("en.lproj/Localizable.strings")) as? [String: String]) ?? [:]
let chinese = (NSDictionary(contentsOf: stringsRoot.appendingPathComponent("zh-Hans.lproj/Localizable.strings")) as? [String: String]) ?? [:]
precondition(!english.isEmpty && Set(english.keys) == Set(chinese.keys), "en and zh-Hans strings have different keys")
for (name, bundlePath) in [("en", "en.lproj"), ("zh-Hans", "zh-Hans.lproj")] {
    Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent(bundlePath))!
    for key in english.keys {
        precondition(Localized.text(key) != key, "no \(name) text for \(key)")
    }
}
print("Finder menu strings verified in English and Chinese")

let shortcutsConfig = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"editors":["Cursor"],"favoriteFolders":[{"id":"archive","path":"/private/tmp","label":"Archive","enabled":true}],"menuItems":[{"id":"app:Cursor","enabled":true,"location":"main","label":"Work Editor"},{"id":"type:json","enabled":true,"location":"main"},{"id":"folder:archive","enabled":true,"location":"main"},{"id":"template:meeting.md","enabled":true,"location":"main"},{"id":"favorite-folders","enabled":true,"location":"submenu"},{"id":"hash","enabled":true,"location":"submenu"},{"id":"copy-name","enabled":true,"location":"submenu"}]}"#.utf8))
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
guard let shortcutsMenu = provider.buildMenu(target: target, config: shortcutsConfig, templates: templates) else { fatalError("missing shortcuts") }
precondition(Array(shortcutsMenu.items.prefix(4)).map(\.title) == ["Work Editor", "New JSON", "Archive", "New 会议记录"], "shortcut order or naming is wrong")
guard let archiveItem = shortcutsMenu.items.first(where: { $0.title == "Archive" }),
      let archiveURL = provider.menuURL(for: archiveItem) else { fatalError("favorite route missing") }
precondition(archiveURL.host == "open-favorite" && archiveURL.query?.contains("favorite=archive") == true, "favorite route must use its saved ID")
guard let shortcutsFileMenu = provider.buildMenu(target: fileTarget, config: shortcutsConfig, templates: templates),
      let appSubmenu = shortcutsFileMenu.items.first(where: { $0.title == "RightClick" })?.submenu else { fatalError("missing file menu") }
let orderedTitles = appSubmenu.items.map(\.title)
precondition(orderedTitles.firstIndex(of: "Compute Hash")! < orderedTitles.firstIndex(of: "Copy Name")!, "menu regrouped a custom order")
guard let copyMenu = appSubmenu.items.first(where: { $0.title == "Copy To…" })?.submenu,
      let copyFavorite = copyMenu.items.first(where: { $0.title == "Archive" }),
      let copyFavoriteURL = provider.menuURL(for: copyFavorite) else { fatalError("copy favorite missing") }
precondition(copyFavoriteURL.host == "copy-to" && copyFavoriteURL.query?.contains("favorite=archive") == true, "copy favorite routes incorrectly")
precondition(copyMenu.items.contains(where: { $0.title == "Choose Another Folder…" }), "folder picker fallback is missing")
precondition(!shortcutsFileMenu.items.contains(where: { $0.title.hasPrefix("New ") }), "creation shortcuts appear on a selected file")
print("Shortcuts, custom ordering, names and favorite destinations verified")

// The destructive action starts switched off and appears only when turned on.
func allTitles(_ menu: NSMenu?) -> [String] {
    (menu?.items ?? []).flatMap { [$0.title] + allTitles($0.submenu) }
}
Localized.bundle = Bundle(url: stringsRoot.appendingPathComponent("en.lproj"))!
precondition(!MenuActions.enabledIDs(from: nil).contains("delete-permanently"), "destructive action is on by default")
let deleteOn = try JSONDecoder().decode(AppConfig.self, from: Data(#"{"menuItems":[{"id":"delete-permanently","enabled":true,"location":"submenu"}]}"#.utf8))
let fileWithDelete = provider.buildMenu(target: fileTarget, config: deleteOn, templates: [])
precondition(allTitles(fileWithDelete).contains("Delete Permanently"), "switched-on action is missing for a file")
let folderWithDelete = provider.buildMenu(target: target, config: deleteOn, templates: [])
precondition(!allTitles(folderWithDelete).contains("Delete Permanently"), "action shown on the blank area of a folder")
print("Permanent delete is off by default and shown for selected items only")

// The language setting chooses the strings; an empty setting follows the system.
let stringsContainer = Bundle(url: stringsRoot)!
Localized.select(language: "en", from: stringsContainer)
precondition(Localized.text("finder.newFile") == "New File", "English setting ignored")
Localized.select(language: "zh-Hans", from: stringsContainer)
precondition(Localized.text("finder.newFile") == "新建文件", "Chinese setting ignored")
print("Finder menu follows the app's language setting")
