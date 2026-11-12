import Cocoa
import FinderSync

/// Finder Sync extension. It shows the app's submenu, built from config.json, and passes
/// each chosen action to the main app as a rightclick:// URL. It does no file work itself.
final class FinderSyncExtension: FIFinderSync {
    private let mainScheme = "rightclick"

    override init() {
        super.init()
        // Finder shows the menu only inside these folders. Read once when the extension starts.
        // With no watch folders configured the set is empty, so Finder shows no menu.
        let folders = AppConfig.load()?.watchFolders ?? []
        FIFinderSyncController.default().directoryURLs = Set(folders.map {
            URL(fileURLWithPath: $0, isDirectory: true)
        })
    }

    override func menu(for menuKind: FIMenuKind) -> NSMenu? {
        let controller = FIFinderSyncController.default()
        let target: MenuTarget
        switch menuKind {
        case .contextualMenuForContainer:
            // Blank area: the current folder is the only target.
            guard let folder = controller.targetedURL() else { return nil }
            target = MenuTarget(items: [MenuTarget.Item(url: folder)], isBlankArea: true)
        case .contextualMenuForItems:
            let urls = controller.selectedItemURLs() ?? []
            guard !urls.isEmpty else { return nil }
            target = MenuTarget(items: urls.map { MenuTarget.Item(url: $0) }, isBlankArea: false)
        default:
            return nil
        }

        // config.json is read on every menu open, so edits in the main app apply without restarting the extension.
        return buildMenu(target: target, config: AppConfig.load(), templates: Templates.load())
    }

    func buildMenu(target: MenuTarget, config: AppConfig?, templates: [TemplateEntry]) -> NSMenu? {
        let enabled = MenuActions.enabledIDs(from: config?.menuItems)
        let fileTypes = FileTypes.resolved(from: config?.fileTypes)
        let editors = config?.editors ?? ["Visual Studio Code", "Cursor"]
        let visibleTemplates = templates.filter { $0.disabled != true }
        let menu = NSMenu(title: "")
        let submenu = NSMenu(title: "右键助手")
        var nestedFiles: [NSMenuItem] = []
        if let folder = target.newItemFolder {
            let folderPaths = pathParams([folder])
            if enabled.contains("new-file") {
                for fileType in FileTypes.visible(fileTypes, at: FileTypeLocation.main) {
                    menu.addItem(fileTypeItem(fileType, title: "新建 \(fileType.label)", folderPaths: folderPaths))
                }
                nestedFiles += FileTypes.visible(fileTypes, at: FileTypeLocation.submenu).map {
                    fileTypeItem($0, title: $0.label, folderPaths: folderPaths)
                }
            }
            if enabled.contains("new-template") {
                for template in visibleTemplates {
                    let label = template.label ?? template.name
                    let item = actionItem(title: template.location == FileTypeLocation.main ? "新建 \(label)" : label,
                                          urlAction: "new-template", params: [("name", template.name)] + folderPaths)
                    item.image = fileTypeImage(URL(fileURLWithPath: template.name).pathExtension)
                    if template.location == FileTypeLocation.main { menu.addItem(item) }
                    else { nestedFiles.append(item) }
                }
            }
        }
        for group in MenuActions.groups {
            var groupItems: [NSMenuItem] = []
            var groupIDs = enabled.filter { group.contains($0) }
            if group.contains("new-file"), !groupIDs.contains("new-file"), !nestedFiles.isEmpty { groupIDs.insert("new-file", at:0) }
            for id in groupIDs where target.allows(id) {
                if id == "new-template" || id == "open-editor" && editors.isEmpty { continue }
                let item: NSMenuItem
                if id == "new-file" {
                    guard !nestedFiles.isEmpty else { continue }
                    item = submenuItem(title: "新建文件", children: nestedFiles, emptyTitle: "")
                    item.image = NSImage(systemSymbolName: "doc.badge.plus", accessibilityDescription: nil)
                } else { item = menuItem(for: id, target: target, editors: editors) }
                let location = config?.menuItems?.first(where: { $0.id == id })?.location ??
                    (["new-file", "copy-path", "open-terminal"].contains(id) ? FileTypeLocation.main : FileTypeLocation.submenu)
                if location == FileTypeLocation.main { menu.addItem(item) }
                else { groupItems.append(item) }
            }
            guard !groupItems.isEmpty else { continue }
            if !submenu.items.isEmpty {
                submenu.addItem(.separator())
            }
            groupItems.forEach { submenu.addItem($0) }
        }
        if !submenu.items.isEmpty {
            let parent = NSMenuItem(title: "右键助手", action: nil, keyEquivalent: "")
            parent.submenu = submenu
            parent.image = NSImage(systemSymbolName: "list.bullet", accessibilityDescription: nil)
            menu.addItem(parent)
        }
        return menu.items.isEmpty ? nil : menu
    }

    /// Builds the item for one action ID. Templates and editors get a submenu.
    private func menuItem(for id: String, target: MenuTarget, editors: [String]) -> NSMenuItem {
        let paths = pathParams(target.items.map(\.url))
        // new-* actions send the folder that receives the new item, not the selected items.
        let folderPaths = pathParams(target.newItemFolder.map { [$0] } ?? [])
        switch id {
        case "new-folder":
            return actionItem(title: title(for: id), urlAction: id, params: folderPaths)
        case "new-template":
            let children = Templates.load().map { entry in
                actionItem(title: entry.name, urlAction: id, params: [("name", entry.name)] + folderPaths)
            }
            return submenuItem(title: title(for: id), children: children, emptyTitle: "无模板")
        case "open-terminal":
            // Terminal opens folders. A selected file stands for the folder that holds it.
            let folders = target.items.map { $0.isDirectory ? $0.url : $0.url.deletingLastPathComponent() }
            return actionItem(title: title(for: id), urlAction: id, params: pathParams(folders))
        case "open-editor":
            let app = editors.first ?? "Visual Studio Code"
            return actionItem(title: "用 \(app) 打开", urlAction: id,
                              params: [("app", app)] + paths)
        case "open-with":
            let children = editors.map { app in
                actionItem(title: app, urlAction: id, params: [("app", app)] + paths)
            }
            return submenuItem(title: title(for: id), children: children, emptyTitle: "无编辑器")
        default:
            return actionItem(title: title(for: id), urlAction: id, params: paths)
        }
    }

    /// Menu titles shown in Finder. They stay in Chinese.
    private func title(for id: String) -> String {
        switch id {
        case "new-file": return "新建文件"
        case "new-folder": return "文件夹"
        case "new-template": return "用户模板"
        case "copy-path": return "拷贝路径"
        case "copy-name": return "拷贝文件名"
        case "copy-to": return "复制到…"
        case "move-to": return "移动到…"
        case "open-terminal": return "在终端打开"
        case "open-editor": return "用默认编辑器打开"
        case "open-with": return "用…打开"
        case "zip": return "压缩为 ZIP"
        case "unzip": return "解压到此处"
        case "hash": return "计算哈希"
        case "compare-hash": return "比较两个文件的哈希"
        case "toggle-hidden": return "显示/隐藏隐藏文件（⌘⇧.）"
        default: return id
        }
    }

    /// A direct file type entry, used when the user pins it to the main menu.
    private func fileTypeItem(_ fileType: FileTypeSetting, title: String,
                              folderPaths: [(String, String)]) -> NSMenuItem {
        let item = actionItem(title: title, urlAction: "new-file",
                              params: [("type", fileType.id)] + folderPaths)
        item.image = fileTypeImage(fileType.ext)
        return item
    }

    /// The default collection of file types shown under New File.
    private func newFileSubmenu(_ fileTypes: [FileTypeSetting],
                                folderPaths: [(String, String)]) -> NSMenuItem {
        let item = submenuItem(
            title: "新建文件",
            children: fileTypes.map {
                fileTypeItem($0, title: $0.label, folderPaths: folderPaths)
            },
            emptyTitle: "无文件类型"
        )
        item.image = fileTypeImage("new-file")
        return item
    }

    /// Draws a small document badge with the type's shorthand and color. A small
    /// bitmap keeps Finder's menu crisp and needs no external image assets.
    private func fileTypeImage(_ id: String) -> NSImage? {
        guard let url = Bundle(for: FinderSyncExtension.self).url(forResource: id, withExtension: "svg", subdirectory: "FileIcons"),
              let image = NSImage(contentsOf: url) else { return NSImage(systemSymbolName: "doc", accessibilityDescription: nil) }
        image.size = NSSize(width: 16, height: 16)
        image.isTemplate = ["txt", "md", "rtf"].contains(id)
        return image
    }

    /// A leaf item. Clicking it opens rightclick://<urlAction>?<params> in the main app.
    private func actionItem(title: String, urlAction: String, params: [(String, String)]) -> NSMenuItem {
        let item = NSMenuItem(title: title, action: #selector(runAction(_:)), keyEquivalent: "")
        item.target = self
        item.representedObject = actionURL(urlAction, params)
        return item
    }

    /// A parent item with a submenu. An empty list gets one disabled placeholder.
    private func submenuItem(title: String, children: [NSMenuItem], emptyTitle: String) -> NSMenuItem {
        let submenu = NSMenu(title: title)
        if children.isEmpty {
            submenu.addItem(NSMenuItem(title: emptyTitle, action: nil, keyEquivalent: ""))
        } else {
            children.forEach { submenu.addItem($0) }
        }
        let item = NSMenuItem(title: title, action: nil, keyEquivalent: "")
        item.submenu = submenu
        return item
    }

    @objc private func runAction(_ sender: NSMenuItem) {
        guard let url = sender.representedObject as? URL else { return }
        NSWorkspace.shared.open(url)
    }

    /// Builds rightclick://<urlAction>?<name=value>&... for the main app.
    /// Values are encoded with the URL query set minus + & ; =. The main app's parser reads
    /// an unencoded + as a space and drops a query pair that contains an unencoded ;,
    /// so those characters are encoded here.
    private func actionURL(_ urlAction: String, _ params: [(String, String)]) -> URL? {
        var allowed = CharacterSet.urlQueryAllowed
        allowed.remove(charactersIn: "+&;=")
        var pairs: [String] = []
        for (name, value) in params {
            guard let encoded = value.addingPercentEncoding(withAllowedCharacters: allowed) else { return nil }
            pairs.append("\(name)=\(encoded)")
        }
        return URL(string: "\(mainScheme)://\(urlAction)?\(pairs.joined(separator: "&"))")
    }

    /// One `path` parameter per URL.
    private func pathParams(_ urls: [URL]) -> [(String, String)] {
        urls.map { ("path", $0.path) }
    }
}

/// What the menu acts on: the blank area of a folder, or the selected items.
struct MenuTarget {
    struct Item {
        let url: URL
        let isDirectory: Bool

        init(url: URL) {
            self.url = url
            self.isDirectory = (try? url.resourceValues(forKeys: [.isDirectoryKey]).isDirectory) ?? false
        }
    }

    let items: [Item]
    /// True for the blank area, where `items` holds only the current folder.
    let isBlankArea: Bool

    var files: [Item] {
        items.filter { !$0.isDirectory }
    }

    /// The folder that receives new items: the blank area's folder, or the one selected folder.
    var newItemFolder: URL? {
        if isBlankArea {
            return items.first?.url
        }
        guard items.count == 1, items[0].isDirectory else { return nil }
        return items[0].url
    }

    /// The URL of the only selected item, when exactly one file is selected.
    var singleFile: URL? {
        guard !isBlankArea, items.count == 1, !items[0].isDirectory else { return nil }
        return items[0].url
    }

    /// Whether an action is shown for this target. The rules come from the product spec.
    func allows(_ id: String) -> Bool {
        switch id {
        case "new-file", "new-folder", "new-template":
            return newItemFolder != nil
        case "copy-path", "copy-name", "copy-to", "move-to":
            return !isBlankArea
        case "open-terminal":
            return newItemFolder != nil
        case "open-editor", "open-with":
            return true
        case "zip":
            return !isBlankArea && !files.isEmpty
        case "unzip":
            return singleFile?.pathExtension.lowercased() == "zip"
        case "hash":
            return singleFile != nil
        case "compare-hash":
            return !isBlankArea && items.count == 2 && files.count == 2
        case "toggle-hidden":
            // A Finder-wide setting, offered wherever a new item would go.
            return newItemFolder != nil
        default:
            return false
        }
    }
}
