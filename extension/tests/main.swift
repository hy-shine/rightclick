import AppKit
import Foundation

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
precondition(menu.items.contains { $0.title == "新建 JSON" }, "pinned file type is not at Finder root")
guard let newFile = menu.items.first(where: { $0.title == "新建文件" }) else { fatalError("missing root New File") }
precondition(newFile.submenu?.items.contains { $0.title == "Text" } == false, "disabled type is still visible")
precondition(newFile.submenu?.items.contains { $0.title == "JSON" } == false, "pinned type duplicated inside submenu")
precondition(newFile.submenu?.items.contains { $0.title == "Word 文档" } == true, "missing Office default")
precondition(newFile.submenu?.items.contains { $0.title == "Go" } == true, "missing code language default")
precondition(newFile.submenu?.items.contains { $0.title == "WPS 文档" } == true, "missing WPS default")
guard let jsonItem = menu.items.first(where: { $0.title == "新建 JSON" }), let url = jsonItem.representedObject as? URL else { fatalError("missing action URL") }
precondition(url.host == "new-file" && url.query?.contains("type=json") == true, "wrong creation route")
let fileTarget = MenuTarget(items: [MenuTarget.Item(url: URL(fileURLWithPath: "/private/tmp/document.txt"))], isBlankArea: false)
let fileMenu = provider.buildMenu(target: fileTarget, config: config, templates: [])
precondition(fileMenu?.items.contains { $0.title == "在终端打开" || $0.title == "新建文件" } == false, "folder actions visible on a file")
print("Finder menu verified: root placement, disabled types, no duplication, Office defaults, URL routing and file context")
