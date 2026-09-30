package application

import "errors"

var (
	// ErrNotFound reports a link that does not exist.
	ErrNotFound = errors.New("link not found")
	// ErrShortNameAlreadyUse reports a short name another link holds.
	ErrShortNameAlreadyUse = errors.New("short name already in use")
	// ErrShortCodeGenerationFailed reports that every generated short name collided.
	ErrShortCodeGenerationFailed = errors.New("short code generation failed")
)

// FieldError associates an application error with a request field.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string { return e.Err.Error() }

func (e *FieldError) Unwrap() error { return e.Err }
