import Foundation

/// Locates the settings that the main app writes to ~/Library/Application Support/RightClick/.
/// The extension is sandboxed, so NSHomeDirectory() points at its container. getpwuid returns
/// the real home directory. The temporary-exception entitlement in
/// RightClickFinderSync.entitlements grants read-only access to this folder.
enum AppSupportFiles {
    /// The RightClick folder in the real home directory.
    static var directory: URL? {
        guard let entry = getpwuid(getuid()) else { return nil }
        let home = String(cString: entry.pointee.pw_dir)
        return URL(fileURLWithPath: home, isDirectory: true)
            .appendingPathComponent("Library/Application Support/RightClick", isDirectory: true)
    }
}

/// One entry of the menuItems array in config.json.
struct MenuItemSetting: Decodable {
    let id: String
    let enabled: Bool
    let location: String?
    var label: String? = nil
}

struct FavoriteFolderSetting: Decodable {
    let id: String
    let path: String
    let label: String
    let enabled: Bool
}

/// One entry of the fileTypes array in config.json.
struct FileTypeSetting: Decodable {
    let id: String
    let label: String
    let ext: String
    let defaultName: String
    let content: String
    let location: String // kept for older config files; the menu ignores it
    var enabled: Bool? = nil

    private enum CodingKeys: String, CodingKey {
        case id, label, defaultName, content, location, enabled
        case ext = "extension"
    }

    /// The name the menu shows. A built-in type stores no label, so its name follows the
    /// language Finder shows. Any other type without a label shows its extension.
    var displayLabel: String {
        if !label.isEmpty { return label }
        switch id {
        case "word": return Localized.text("finder.type.word")
        case "excel": return Localized.text("finder.type.excel")
        case "powerpoint": return Localized.text("finder.type.powerpoint")
        case "wps-document": return Localized.text("finder.type.wpsDocument")
        case "wps-spreadsheet": return Localized.text("finder.type.wpsSpreadsheet")
        case "wps-presentation": return Localized.text("finder.type.wpsPresentation")
        case "rtf": return Localized.text("finder.type.rtf")
        case "text": return Localized.text("finder.type.text")
        default: return ext
        }
    }
}

enum FileTypeLocation {
    static let hidden = "hidden"
    static let submenu = "submenu"
    static let main = "main"
}

enum FileTypes {
    /// The built-in text types. Keep this order and these IDs in sync with
    /// config.defaultFileTypes in Go.
    static let defaults: [FileTypeSetting] = [
        FileTypeSetting(id: "word", label: "", ext: "docx", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "excel", label: "", ext: "xlsx", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "powerpoint", label: "", ext: "pptx", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "wps-document", label: "", ext: "docx", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "wps-spreadsheet", label: "", ext: "xlsx", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "wps-presentation", label: "", ext: "pptx", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "rtf", label: "", ext: "rtf", defaultName: "", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "text", label: "", ext: "txt", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "markdown", label: "Markdown", ext: "md", defaultName: "",
                        content: "# Title\n\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "json", label: "JSON", ext: "json", defaultName: "",
                        content: "{}\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "csv", label: "CSV", ext: "csv", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "html", label: "HTML", ext: "html", defaultName: "",
                        content: "<!doctype html>\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "css", label: "CSS", ext: "css", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "javascript", label: "JavaScript", ext: "js", defaultName: "",
                        content: "\"use strict\";\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "swift", label: "Swift", ext: "swift", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "python", label: "Python", ext: "py", defaultName: "",
                        content: "#!/usr/bin/env python3\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "go", label: "Go", ext: "go", defaultName: "",
                        content: "package main\n\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "rust", label: "Rust", ext: "rs", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "java", label: "Java", ext: "java", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "c", label: "C", ext: "c", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "cpp", label: "C++", ext: "cpp", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "sql", label: "SQL", ext: "sql", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "typescript", label: "TypeScript", ext: "ts", defaultName: "",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "shell", label: "Shell", ext: "sh", defaultName: "",
                        content: "#!/usr/bin/env bash\n", location: FileTypeLocation.submenu),
    ]

    /// Returns the configured types in their saved order, appending defaults
    /// introduced in a newer version. Duplicate IDs from a damaged config are
    /// ignored.
    static func resolved(from settings: [FileTypeSetting]?) -> [FileTypeSetting] {
        var seen = Set<String>()
        var types: [FileTypeSetting] = []
        for entry in settings ?? [] where !entry.id.isEmpty && !seen.contains(entry.id) {
            seen.insert(entry.id)
            types.append(entry)
        }
        for entry in defaults where !seen.contains(entry.id) {
            types.append(entry)
        }
        return types
    }

    /// The types that belong in the New File submenu. The stored location is
    /// ignored: file types always live inside that submenu.
    static func enabled(_ types: [FileTypeSetting]) -> [FileTypeSetting] {
        types.filter { $0.enabled != false && $0.location != FileTypeLocation.hidden }
    }
}

/// The parts of config.json that the extension reads. Other keys are ignored.
struct AppConfig: Decodable {
    /// "en", "zh-Hans", "zh-Hant", or nil and empty for the system language.
    let language: String?
    let menuItems: [MenuItemSetting]?
    let fileTypes: [FileTypeSetting]?
    let editors: [String]?
    let programs: [String]?
    let terminals: [String]?
    let favoriteFolders: [FavoriteFolderSetting]?

    /// Returns nil when config.json is missing or cannot be decoded.
    static func load() -> AppConfig? {
        guard let url = AppSupportFiles.directory?.appendingPathComponent("config.json"),
              let data = try? Data(contentsOf: url) else { return nil }
        return try? JSONDecoder().decode(AppConfig.self, from: data)
    }
}

/// One item of templates.json, which is an array of {"name": ..., "content": ...} objects.
/// The menu needs only the name. The main app reads the content.
struct TemplateEntry: Decodable {
    let name: String
    let disabled: Bool?
    let location: String? // kept for older files; templates stay inside New File
    let label: String?
}

enum Templates {
    /// Returns an empty list when templates.json is missing or cannot be decoded.
    static func load() -> [TemplateEntry] {
        guard let url = AppSupportFiles.directory?.appendingPathComponent("templates.json"),
              let data = try? Data(contentsOf: url),
              let entries = try? JSONDecoder().decode([TemplateEntry].self, from: data) else { return [] }
        return entries
    }
}

/// Action IDs. Each ID is also the action name in the rightclick:// URL.
enum MenuActions {
    /// Groups in the order they appear in the submenu. Within a group, config order applies.
    static let groups: [[String]] = [
        ["new-file", "new-folder", "new-template"],
        ["copy-path", "copy-name", "copy-basename", "copy-to", "move-to"],
        ["open-terminal", "open-editor"],
        ["zip", "unzip", "hash"],
        ["compare-hash"],
        ["toggle-hidden"],
        ["delete-permanently"],
        ["add-template"],
    ]

    /// Every ID in default order. It is also the order used when config.json is missing.
    static let defaultOrder: [String] = MenuActions.groups.flatMap { $0 }

    /// Actions that start switched off. Like the Go side's defaultOffMenuIDs, they stay out of
    /// the menu until the user turns them on.
    static let defaultOff: Set<String> = ["delete-permanently", "open-terminal"]

    /// The IDs to show, in display order, from the menuItems array. Entries with enabled = false
    /// are dropped. An ID missing from the array (for example, one added in a later version)
    /// stays enabled and follows the listed IDs in default order.
    static func enabledIDs(from settings: [MenuItemSetting]?) -> [String] {
        var listed = Set<String>()
        var ids: [String] = []
        for setting in settings ?? [] where (defaultOrder.contains(setting.id) || setting.id == "favorite-folders" ||
            ["type:", "template:", "app:", "folder:"].contains(where: { setting.id.hasPrefix($0) })) && !listed.contains(setting.id) {
            listed.insert(setting.id)
            if setting.enabled {
                ids.append(setting.id)
            }
        }
        ids += defaultOrder.filter { !listed.contains($0) && !defaultOff.contains($0) }
        return ids
    }
}
