package links

import (
	"fmt"
	"regexp"
	"strings"
)

var shortCodeRe = regexp.MustCompile(`^[a-zA-Z0-9-]{3,32}$`)

// NormalizeShortCode validates a URL-safe link identifier and returns it trimmed.
func NormalizeShortCode(raw string) (string, error) {
	code := strings.TrimSpace(raw)
	if !shortCodeRe.MatchString(code) {
		return "", fmt.Errorf("%w: %q", ErrInvalidShortCode, code)
	}

	return code, nil
}
