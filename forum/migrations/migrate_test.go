package migrations_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mikl0s/debateos/forum/migrations"
)

// TestApply verifies that Apply creates all expected tables.
func TestApply(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	defer db.Close()

	if err := migrations.Apply(db); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Verify core tables exist by querying sqlite_master.
	tables := []string{"points", "subscriptions", "ratings", "conflict_threads"}
	for _, tbl := range tables {
		var name string
		row := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, tbl,
		)
		if err := row.Scan(&name); err != nil {
			t.Errorf("table %q not found after Apply: %v", tbl, err)
		}
	}

	// FTS5 virtual table should also exist.
	var ftsName string
	row := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='points_fts'`,
	)
	if err := row.Scan(&ftsName); err != nil {
		t.Errorf("points_fts virtual table not found after Apply: %v", err)
	}
}

// TestApplyIdempotent verifies that calling Apply twice does not fail.
func TestApplyIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	defer db.Close()

	if err := migrations.Apply(db); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	if err := migrations.Apply(db); err != nil {
		t.Fatalf("second Apply (idempotent): %v", err)
	}
}

// TestApplyClosedDB verifies that Apply returns an error when the database is closed.
func TestApplyClosedDB(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	db.Close() // close before Apply

	if err := migrations.Apply(db); err == nil {
		t.Error("expected error when applying migrations to closed DB, got nil")
	}
}
