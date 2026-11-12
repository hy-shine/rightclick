# AGENTS.md

Instructions for AI coding agents working in this repository. Users and contributors should start with README.md.

These rules apply to any clone of the repository. Do not rely on one machine's paths, installed apps, or local files.

## Project

RightClick is a macOS Finder context-menu helper for everyday file actions.

- The app is written in Go and built with mygo. The version is pinned in `go.mod`. The UI is native, through mygo's `ui` package.
- The Finder menu is a Swift Finder Sync extension in `extension/`. See "Finder extension" below.
- Requirements: macOS 15.0 or later, the Go version in `go.mod`, and the Xcode Command Line Tools (`swiftc`, `codesign`).
- Publishing follows the Git rule under Working rules. The remote `origin` is github.com/hy-shine/rightclick.

## Commands

Run from the repository root. Every Go check must exit 0 before you report a Go change as done.

```sh
go test ./...
CGO_ENABLED=0 go test ./...
go vet ./...
go tool mygo vet .
gofmt -l .        # must print nothing
```

Finder extension tests. Run them after changing `extension/` or any menu action:

```sh
swiftc -application-extension -o "${TMPDIR:-/tmp}/rc_ext_test" \
  extension/RightClickFinderSync/FinderSyncExtension.swift \
  extension/RightClickFinderSync/MenuConfig.swift \
  extension/RightClickFinderSync/Localized.swift \
  extension/RightClickFinderSync/ActionIcons.swift \
  extension/tests/main.swift
"${TMPDIR:-/tmp}/rc_ext_test"
```

Other commands:

- `go tool mygo dev`: run the app with hot reload.
- `go tool mygo generate`: regenerate `internal/locales/catalog_gen.go` after editing the locale JSON.
- `go tool mygo build -skip-dmg`: build the app bundle under `build/`.
- `sh extension/build.sh <path to RightClick.app>`: build the Finder extension and install it into that bundle. Run it after `go tool mygo build`. `ARCH=arm64` or `ARCH=x86_64` picks the Mac architecture; by default it is this Mac's.
- `sh scripts/release-dmg.sh`: build two release disk images under `build/release/`, one for Apple silicon (arm64) and one for Intel (x86_64). Each has the Finder extension inside the app and its SHA-256 beside it.

Do not quit, restart or re-register the app or the Finder extension on your own. If the user asks for that: the app registers the extension at every start, so opening it again after a rebuild is enough. If the menu does not appear, the user clicks Turn On on the Finder Extension page, or restarts Finder.

## Layout

```
cmd/rightclick/                   app entry only: start-up and which window opens
internal/app/                     menu action handlers, rightclick:// link parsing and the --action run, with their tests
internal/config/                  saved settings, defaults and migrations
internal/files/                   file operations; no UI code
internal/finder/                  Finder helpers: restart, hidden files, the action script, extension registration
internal/settings/                the settings window; pane_*.go holds the panes, and pane_system.go holds General and Diagnostics
internal/locales/                 en.json (source), zh-Hans.json, zh-Hant.json; catalog_gen.go is generated
internal/actionicons/             Lucide icons for menu actions; bundled into the app and the extension
internal/catalog/                 blank document templates for New File, with their icons
extension/                        Finder Sync extension in Swift, build.sh, tests/main.swift
scripts/                          release-dmg.sh builds the release disk image; document_templates.py regenerates the blank templates
resources/                        app icon used by the README
```

- Put new Go code in a package directory, not in the repository root.
- Do not depend on files that are not in git (local notes, scratch output, another machine's build). If a task needs such a file, ask for it.

## Working rules

- **Scope.** Change only what the task asks. Do not rename, reformat or tidy unrelated code. If the change needs something outside the request, ask first.
- **Analysis means no edits.** When the request is to read, check, review or explain, change no files and leave no scratch files behind.
- **Git.** Do not commit, push, tag or create a release unless asked. When you commit, use Conventional Commits with a scope where it fits, as the history does: `feat(settings): ...`, `refactor: ...`.
- **Identity.** `mygo.json` `name` and `identifier` are placeholders. Ask before changing them.
- **Generated files.** Do not hand-edit `internal/locales/catalog_gen.go`. Change the locale JSON, then run `go tool mygo generate`.
- **Dependencies.** Pin exact versions, such as `@v0.6.4`. `mygo init` fetches the latest mygo; if you run it, restore the pinned version. Keep `CGO_ENABLED=0` builds working.
- **Tests.** Wait on a condition, for example with `ui.NewTester`. Do not use fixed sleeps.

## Language and copy

- Code comments, test names and test strings are in English. The one exception is test data that checks a Chinese locale, such as `zh-Hans` or `zh-Hant` strings or Chinese file names.
- Chinese text belongs in `internal/locales/zh-Hans.json`, `internal/locales/zh-Hant.json`, `README.zh-CN.md` and that test data. Nowhere else.
- User-visible text goes in `internal/locales/en.json`, `zh-Hans.json` and `zh-Hant.json`, with the same keys in every file.
- Finder extension text goes in the `*.lproj/Localizable.strings` files under `extension/RightClickFinderSync/`, one per language (en, zh-Hans, zh-Hant), with the same keys in each.

## Icons

- Do not hand-draw action icons. Use the Lucide SVGs (ISC) in `internal/actionicons/icons/`.
- A new menu action needs an icon whose meaning matches the action. Extension tests in `extension/tests/main.swift` must cover it.

## Menu actions

- An action's ID is defined in two places: `defaultMenuIDs` in `internal/config/config.go`, and `MenuActions` in `extension/RightClickFinderSync/MenuConfig.swift`. Change both together, and add or update a test on each side.
- When an action is removed, drop it from both sides and add a migration in `internal/config/` so saved settings do not keep it.

## Settings UI

- Reuse the shared controls in `internal/settings/style.go`, such as `actionButton`, `iconButton`, `settingRow` and `itemRow`, before adding a new variant.
- Check every layout visually. Render the pane to PNG with `ui.NewTester(...).Image()` in a temporary `_test.go` file, look at the image, then delete both the test file and the PNG. Judge the result as a designer would: hierarchy, spacing, alignment and wording. Compiling is not enough.

## Finder extension

- Keep menu behaviour in Swift. Keep file operations in Go. The extension only builds the menu. To run an action it calls the `rightclick-action` script in its Application Scripts folder. The app writes that script at every start, and the script starts the app's executable with `--action`, so no window opens.
- An action's ID is defined in `MenuActions` (Swift) and `defaultMenuIDs` (Go); see "Menu actions" above.
- Each menu action needs an icon in `internal/actionicons/icons/` whose meaning matches the action. `extension/tests/main.swift` must check it.
- After changing the extension, run the Swift tests and `extension/build.sh`. Do not register or restart anything unless asked.
- The user must turn the extension on in System Settings. Do not assume it is on.

## Done means

Before you report success:

1. Run the commands that match what you changed. Go changes need every Go check. Extension or menu changes need the Swift tests.
2. For UI changes, look at a rendered image of the affected pane.
3. Report failures with the actual output. Do not say a check passed unless you ran it in this session.
