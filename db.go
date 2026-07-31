package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func connectDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return conn, nil
}
