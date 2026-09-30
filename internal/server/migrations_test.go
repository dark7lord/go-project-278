package server

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/db/migrations"
)

// freshDatabase creates an empty database in the test container, so migrations
// can go up and down without touching the one the other tests share.
func freshDatabase(t *testing.T, td *testDB) *pgxpool.Pool {
	t.Helper()

	name := fmt.Sprintf("migrations_%d", time.Now().UnixNano())
	_, err := td.conn.Exec(t.Context(), "CREATE DATABASE "+name)
	require.NoError(t, err)

	config := td.conn.Config().Copy()
	config.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	require.NoError(t, err)

	// Not t.Context(): it is already canceled when cleanups run
	t.Cleanup(func() {
		pool.Close()
		_, _ = td.conn.Exec(context.Background(), "DROP DATABASE "+name+" WITH (FORCE)")
	})

	return pool
}

func TestMigrationsRoundTrip(t *testing.T) {
	td := setupTestDB(t)
	pool := freshDatabase(t, td)
	ctx := t.Context()

	status := func() string {
		var out bytes.Buffer
		require.NoError(t, migrations.Status(ctx, pool, &out))

		return out.String()
	}

	assert.Contains(t, status(), "pending")

	require.NoError(t, migrations.Up(ctx, pool))
	assert.Contains(t, status(), "applied")
	require.NoError(t, migrations.Up(ctx, pool), "nothing left to apply is not an error")

	require.NoError(t, migrations.Redo(ctx, pool))
	assert.Contains(t, status(), "applied")

	require.NoError(t, migrations.Down(ctx, pool))
	assert.Contains(t, status(), "pending")
}
