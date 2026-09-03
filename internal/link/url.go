package link

import (
	domainlinks "code/internal/domain/links"
)

// ErrInvalidURL indicates the URL is not a valid http(s) URL.
var ErrInvalidURL = domainlinks.ErrInvalidURL

// normalizeURL validates a URL and prepends https:// when the scheme is missing.
func normalizeURL(raw string) (string, error) {
	u, err := domainlinks.NewURL(raw)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
