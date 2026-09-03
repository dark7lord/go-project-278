package links

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func newShortCodeValidator() *validator.Validate {
	v := validator.New()
	if err := v.RegisterValidation("safe_short_code", func(fl validator.FieldLevel) bool {
		code := strings.TrimSpace(fl.Field().String())
		if code == "" || len(code) < 3 || len(code) > 32 {
			return false
		}

		return regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(code)
	}); err != nil {
		panic(err)
	}

	return v
}

type ShortCode struct {
	value string
}

func NewShortCode(raw string) (ShortCode, error) {
	code := strings.TrimSpace(raw)
	v := newShortCodeValidator()
	if err := v.Var(code, "required,safe_short_code"); err != nil {
		return ShortCode{}, fmt.Errorf("%w: %v", ErrInvalidShortCode, err)
	}

	return ShortCode{value: code}, nil
}

func (s ShortCode) String() string {
	return s.value
}
