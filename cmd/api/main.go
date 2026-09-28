// Package main is the entry point for the URL shortener service.
package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"code/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file, using the environment")
	}

	if err := app.Run(); err != nil {
		slog.Error("app stopped", "error", err)
		os.Exit(1)
	}
}
