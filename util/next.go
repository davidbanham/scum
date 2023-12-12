package util

import "net/url"

func NextFlow(defaultURL string, form url.Values) string {
	ret, _ := url.Parse(defaultURL)
	next := form.Get("next")
	if next != "" {
		parsed, _ := url.Parse(next)
		ret.Path = parsed.Path
		for k, v := range parsed.Query() {
			q := ret.Query()
			q[k] = v
			ret.RawQuery = q.Encode()
		}
	}
	if form.Get("next_fragment") != "" {
		ret.Fragment = form.Get("next_fragment")
	}
	return ret.String()
}
