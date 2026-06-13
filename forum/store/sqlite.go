package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mikl0s/debateos/forum/migrations"
	"github.com/mikl0s/debateos/forum/store/generated"

	_ "modernc.org/sqlite" // register "sqlite" driver
)

// SQLiteStore implements Store backed by modernc.org/sqlite via sqlc-generated queries.
type SQLiteStore struct {
	db *sql.DB
	q  *generated.Queries
}

// Open opens a SQLite database at dsn (e.g. "forum.db" or ":memory:"),
// enables WAL mode and foreign keys, then applies embedded migrations.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store.Open: sql.Open: %w", err)
	}
	// WAL mode for concurrent reads; foreign keys for referential integrity.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.Open: PRAGMA journal_mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.Open: PRAGMA foreign_keys: %w", err)
	}
	if err := migrations.Apply(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.Open: migrations: %w", err)
	}
	return db, nil
}

// New wraps an already-opened *sql.DB with the sqlc query layer.
func New(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db, q: generated.New(db)}
}

// DB exposes the underlying *sql.DB.
func (s *SQLiteStore) DB() *sql.DB { return s.db }

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error { return s.db.Close() }

// --- Search and retrieval ---

// SearchPoints performs an FTS5 full-text search over points.
// If q is empty, it falls back to ListPoints.
// foundation filters by JSON-array membership (simple substring check on stored JSON).
// Security: query is parameterized via FTS5 MATCH ? — no string interpolation (T-05-06).
func (s *SQLiteStore) SearchPoints(ctx context.Context, q, foundation string, limit int) ([]PointEntry, error) {
	if q == "" {
		return s.ListPoints(ctx, limit, 0)
	}

	const ftsSQL = `
		SELECT p.id, p.name, p.intent, p.curator, p.foundation_compat, p.commit_date,
		       p.subscribers, p.avg_rating, p.rating_count, p.tags
		FROM points_fts
		JOIN points p ON p.rowid = points_fts.rowid
		WHERE points_fts MATCH ?
		LIMIT ?`

	rows, err := s.db.QueryContext(ctx, ftsSQL, q, limit)
	if err != nil {
		return nil, fmt.Errorf("SearchPoints: FTS5 query: %w", err)
	}
	defer rows.Close()

	var results []PointEntry
	for rows.Next() {
		var p PointEntry
		if err := rows.Scan(&p.ID, &p.Name, &p.Intent, &p.Curator, &p.FoundationCompat,
			&p.CommitDate, &p.Subscribers, &p.AvgRating, &p.RatingCount, &p.Tags); err != nil {
			return nil, fmt.Errorf("SearchPoints: scan: %w", err)
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SearchPoints: rows: %w", err)
	}

	// Filter by foundation: parse the JSON array and check exact membership (CR-04).
	// Substring-contains on the raw JSON string was injection-prone — e.g.
	// foundation=arch","debian would match any record containing both adjacently.
	if foundation != "" {
		filtered := results[:0]
		for _, p := range results {
			var ids []string
			if err := json.Unmarshal([]byte(p.FoundationCompat), &ids); err == nil {
				for _, id := range ids {
					if id == foundation {
						filtered = append(filtered, p)
						break
					}
				}
			}
		}
		results = filtered
	}

	if results == nil {
		results = []PointEntry{}
	}
	return results, nil
}

// GetPoint retrieves a single point by ID. Returns nil, nil if not found.
func (s *SQLiteStore) GetPoint(ctx context.Context, id string) (*PointEntry, error) {
	row, err := s.q.GetPoint(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetPoint %q: %w", id, err)
	}
	p := pointFromGenerated(row)
	return &p, nil
}

// ListPoints returns points ordered by ID.
func (s *SQLiteStore) ListPoints(ctx context.Context, limit, offset int) ([]PointEntry, error) {
	rows, err := s.q.ListPoints(ctx, generated.ListPointsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("ListPoints: %w", err)
	}
	out := make([]PointEntry, len(rows))
	for i, r := range rows {
		out[i] = pointFromGenerated(r)
	}
	return out, nil
}

// UpsertPoint inserts or updates a point and rebuilds the FTS5 index (Pitfall 5).
func (s *SQLiteStore) UpsertPoint(ctx context.Context, p PointEntry) error {
	if err := s.q.UpsertPoint(ctx, generated.UpsertPointParams{
		ID:               p.ID,
		Name:             p.Name,
		Intent:           p.Intent,
		Curator:          p.Curator,
		FoundationCompat: p.FoundationCompat,
		CommitDate:       p.CommitDate,
		Tags:             p.Tags,
	}); err != nil {
		return fmt.Errorf("UpsertPoint %q: upsert: %w", p.ID, err)
	}
	// Rebuild FTS5 external-content table to sync with points (Pitfall 5).
	if err := s.q.RebuildFTS(ctx); err != nil {
		return fmt.Errorf("UpsertPoint %q: RebuildFTS: %w", p.ID, err)
	}
	return nil
}

// --- Subscriptions ---

func (s *SQLiteStore) AddSubscription(ctx context.Context, userID, pointID string) error {
	if err := s.q.AddSubscription(ctx, generated.AddSubscriptionParams{
		UserID:  userID,
		PointID: pointID,
	}); err != nil {
		return fmt.Errorf("AddSubscription: %w", err)
	}
	// Update subscriber aggregate on the point.
	_ = s.q.UpdatePointAggregates(ctx, pointID)
	return nil
}

func (s *SQLiteStore) RemoveSubscription(ctx context.Context, userID, pointID string) error {
	if err := s.q.RemoveSubscription(ctx, generated.RemoveSubscriptionParams{
		UserID:  userID,
		PointID: pointID,
	}); err != nil {
		return fmt.Errorf("RemoveSubscription: %w", err)
	}
	_ = s.q.UpdatePointAggregates(ctx, pointID)
	return nil
}

func (s *SQLiteStore) GetSubscriptions(ctx context.Context, userID string) ([]PointEntry, error) {
	rows, err := s.q.GetSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetSubscriptions: %w", err)
	}
	out := make([]PointEntry, len(rows))
	for i, r := range rows {
		out[i] = pointFromGenerated(r)
	}
	return out, nil
}

// --- Ratings ---

// SetRating writes or updates a rating for (userID, pointID) with stars in [1,5].
// Returns an error if stars is out of range (defense-in-depth alongside CHECK constraint).
func (s *SQLiteStore) SetRating(ctx context.Context, userID, pointID string, stars int) error {
	if stars < 1 || stars > 5 {
		return fmt.Errorf("SetRating: stars must be between 1 and 5, got %d", stars)
	}
	if err := s.q.SetRating(ctx, generated.SetRatingParams{
		UserID:  userID,
		PointID: pointID,
		Stars:   int64(stars),
	}); err != nil {
		return fmt.Errorf("SetRating: %w", err)
	}
	_ = s.q.UpdatePointAggregates(ctx, pointID)
	return nil
}

// GetRatings returns the aggregate rating summary for a point.
func (s *SQLiteStore) GetRatings(ctx context.Context, pointID string) (RatingSummary, error) {
	row, err := s.q.GetRatingSummary(ctx, pointID)
	if err != nil {
		return RatingSummary{}, fmt.Errorf("GetRatings: %w", err)
	}
	var avg float64
	switch v := row.AvgRating.(type) {
	case float64:
		avg = v
	case int64:
		avg = float64(v)
	case nil:
		avg = 0.0
	}
	return RatingSummary{Avg: avg, Count: row.RatingCount}, nil
}

// --- Conflict threads (stubs for 05-05) ---

func (s *SQLiteStore) GetConflicts(ctx context.Context, pointA, pointB string) ([]ConflictThread, error) {
	rows, err := s.q.GetConflicts(ctx, generated.GetConflictsParams{
		PointA:   pointA,
		PointB:   pointB,
		PointA_2: pointB, // symmetric lookup
		PointB_2: pointA,
	})
	if err != nil {
		return nil, fmt.Errorf("GetConflicts: %w", err)
	}
	out := make([]ConflictThread, len(rows))
	for i, r := range rows {
		out[i] = ConflictThread{
			ID: r.ID, PointA: r.PointA, PointB: r.PointB,
			Status: r.Status, PatchPRURL: r.PatchPrUrl, CreatedAt: r.CreatedAt,
		}
	}
	return out, nil
}

func (s *SQLiteStore) UpsertConflictThread(ctx context.Context, t ConflictThread) error {
	return s.q.UpsertConflictThread(ctx, generated.UpsertConflictThreadParams{
		ID:         t.ID,
		PointA:     t.PointA,
		PointB:     t.PointB,
		Status:     t.Status,
		PatchPrUrl: t.PatchPRURL,
	})
}

// Reindex rebuilds the FTS5 index — stub for 05-05 (full re-index from registry).
func (s *SQLiteStore) Reindex(ctx context.Context) error {
	return s.q.RebuildFTS(ctx)
}

// Truncate wipes all data (tests only).
func (s *SQLiteStore) Truncate(ctx context.Context) error {
	// Must delete in FK order: ratings, subscriptions, conflict_threads, then points.
	if err := s.q.TruncateRatings(ctx); err != nil {
		return err
	}
	if err := s.q.TruncateSubscriptions(ctx); err != nil {
		return err
	}
	if err := s.q.TruncateAll(ctx); err != nil {
		return err
	}
	return s.q.TruncatePoints(ctx)
}

// pointFromGenerated converts a generated.Point to a store.PointEntry.
func pointFromGenerated(p generated.Point) PointEntry {
	return PointEntry{
		ID:               p.ID,
		Name:             p.Name,
		Intent:           p.Intent,
		Curator:          p.Curator,
		FoundationCompat: p.FoundationCompat,
		CommitDate:       p.CommitDate,
		Subscribers:      p.Subscribers,
		AvgRating:        p.AvgRating,
		RatingCount:      p.RatingCount,
		Tags:             p.Tags,
	}
}
