package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/db/migrations"
)

// freshDatabase creates an empty database next to the shared one, so migrations
// can go down and up without touching it.
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

// TestMigrationsRoundTrip walks every version down to an empty schema and
// back up, so each Down runs, not only the latest one.
func TestMigrationsRoundTrip(t *testing.T) {
	td := setupTestDB(t)
	pool := freshDatabase(t, td)
	ctx := t.Context()

	files, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)

	require.NoError(t, migrations.Up(ctx, pool))
	require.NoError(t, migrations.Up(ctx, pool), "nothing left to apply is not an error")
	require.NoError(t, migrations.Redo(ctx, pool))

	for range files {
		require.NoError(t, migrations.Down(ctx, pool))
	}
	assert.Empty(t, userTables(t, pool), "every Down must drop what its Up created")

	require.NoError(t, migrations.Up(ctx, pool), "the empty schema applies again")

	var status bytes.Buffer
	require.NoError(t, migrations.Status(ctx, pool, &status))
	assert.NotContains(t, status.String(), "pending")
}

// userTables lists the tables in the public schema besides goose's own.
func userTables(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()

	rows, err := pool.Query(t.Context(), `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'goose_db_version'`)
	require.NoError(t, err)

	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	return tables
}
