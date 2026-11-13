<p align="center">
  <img src="resources/icon.png" width="128" alt="RightClick app icon">
</p>

# RightClick

Make the Finder right-click menu do more: create files from types and templates, copy names and paths, copy or move items to your favorite folders, open them in Terminal or any app, zip and unzip, and compute file hashes.

English | [简体中文](README.zh-CN.md)

## Features

- **New File and New from Template**: Word, Excel, PowerPoint, Markdown, JSON and many programming languages, plus your own templates.
- **Open with**: Terminal, the default app, or any app you add.
- **Copy**: the name, the name without its extension, or the path. Enable Copy as Terminal Path, Copy Relative Path, or Copy Markdown File Link for other formats. Relative paths and Markdown links ask you to choose a base folder.
- **Copy To / Move To** and **Favorite Folders**: send items to a folder you choose or to a favorite.
- **Archives and hashes**: Compress and extract ZIP archives, and compute or compare file hashes.
- **Also**: show or hide hidden files, and Delete Permanently (off by default).
- **Menu layout and languages**: choose where each entry sits and drag to reorder. English, Simplified Chinese and Traditional Chinese.

## Install

Requires macOS 15.0 or later.

1. Download the disk image from GitHub Releases: `RightClick-0.1.3-arm64.dmg` for Apple silicon, `RightClick-0.1.3-x86_64.dmg` for Intel.
2. Open the disk image and drag RightClick into `/Applications`.
3. Double-click RightClick. macOS shows a warning. Close it.

RightClick 0.1.2 and later update from the app menu. If you have 0.1.1 or earlier, install 0.1.3 manually once.

If RightClick still won't open, use one of these two ways.

**Way 1: System Settings**

- Open **System Settings › Privacy & Security**. Under **Security**, find RightClick and click **Open Anyway**. Enter your password or use Touch ID, then click **Open Anyway** again.

**Way 2: Terminal**

- Open **Terminal**, paste this line, and press Return:

```sh
sudo xattr -dr com.apple.quarantine /Applications/RightClick.app
```

- Enter your password if asked. Then double-click RightClick.

Then turn on the Finder extension. The RightClick window shows you where.

## Build from Source

Requires Go (the version in `go.mod`) and Xcode or the Xcode Command Line Tools. Clone this repository, then run:

```sh
go tool mygo build -skip-dmg
sh extension/build.sh build/darwin-arm64/RightClick.app
open build/darwin-arm64/RightClick.app
```

## License

RightClick is released under the [MIT License](LICENSE). The bundled Lucide icons are under the ISC license; see `THIRD_PARTY_LICENSES.txt`.
