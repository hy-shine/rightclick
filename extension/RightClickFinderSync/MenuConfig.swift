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
}

/// One entry of the fileTypes array in config.json.
struct FileTypeSetting: Decodable {
    let id: String
    let label: String
    let ext: String
    let defaultName: String
    let content: String
    let location: String
    var enabled: Bool? = nil

    private enum CodingKeys: String, CodingKey {
        case id, label, defaultName, content, location, enabled
        case ext = "extension"
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
        FileTypeSetting(id: "word", label: "Word 文档", ext: "docx", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "excel", label: "Excel 工作簿", ext: "xlsx", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "powerpoint", label: "PowerPoint 演示", ext: "pptx", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "wps-document", label: "WPS 文档", ext: "docx", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "wps-spreadsheet", label: "WPS 表格", ext: "xlsx", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "wps-presentation", label: "WPS 演示", ext: "pptx", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "rtf", label: "富文本文档", ext: "rtf", defaultName: "未命名", content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "text", label: "文本文件", ext: "txt", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "markdown", label: "Markdown", ext: "md", defaultName: "未命名",
                        content: "# 标题\n\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "json", label: "JSON", ext: "json", defaultName: "未命名",
                        content: "{}\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "csv", label: "CSV", ext: "csv", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "html", label: "HTML", ext: "html", defaultName: "未命名",
                        content: "<!doctype html>\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "css", label: "CSS", ext: "css", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "javascript", label: "JavaScript", ext: "js", defaultName: "未命名",
                        content: "\"use strict\";\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "swift", label: "Swift", ext: "swift", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "python", label: "Python", ext: "py", defaultName: "未命名",
                        content: "#!/usr/bin/env python3\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "go", label: "Go", ext: "go", defaultName: "未命名",
                        content: "package main\n\n", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "rust", label: "Rust", ext: "rs", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "java", label: "Java", ext: "java", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "c", label: "C", ext: "c", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "cpp", label: "C++", ext: "cpp", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "sql", label: "SQL", ext: "sql", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "typescript", label: "TypeScript", ext: "ts", defaultName: "未命名",
                        content: "", location: FileTypeLocation.submenu),
        FileTypeSetting(id: "shell", label: "Shell", ext: "sh", defaultName: "未命名",
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

    static func visible(_ types: [FileTypeSetting], at location: String) -> [FileTypeSetting] {
        types.filter { entry in
            guard entry.enabled != false && entry.location != FileTypeLocation.hidden else { return false }
            switch entry.location {
            case FileTypeLocation.hidden, FileTypeLocation.submenu, FileTypeLocation.main:
                return entry.location == location
            default:
                return location == FileTypeLocation.submenu
            }
        }
    }
}

/// The parts of config.json that the extension reads. Other keys are ignored.
struct AppConfig: Decodable {
    let menuItems: [MenuItemSetting]?
    let fileTypes: [FileTypeSetting]?
    let watchFolders: [String]?
    let editors: [String]?

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
    let location: String?
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
        ["copy-path", "copy-name", "copy-to", "move-to"],
        ["open-terminal", "open-editor", "open-with"],
        ["zip", "unzip", "hash"],
        ["compare-hash"],
        ["toggle-hidden"],
    ]

    /// Every ID in default order. It is also the order used when config.json is missing.
    static let defaultOrder: [String] = MenuActions.groups.flatMap { $0 }

    /// The IDs to show, in display order, from the menuItems array. Entries with enabled = false
    /// are dropped. An ID missing from the array (for example, one added in a later version)
    /// stays enabled and follows the listed IDs in default order.
    static func enabledIDs(from settings: [MenuItemSetting]?) -> [String] {
        var listed = Set<String>()
        var ids: [String] = []
        for setting in settings ?? [] where defaultOrder.contains(setting.id) && !listed.contains(setting.id) {
            listed.insert(setting.id)
            if setting.enabled {
                ids.append(setting.id)
            }
        }
        ids += defaultOrder.filter { !listed.contains($0) }
        return ids
    }
}
