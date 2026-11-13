package settings

import (
	"math"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/locales"
)

func TestFinderStatusCheckingDoesNotClaimSuccess(t *testing.T) {
	for _, language := range []string{"zh-Hans", "zh-Hant", "en"} {
		t.Run(language, func(t *testing.T) {
			useLocale(t, language)
			model := readyApp(t)
			selectGeneral(model)
			model.finder = finderChecking
			tester := ui.NewTester(model.view, mainWidth, mainHeight)
			texts := tester.Texts()
			if !slices.Contains(texts, locales.CheckingItem()) {
				t.Fatal("checking state is missing")
			}
			if slices.Contains(texts, locales.CheckExtension()) || slices.Contains(texts, locales.ReadyInFinder()) {
				t.Fatalf("checking state claims success: %q", texts)
			}
			model.finder = finderOn
			tester.Frame()
			if !slices.Contains(tester.Texts(), locales.ReadyInFinder()) {
				t.Fatal("completed check does not show readiness")
			}
			model.finder = finderOff
			tester.Frame()
			if slices.Contains(tester.Texts(), locales.CheckExtension()) {
				t.Fatal("disabled extension still claims success")
			}
			if _, found := tester.Find(locales.GoEnable()); !found {
				t.Fatal("disabled extension has no enable action")
			}
		})
	}
}

func TestWindowThemeFollowsPreferenceChanges(t *testing.T) {
	for _, dark := range []bool{false, true} {
		model := readyApp(t)
		model.selectPane(paneTemplates)
		var theme ui.Theme
		tester := ui.NewTester(func(context *ui.Context) {
			context.DisableAnimations()
			model.view(context)
			theme = *context.Theme()
		}, mainWidth, mainHeight)
		tester.SetDark(dark)
		for _, accent := range []ui.Color{ui.Transparent, ui.Hex("#af52de"), ui.Hex("#ffcc00")} {
			tester.SetPreferences(ui.Preferences{Accent: accent})
			if accent.A > 0 && theme.Accent != accent {
				t.Fatalf("dark=%v: accent = %+v, want %+v", dark, theme.Accent, accent)
			}
			bounds, found := tester.Find(locales.EnableLabel(locales.FileTypeWord()))
			if !found {
				t.Fatal("enabled file type switch is missing")
			}
			track := tester.Image().RGBAAt(int(bounds.X+5), int(bounds.Y+bounds.H/2))
			if track.R != theme.Accent.R || track.G != theme.Accent.G || track.B != theme.Accent.B {
				t.Errorf("selected-row switch lost system accent: got %+v, want %+v", track, theme.Accent)
			}
			selected, _ := selectionColors(&theme)
			secondary := theme.TextMuted.Mix(theme.Text, 0.3)
			if ratio := contrastRatio(secondary, selected); ratio < 4.5 {
				t.Errorf("dark=%v accent=%+v: selected secondary contrast = %.2f, want >= 4.5", dark, accent, ratio)
			}
			regularBorder, regularText := theme.Border, theme.TextMuted
			tester.SetPreferences(ui.Preferences{Accent: accent, HighContrast: true})
			if theme.Border == regularBorder || theme.TextMuted == regularText || theme.Focus.A != 255 {
				t.Errorf("dark=%v: increased contrast did not strengthen borders, text and focus", dark)
			}
		}
	}
}

func contrastRatio(first, second ui.Color) float64 {
	luminance := func(color ui.Color) float64 {
		linear := func(channel uint8) float64 {
			value := float64(channel) / 255
			if value <= 0.04045 {
				return value / 12.92
			}
			return math.Pow((value+0.055)/1.055, 2.4)
		}
		return 0.2126*linear(color.R) + 0.7152*linear(color.G) + 0.0722*linear(color.B)
	}
	firstLight, secondLight := luminance(first), luminance(second)
	return (max(firstLight, secondLight) + 0.05) / (min(firstLight, secondLight) + 0.05)
}

func TestCreationColumnsStayBalanced(t *testing.T) {
	for _, language := range []string{"zh-Hans", "zh-Hant", "en"} {
		t.Run(language, func(t *testing.T) {
			useLocale(t, language)
			for _, width := range []int{minWidth, mainWidth, 1200} {
				model := readyApp(t)
				model.selectPane(paneTemplates)
				tester := ui.NewTester(model.view, width, mainHeight)
				table, found := tester.Find(locales.FileTypesList())
				if !found {
					t.Fatal("file types table is missing")
				}
				format, found := tester.Find(locales.ColFormat())
				if !found {
					t.Fatal("format column is missing")
				}
				if offset := format.X - table.X; offset > 280 {
					t.Errorf("%dpt: format column starts %.1fpt into the table, want at most 280", width, offset)
				}
				edit, found := tester.Find(locales.EditItem(locales.FileTypeWord()))
				if !found || edit.W < 32 || edit.X+edit.W > table.X+table.W {
					t.Errorf("%dpt: edit target is clipped or undersized: %+v in %+v", width, edit, table)
				}
			}
		})
	}
}
