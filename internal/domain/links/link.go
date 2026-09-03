package links

import (
	"fmt"
)

type Link struct {
	ID          int64
	OriginalURL URL
	ShortName   ShortCode
}

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
