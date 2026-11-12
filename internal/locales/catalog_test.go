package locales

import (
	"encoding/json"
	"os"
	"testing"
	"unicode"
)

// readCatalog returns the messages of one catalog file, without its @@locale entry.
func readCatalog(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var all map[string]string
	if err := json.Unmarshal(data, &all); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	delete(all, "@@locale")
	return all
}

// TestCatalogsMatch checks that the English catalog and both Chinese catalogs have
// the same keys, that no message is empty, and that English messages contain no
// Chinese characters.
func TestCatalogsMatch(t *testing.T) {
	en := readCatalog(t, "en.json")
	for _, name := range []string{"zh-Hans.json", "zh-Hant.json"} {
		zh := readCatalog(t, name)
		for key := range en {
			if _, ok := zh[key]; !ok {
				t.Errorf("%s lacks %q", name, key)
			}
		}
		for key := range zh {
			if _, ok := en[key]; !ok {
				t.Errorf("en lacks %q, which %s has", key, name)
			}
		}
		for key, message := range zh {
			if message == "" {
				t.Errorf("%s %q is empty", name, key)
			}
		}
	}
	for key, message := range en {
		if message == "" {
			t.Errorf("en %q is empty", key)
		}
		for _, r := range message {
			if unicode.Is(unicode.Han, r) {
				t.Errorf("en %q has Chinese text: %q", key, message)
				break
			}
		}
	}
}

// TestEnglishMessages checks a few messages in English, the source language.
func TestEnglishMessages(t *testing.T) {
	msgs := In("en")
	if got := msgs.FileTypeWord(); got != "Word Document" {
		t.Errorf("FileTypeWord() = %q, want %q", got, "Word Document")
	}
	if got := ErrUnknownAction("bogus"); got != "Unknown action: bogus" {
		t.Errorf("ErrUnknownAction() = %q, want the English message", got)
	}
}
