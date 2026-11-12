import Cocoa
import FinderSync

/// Finder Sync extension. It shows the app's submenu, built from config.json. Each chosen
/// action runs through a script that the app installs; the script starts the app in the
/// background, so no window opens. The extension does no file work itself.
final class FinderSyncExtension: FIFinderSync {
    private let mainScheme = "rightclick"
    /// The script the app installs in this extension's Application Scripts folder. The name
    /// must match finder.ActionScriptName in internal/finder.
    private let actionScriptName = "rightclick-action"
    /// The rightclick:// URL of each action item in the menu that buildMenu built last.
    /// Finder copies menu items and drops representedObject, so an item carries its index here in tag.
    private var actionURLs: [URL] = []

    override init() {
        super.init()
        // Finder shows the menu in every folder.
        FIFinderSyncController.default().directoryURLs = [URL(fileURLWithPath: "/")]
        // The menu must not load icons on the main thread, so they warm up here.
        DispatchQueue.global(qos: .utility).async { ActionIcons.warm() }
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
        let config = AppConfig.load()
        Localized.select(language: config?.language)
        return buildMenu(target: target, config: config, templates: Templates.load())
    }

    func buildMenu(target: MenuTarget, config: AppConfig?, templates: [TemplateEntry]) -> NSMenu? {
        actionURLs = []
        let enabled = MenuActions.enabledIDs(from: config?.menuItems)
        let fileTypes = FileTypes.resolved(from: config?.fileTypes)
        // Older files kept editors and terminals separately; merge them until the
        // app rewrites the file with one program list.
        let legacyPrograms = (config?.editors ?? []) + (config?.terminals ?? [])
        let programs = config?.programs ?? legacyPrograms
        let visibleTemplates = templates.filter { $0.disabled != true }
        let favorites = (config?.favoriteFolders ?? []).filter { $0.enabled }
        let menu = NSMenu(title: "")
        let submenu = NSMenu(title: Localized.text("finder.appName"))
        var nestedFiles: [NSMenuItem] = []
        if let folder = target.newItemFolder, enabled.contains("new-file") {
            let folderPaths = pathParams([folder])
            if enabled.contains("new-folder"), target.allows("new-folder") {
                let folderItem = menuItem(for: "new-folder", target: target, programs: programs)
                if let icon = ActionIcons.image(for: "new-folder") { folderItem.image = icon }
                let folderSetting = config?.menuItems?.first(where: { $0.id == "new-folder" })
                if let label = folderSetting?.label, !label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                    folderItem.title = label
                }
                nestedFiles.append(folderItem)
            }
            nestedFiles += FileTypes.enabled(fileTypes).map {
                    fileTypeItem($0, title: $0.displayLabel, folderPaths: folderPaths)
                }
            if enabled.contains("new-template") {
                for template in visibleTemplates {
                    let label = template.label ?? template.name
                    let item = actionItem(title: label, urlAction: "new-template",
                                          params: [("name", template.name)] + folderPaths)
                    item.image = fileTypeImage(URL(fileURLWithPath: template.name).pathExtension)
                    nestedFiles.append(item)
                }
            }
        }
        for id in enabled {
            let item: NSMenuItem
            if id.contains(":") {
                guard let shortcut = shortcutItem(id, target: target, programs: programs, favorites: favorites) else { continue }
                item = shortcut
            } else {
                guard target.allows(id), id != "new-template", id != "new-folder" else { continue }
                switch id {
                case "new-file":
                    guard !nestedFiles.isEmpty else { continue }
                    item = submenuItem(title: Localized.text("finder.newFile"), children: nestedFiles, emptyTitle: "")
                case "favorite-folders":
                    guard !favorites.isEmpty else { continue }
                    let paths = pathParams(target.items.map(\.url))
                    let children = favorites.map { folder -> NSMenuItem in
                        let child = actionItem(title: folder.label, urlAction: "open-favorite", params: [("favorite", folder.id)] + paths)
                        child.image = ActionIcons.image(for: "folder")
                        return child
                    }
                    item = submenuItem(title: Localized.text("finder.favoriteFolders"), children: children, emptyTitle: "")
                case "copy-to", "move-to":
                    if favorites.isEmpty {
                        item = menuItem(for: id, target: target, programs: programs)
                    } else {
                        let paths = pathParams(target.items.map(\.url))
                        var children = favorites.map { folder -> NSMenuItem in
                            let child = actionItem(title: folder.label, urlAction: id, params: [("favorite", folder.id)] + paths)
                            child.image = ActionIcons.image(for: "folder")
                            return child
                        }
                        let choose = actionItem(title: Localized.text("finder.chooseFolder"), urlAction: id, params: paths)
                        choose.image = ActionIcons.image(for: "choose-folder")
                        children.append(choose)
                        item = submenuItem(title: title(for: id), children: children, emptyTitle: "")
                    }
                default:
                    item = menuItem(for: id, target: target, programs: programs)
                }
            }
            if let icon = ActionIcons.image(for: id) {
                item.image = icon
            }
            let setting = config?.menuItems?.first(where: { $0.id == id })
            if let label = setting?.label, !label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                item.title = label
            }
            let location = setting?.location ??
                (["new-file", "copy-path", "open-terminal"].contains(id) ? FileTypeLocation.main : FileTypeLocation.submenu)
            if location == FileTypeLocation.main { menu.addItem(item) }
            else { submenu.addItem(item) }
        }
        if !submenu.items.isEmpty {
            let parent = NSMenuItem(title: Localized.text("finder.appName"), action: nil, keyEquivalent: "")
            parent.submenu = submenu
            parent.image = ActionIcons.image(for: "menu")
            // The group sits above the rows of the menu, so it is the first item.
            menu.insertItem(parent, at: 0)
        }
        return menu.items.isEmpty ? nil : menu
    }

    private func shortcutItem(_ id: String, target: MenuTarget, programs: [String],
                              favorites: [FavoriteFolderSetting]) -> NSMenuItem? {
        let parts = id.split(separator: ":", maxSplits: 1, omittingEmptySubsequences: false)
        guard parts.count == 2 else { return nil }
        let reference = String(parts[1])
        switch parts[0] {
        case "app":
            // One program list: the app decides whether it sees a file or a
            // folder, and the action routes it.
            guard programs.contains(reference) else { return nil }
            let item = actionItem(title: Localized.text("finder.openWithApp", reference), urlAction: "open-with",
                                  params: [("app", reference)] + pathParams(target.items.map(\.url)))
            item.image = ProgramIcons.image(for: reference) ?? ActionIcons.image(for: "app")
            return item
        case "folder":
            guard let folder = favorites.first(where: { $0.id == reference }) else { return nil }
            let item = actionItem(title: folder.label, urlAction: "open-favorite",
                                  params: [("favorite", folder.id)] + pathParams(target.items.map(\.url)))
            item.image = ActionIcons.image(for: "folder")
            return item
        default:
            return nil
        }
    }

    /// Builds the item for one action ID. Templates and editors get a submenu.
    private func menuItem(for id: String, target: MenuTarget, programs: [String]) -> NSMenuItem {
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
            return submenuItem(title: title(for: id), children: children, emptyTitle: Localized.text("finder.noTemplates"))
        case "open-terminal":
            // Terminal opens folders. A selected file stands for the folder that holds it.
            let folders = target.items.map { $0.isDirectory ? $0.url : $0.url.deletingLastPathComponent() }
            return actionItem(title: title(for: id), urlAction: id, params: pathParams(folders))
        case "hash":
            // One entry: the window hashes with SHA-256 first and offers the other
            // two algorithms, so the menu does not ask for one.
            return actionItem(title: title(for: id), urlAction: id, params: paths)
        default:
            return actionItem(title: title(for: id), urlAction: id, params: paths)
        }
    }

    /// Menu titles shown in Finder, in the language Finder shows.
    private func title(for id: String) -> String {
        switch id {
        case "new-file": return Localized.text("finder.newFile")
        case "new-folder": return Localized.text("finder.newFolder")
        case "new-template": return Localized.text("finder.userTemplate")
        case "add-template": return Localized.text("finder.captureTemplate")
        case "copy-path": return Localized.text("finder.copyPath")
        case "copy-name": return Localized.text("finder.copyName")
        case "copy-basename": return Localized.text("finder.copyBaseName")
        case "copy-to": return Localized.text("finder.copyTo")
        case "move-to": return Localized.text("finder.moveTo")
        case "open-terminal": return Localized.text("finder.openTerminal")
        case "zip": return Localized.text("finder.zip")
        case "unzip": return Localized.text("finder.unzip")
        case "hash": return Localized.text("finder.hash")
        case "compare-hash": return Localized.text("finder.compareHash")
        case "toggle-hidden": return Localized.text("finder.toggleHidden")
        case "delete-permanently": return Localized.text("finder.deletePermanently")
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
            title: Localized.text("finder.newFile"),
            children: fileTypes.map {
                fileTypeItem($0, title: $0.displayLabel, folderPaths: folderPaths)
            },
            emptyTitle: Localized.text("finder.noFileTypes")
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
    /// The URL goes into actionURLs; the item's tag is its index there. A URL that cannot be
    /// built gets tag -1, which menuURL(for:) never matches.
    private func actionItem(title: String, urlAction: String, params: [(String, String)]) -> NSMenuItem {
        let item = NSMenuItem(title: title, action: #selector(runAction(_:)), keyEquivalent: "")
        item.target = self
        if let url = actionURL(urlAction, params) {
            item.tag = actionURLs.count
            actionURLs.append(url)
        } else {
            item.tag = -1
        }
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
        guard let url = menuURL(for: sender) else { return }
        runScript(url)
    }

    /// Runs one action through the script in this extension's Application Scripts folder.
    /// macOS runs that script outside the sandbox. The script starts the app's executable
    /// with --action, so no window or Dock icon appears.
    private func runScript(_ url: URL) {
        do {
            let folder = try FileManager.default.url(
                for: .applicationScriptsDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
            let task = try NSUserUnixTask(url: folder.appendingPathComponent(actionScriptName))
            task.execute(withArguments: [url.absoluteString]) { error in
                _ = task // keeps the task alive until the script finishes
                if let error {
                    NSLog("RightClick action failed: %@", error.localizedDescription)
                }
            }
        } catch {
            NSLog("RightClick could not start an action: %@", error.localizedDescription)
        }
    }

    /// The URL for an item from the menu built last, found by its tag. Nil for any other item.
    func menuURL(for item: NSMenuItem) -> URL? {
        actionURLs.indices.contains(item.tag) ? actionURLs[item.tag] : nil
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
        case "copy-path", "copy-name", "copy-basename":
            // On the blank area these copy the current folder itself, as a selected folder would.
            return true
        case "copy-to", "move-to", "delete-permanently":
            return !isBlankArea
        case "open-terminal":
            return newItemFolder != nil
        case "favorite-folders":
            return true
        case "zip":
            return !isBlankArea && !files.isEmpty
        case "unzip":
            return singleFile?.pathExtension.lowercased() == "zip"
        case "hash":
            return singleFile != nil
        case "add-template":
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
