// Package updatepack builds the signed archives and manifests that the
// in-app updater downloads.
package updatepack

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Options describe one platform's update.
type Options struct {
	// App is the .app bundle to archive.
	App string
	// Name is the file name prefix, usually the app name.
	Name string
	// Version is the version the manifest advertises.
	Version string
	// Target is the mygo update target, such as "darwin-arm64" or "darwin-amd64".
	Target string
	// BaseURL is the HTTPS directory that serves the archive.
	BaseURL string
	// Notes are the release notes in Markdown.
	Notes string
	// Appex is the Finder extension bundle name that must be in Contents/PlugIns.
	// Empty skips the check.
	Appex string
	// PrivateKey signs the archive; PublicKey must be its public half.
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	// OutDir receives the archive and update-<target>.json.
	OutDir string
	// Date is the manifest date. Zero means now, UTC.
	Date time.Time
}

// Result names the files written.
type Result struct {
	Archive  string
	Manifest string
	Size     int64
}

type manifest struct {
	Version   string `json:"version"`
	Notes     string `json:"notes,omitempty"`
	Date      string `json:"date,omitempty"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	Signature string `json:"signature"`
}

// Create writes the archive, signs its SHA-256 digest with the Ed25519 key, and
// writes the manifest next to it.
func Create(opts Options) (Result, error) {
	if err := validate(opts); err != nil {
		return Result{}, err
	}
	if opts.Date.IsZero() {
		opts.Date = time.Now().UTC()
	}
	if opts.Appex != "" {
		appex := filepath.Join(opts.App, "Contents", "PlugIns", opts.Appex)
		info, err := os.Stat(appex)
		if err != nil || !info.IsDir() {
			return Result{}, fmt.Errorf("update: %s is missing", appex)
		}
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("update: create output directory: %w", err)
	}

	archiveName := fmt.Sprintf("%s-%s-%s.tar.gz", safeName(opts.Name), opts.Version, opts.Target)
	archivePath := filepath.Join(opts.OutDir, archiveName)
	size, digest, err := writeArchive(archivePath, opts.App)
	if err != nil {
		return Result{}, err
	}
	manifestPath := filepath.Join(opts.OutDir, "update-"+opts.Target+".json")
	entry := manifest{
		Version:   opts.Version,
		Notes:     opts.Notes,
		Date:      opts.Date.UTC().Format(time.RFC3339),
		URL:       strings.TrimSuffix(opts.BaseURL, "/") + "/" + url.PathEscape(archiveName),
		Size:      size,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(opts.PrivateKey, digest)),
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("update: encode manifest: %w", err)
	}
	if err := writeFileAtomic(manifestPath, append(data, '\n')); err != nil {
		return Result{}, fmt.Errorf("update: write manifest: %w", err)
	}
	return Result{Archive: archivePath, Manifest: manifestPath, Size: size}, nil
}

func validate(opts Options) error {
	if opts.App == "" {
		return errors.New("update: app is required")
	}
	info, err := os.Stat(opts.App)
	if err != nil || !info.IsDir() || filepath.Ext(opts.App) != ".app" {
		return fmt.Errorf("update: %s is not an app bundle", opts.App)
	}
	if opts.Name == "" || opts.Version == "" || opts.Target == "" {
		return errors.New("update: name, version and target are required")
	}
	if strings.ContainsAny(opts.Target, `/\`) || filepath.Base(opts.Target) != opts.Target {
		return fmt.Errorf("update: invalid target %q", opts.Target)
	}
	u, err := url.Parse(opts.BaseURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("update: base URL %q is not HTTPS", opts.BaseURL)
	}
	if opts.OutDir == "" {
		return errors.New("update: output directory is required")
	}
	if len(opts.PrivateKey) != ed25519.PrivateKeySize {
		return errors.New("update: invalid private key")
	}
	if len(opts.PublicKey) != ed25519.PublicKeySize {
		return errors.New("update: invalid public key")
	}
	derived := opts.PrivateKey.Public().(ed25519.PublicKey)
	if !bytes.Equal(derived, opts.PublicKey) {
		return errors.New("update: private key does not match the public key")
	}
	return nil
}

func writeArchive(path, app string) (int64, []byte, error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return 0, nil, fmt.Errorf("update: create archive: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()
	sum := sha256.New()
	if err := writeTarGz(io.MultiWriter(tmp, sum), app); err != nil {
		tmp.Close()
		return 0, nil, err
	}
	if err := tmp.Close(); err != nil {
		return 0, nil, fmt.Errorf("update: close archive: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return 0, nil, fmt.Errorf("update: write archive: %w", err)
	}
	tmpName = ""
	if err := os.Chmod(path, 0o644); err != nil {
		return 0, nil, fmt.Errorf("update: set archive permissions: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, nil, fmt.Errorf("update: stat archive: %w", err)
	}
	return info.Size(), sum.Sum(nil), nil
}

func writeTarGz(w io.Writer, app string) error {
	gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("update: compress archive: %w", err)
	}
	tw := tar.NewWriter(gz)
	parent := filepath.Dir(app)
	err = filepath.WalkDir(app, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(parent, path)
		if err != nil {
			return err
		}
		link := ""
		if info.Mode()&fs.ModeSymlink != 0 {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
		}
		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if d.IsDir() {
			hdr.Name += "/"
		}
		hdr.Uid, hdr.Gid, hdr.Uname, hdr.Gname = 0, 0, "", ""
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(tw, f)
			f.Close()
			return err
		}
		return nil
	})
	if err != nil {
		tw.Close()
		gz.Close()
		return fmt.Errorf("update: archive %s: %w", app, err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("update: close archive: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("update: close archive: %w", err)
	}
	return nil
}

func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	tmpName = ""
	return os.Chmod(path, 0o644)
}

// ReleaseNotes returns the Markdown section under "## <version>" or
// "## [<version>]" in changelog.
func ReleaseNotes(changelog, version string) string {
	var notes []string
	in := false
	version = strings.TrimPrefix(version, "v")
	for _, line := range strings.Split(strings.ReplaceAll(changelog, "\r\n", "\n"), "\n") {
		if rest, ok := strings.CutPrefix(line, "## "); ok {
			if in {
				break
			}
			fields := strings.Fields(rest)
			in = len(fields) > 0 && strings.TrimPrefix(headingVersion(fields[0]), "v") == version
			continue
		}
		if in {
			notes = append(notes, line)
		}
	}
	return strings.TrimSpace(strings.Join(notes, "\n"))
}

func headingVersion(s string) string {
	if rest, ok := strings.CutPrefix(s, "["); ok {
		s, _, _ = strings.Cut(rest, "]")
	}
	return s
}

// ParsePrivateKey decodes a base64 Ed25519 private key.
func ParsePrivateKey(s string) (ed25519.PrivateKey, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil || len(b) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid update private key")
	}
	return ed25519.PrivateKey(b), nil
}

// ParsePublicKey decodes a base64 Ed25519 public key.
func ParsePublicKey(s string) (ed25519.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil || len(b) != ed25519.PublicKeySize {
		return nil, errors.New("invalid update public key")
	}
	return ed25519.PublicKey(b), nil
}

func safeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteByte('-')
		default:
			b.WriteByte('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		return "app"
	}
	return name
}
