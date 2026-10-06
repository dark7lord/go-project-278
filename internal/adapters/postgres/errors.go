package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"code/internal/application"
)

const (
	// uniqueViolation is the PostgreSQL error code for a broken UNIQUE constraint.
	uniqueViolation     = "23505"
	shortNameConstraint = "links_short_name_key"
)

// mapStorageError turns a missing row and a taken short name into application
// errors and passes anything else through.
func mapStorageError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrNotFound
	}

	pgErr, isPgErr := errors.AsType[*pgconn.PgError](err)
	isShortNameTaken := isPgErr &&
		pgErr.Code == uniqueViolation &&
		pgErr.ConstraintName == shortNameConstraint
	if isShortNameTaken {
		return &application.FieldError{
			Field: "short_name",
			Err:   application.ErrShortNameAlreadyUse,
		}
	}

	return err
}
