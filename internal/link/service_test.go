package link

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUniqueViolation(t *testing.T) {
	t.Run("short_name constraint", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505", ConstraintName: shortNameConstraint}
		got := mapUniqueViolation(err)
		var fe *FieldError
		require.ErrorAs(t, got, &fe)
		assert.Equal(t, fieldShortName, fe.Field)
		assert.ErrorIs(t, fe.Err, ErrShortNameAlreadyUse)
	})

	t.Run("short_url constraint", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505", ConstraintName: shortURLConstraint}
		got := mapUniqueViolation(err)
		var fe *FieldError
		require.ErrorAs(t, got, &fe)
		assert.Equal(t, fieldShortURL, fe.Field)
	})

	t.Run("non-unique pg error passes through", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.Same(t, err, mapUniqueViolation(err))
	})

	t.Run("non-pg error passes through", func(t *testing.T) {
		err := errors.New("boom")
		assert.Same(t, err, mapUniqueViolation(err))
	})
}
