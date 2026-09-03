package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func TestMapStorageError(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		assert.ErrorIs(t, mapStorageError(pgx.ErrNoRows), application.ErrNotFound)
	})

	t.Run("short name constraint", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505", ConstraintName: shortNameConstraint}
		mapped := mapStorageError(err)
		var fieldErr *application.FieldError
		require.ErrorAs(t, mapped, &fieldErr)
		assert.Equal(t, "short_name", fieldErr.Field)
		assert.ErrorIs(t, fieldErr.Err, application.ErrShortNameAlreadyUse)
	})

	t.Run("short url constraint", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505", ConstraintName: shortURLConstraint}
		mapped := mapStorageError(err)
		var fieldErr *application.FieldError
		require.ErrorAs(t, mapped, &fieldErr)
		assert.Equal(t, "short_url", fieldErr.Field)
		assert.ErrorIs(t, fieldErr.Err, application.ErrShortURLAlreadyUse)
	})

	t.Run("other errors pass through", func(t *testing.T) {
		err := errors.New("boom")
		assert.Same(t, err, mapStorageError(err))
	})
}
