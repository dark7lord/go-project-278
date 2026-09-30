package links

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeURL validates an HTTP or HTTPS URL and returns it normalized: trimmed,
// with https:// added when the scheme is missing.
func NormalizeURL(raw string) (string, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return "", fmt.Errorf("%w: empty url", ErrInvalidURL)
	}

	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	u, err := url.Parse(candidate)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: unsupported scheme %q", ErrInvalidURL, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	return u.String(), nil
}
