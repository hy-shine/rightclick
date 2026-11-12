package settings

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// icnsEntry is one entry of an .icns file: a four-byte type, a length that counts the header, and the data.
func icnsEntry(kind string, body []byte) []byte {
	out := []byte(kind)
	out = binary.BigEndian.AppendUint32(out, uint32(8+len(body)))
	return append(out, body...)
}

func icnsFile(entries ...[]byte) []byte {
	body := bytes.Join(entries, nil)
	out := []byte("icns")
	out = binary.BigEndian.AppendUint32(out, uint32(8+len(body)))
	return append(out, body...)
}

func pngOfSize(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// makeBundle writes an application bundle called name under folder, whose Info.plist names
// iconFile as its icon and whose Resources folder holds the given .icns data.
func makeBundle(t *testing.T, folder, name, iconFile string, icns []byte) {
	t.Helper()
	bundle := filepath.Join(folder, name+".app")
	resources := filepath.Join(bundle, "Contents", "Resources")
	if err := os.MkdirAll(resources, 0o755); err != nil {
		t.Fatal(err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>CFBundleIconFile</key><string>` + iconFile + `</string></dict></plist>`
	if err := os.WriteFile(filepath.Join(bundle, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(resources, iconFile+".icns"), icns, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAppIconReadsTheLargestPNGOfTheBundle(t *testing.T) {
	folder := t.TempDir()
	// The non-PNG entry comes first and must be skipped; the 16px PNG is smaller than the 64px one.
	icns := icnsFile(
		icnsEntry("il32", []byte("not a png")),
		icnsEntry("ic04", pngOfSize(t, 16)),
		icnsEntry("ic07", pngOfSize(t, 64)),
	)
	makeBundle(t, folder, "Demo", "Demo", icns)

	icon := appIcon([]string{folder}, "Demo")
	if icon == nil {
		t.Fatal("appIcon returned nil for a bundle with a PNG icon")
	}
	if w, h := icon.Size(); w != 64 || h != 64 {
		t.Fatalf("icon is %dx%d, want the 64x64 PNG", w, h)
	}
}

func TestAppIconIsNilWithoutABundleOrAPNG(t *testing.T) {
	folder := t.TempDir()
	makeBundle(t, folder, "NoPNG", "NoPNG", icnsFile(icnsEntry("il32", []byte("not a png"))))

	if icon := appIcon([]string{folder}, "Missing"); icon != nil {
		t.Fatal("appIcon found an application that is not installed")
	}
	if icon := appIcon([]string{folder}, "NoPNG"); icon != nil {
		t.Fatal("appIcon returned an icon with no PNG entry")
	}
}

func TestEditorIconIsReadOnce(t *testing.T) {
	folder := t.TempDir()
	makeBundle(t, folder, "Demo", "Demo", icnsFile(icnsEntry("ic07", pngOfSize(t, 64))))
	saved := applicationFolders
	applicationFolders = []string{folder}
	t.Cleanup(func() { applicationFolders = saved })

	m := &app{}
	if m.editorIcon("Demo") == nil {
		t.Fatal("first lookup found no icon")
	}
	// Once read, the icon stays even when the bundle is gone, so the settings pane never reads the disk again.
	if err := os.RemoveAll(filepath.Join(folder, "Demo.app")); err != nil {
		t.Fatal(err)
	}
	if m.editorIcon("Demo") == nil {
		t.Fatal("second lookup read the disk again instead of using the cached icon")
	}
}

// TestEntryAppIconTakesOnlyTheAppEntries checks that an entry that opens an app
// gets that app's icon, and that other entries and missing apps get none.
func TestEntryAppIconTakesOnlyTheAppEntries(t *testing.T) {
	folder := t.TempDir()
	makeBundle(t, folder, "Demo", "Demo", icnsFile(icnsEntry("ic07", pngOfSize(t, 64))))
	saved := applicationFolders
	applicationFolders = []string{folder}
	t.Cleanup(func() { applicationFolders = saved })

	m := &app{}
	if m.entryAppIcon("app:Demo") == nil {
		t.Error("the entry that opens Demo has no icon, want Demo's icon")
	}
	if m.entryAppIcon("app:Missing") != nil {
		t.Error("an app that is not installed got an icon")
	}
	if m.entryAppIcon("copy-path") != nil {
		t.Error("a built-in action got an app icon")
	}
}
