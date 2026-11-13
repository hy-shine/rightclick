import AppKit
import UniformTypeIdentifiers

/// Icons for the built-in actions. Most come from the Lucide icon set (ISC
/// license, see THIRD_PARTY_LICENSES.txt at the repository root). Folder rows
/// use the system folder artwork, and the destructive action uses the system
/// red, so those two read like the rest of macOS. The files are loaded once,
/// off the main thread: Finder builds the menu on its main thread every time it
/// opens, and a cold load of every icon costs about 13 ms, which the menu itself
/// must not pay.
enum ActionIcons {
    /// The icon file of each action ID or shortcut kind, without the extension.
    /// Keep it in step with MenuActions.defaultOrder.
    static let names: [String: String] = [
        "new-file": "new-file",
        "new-folder": "new-folder",
        "new-template": "new-template",
        "paste-as-file": "clipboard-paste",
        "add-template": "add-template",
        "copy-path": "copy-path",
        "copy-terminal-path": "open-terminal",
        "copy-relative-path": "copy-path",
        "copy-markdown-link": "copy-path",
        "copy-content": "file-text",
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
        "chmod-x": "play",
        "toggle-hidden": "toggle-hidden",
        "delete-permanently": "delete-permanently",
        "favorite-folders": "favorite-folders",
        // Shortcut kinds. "type" and "template" use the file type icon instead.
        "app": "app",
        "folder": "folder",
        // The menu of the app itself: a pointer clicking, the right-click idea.
        "menu": "menu",
        // Not an action of its own: the "Choose Another Folder…" row uses the
        // folder artwork, so every row in that submenu keeps one icon column.
        "choose-folder": "folder",
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
        if let folder = NSWorkspace.shared.icon(for: .folder).copy() as? NSImage {
            folder.size = NSSize(width: 16, height: 16)
            folder.isTemplate = false
            loaded["folder"] = folder
        }
        if let delete = loaded["delete-permanently"] {
            loaded["delete-permanently"] = tinted(delete, color: .systemRed)
        }
        lock.lock()
        cache = loaded
        lock.unlock()
    }

    /// Draws a template image in one fixed color. The drawing handler runs when
    /// Finder renders the menu, so the system color resolves in the current
    /// appearance instead of being frozen when the extension starts.
    private static func tinted(_ image: NSImage, color: NSColor) -> NSImage {
        let tinted = NSImage(size: image.size, flipped: false) { rect in
            image.draw(in: rect, from: NSRect(origin: .zero, size: image.size),
                       operation: .sourceOver, fraction: 1)
            color.set()
            rect.fill(using: .sourceAtop)
            return true
        }
        tinted.isTemplate = false
        return tinted
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
