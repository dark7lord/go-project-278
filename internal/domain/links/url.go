package links

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
)

var urlValidator = validator.New()

type urlInput struct {
	Value string `validate:"required,normalized_url"`
}

func init() {
	_ = urlValidator.RegisterValidation("normalized_url", func(fl validator.FieldLevel) bool {
		raw := strings.TrimSpace(fl.Field().String())
		if raw == "" {
			return false
		}

		candidate := raw
		if !strings.Contains(candidate, "://") {
			candidate = "https://" + candidate
		}

		u, err := url.Parse(candidate)
		if err != nil {
			return false
		}
		return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
	})
}

type URL struct {
	value string
}

func NewURL(raw string) (URL, error) {
	input := urlInput{Value: raw}
	if err := urlValidator.Struct(input); err != nil {
		return URL{}, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	candidate := raw
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	u, err := url.Parse(candidate)
	if err != nil {
		return URL{}, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return URL{}, fmt.Errorf("%w: unsupported scheme %q", ErrInvalidURL, u.Scheme)
	}
	if u.Host == "" {
		return URL{}, fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	return URL{value: u.String()}, nil
}

func (u URL) String() string {
	return u.value
}
