#!/bin/sh
# Builds the release disk images for the version in mygo.json: one for Apple silicon
# (arm64) and one for Intel (x86_64). Each holds RightClick.app with the Finder Sync
# extension inside Contents/PlugIns, signed ad-hoc, under build/release/, with its
# SHA-256 beside it. The images are not notarized, so the README tells users to allow
# the app once in System Settings.
#
# Usage: sh scripts/release-dmg.sh   (from anywhere inside the repository)
set -eu

cd "$(dirname "$0")/.."

version=$(python3 -I -c 'import json; print(json.load(open("mygo.json"))["version"])')
out=build/release
rm -rf "$out"
mkdir -p "$out"

for arch in arm64 x86_64; do
  case "$arch" in
    arm64) goarch=arm64 ;;
    x86_64) goarch=amd64 ;;
  esac
  work="$out/work-$arch"

  # mygo builds the app only; its own disk image would lack the extension.
  go tool mygo build -skip-dmg -platform "darwin/$goarch" -o "$work"
  app=$(find "$work" -mindepth 2 -maxdepth 2 -name RightClick.app -type d | head -1)
  test -n "$app" || { echo "release: no RightClick.app for $arch under $work" >&2; exit 1; }

  # The licenses ship inside the app and at the root of the disk image.
  cp LICENSE THIRD_PARTY_LICENSES.txt "$app/Contents/Resources/"
  # Installs the extension for this architecture and signs the app again.
  ARCH="$arch" sh extension/build.sh "$app"
  codesign --verify --deep --strict "$app"
  test -d "$app/Contents/PlugIns/RightClickFinderSync.appex" || { echo "release: extension missing from the $arch app" >&2; exit 1; }

  stage="$work/dmg"
  mkdir -p "$stage"
  ditto "$app" "$stage/RightClick.app"
  ln -s /Applications "$stage/Applications"
  cp LICENSE THIRD_PARTY_LICENSES.txt "$stage/"
  dmg="$out/RightClick-$version-$arch.dmg"
  hdiutil create -quiet -volname "RightClick $version $arch" -srcfolder "$stage" -ov -format ULMO "$dmg"

  # Only the image is the release output: remove the app and the staging folders.
  rm -rf "$work"

  # The checksum names the image without its folder, so it also checks from a download folder.
  (cd "$out" && shasum -a 256 "$(basename "$dmg")" > "$(basename "$dmg").sha256")
  echo "release: $dmg"
  cat "$dmg.sha256"
done
