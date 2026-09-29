package migrations

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// Up applies every pending migration. The app calls it on start.
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	provider, err := newProvider(pool)
	if err != nil {
		return err
	}

	results, err := provider.Up(ctx)
	logResults(results...)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	slog.Info("migrations applied", "count", len(results))

	return nil
}

// Down rolls back the latest migration.
func Down(ctx context.Context, pool *pgxpool.Pool) error {
	provider, err := newProvider(pool)
	if err != nil {
		return err
	}

	result, err := provider.Down(ctx)
	logResults(result)
	if err != nil {
		return fmt.Errorf("roll back migration: %w", err)
	}

	return nil
}

// Redo rolls back the latest migration and applies it again.
func Redo(ctx context.Context, pool *pgxpool.Pool) error {
	provider, err := newProvider(pool)
	if err != nil {
		return err
	}

	down, err := provider.Down(ctx)
	logResults(down)
	if err != nil {
		return fmt.Errorf("roll back migration: %w", err)
	}

	up, err := provider.UpByOne(ctx)
	logResults(up)
	if err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	return nil
}

// Status writes one line per migration to w: state, when applied, file.
func Status(ctx context.Context, pool *pgxpool.Pool, w io.Writer) error {
	provider, err := newProvider(pool)
	if err != nil {
		return err
	}

	statuses, err := provider.Status(ctx)
	if err != nil {
		return fmt.Errorf("migration status: %w", err)
	}

	for _, status := range statuses {
		applied := "-"
		if status.State == goose.StateApplied {
			applied = status.AppliedAt.Format("2006-01-02 15:04:05")
		}
		if _, err := fmt.Fprintf(w, "%-8s %-20s %s\n", status.State, applied, status.Source.Path); err != nil {
			return fmt.Errorf("write migration status: %w", err)
		}
	}

	return nil
}

// newProvider runs the migrations embedded in FS against pool. A PostgreSQL
// advisory lock keeps two instances from migrating at once.
func newProvider(pool *pgxpool.Pool) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("migration lock: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		stdlib.OpenDBFromPool(pool),
		FS,
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return provider, nil
}

// logResults logs each applied or rolled back migration.
func logResults(results ...*goose.MigrationResult) {
	for _, result := range results {
		if result == nil {
			continue
		}
		slog.Info(
			"migration",
			"direction", result.Direction,
			"file", result.Source.Path,
			"took", result.Duration,
		)
	}
}
