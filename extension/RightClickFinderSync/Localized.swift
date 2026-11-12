import Foundation

/// Menu text in the language Finder shows. The strings are in the .lproj folders
/// next to this file; build.sh copies them all into the extension.
enum Localized {
    /// The bundle that holds the strings. The extension uses the main bundle, which
    /// is the extension itself. Tests point it at one language.
    static var bundle: Bundle = .main

    /// Chooses the language from the app's language setting in config.json. An empty setting
    /// follows the system, as Bundle.main does. container holds the .lproj folders.
    static func select(language: String?, from container: Bundle = .main) {
        if let language, !language.isEmpty,
           let path = container.path(forResource: language, ofType: "lproj"),
           let chosen = Bundle(path: path) {
            bundle = chosen
        } else {
            bundle = .main
        }
    }

    /// Returns the text for key, or key itself when no string has that key.
    static func text(_ key: String) -> String {
        bundle.localizedString(forKey: key, value: key, table: nil)
    }

    /// Returns the text for key with value put in place of its %@.
    static func text(_ key: String, _ value: String) -> String {
        String(format: text(key), value)
    }
}
