// Package update checks GitHub for a newer release of RightClick. It reads the
// newest release and compares version numbers. It never downloads or installs
// anything: the user opens the release page to get a new version.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository whose releases are checked.
const Repo = "hy-shine/rightclick"

// apiURL is the address of the newest release. Tests replace it.
var apiURL = "https://api.github.com/repos/" + Repo + "/releases/latest"

// timeout bounds one check, so a slow network cannot keep the pane waiting.
const timeout = 10 * time.Second

// Release is the newest published release of the app.
type Release struct {
	Version string // the tag without its leading "v"
	URL     string // the page where the release can be downloaded
}

// Latest reads the newest release from GitHub. It fails when the network is
// unavailable, when the repository has no published release, or when the answer
// has no tag.
func Latest() (Release, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "RightClick")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("github answered %s", resp.Status)
	}
	var body struct {
		Tag string `json:"tag_name"`
		URL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return Release{}, err
	}
	if body.Tag == "" {
		return Release{}, errors.New("the release has no tag")
	}
	return Release{Version: strings.TrimPrefix(body.Tag, "v"), URL: body.URL}, nil
}

// Newer reports whether version latest is higher than current. Both are dotted
// numbers such as "0.2.1", with an optional leading "v". Missing parts count as
// zero, so "0.2" equals "0.2.0". A part that is not a plain number is an error,
// so a pre-release tag is never taken for a newer version.
func Newer(latest, current string) (bool, error) {
	a, err := parts(latest)
	if err != nil {
		return false, err
	}
	b, err := parts(current)
	if err != nil {
		return false, err
	}
	for i := range max(len(a), len(b)) {
		if x, y := part(a, i), part(b, i); x != y {
			return x > y, nil
		}
	}
	return false, nil
}

// parts splits a version into its numbers.
func parts(version string) ([]int, error) {
	var numbers []int
	for _, field := range strings.Split(strings.TrimPrefix(version, "v"), ".") {
		n, err := strconv.Atoi(field)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("not a version number: %q", version)
		}
		numbers = append(numbers, n)
	}
	return numbers, nil
}

// part returns the i-th number of a version, or zero past its end.
func part(numbers []int, i int) int {
	if i < len(numbers) {
		return numbers[i]
	}
	return 0
}
