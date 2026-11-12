package main

import (
	"errors"
	"fmt"
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
		return "", nil, nil, fmt.Errorf("不认识的网址：%s", raw)
	}
	if _, ok := actionHandlers()[u.Host]; !ok {
		return "", nil, nil, fmt.Errorf("不认识的操作：%s", u.Host)
	}
	params = u.Query()
	paths = params["path"]
	if len(paths) == 0 {
		return "", nil, nil, errors.New("网址里没有路径")
	}
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			return "", nil, nil, fmt.Errorf("路径必须是绝对路径：%s", p)
		}
	}
	return u.Host, params, paths, nil
}
