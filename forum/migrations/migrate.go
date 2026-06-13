// Package migrations provides embedded SQL migrations for the Forum database.
package migrations

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed 001_init.sql
var initSQL string

// Apply applies all embedded migrations to db in order.
// It is safe to call multiple times — all statements use IF NOT EXISTS.
func Apply(db *sql.DB) error {
	if _, err := db.Exec(initSQL); err != nil {
		return fmt.Errorf("migrations.Apply: %w", err)
	}
	return nil
}
