package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"code/internal/application"
)

const shortNameConstraint = "links_short_name_key"

// mapStorageError turns a missing row and a taken short name into application
// errors and passes anything else through.
func mapStorageError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == shortNameConstraint {
		return &application.FieldError{
			Field: "short_name",
			Err:   application.ErrShortNameAlreadyUse,
		}
	}

	return err
}
