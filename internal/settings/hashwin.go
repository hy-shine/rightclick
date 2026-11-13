package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/files"
	"github.com/hy-shine/rightclick/internal/locales"
)

// The hash window shows one file's digest, with a control for the algorithm. Its
// size is that of its content, and it cannot be resized.
const (
	hashWidth  = 600
	hashHeight = 226
	hashMargin = 24
	// hashDirLimit is how wide the folder line may be, in characters, with a wide
	// character such as a Chinese one counted as two.
	hashDirLimit = 68
	// hashDigestSize is the monospace size of the digest. At it a 64-character
	// digest fits the block on one line. A digest is one value, so it never wraps:
	// a second line would read as a second hash.
	hashDigestSize = 13
	// hashGroup is how many characters of the digest share a tone. The tones
	// alternate, so the eye keeps its place without spaces the copy would not have.
	hashGroup = 8
)

// hashAlgorithms are the choices of the algorithm control, in the order the
// window shows them.
var hashAlgorithms = []files.Algorithm{files.SHA256, files.SHA1, files.MD5}

// writeClipboard puts text on the clipboard. Tests replace it.
var writeClipboard = func(text string) error {
	return mygo.Clipboard.Write(transfer.TextData(text))
}

// pauseAfterCopy is how long the copy button keeps saying that it copied. Tests
// replace it so they do not sleep.
var pauseAfterCopy = func() { time.Sleep(1500 * time.Millisecond) }

// hashModel is the state of the hash window: the file, the chosen algorithm and
// the digest the last run produced.
type hashModel struct {
	win       *mygo.Window
	path      string
	size      int64 // the file's size in bytes, when sized is true
	sized     bool
	algo      files.Algorithm
	algoIndex int // the control's choice, kept in step with algo
	digest    string
	busy      bool // a digest is being computed
	failed    bool
	copied    bool
}

var (
	hashWindow *mygo.Window
	hashApp    *hashModel
)

// OpenHashWindow shows the digest of path, computed with algoName. Finder actions
// supply the saved default algorithm; an empty name falls back to SHA-256.
// The window is reused, so a second request replaces what it shows.
// Call it on the main thread.
func OpenHashWindow(path, algoName string) error {
	algo, ok := windowAlgorithm(algoName)
	if !ok {
		return &unknownAlgorithmError{name: algoName}
	}
	if hashWindow == nil || hashWindow.IsDestroyed() {
		m := &hashModel{}
		m.win = mygo.NewWindow(mygo.WindowOptions{
			Title:           locales.HashWindowTitle(),
			Width:           hashWidth,
			Height:          hashHeight,
			UseContentSize:  true,
			DisableResize:   true,
			DisableMaximize: true,
			Content:         ui.View(m.view),
		})
		hashWindow, hashApp = m.win, m
	}
	hashApp.path, hashApp.sized = path, false
	hashApp.reset(algo)
	hashWindow.Show()
	hashWindow.Invalidate()
	hashApp.compute()
	return nil
}

// windowAlgorithm maps the requested algorithm to its constant, with SHA-256
// as the fallback when no algorithm is supplied.
func windowAlgorithm(name string) (files.Algorithm, bool) {
	if name == "" {
		return files.SHA256, true
	}
	return parseAlgorithm(name)
}

// parseAlgorithm maps a menu algorithm name to its constant.
func parseAlgorithm(name string) (files.Algorithm, bool) {
	for _, algo := range hashAlgorithms {
		if strings.EqualFold(string(algo), name) {
			return algo, true
		}
	}
	return "", false
}

// unknownAlgorithmError reports a name the menu does not offer.
type unknownAlgorithmError struct{ name string }

func (e *unknownAlgorithmError) Error() string { return locales.UnsupportedHash(e.name) }

// reset chooses algo and clears the last result, ready for a new run.
func (m *hashModel) reset(algo files.Algorithm) {
	m.algo = algo
	for i, choice := range hashAlgorithms {
		if choice == algo {
			m.algoIndex = i
		}
	}
	m.digest, m.busy, m.failed, m.copied = "", true, false, false
}

// compute sizes and hashes the file off the main thread and shows the result. A
// newer request replaces an older one, so a slow run never overwrites a newer
// digest.
func (m *hashModel) compute() {
	path, algo := m.path, m.algo
	go func() {
		if info, err := os.Stat(path); err == nil {
			m.update(func() {
				if m.path == path {
					m.size, m.sized = info.Size(), true
				}
			})
		}
		digest, err := files.Hash(path, algo)
		m.update(func() {
			if m.path != path || m.algo != algo {
				return
			}
			m.digest, m.failed, m.busy = digest, err != nil, false
		})
	}()
}

// setAlgorithm switches the window to another algorithm and recomputes.
func (m *hashModel) setAlgorithm(algo files.Algorithm) {
	if m.algo == algo {
		return
	}
	m.rerun(algo)
}

// rerun computes the digest again, with algo: after a switch, or after a failure.
func (m *hashModel) rerun(algo files.Algorithm) {
	m.reset(algo)
	m.invalidate()
	m.compute()
}

// copyDigest puts the digest on the clipboard and shows that it was copied.
func (m *hashModel) copyDigest() {
	if m.digest == "" || m.busy {
		return
	}
	if err := writeClipboard(m.digest); err != nil {
		return
	}
	m.copied = true
	m.invalidate()
	go func() {
		pauseAfterCopy()
		m.update(func() { m.copied = false })
	}()
}

// invalidate redraws the window, when the model has one. Tests build a model
// without a window and check the logic alone.
func (m *hashModel) invalidate() {
	if m.win != nil {
		m.win.Invalidate()
	}
}

// update runs fn on the main thread and redraws the window.
func (m *hashModel) update(fn func()) {
	if m.win == nil {
		fn()
		return
	}
	m.win.Update(fn)
}

// view draws the window: the file, the algorithm, the digest and the actions.
func (m *hashModel) view(c *ui.Context) {
	c.SetTheme(windowTheme(c))
	ui.Column(c).Padding(20, hashMargin, hashMargin).Children(func() {
		m.fileHeader(c)
		ui.Spacer(c).Height(16)
		m.algorithmControl(c)
		ui.Spacer(c).Height(8)
		m.digestBlock(c)
		ui.Spacer(c).Height(20)
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Spacer(c).Grow(1)
			if actionButtonEnabled(c, quietKind, true, nil, locales.Close(), ui.End) {
				m.win.Close()
			}
			label := locales.HashCopy()
			if m.copied {
				label = locales.HashCopiedButton()
			}
			if actionButtonEnabled(c, primaryKind, m.digest != "" && !m.busy, nil, label, ui.End) {
				m.copyDigest()
			}
		})
	})
}

// fileHeader names the file: its name, then the folder it is in and its size.
func (m *hashModel) fileHeader(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Gap(2).Children(func() {
		ui.Text(c, filepath.Base(m.path)).FontSize(14).FontWeight(600).TextColor(t.Text).SingleLine()
		ui.Row(c).Gap(5).Children(func() {
			ui.Text(c, shortDir(displayDir(m.path), hashDirLimit)).FontSize(sizeNote).TextColor(t.TextMuted).SingleLine().MinWidth(0)
			if m.sized {
				ui.Text(c, "·").FontSize(sizeNote).TextColor(t.TextMuted).Shrink(0)
				ui.Text(c, formatSize(m.size)).FontSize(sizeNote).TextColor(t.TextMuted).Shrink(0)
			}
		})
	})
}

// algorithmControl is the segmented control that picks the algorithm. Its chosen
// segment is also the heading of the digest under it.
func (m *hashModel) algorithmControl(c *ui.Context) {
	labels := make([]string, len(hashAlgorithms))
	for i, algo := range hashAlgorithms {
		labels[i] = string(algo)
	}
	ui.Row(c).Children(func() {
		control := ui.Segmented(c.Key("hash-algorithm"), &m.algoIndex, labels...).Label(locales.HashAlgorithm())
		if control.Changed() {
			algo := hashAlgorithms[m.algoIndex]
			// The redraw that runs the switch must not start inside this build.
			go func() { m.update(func() { m.setAlgorithm(algo) }) }()
		}
	})
}

// digestBlock shows the digest on one line. A click anywhere on it copies the
// digest. While the digest is computed it shows a spinner, and after a failure
// the reason with a way to try again.
func (m *hashModel) digestBlock(c *ui.Context) {
	t := c.Theme()
	block := ui.Row(c).Gap(8).AlignItems(ui.Center).Height(44).Padding(0, 16).Radius(10)
	switch {
	case m.busy:
		block.Background(t.Surface).Children(func() {
			ui.Spinner(c).FontSize(14).Label(locales.HashComputing())
			ui.Text(c, locales.HashComputing()).FontSize(sizeBody).TextColor(t.TextMuted)
		})
	case m.failed:
		block.Background(t.Surface).Children(func() {
			ui.Text(c, locales.HashFailed()).FontSize(sizeBody).TextColor(t.Danger).Grow(1).SingleLine()
			if ghostButton(c, locales.HashRetry(), false) {
				algo := m.algo
				go func() { m.update(func() { m.rerun(algo) }) }()
			}
		})
	default:
		block.OnClick(m.copyDigest).Children(func() {
			ui.RichText(c, digestSpans(m.digest, t.Text, t.Text.Alpha(0.58))...).SingleLine()
		})
		tint(block, t.Surface, t.SurfaceHover, t.SurfacePressed)
	}
}

// digestSpans splits a digest into groups of hashGroup characters that alternate
// between two tones, in the monospace font.
func digestSpans(digest string, first, second ui.Color) []ui.Span {
	var spans []ui.Span
	for start := 0; start < len(digest); start += hashGroup {
		end := min(start+hashGroup, len(digest))
		color := first
		if (start/hashGroup)%2 == 1 {
			color = second
		}
		spans = append(spans, ui.Span{
			Text:  digest[start:end],
			Font:  "SF Mono, Menlo, monospace",
			Size:  hashDigestSize,
			Color: color,
		})
	}
	return spans
}

// displayDir returns the folder a file is in, with the home folder written as ~.
func displayDir(path string) string {
	dir := filepath.Dir(path)
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return dir
	}
	if dir == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(dir, home+string(filepath.Separator)); ok {
		return "~" + string(filepath.Separator) + rest
	}
	return dir
}

// shortDir keeps the end of a folder path wider than limit, the part nearest the
// file, with an ellipsis for the folders left out. It cuts between folders, so a
// name is never cut inside a character.
func shortDir(dir string, limit int) string {
	if textWidth(dir) <= limit {
		return dir
	}
	separator := string(filepath.Separator)
	parts := strings.Split(dir, separator)
	kept := parts[len(parts)-1]
	for i := len(parts) - 2; i > 0; i-- {
		next := parts[i] + separator + kept
		if textWidth(next)+2 > limit {
			break
		}
		kept = next
	}
	return "…" + separator + kept
}

// textWidth estimates how wide s is, in characters of the Latin alphabet: a
// character outside it, such as a Chinese one, counts as two.
func textWidth(s string) int {
	width := 0
	for _, r := range s {
		width++
		if r > unicode.MaxLatin1 {
			width++
		}
	}
	return width
}

// formatSize writes a size in bytes the way Finder does, in decimal units.
func formatSize(bytes int64) string {
	switch {
	case bytes < 1000:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1000*1000:
		return fmt.Sprintf("%.0f KB", float64(bytes)/1e3)
	case bytes < 1000*1000*1000:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1e6)
	default:
		return fmt.Sprintf("%.2f GB", float64(bytes)/1e9)
	}
}
