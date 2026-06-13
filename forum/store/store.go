// Package store defines the Forum's storage interface and data types.
// The SQLite implementation (SQLiteStore) uses sqlc-generated queries over modernc.org/sqlite with FTS5.
// The in-memory implementation (NewInMemory) is used in tests.
package store

import (
	"context"
	"database/sql"
)

// PointEntry is a single point in the forum index.
// FoundationCompat and Tags are stored as JSON arrays.
type PointEntry struct {
	ID               string
	Name             string
	Intent           string
	Curator          string
	FoundationCompat string // JSON array: ["arch","debian"]
	CommitDate       string // ISO 8601
	Subscribers      int64
	AvgRating        float64
	RatingCount      int64
	Tags             string // JSON array: ["tag1","tag2"]
}

// RatingSummary holds the aggregate rating for a point.
type RatingSummary struct {
	Avg   float64
	Count int64
}

// ConflictThread represents a known conflict between two points with an optional patch PR.
type ConflictThread struct {
	ID         string
	PointA     string
	PointB     string
	Status     string // "open" | "resolved"
	PatchPRURL string
	CreatedAt  string
}

// Store is the Forum's storage abstraction.
// All write methods that involve user identity accept a userID string.
// The OAuth flow (05-05) supplies the real GitHub user ID; tests supply a fake string.
//
// Methods marked "stub for 05-05" return nil/empty and will be implemented in plan 05-05.
type Store interface {
	// Search and retrieval (FORM-01)
	SearchPoints(ctx context.Context, q string, foundation string, limit int) ([]PointEntry, error)
	GetPoint(ctx context.Context, id string) (*PointEntry, error)
	ListPoints(ctx context.Context, limit, offset int) ([]PointEntry, error)
	UpsertPoint(ctx context.Context, p PointEntry) error

	// Subscriptions (FORM-02)
	AddSubscription(ctx context.Context, userID, pointID string) error
	RemoveSubscription(ctx context.Context, userID, pointID string) error
	GetSubscriptions(ctx context.Context, userID string) ([]PointEntry, error)

	// Ratings (FORM-03) — write requires a real userID (OAuth-backed in production)
	SetRating(ctx context.Context, userID, pointID string, stars int) error
	GetRatings(ctx context.Context, pointID string) (RatingSummary, error)

	// Conflict threads (FORM-04) — stub for 05-05
	GetConflicts(ctx context.Context, pointA, pointB string) ([]ConflictThread, error)
	UpsertConflictThread(ctx context.Context, t ConflictThread) error

	// Reindex rebuilds the FTS5 index from the points table — stub for 05-05.
	Reindex(ctx context.Context) error

	// Truncate wipes all data — for tests only.
	Truncate(ctx context.Context) error

	// Close releases the underlying database connection.
	Close() error

	// DB exposes the underlying *sql.DB for raw queries in tests (e.g. FTS5 smoke).
	DB() *sql.DB
}
