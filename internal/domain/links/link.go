package links

import (
	"fmt"
)

// Link is a shortened link with validated URL and short code values.
type Link struct {
	ID          int64
	OriginalURL URL
	ShortName   ShortCode
}

// NewLink creates a link after validating its URL and short code.
func NewLink(rawURL, code string) (Link, error) {
	parsedURL, err := NewURL(rawURL)
	if err != nil {
		return Link{}, err
	}

	shortCode, err := NewShortCode(code)
	if err != nil {
		return Link{}, err
	}

	return Link{
		ID:          0,
		OriginalURL: parsedURL,
		ShortName:   shortCode,
	}, nil
}

func (l Link) String() string {
	return fmt.Sprintf("link{id=%d, original_url=%s, short_name=%s}", l.ID, l.OriginalURL.String(), l.ShortName.String())
}
