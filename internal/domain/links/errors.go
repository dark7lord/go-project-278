package links

import "errors"

var (
	ErrLinkNotFound     = errors.New("link not found")
	ErrInvalidURL       = errors.New("invalid url")
	ErrInvalidShortCode = errors.New("invalid short code")
	ErrShortCodeTaken   = errors.New("short code already taken")
)
