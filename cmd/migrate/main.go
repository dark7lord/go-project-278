// Package main applies the embedded migrations: run.sh calls "up" before the
// API starts, and the db-* targets also roll back and report.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"code/db/migrations"
)

const usage = "usage: migrate up|down|redo|status"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	_ = godotenv.Load()

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	command := os.Args[1]

	if err := run(context.Background(), command); err != nil {
		slog.Error("migrate failed", "command", command, "error", err)
		os.Exit(1)
	}
	slog.Info("migrate done", "command", command)
}

func run(ctx context.Context, command string) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	switch command {
	case "up":
		err = migrations.Up(ctx, pool)
	case "down":
		err = migrations.Down(ctx, pool)
	case "redo":
		err = migrations.Redo(ctx, pool)
	case "status":
		err = migrations.Status(ctx, pool, os.Stdout)
	default:
		err = errors.New(usage)
	}

	return err
}
