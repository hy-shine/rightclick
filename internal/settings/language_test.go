package settings

import (
	"slices"
	"testing"
	"unicode"

	"github.com/egoist/mygo/ui"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

// TestLanguagePickerSavesTheChoice checks that choosing a language in the dropdown
// saves it to the settings file and asks the app to show that language. The dropdown
// shows the current choice on its button, so each pick opens it first.
func TestLanguagePickerSavesTheChoice(t *testing.T) {
	store := testStore(t)
	var shown []string
	e := testEnv()
	e.setLocale = func(tag string) { shown = append(shown, tag) }
	m, err := newApp(store, e)
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, 760, 560)
	if err := tt.Click("General"); err != nil {
		t.Fatal(err)
	}
	// The button shows Same as System, the saved choice. Open the dropdown and pick English.
	if err := tt.Click("Same as System"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("English"); err != nil {
		t.Fatal(err)
	}
	// The button now shows English. Open the dropdown and pick the system choice.
	if err := tt.Click("English"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Same as System"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"en", ""}; !slices.Equal(shown, want) {
		t.Fatalf("the app was asked to show %q, want %q", shown, want)
	}
	saved, err := config.Load(store.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Language != "" {
		t.Errorf("saved language = %q, want the system choice", saved.Language)
	}
}

// TestSettingsWindowShowsEnglish checks that the settings window uses the English
// catalog when the app shows English, and that no Chinese text reaches the window.
func TestSettingsWindowShowsEnglish(t *testing.T) {
	useLocale(t, "en")
	m, err := newApp(testStore(t), testEnv())
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(m.view, 760, 560)
	if err := tt.Click("General"); err != nil {
		t.Fatal(err)
	}
	texts := tt.Texts()
	if !slices.Contains(texts, "Open at Login") {
		t.Fatalf("the English general pane is not shown; texts %q", texts)
	}
	for _, text := range texts {
		for _, r := range text {
			if unicode.Is(unicode.Han, r) {
				t.Errorf("Chinese text in the English window: %q", text)
				break
			}
		}
	}
}

// TestLanguageLabelsMatchTheTags checks that the dropdown's label at each position
// names the tag at the same position in config.Languages.
func TestLanguageLabelsMatchTheTags(t *testing.T) {
	labels := languageLabels()
	if len(labels) != len(config.Languages) {
		t.Fatalf("%d labels for %d tags", len(labels), len(config.Languages))
	}
	want := map[string]string{
		"":        locales.LanguageSystem(),
		"zh-Hans": locales.LanguageChinese(),
		"zh-Hant": locales.LanguageTraditionalChinese(),
		"en":      locales.LanguageEnglish(),
	}
	for i, tag := range config.Languages {
		if labels[i] != want[tag] {
			t.Errorf("label for %q = %q, want %q", tag, labels[i], want[tag])
		}
	}
}

// TestLanguageChangeRenamesTheWindow checks that the main window takes the name of
// the language the app now shows. The window takes its title when it is created, so
// a change of language has to rename it again.
func TestLanguageChangeRenamesTheWindow(t *testing.T) {
	var titles []string
	e := testEnv()
	e.setLocale = func(tag string) { useLocale(t, tag) }
	e.setWindowTitle = func(title string) { titles = append(titles, title) }
	m, err := newApp(testStore(t), e)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"zh-Hans", "zh-Hant", "en"} {
		m.cfg.Language = tag
		m.applyLanguage()
	}
	if want := []string{"右键助手", "右鍵助手", "RightClick"}; !slices.Equal(titles, want) {
		t.Fatalf("the window was renamed to %q, want %q", titles, want)
	}
}
