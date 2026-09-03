package links

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLink(t *testing.T) {
	link, err := NewLink("example.com", "abc123")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", link.OriginalURL.String())
	assert.Equal(t, "abc123", link.ShortName.String())
	assert.Equal(t, int64(0), link.ID)
}

func TestNewLinkInvalidURL(t *testing.T) {
	_, err := NewLink("", "abc123")
	assert.ErrorIs(t, err, ErrInvalidURL)
}

func TestNewLinkInvalidShortCode(t *testing.T) {
	_, err := NewLink("https://example.com", "")
	assert.ErrorIs(t, err, ErrInvalidShortCode)
}

func TestNewLinkRejectsShortCodeLengthViolation(t *testing.T) {
	_, err := NewLink("https://example.com", "x")
	assert.ErrorIs(t, err, ErrInvalidShortCode)

	longCode := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err = NewLink("https://example.com", longCode)
	assert.ErrorIs(t, err, ErrInvalidShortCode)
}

func TestNewLinkRejectsUnsupportedScheme(t *testing.T) {
	_, err := NewLink("ftp://example.com", "valid-code")
	assert.ErrorIs(t, err, ErrInvalidURL)
}
