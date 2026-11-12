#!/bin/sh
# Builds the release disk image for the version in mygo.json: RightClick.app with the
# Finder Sync extension inside Contents/PlugIns, signed ad-hoc, in a compressed DMG
# under build/release/. The image is not notarized, so the README tells users to allow
# it once in System Settings.
#
# Usage: sh scripts/release-dmg.sh   (from anywhere inside the repository)
set -eu

cd "$(dirname "$0")/.."
version=$(python3 -I -c 'import json; print(json.load(open("mygo.json"))["version"])')
out=build/release
rm -rf "$out"

# mygo builds the app only; its own disk image would lack the extension.
go tool mygo build -skip-dmg -platform darwin/arm64 -o "$out"
app=$(find "$out" -mindepth 2 -maxdepth 2 -name RightClick.app -type d | head -1)
test -n "$app" || { echo "release: no RightClick.app under $out" >&2; exit 1; }

# Installs the extension into the app and signs the app again.
sh extension/build.sh "$app"
codesign --verify --deep --strict "$app"
test -d "$app/Contents/PlugIns/RightClickFinderSync.appex" || { echo "release: extension missing from the app" >&2; exit 1; }

stage="$out/dmg"
mkdir -p "$stage"
ditto "$app" "$stage/RightClick.app"
ln -s /Applications "$stage/Applications"
dmg="$out/RightClick-$version-arm64.dmg"
hdiutil create -quiet -volname "RightClick $version" -srcfolder "$stage" -ov -format ULMO "$dmg"
rm -rf "$stage"
# mygo leaves the app in its own folder. Only the image is the release output.
test "$(dirname "$app")" != "$out" && rm -rf "$(dirname "$app")"

# The checksum names the image without its folder, so it also checks from a download folder.
(cd "$out" && shasum -a 256 "$(basename "$dmg")" > "$(basename "$dmg").sha256")
echo "release: $dmg"
cat "$dmg.sha256"
