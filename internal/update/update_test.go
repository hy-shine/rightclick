package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// serve points apiURL at a test server that answers with status and body.
func serve(t *testing.T, status int, body string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	old := apiURL
	apiURL = server.URL
	t.Cleanup(func() { apiURL = old })
}

func TestNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"0.1.1", "0.1.0", true},
		{"v0.2.0", "0.1.9", true},
		{"0.10.0", "0.9.0", true}, // numbers, not text
		{"0.1.0", "0.1.0", false},
		{"0.1", "0.1.0", false},
		{"0.1.0", "0.2.0", false},
	}
	for _, tt := range tests {
		got, err := Newer(tt.latest, tt.current)
		if err != nil {
			t.Errorf("Newer(%q, %q) error = %v", tt.latest, tt.current, err)
		}
		if got != tt.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
	if _, err := Newer("0.2.0-beta", "0.1.0"); err == nil {
		t.Error("a pre-release tag was compared as if it were a plain version")
	}
}

func TestLatestReadsTheNewestRelease(t *testing.T) {
	serve(t, http.StatusOK, `{"tag_name":"v0.2.0","html_url":"https://example.com/release"}`)
	release, err := Latest()
	if err != nil {
		t.Fatal(err)
	}
	if release.Version != "0.2.0" || release.URL != "https://example.com/release" {
		t.Errorf("Latest() = %+v", release)
	}
}

func TestLatestFailsWhenNoReleaseIsPublished(t *testing.T) {
	serve(t, http.StatusNotFound, `{"message":"Not Found"}`)
	if _, err := Latest(); err == nil {
		t.Error("a 404 answer was accepted as a release")
	}
}

func TestLatestRefusesAnAnswerWithoutATag(t *testing.T) {
	serve(t, http.StatusOK, `{"html_url":"https://example.com/release"}`)
	if _, err := Latest(); err == nil {
		t.Error("an answer without a tag was accepted")
	}
}
