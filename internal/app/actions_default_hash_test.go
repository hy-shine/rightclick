package app

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hy-shine/rightclick/internal/config"
	"github.com/hy-shine/rightclick/internal/locales"
)

func TestHashActionUsesDefaultOrExplicitAlgorithm(t *testing.T) {
	previous := hashWindow
	t.Cleanup(func() { hashWindow = previous })
	for _, sample := range []struct {
		name      string
		defaulted string
		explicit  string
		want      string
	}{
		{"SHA-256 default", "SHA-256", "", "SHA-256"},
		{"SHA-1 default", "SHA-1", "", "SHA-1"},
		{"MD5 default", "MD5", "", "MD5"},
		{"explicit wins", "MD5", "SHA-256", "SHA-256"},
		{"missing default", "", "", "SHA-256"},
		{"invalid default", "SHA-512", "", "SHA-256"},
		{"invalid explicit preserved", "MD5", "SHA-512", "SHA-512"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			var gotAlgorithm string
			var gotPaths []string
			hashWindow = func(algorithm string, paths []string) error {
				gotAlgorithm, gotPaths = algorithm, paths
				return nil
			}
			path := filepath.Join(t.TempDir(), "file.txt")
			params := url.Values{"algo": {sample.explicit}}
			cfg := config.Config{DefaultHashAlgorithm: sample.defaulted}
			if err := runAction("hash", params, []string{path}, cfg); err != nil {
				t.Fatal(err)
			}
			if gotAlgorithm != sample.want || len(gotPaths) != 1 || gotPaths[0] != path {
				t.Fatalf("hash window received %q, %q; want %q, %q", gotAlgorithm, gotPaths, sample.want, []string{path})
			}
			if params.Get("algo") != sample.explicit {
				t.Fatal("hash action changed the link parameters")
			}
		})
	}
}

func TestComparisonActionHashesWithSavedDefault(t *testing.T) {
	previous := compareWindow
	t.Cleanup(func() { compareWindow = previous })
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "first.txt"), filepath.Join(dir, "second.txt")}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, sample := range []struct {
		name      string
		defaulted string
		explicit  string
		want      string
	}{
		{"SHA-256", "SHA-256", "", "SHA-256"},
		{"SHA-1", "SHA-1", "", "SHA-1"},
		{"MD5", "MD5", "", "MD5"},
		{"explicit wins", "MD5", "SHA-256", "SHA-256"},
		{"missing default", "", "", "SHA-256"},
		{"invalid default", "invalid", "", "SHA-256"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			var gotAlgorithm string
			var gotPaths []string
			compareWindow = func(algorithm string, selected []string) error {
				gotAlgorithm, gotPaths = algorithm, selected
				return nil
			}
			params := url.Values{"algo": {sample.explicit}}
			cfg := config.Config{DefaultHashAlgorithm: sample.defaulted}
			if err := runAction("compare-hash", params, paths, cfg); err != nil {
				t.Fatal(err)
			}
			if gotAlgorithm != sample.want || len(gotPaths) != 2 || gotPaths[0] != paths[0] || gotPaths[1] != paths[1] {
				t.Fatalf("compare window received %q, %q; want %q, %q", gotAlgorithm, gotPaths, sample.want, paths)
			}
			title, body, err := compareFiles(gotAlgorithm, paths)
			if err != nil {
				t.Fatal(err)
			}
			if title != locales.HashCompareTitle(sample.want, locales.VerdictSame()) || strings.Count(body, helloDigests[sample.want]) != 2 {
				t.Fatalf("comparison result = %q, %q; want algorithm %q and both corresponding digests", title, body, sample.want)
			}
		})
	}
	if _, _, err := compareFiles("SHA-512", paths); err == nil {
		t.Fatal("unsupported explicit comparison algorithm was accepted")
	}
}
