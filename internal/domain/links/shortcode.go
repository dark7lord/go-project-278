package links

import (
	"fmt"
	"regexp"
	"strings"
)

var shortCodeRe = regexp.MustCompile(`^[a-zA-Z0-9-]{3,32}$`)

// ShortCode is a validated, URL-safe link identifier.
type ShortCode struct {
	value string
}

// NewShortCode creates a short code after validating and trimming its value.
func NewShortCode(raw string) (ShortCode, error) {
	code := strings.TrimSpace(raw)
	if !shortCodeRe.MatchString(code) {
		return ShortCode{}, fmt.Errorf("%w: %q", ErrInvalidShortCode, code)
	}

	return ShortCode{value: code}, nil
}

func (s ShortCode) String() string {
	return s.value
}
