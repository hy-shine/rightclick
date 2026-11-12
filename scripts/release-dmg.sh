#!/bin/sh
# Builds the release disk images for the version in mygo.json: one for Apple silicon
# (arm64) and one for Intel (x86_64). Each holds RightClick.app with the Finder Sync
# extension inside Contents/PlugIns, signed ad-hoc, under build/release/, with its
# SHA-256 beside it. It also writes the signed update archive and manifest for each
# architecture. The images are not notarized, so the README tells users to allow the
# app once in System Settings.
#
# The update signing key is never read by `mygo build`; set its path in
# MYGO_UPDATER_PRIVATE_KEY_FILE. The public key and update location come from
# mygo.json updates. Usage: sh scripts/release-dmg.sh
set -eu

cd "$(dirname "$0")/.."

version=$(python3 -I -c 'import json; print(json.load(open("mygo.json"))["version"])')
public_key=$(python3 -I -c 'import json; print((json.load(open("mygo.json")).get("updates") or {}).get("publicKey", ""))')
configured_private_key=$(python3 -I -c 'import json; print((json.load(open("mygo.json")).get("updates") or {}).get("privateKey", ""))')
private_key=${MYGO_UPDATER_PRIVATE_KEY_FILE:-}
test -n "$public_key" || { echo "release: add updates.publicKey to mygo.json" >&2; exit 1; }
test -n "$private_key" || { echo "release: set MYGO_UPDATER_PRIVATE_KEY_FILE to the update private key" >&2; exit 1; }
test -f "$private_key" || { echo "release: no private key at $private_key" >&2; exit 1; }
test -z "$configured_private_key" || { echo "release: remove updates.privateKey from mygo.json; the build must not sign its own archive" >&2; exit 1; }

base_url=${RELEASE_UPDATE_BASE_URL:-}
if [ -z "$base_url" ]; then
  base_url=$(python3 -I -c '
import json, sys
updates = json.load(open("mygo.json")).get("updates") or {}
repo = updates.get("github", "")
url = updates.get("url", "")
prefix = updates.get("tagPrefix", "v")
print(url or ("https://github.com/" + repo + "/releases/download/" + prefix + sys.argv[1] if repo else ""))
' "$version")
fi
test -n "$base_url" || { echo "release: mygo.json updates needs github or url, or set RELEASE_UPDATE_BASE_URL" >&2; exit 1; }

base_version=${version%%-*}
build_number=$(printf '%s' "$base_version" | awk -F. '{printf "%d", $1*10000 + $2*100 + $3}')
out=build/release
rm -rf "$out"
mkdir -p "$out"

for arch in arm64 x86_64; do
  case "$arch" in
    arm64) goarch=arm64 ;;
    x86_64) goarch=amd64 ;;
  esac
  work="$out/work-$arch"

  # mygo builds the app only; its own update archive would lack the extension.
  # Keep the signing key out of this step so mygo does not write that archive.
  env -u MYGO_UPDATER_PRIVATE_KEY go tool mygo build -skip-dmg -platform "darwin/$goarch" -o "$work"
  app=$(find "$work" -mindepth 2 -maxdepth 2 -name RightClick.app -type d | head -1)
  test -n "$app" || { echo "release: no RightClick.app for $arch under $work" >&2; exit 1; }
  target=$(basename "$(dirname "$app")")
  test -n "$target" || { echo "release: no mygo target for $arch" >&2; exit 1; }

  # The licenses stay inside the app bundle; the image holds the app and the link.
  cp LICENSE THIRD_PARTY_LICENSES.txt "$app/Contents/Resources/"
  # Installs the extension for this architecture and signs the app again.
  ARCH="$arch" sh extension/build.sh "$app"
  appex="$app/Contents/PlugIns/RightClickFinderSync.appex"
  test -d "$appex" || { echo "release: extension missing from the $arch app" >&2; exit 1; }
  # Finder caches the extension by its version. Keep it in step with the app.
  /usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $version" "$appex/Contents/Info.plist"
  /usr/libexec/PlistBuddy -c "Set :CFBundleVersion $build_number" "$appex/Contents/Info.plist"
  codesign --force --sign - --entitlements extension/RightClickFinderSync/RightClickFinderSync.entitlements "$appex"
  codesign --force --sign - "$app"
  codesign --verify --deep --strict "$app"

  # The archive and manifest are built from the signed app that contains the
  # extension, not from mygo's earlier staging copy.
  MYGO_UPDATER_PUBLIC_KEY="$public_key" go run ./cmd/updatepack \
    -app "$app" \
    -name RightClick \
    -version "$version" \
    -target "$target" \
    -base-url "$base_url" \
    -notes-file CHANGELOG.md \
    -private-key "$private_key" \
    -out "$out"
  update_archive="$out/RightClick-$version-$target.tar.gz"
  test -f "$update_archive" || { echo "release: update archive missing for $arch" >&2; exit 1; }
  (cd "$out" && shasum -a 256 "$(basename "$update_archive")" > "$(basename "$update_archive").sha256")

  stage="$work/dmg"
  mkdir -p "$stage"
  ditto "$app" "$stage/RightClick.app"
  ln -s /Applications "$stage/Applications"
  dmg="$out/RightClick-$version-$arch.dmg"
  hdiutil create -quiet -volname "RightClick $version $arch" -srcfolder "$stage" -ov -format ULMO "$dmg"
  test -f "$dmg" || { echo "release: no disk image at $dmg" >&2; exit 1; }

  # Only the image is the release output: remove the app and the staging folders.
  rm -rf "$work"

  # The checksum names the image without its folder, so it also checks from a download folder.
  (cd "$out" && shasum -a 256 "$(basename "$dmg")" > "$(basename "$dmg").sha256")
  echo "release: $dmg"
  cat "$dmg.sha256"
done
