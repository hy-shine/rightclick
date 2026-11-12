#!/bin/sh
# Builds the Finder Sync extension with swiftc, signs it, and installs it into
# an app bundle's Contents/PlugIns. The app is signed again afterwards, because
# adding a plug-in changes what the app seals.
#
# Usage: extension/build.sh path/to/RightClick.app
set -eu

here=$(cd "$(dirname "$0")" && pwd)
name=RightClickFinderSync
app=${1:?usage: extension/build.sh <App.app>}
out="$here/../.build/$name.appex"

rm -rf "$out"
mkdir -p "$out/Contents/MacOS"
cp "$here/$name/Info.plist" "$out/Contents/Info.plist"
mkdir -p "$out/Contents/Resources/FileIcons"
cp "$here/../internal/catalog/icons/"*.svg "$out/Contents/Resources/FileIcons/"
cp "$here/../internal/catalog/icons/"*.txt "$out/Contents/Resources/FileIcons/"
# The menu text: one .lproj folder per language, read through Bundle.main.
for lproj in en.lproj zh-Hans.lproj zh-Hant.lproj; do
  cp -R "$here/$name/$lproj" "$out/Contents/Resources/"
done
# Action icons, collected from the Lucide set (ISC); see ActionIcons/LICENSE.txt.
cp -R "$here/../internal/actionicons/icons" "$out/Contents/Resources/ActionIcons"

# -application-extension restricts the code to APIs that extensions may use.
# _NSExtensionMain is the entry point that the extension host calls.
# Compiles every Swift file in the extension folder into one binary.
swiftc -application-extension \
  -module-name "$name" \
  -target arm64-apple-macosx15.0 \
  -framework FinderSync -framework AppKit \
  -Xlinker -e -Xlinker _NSExtensionMain \
  -o "$out/Contents/MacOS/$name" \
  "$here/$name"/*.swift

codesign --force --sign - --entitlements "$here/$name/$name.entitlements" "$out"

rm -rf "$app/Contents/PlugIns/$name.appex"
mkdir -p "$app/Contents/PlugIns"
cp -R "$out" "$app/Contents/PlugIns/"
codesign --force --sign - "$app"
