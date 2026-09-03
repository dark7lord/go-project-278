// Package links contains link domain entities and value objects.
package links

import "errors"

var (
	// ErrLinkNotFound indicates that a link does not exist.
	ErrLinkNotFound = errors.New("link not found")
	// ErrInvalidURL indicates that a URL violates domain rules.
	ErrInvalidURL = errors.New("invalid url")
	// ErrInvalidShortCode indicates that a short code violates domain rules.
	ErrInvalidShortCode = errors.New("invalid short code")
	// ErrShortCodeTaken indicates that a short code is already used.
	ErrShortCodeTaken = errors.New("short code already taken")
)
