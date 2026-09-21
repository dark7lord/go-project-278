package application

import "errors"

var (
	// ErrNotFound indicates that the requested entity does not exist.
	ErrNotFound = errors.New("link not found")
	// ErrShortNameAlreadyUse indicates that a short name is already taken.
	ErrShortNameAlreadyUse = errors.New("short name already in use")
)

// FieldError associates an application error with a request field.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string { return e.Err.Error() }

func (e *FieldError) Unwrap() error { return e.Err }
