import AppKit

/// Icons for the built-in actions, taken from the Lucide icon set (ISC license,
/// see THIRD_PARTY_LICENSES.txt at the repository root). The files are
/// loaded once, off the main thread: Finder builds the menu on its main thread
/// every time it opens, and a cold load of every icon costs about 13 ms, which
/// the menu itself must not pay.
enum ActionIcons {
    /// The icon file of each action ID or shortcut kind, without the extension.
    /// Keep it in step with MenuActions.defaultOrder.
    static let names: [String: String] = [
        "new-file": "new-file",
        "new-folder": "new-folder",
        "new-template": "new-template",
        "add-template": "add-template",
        "copy-path": "copy-path",
        "copy-terminal-path": "open-terminal",
        "copy-relative-path": "copy-path",
        "copy-markdown-link": "copy-path",
        "copy-name": "copy-name",
        "copy-basename": "copy-basename",
        "copy-to": "copy-to",
        "move-to": "move-to",
        "open-terminal": "open-terminal",
        "zip": "zip",
        "unzip": "unzip",
        "unzip-to-folder": "unzip",
        "hash": "hash",
        "compare-hash": "compare-hash",
        "toggle-hidden": "toggle-hidden",
        "delete-permanently": "delete-permanently",
        "favorite-folders": "favorite-folders",
        // Shortcut kinds. "type" and "template" use the file type icon instead.
        "app": "app",
        "folder": "folder",
        // The menu of the app itself: a pointer clicking, the right-click idea.
        "menu": "menu",
        // Not an action of its own: the "Choose Another Folder…" row reuses the
        // New Folder artwork, so every row in that submenu keeps one icon column.
        "choose-folder": "new-folder",
    ]

    private static var cache: [String: NSImage] = [:]
    private static let lock = NSLock()

    /// Loads every icon into the cache. The extension calls it from init on a
    /// background queue; tests pass the source folder, which is not bundled.
    static func warm(from root: URL? = nil) {
        let folder = root ?? Bundle(for: FinderSyncExtension.self).resourceURL?
            .appendingPathComponent("ActionIcons")
        guard let folder else { return }
        var loaded: [String: NSImage] = [:]
        for name in Set(names.values) {
            guard let image = NSImage(contentsOf: folder.appendingPathComponent("\(name).svg")) else { continue }
            image.size = NSSize(width: 16, height: 16)
            image.isTemplate = true
            loaded[name] = image
        }
        lock.lock()
        cache = loaded
        lock.unlock()
    }

    /// The icon of an action, or nil while the cache is still empty.
    static func image(for id: String) -> NSImage? {
        guard let name = names[id] else { return nil }
        lock.lock()
        defer { lock.unlock() }
        return cache[name]
    }
}

/// The icons of the apps that the "Open With" items name. Each app's own artwork
/// is used, not a generic glyph, so the menu matches the app list in settings.
/// Looked up once per app, found or not; an app that is not installed keeps the
/// generic icon.
enum ProgramIcons {
    /// nil records an app that was looked up and not found.
    private static var cache: [String: NSImage?] = [:]
    private static let lock = NSLock()

    static func image(for program: String) -> NSImage? {
        lock.lock()
        defer { lock.unlock() }
        if let known = cache[program] { return known }
        var found: NSImage?
        if let path = NSWorkspace.shared.fullPath(forApplication: program) {
            let icon = NSWorkspace.shared.icon(forFile: path)
            icon.size = NSSize(width: 16, height: 16)
            found = icon
        }
        cache[program] = found
        return found
    }
}
