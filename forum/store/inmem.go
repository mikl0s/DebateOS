package store

import "fmt"

// NewInMemory opens an in-memory SQLite database (":memory:"), applies migrations,
// and returns a fully initialized SQLiteStore. Used exclusively in tests.
func NewInMemory() (*SQLiteStore, error) {
	db, err := Open(":memory:")
	if err != nil {
		return nil, fmt.Errorf("store.NewInMemory: %w", err)
	}
	return New(db), nil
}
