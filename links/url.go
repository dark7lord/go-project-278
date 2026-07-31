package links

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalidURL indicates the URL is not a valid http(s) URL.
var ErrInvalidURL = errors.New("invalid url")

// normalizeURL validates a URL and prepends https:// when the scheme is missing.
func normalizeURL(raw string) (string, error) {
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("missing host")
	}

	return u.String(), nil
}
