package settings

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
)

// applicationFolders are the folders searched for an application bundle, in order.
var applicationFolders = func() []string {
	folders := []string{"/Applications", "/System/Applications"}
	if home, err := os.UserHomeDir(); err == nil {
		folders = append(folders, filepath.Join(home, "Applications"))
	}
	return folders
}()

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// editorIcon returns the icon of the editor named name, reading its bundle once.
// Nil means there is no icon to show, and the row keeps the generic glyph.
func (m *app) editorIcon(name string) *ui.Bitmap {
	if icon, ok := m.editorIcons[name]; ok {
		return icon
	}
	if m.editorIcons == nil {
		m.editorIcons = make(map[string]*ui.Bitmap)
	}
	icon := appIcon(applicationFolders, name)
	m.editorIcons[name] = icon
	return icon
}

// appIcon returns the icon of the application named name, as open -a takes the name.
// It returns nil when no bundle with that name is in folders, or the bundle has no
// readable PNG icon.
func appIcon(folders []string, name string) *ui.Bitmap {
	bundle, err := findBundle(folders, name)
	if err != nil {
		return nil
	}
	file, err := iconFile(bundle)
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	png := pngFromICNS(data)
	if png == nil {
		return nil
	}
	icon, err := ui.DecodeBitmap(png)
	if err != nil {
		return nil
	}
	return icon
}

// findBundle returns the path of the .app bundle called name in the first folder that has it.
func findBundle(folders []string, name string) (string, error) {
	for _, folder := range folders {
		path := filepath.Join(folder, name+".app")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}
	}
	return "", errors.New("application not found: " + name)
}

// iconFile returns the .icns file that the bundle's Info.plist names in CFBundleIconFile.
// plutil reads the plist, which may be in binary form.
func iconFile(bundle string) (string, error) {
	out, err := exec.Command("/usr/bin/plutil", "-extract", "CFBundleIconFile", "raw", "-o", "-",
		filepath.Join(bundle, "Contents", "Info.plist")).Output()
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "", errors.New("no CFBundleIconFile in " + bundle)
	}
	if filepath.Ext(name) == "" {
		name += ".icns"
	}
	return filepath.Join(bundle, "Contents", "Resources", name), nil
}

// pngFromICNS returns the largest PNG image in an .icns file, or nil when it has none.
// An .icns file is a list of entries, each a four-byte type, a four-byte big-endian
// length that counts the eight header bytes, and the data. Large sizes are stored as
// PNG; the small sizes use an older encoding, which is skipped.
func pngFromICNS(data []byte) []byte {
	if len(data) < 8 || string(data[:4]) != "icns" {
		return nil
	}
	var best []byte
	for rest := data[8:]; len(rest) >= 8; {
		size := int(binary.BigEndian.Uint32(rest[4:8]))
		if size < 8 || size > len(rest) {
			break
		}
		entry := rest[8:size]
		if bytes.HasPrefix(entry, pngSignature) && len(entry) > len(best) {
			best = entry
		}
		rest = rest[size:]
	}
	return best
}
