package settings

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/files"
	"github.com/hy-shine/rightclick/internal/locales"
)

const (
	compareWidth    = 720
	compareHeight   = 380
	compareMargin   = 24
	compareDirLimit = 56
)

type compareSlot struct {
	path   string
	size   int64
	sized  bool
	digest string
	busy   bool
	failed bool
	copied bool
}

type compareModel struct {
	win       *mygo.Window
	algo      files.Algorithm
	algoIndex int
	runSeq    int
	slots     [2]compareSlot
	copiedAll bool
}

var (
	compareWindow *mygo.Window
	compareApp    *compareModel
)

// OpenCompareWindow opens or updates the dedicated two-file hash comparison window.
// Call on the main thread.
func OpenCompareWindow(pathA, pathB, algoName string) error {
	if pathA == "" || pathB == "" {
		return errors.New(locales.CompareNeedsTwo())
	}
	algo, ok := windowAlgorithm(algoName)
	if !ok {
		return &unknownAlgorithmError{name: algoName}
	}
	if compareWindow == nil || compareWindow.IsDestroyed() {
		m := &compareModel{}
		m.win = mygo.NewWindow(mygo.WindowOptions{
			Title:           locales.CompareWindowTitle(),
			Width:           compareWidth,
			Height:          compareHeight,
			UseContentSize:  true,
			DisableResize:   true,
			DisableMaximize: true,
			Content:         ui.View(m.view),
		})
		compareWindow, compareApp = m.win, m
	}
	compareApp.slots = [2]compareSlot{{path: pathA}, {path: pathB}}
	compareApp.reset(algo)
	compareWindow.Show()
	compareWindow.Invalidate()
	compareApp.compute()
	return nil
}

func (m *compareModel) reset(algo files.Algorithm) {
	m.algo = algo
	for i, choice := range hashAlgorithms {
		if choice == algo {
			m.algoIndex = i
		}
	}
	m.copiedAll = false
	for i := range m.slots {
		m.slots[i].digest = ""
		m.slots[i].busy = true
		m.slots[i].failed = false
		m.slots[i].copied = false
	}
}

func (m *compareModel) setAlgorithm(algo files.Algorithm) {
	if m.algo == algo {
		return
	}
	m.reset(algo)
	m.invalidate()
	m.compute()
}

func (m *compareModel) compute() {
	m.runSeq++
	seq := m.runSeq
	algo := m.algo
	paths := [2]string{m.slots[0].path, m.slots[1].path}
	var wg sync.WaitGroup
	for idx := range 2 {
		path := paths[idx]
		wg.Go(func() {
			if info, err := os.Stat(path); err == nil {
				m.update(func() {
					if m.runSeq == seq {
						m.slots[idx].size, m.slots[idx].sized = info.Size(), true
					}
				})
			}
			digest, err := files.Hash(path, algo)
			m.update(func() {
				if m.runSeq != seq {
					return
				}
				m.slots[idx].digest = digest
				m.slots[idx].failed = err != nil
				m.slots[idx].busy = false
			})
		})
	}
	go func() {
		wg.Wait()
	}()
}

func (m *compareModel) copySlot(idx int) {
	slot := &m.slots[idx]
	if slot.digest == "" || slot.busy {
		return
	}
	if err := writeClipboard(slot.digest); err != nil {
		return
	}
	slot.copied = true
	m.invalidate()
	seq := m.runSeq
	go func() {
		pauseAfterCopy()
		m.update(func() {
			if m.runSeq == seq {
				m.slots[idx].copied = false
			}
		})
	}()
}

func (m *compareModel) copySummary() {
	if !m.ready() {
		return
	}
	text := fmt.Sprintf("%s (%s): %s\n%s (%s): %s",
		filepath.Base(m.slots[0].path), m.algo, m.slots[0].digest,
		filepath.Base(m.slots[1].path), m.algo, m.slots[1].digest,
	)
	if err := writeClipboard(text); err != nil {
		return
	}
	m.copiedAll = true
	m.invalidate()
	seq := m.runSeq
	go func() {
		pauseAfterCopy()
		m.update(func() {
			if m.runSeq == seq {
				m.copiedAll = false
			}
		})
	}()
}

func (m *compareModel) ready() bool {
	return !m.slots[0].busy && !m.slots[1].busy &&
		!m.slots[0].failed && !m.slots[1].failed &&
		m.slots[0].digest != "" && m.slots[1].digest != ""
}

func (m *compareModel) match() (decided, identical bool) {
	if !m.ready() {
		return false, false
	}
	return true, m.slots[0].digest == m.slots[1].digest
}

func (m *compareModel) invalidate() {
	if m.win != nil {
		m.win.Invalidate()
	}
}

func (m *compareModel) update(fn func()) {
	if m.win == nil {
		fn()
		return
	}
	m.win.Update(fn)
}

func (m *compareModel) view(c *ui.Context) {
	c.SetTheme(windowTheme(c))
	ui.Column(c).Fill().Padding(20, compareMargin, compareMargin).Children(func() {
		m.headerBar(c)
		ui.Spacer(c).Height(16)
		ledger(c, ui.Column(c)).Children(func() {
			m.slotCard(c, 0, locales.CompareFileFirst())
			m.slotCard(c, 1, locales.CompareFileSecond())
		})
		ui.Spacer(c).Grow(1)
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Spacer(c).Grow(1)
			if actionButtonEnabled(c, quietKind, true, nil, locales.Close(), ui.End) {
				m.win.Close()
			}
			label := locales.CompareCopyBoth()
			if m.copiedAll {
				label = locales.HashCopiedButton()
			}
			if actionButtonEnabled(c, primaryKind, m.ready(), nil, label, ui.End) {
				m.copySummary()
			}
		})
	})
}

func (m *compareModel) headerBar(c *ui.Context) {
	t := c.Theme()
	labels := make([]string, len(hashAlgorithms))
	for i, algo := range hashAlgorithms {
		labels[i] = string(algo)
	}
	ui.Row(c).AlignItems(ui.Center).Gap(12).Children(func() {
		control := ui.Segmented(c.Key("compare-algorithm"), &m.algoIndex, labels...).Label(locales.HashAlgorithm())
		if control.Changed() {
			algo := hashAlgorithms[m.algoIndex]
			go func() { m.update(func() { m.setAlgorithm(algo) }) }()
		}
		ui.Spacer(c)
		decided, identical := m.match()
		switch {
		case m.slots[0].busy || m.slots[1].busy:
			statusPill(c, t.TextMuted, nil, locales.HashComputing(), true)
		case m.slots[0].failed || m.slots[1].failed:
			statusPill(c, t.Danger, iconAlert, locales.HashFailed(), false)
		case decided && identical:
			statusPill(c, t.Success, iconCheck, locales.CompareMatch(), false)
		case decided && !identical:
			statusPill(c, t.Danger, iconAlert, locales.CompareMismatch(), false)
		}
	})
}

func statusPill(c *ui.Context, tone ui.Color, s *ui.SVG, text string, spinning bool) {
	ui.Row(c).Gap(6).Padding(5, 12).Radius(999).AlignItems(ui.Center).
		Background(tone.Alpha(0.14)).Border(1, tone.Alpha(0.32)).Children(func() {
		if spinning {
			ui.Spinner(c).FontSize(12).Label(text)
		} else if s != nil {
			glyph(c, s, 13, tone)
		}
		ui.Text(c, text).FontSize(sizeNote).FontWeight(600).TextColor(tone).SingleLine()
	})
}

func (m *compareModel) slotCard(c *ui.Context, idx int, tag string) {
	t := c.Theme()
	slot := m.slots[idx]
	ui.Column(c).Gap(8).Padding(14, 16).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Padding(2, 7).Radius(5).Background(t.Background).Children(func() {
				ui.Text(c, tag).FontSize(sizeCaption).FontWeight(600).TextColor(t.TextMuted).SingleLine()
			})
			ui.Text(c, filepath.Base(slot.path)).FontSize(14).FontWeight(600).TextColor(t.Text).SingleLine().Grow(1).MinWidth(0).Tooltip(slot.path)
			meta := shortDir(displayDir(slot.path), compareDirLimit)
			if slot.sized {
				meta += " · " + formatSize(slot.size)
			}
			ui.Text(c, meta).FontSize(sizeCaption).TextColor(t.TextMuted).SingleLine().Shrink(0)
		})
		block := ui.Row(c).Gap(8).AlignItems(ui.Center).Height(38).Padding(0, 12).Radius(8)
		switch {
		case slot.busy:
			block.Background(t.Background.Alpha(0.6)).Children(func() {
				ui.Spinner(c).FontSize(13).Label(locales.HashComputing())
				ui.Text(c, locales.HashComputing()).FontSize(sizeNote).TextColor(t.TextMuted)
			})
		case slot.failed:
			block.Background(t.Danger.Alpha(0.08)).Children(func() {
				ui.Text(c, locales.HashFailed()).FontSize(sizeNote).TextColor(t.Danger).Grow(1).SingleLine()
				if ghostButton(c.Key("retry"), locales.HashRetry(), false) {
					algo := m.algo
					go func() {
						m.update(func() {
							m.reset(algo)
							m.invalidate()
							m.compute()
						})
					}()
				}
			})
		default:
			block.OnClick(func() { m.copySlot(idx) }).Children(func() {
				ui.Row(c).Grow(1).MinWidth(0).Children(func() {
					spans := digestSpans(slot.digest, t.Text, t.Text.Alpha(0.58))
					for i := range spans {
						spans[i].Size = 12.5
					}
					ui.RichText(c, spans...).SingleLine()
				})
				badgeText := locales.HashCopy()
				badgeColor := t.TextMuted
				if slot.copied {
					badgeText = locales.HashCopiedButton()
					badgeColor = t.Success
				}
				ui.Text(c, badgeText).FontSize(sizeCaption).FontWeight(600).TextColor(badgeColor).Shrink(0)
			})
			tint(block, t.Background.Alpha(0.65), t.SurfaceHover, t.SurfacePressed)
		}
	})
}
