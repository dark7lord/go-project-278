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

func TestPageRange(t *testing.T) {
	tests := []struct {
		name       string
		start      int64
		end        int64
		wantLimit  int64
		wantOffset int64
	}{
		{name: "full page", start: 5, end: 9, wantLimit: 5, wantOffset: 5},
		{name: "single item", start: 3, end: 3, wantLimit: 1, wantOffset: 3},
		{name: "maximum page", start: 0, end: 999, wantLimit: 1000, wantOffset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset := pageRange(tt.start, tt.end)
			assert.Equal(t, tt.wantLimit, limit)
			assert.Equal(t, tt.wantOffset, offset)
		})
	}
}

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

	t.Run("other errors pass through", func(t *testing.T) {
		err := errors.New("boom")
		assert.Same(t, err, mapStorageError(err))
	})
}
