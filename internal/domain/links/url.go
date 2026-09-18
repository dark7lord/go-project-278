package links

import (
	"fmt"
	"net/url"
	"strings"
)

// URL is a validated HTTP or HTTPS URL.
type URL struct {
	value string
}

// NewURL creates a URL after normalizing and validating its value.
func NewURL(raw string) (URL, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return URL{}, fmt.Errorf("%w: empty url", ErrInvalidURL)
	}

	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	u, err := url.Parse(candidate)
	if err != nil {
		return URL{}, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return URL{}, fmt.Errorf("%w: unsupported scheme %q", ErrInvalidURL, u.Scheme)
	}
	if u.Host == "" {
		return URL{}, fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	return URL{value: u.String()}, nil
}

func (u URL) String() string {
	return u.value
}
