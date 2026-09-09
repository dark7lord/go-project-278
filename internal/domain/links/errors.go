// Package links contains link domain entities and value objects.
package links

import "errors"

var (
	// ErrInvalidURL indicates that a URL violates domain rules.
	ErrInvalidURL = errors.New("invalid url")
	// ErrInvalidShortCode indicates that a short code violates domain rules.
	ErrInvalidShortCode = errors.New("invalid short code")
)
