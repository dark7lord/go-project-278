// Package migrations holds the SQL migrations, embedded in the binary, and applies them.
package migrations

import "embed"

// FS contains the embedded SQL migration files.
//
//go:embed *.sql
var FS embed.FS
