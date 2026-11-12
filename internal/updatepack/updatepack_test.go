package updatepack

import (
	"archive/tar"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateWritesSignedArchiveAndManifest(t *testing.T) {
	app := makeApp(t, true)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	result, err := Create(Options{
		App:        app,
		Name:       "RightClick",
		Version:    "0.1.2",
		Target:     "darwin-arm64",
		BaseURL:    "https://example.com/updates",
		Notes:      "- Faster",
		Appex:      "RightClickFinderSync.appex",
		PrivateKey: priv,
		PublicKey:  pub,
		OutDir:     out,
		Date:       time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := filepath.Base(result.Archive), "RightClick-0.1.2-darwin-arm64.tar.gz"; got != want {
		t.Errorf("archive = %s, want %s", got, want)
	}
	if got, want := filepath.Base(result.Manifest), "update-darwin-arm64.json"; got != want {
		t.Errorf("manifest = %s, want %s", got, want)
	}
	data, err := os.ReadFile(result.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	var entry manifest
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Version != "0.1.2" || entry.Notes != "- Faster" || entry.Date != "2026-11-12T10:00:00Z" {
		t.Errorf("manifest = %+v", entry)
	}
	if entry.Size != result.Size || entry.URL != "https://example.com/updates/RightClick-0.1.2-darwin-arm64.tar.gz" {
		t.Errorf("manifest size/url = %d/%s", entry.Size, entry.URL)
	}
	for _, path := range []string{result.Archive, result.Manifest} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o644 {
			t.Errorf("%s mode = %o, want 644", filepath.Base(path), got)
		}
	}
	signature, err := base64.StdEncoding.DecodeString(entry.Signature)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(result.Archive)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(archive)
	if !ed25519.Verify(pub, sum[:], signature) {
		t.Error("manifest signature does not verify")
	}
	entries := readArchive(t, result.Archive)
	for _, name := range []string{
		"RightClick.app/",
		"RightClick.app/Contents/Info.plist",
		"RightClick.app/Contents/PlugIns/RightClickFinderSync.appex/Contents/Info.plist",
	} {
		if _, ok := entries[name]; !ok {
			t.Errorf("archive has no %s", name)
		}
	}
	link, ok := entries["RightClick.app/Contents/Resources/info-link"]
	if !ok || link.Typeflag != tar.TypeSymlink || link.Linkname != "../Info.plist" {
		t.Errorf("symlink entry = %+v, %v", link, ok)
	}
}

func TestCreateRequiresTheFinderExtension(t *testing.T) {
	app := makeApp(t, false)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(Options{
		App:        app,
		Name:       "RightClick",
		Version:    "0.1.2",
		Target:     "darwin-arm64",
		BaseURL:    "https://example.com/updates",
		Appex:      "RightClickFinderSync.appex",
		PrivateKey: priv,
		PublicKey:  pub,
		OutDir:     t.TempDir(),
	})
	if err == nil || !strings.Contains(err.Error(), "RightClickFinderSync.appex is missing") {
		t.Fatalf("Create = %v, want a missing extension error", err)
	}
}

func TestCreateRejectsAKeyThatDoesNotMatch(t *testing.T) {
	app := makeApp(t, true)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	other, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(Options{
		App:        app,
		Name:       "RightClick",
		Version:    "0.1.2",
		Target:     "darwin-arm64",
		BaseURL:    "https://example.com/updates",
		Appex:      "RightClickFinderSync.appex",
		PrivateKey: priv,
		PublicKey:  other,
		OutDir:     t.TempDir(),
	})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("Create = %v, want a key mismatch error", err)
	}
}

func TestReleaseNotes(t *testing.T) {
	changelog := "# Changelog\n\n## 0.1.2 - 2026-11-12\n\n### Added\n\n- Signed updates.\n\n## [0.1.3](https://example.com) - 2026-11-13\n\n- Next.\n"
	cases := []struct {
		version string
		want    string
	}{
		{"0.1.2", "### Added\n\n- Signed updates."},
		{"v0.1.3", "- Next."},
		{"0.1.4", ""},
	}
	for _, tt := range cases {
		if got := ReleaseNotes(changelog, tt.version); got != tt.want {
			t.Errorf("ReleaseNotes(%q) = %q, want %q", tt.version, got, tt.want)
		}
	}
}

func TestParseKeys(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	gotPriv, err := ParsePrivateKey(base64.StdEncoding.EncodeToString(priv))
	if err != nil || !gotPriv.Equal(priv) {
		t.Errorf("ParsePrivateKey = %v, %v", gotPriv, err)
	}
	gotPub, err := ParsePublicKey(base64.StdEncoding.EncodeToString(pub))
	if err != nil || !gotPub.Equal(pub) {
		t.Errorf("ParsePublicKey = %v, %v", gotPub, err)
	}
	if _, err := ParsePrivateKey("bad"); err == nil {
		t.Error("ParsePrivateKey accepted an invalid key")
	}
}

func makeApp(t *testing.T, extension bool) string {
	t.Helper()
	app := filepath.Join(t.TempDir(), "RightClick.app")
	writeTestFile(t, filepath.Join(app, "Contents", "Info.plist"), "<plist>app</plist>\n", 0o644)
	writeTestFile(t, filepath.Join(app, "Contents", "MacOS", "RightClick"), "binary\n", 0o755)
	if extension {
		writeTestFile(t, filepath.Join(app, "Contents", "PlugIns", "RightClickFinderSync.appex", "Contents", "Info.plist"), "<plist>extension</plist>\n", 0o644)
	}
	if err := os.MkdirAll(filepath.Join(app, "Contents", "Resources"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../Info.plist", filepath.Join(app, "Contents", "Resources", "info-link")); err != nil {
		t.Fatal(err)
	}
	return app
}

func writeTestFile(t *testing.T, path, text string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), mode); err != nil {
		t.Fatal(err)
	}
}

func readArchive(t *testing.T, path string) map[string]tar.Header {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	entries := map[string]tar.Header{}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return entries
		}
		if err != nil {
			t.Fatal(err)
		}
		entries[hdr.Name] = *hdr
	}
}
