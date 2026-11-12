# Changelog

## 0.1.0 - 2026-11-12

The first release.

### Added

- A right-click menu in Finder, in every folder. Each action runs without opening the app.
- Actions: new file, new folder and new file from a template; copy path, copy name and copy name without extension; copy to and move to a favorite folder; zip and unzip; hash with SHA-256, SHA-1 or MD5; compare hashes; show or hide hidden files; open in an editor; open in Terminal, off by default.
- Copy Path, Copy Name and Copy Name Without Extension also work on blank space, for the current folder.
- Permanent delete, off by default. It works only on selected items and asks first.
- A program list for the Open With items, and favorite folders for Copy To and Move To.
- A settings window for the menu, programs, folders and new file types.
- Languages: System, Simplified Chinese, Traditional Chinese and English.
- Check for Updates, in the app menu and in Settings. It opens the newest release page in your browser. The app makes no network requests of its own.

### Notes

- Requires macOS 15.0 or later.
- Two disk images: `RightClick-0.1.0-arm64.dmg` for Apple silicon Macs, and `RightClick-0.1.0-x86_64.dmg` for Intel Macs.
- The app is not notarized. The first time you open it, allow it under Privacy & Security, with Open Anyway.
- Files are copied, archived and unpacked with their download mark, so macOS still asks before a downloaded app runs.
- Unpacking refuses an archive that would unpack to more than 2 GiB or hold more than 10,000 items, and a file of 10 MiB or more that unpacks to over 200 times its packed size. If the data written reaches 2 GiB, unpacking stops and keeps the files already written.
