// Package links holds the rules a link's URL and short name follow.
package links

import "errors"

var (
	// ErrInvalidURL reports a URL that is not an http(s) link.
	ErrInvalidURL = errors.New("invalid url")
	// ErrInvalidShortCode reports a short name outside [a-zA-Z0-9-]{3,32}.
	ErrInvalidShortCode = errors.New("invalid short code")
)
