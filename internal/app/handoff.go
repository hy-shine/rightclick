package app

import (
	"errors"
	"github.com/hy-shine/rightclick/internal/locales"
	"net/url"
	"path/filepath"
)

// parseURL splits a rightclick://<action>?<params> URL into its action, its query
// parameters, and the values of the repeated path parameter, in URL order. The
// Finder Sync extension percent-encodes every value, so a + or ; in a path
// arrives intact. The action must be a known one, and every path must be absolute,
// because the app's working directory is not the user's folder.
func parseURL(raw string) (action string, params url.Values, paths []string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", nil, nil, err
	}
	if u.Scheme != "rightclick" {
		return "", nil, nil, errors.New(locales.ErrUnknownURL(raw))
	}
	if _, ok := actionHandlers()[u.Host]; !ok {
		return "", nil, nil, errors.New(locales.ErrUnknownAction(u.Host))
	}
	params = u.Query()
	paths = params["path"]
	if len(paths) == 0 {
		return "", nil, nil, errors.New(locales.ErrURLNoPath())
	}
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			return "", nil, nil, errors.New(locales.ErrPathNotAbsolute(p))
		}
	}
	return u.Host, params, paths, nil
}
