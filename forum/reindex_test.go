package forum_test

// Reindex tests (FORM-05 — total DB loss recoverable by re-reading the registry index).
// RED phase: Reindex function does not exist yet.
// TestReindex: starting from empty store, Reindex upserts all points from a RegistryIndex;
//              second Reindex is idempotent (no duplicates).

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mikl0s/debateos/forum"
	"github.com/mikl0s/debateos/forum/store"
	"github.com/mikl0s/debateos/registry/index"
)

// errStore is a minimal store.Store implementation that returns errors on UpsertPoint
// and Reindex, for testing error paths in forum.Reindex.
type errStore struct {
	upsertErr  error
	reindexErr error
}

func (e *errStore) SearchPoints(_ context.Context, _, _ string, _ int) ([]store.PointEntry, error) {
	return nil, nil
}
func (e *errStore) GetPoint(_ context.Context, _ string) (*store.PointEntry, error) {
	return nil, nil
}
func (e *errStore) ListPoints(_ context.Context, _, _ int) ([]store.PointEntry, error) {
	return nil, nil
}
func (e *errStore) UpsertPoint(_ context.Context, _ store.PointEntry) error {
	return e.upsertErr
}
func (e *errStore) AddSubscription(_ context.Context, _, _ string) error   { return nil }
func (e *errStore) RemoveSubscription(_ context.Context, _, _ string) error { return nil }
func (e *errStore) GetSubscriptions(_ context.Context, _ string) ([]store.PointEntry, error) {
	return nil, nil
}
func (e *errStore) SetRating(_ context.Context, _, _ string, _ int) error { return nil }
func (e *errStore) GetRatings(_ context.Context, _ string) (store.RatingSummary, error) {
	return store.RatingSummary{}, nil
}
func (e *errStore) GetConflicts(_ context.Context, _, _ string) ([]store.ConflictThread, error) {
	return nil, nil
}
func (e *errStore) UpsertConflictThread(_ context.Context, _ store.ConflictThread) error {
	return nil
}
func (e *errStore) Reindex(_ context.Context) error { return e.reindexErr }
func (e *errStore) Truncate(_ context.Context) error { return nil }
func (e *errStore) Close() error                    { return nil }
func (e *errStore) DB() *sql.DB                     { return nil }

// TestReindexUpsertError: Reindex returns error when UpsertPoint fails.
func TestReindexUpsertError(t *testing.T) {
	ctx := context.Background()
	s := &errStore{upsertErr: errors.New("upsert failed")}
	idx := sampleRegistryIndex(1)
	if err := forum.Reindex(ctx, s, idx); err == nil {
		t.Error("expected error when UpsertPoint fails, got nil")
	}
}

// TestReindexFTSError: Reindex returns error when FTS rebuild fails.
func TestReindexFTSError(t *testing.T) {
	ctx := context.Background()
	s := &errStore{reindexErr: errors.New("fts rebuild failed")}
	idx := sampleRegistryIndex(1)
	if err := forum.Reindex(ctx, s, idx); err == nil {
		t.Error("expected error when Reindex FTS fails, got nil")
	}
}

// sampleRegistryIndex builds a minimal RegistryIndex with n points.
func sampleRegistryIndex(n int) *index.RegistryIndex {
	ids := []string{"point-alpha", "point-beta", "point-gamma", "point-delta", "point-epsilon"}
	points := make([]index.PointEntry, 0, n)
	for i := 0; i < n && i < len(ids); i++ {
		points = append(points, index.PointEntry{
			ID:      ids[i],
			Name:    "Point " + ids[i],
			Intent:  "Intent for " + ids[i],
			Curator: "curator-x",
			Members: []string{"op-1", "op-2"},
			FoundationCompat: []index.FoundationCompat{
				{Foundation: "arch", Compatible: true},
				{Foundation: "debian", Compatible: true},
			},
			CommitDate: "2026-01-01T00:00:00Z",
			Tags:       []string{"test", "demo"},
		})
	}
	return &index.RegistryIndex{
		Schema:      index.SchemaVersion,
		GeneratedAt: "2026-01-01T00:00:00Z",
		Points:      points,
	}
}

// TestReindex: Reindex from RegistryIndex → store; idempotent on second call.
func TestReindex(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	idx := sampleRegistryIndex(3)

	// First Reindex: empty store → all 3 points upserted.
	if err := forum.Reindex(ctx, s, idx); err != nil {
		t.Fatalf("Reindex (first): %v", err)
	}

	points, err := s.ListPoints(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListPoints after first Reindex: %v", err)
	}
	if len(points) != 3 {
		t.Errorf("after first Reindex: expected 3 points, got %d", len(points))
	}

	// Verify a specific point was stored correctly.
	got, err := s.GetPoint(ctx, "point-alpha")
	if err != nil {
		t.Fatalf("GetPoint point-alpha: %v", err)
	}
	if got == nil {
		t.Fatal("point-alpha not found after Reindex")
	}
	if got.Name != "Point point-alpha" {
		t.Errorf("Name: got %q, want %q", got.Name, "Point point-alpha")
	}
	if got.Curator != "curator-x" {
		t.Errorf("Curator: got %q, want curator-x", got.Curator)
	}

	// FoundationCompat stored as JSON array.
	var fc []map[string]interface{}
	if err := json.Unmarshal([]byte(got.FoundationCompat), &fc); err != nil {
		t.Errorf("FoundationCompat is not valid JSON: %v (got %q)", err, got.FoundationCompat)
	}

	// Second Reindex: idempotent — no duplicates (FORM-05 DB-loss recovery proof).
	if err := forum.Reindex(ctx, s, idx); err != nil {
		t.Fatalf("Reindex (second): %v", err)
	}

	points2, err := s.ListPoints(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListPoints after second Reindex: %v", err)
	}
	if len(points2) != 3 {
		t.Errorf("after second Reindex (idempotent): expected 3 points, got %d", len(points2))
	}
}

// TestReindexFromLargerIndex: Reindex with 5 points all stored; idempotent.
func TestReindexFromLargerIndex(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	idx := sampleRegistryIndex(5)

	if err := forum.Reindex(ctx, s, idx); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	points, err := s.ListPoints(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListPoints: %v", err)
	}
	if len(points) != 5 {
		t.Errorf("expected 5 points, got %d", len(points))
	}

	// Second pass (idempotent).
	if err := forum.Reindex(ctx, s, idx); err != nil {
		t.Fatalf("second Reindex: %v", err)
	}
	points2, err := s.ListPoints(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListPoints after idempotent: %v", err)
	}
	if len(points2) != 5 {
		t.Errorf("idempotent: expected 5 points, got %d", len(points2))
	}
}

// TestReindexNilIndex: Reindex with nil index returns error.
func TestReindexNilIndex(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	if err := forum.Reindex(ctx, s, nil); err == nil {
		t.Error("expected error for nil index, got nil")
	}
}

// TestReindexEmptyIndex: Reindex with empty index succeeds with no points.
func TestReindexEmptyIndex(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	idx := &index.RegistryIndex{
		Schema:      index.SchemaVersion,
		GeneratedAt: "2026-01-01T00:00:00Z",
		Points:      []index.PointEntry{},
	}
	if err := forum.Reindex(ctx, s, idx); err != nil {
		t.Fatalf("Reindex empty index: %v", err)
	}

	points, err := s.ListPoints(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListPoints: %v", err)
	}
	if len(points) != 0 {
		t.Errorf("expected 0 points for empty index, got %d", len(points))
	}
}

// TestReindexPreservesFoundationCompat: FoundationCompat from RegistryIndex
// is stored as a JSON array on the PointEntry in the store.
func TestReindexPreservesFoundationCompat(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	idx := &index.RegistryIndex{
		Schema:      index.SchemaVersion,
		GeneratedAt: "2026-01-01T00:00:00Z",
		Points: []index.PointEntry{
			{
				ID:      "compat-point",
				Name:    "Compat Test",
				Intent:  "testing compat",
				Curator: "alice",
				FoundationCompat: []index.FoundationCompat{
					{Foundation: "arch", Compatible: true},
					{Foundation: "debian", Compatible: false, Missing: []string{"deploy-config-file-tree"}},
				},
				Tags: []string{"net"},
			},
		},
	}

	if err := forum.Reindex(ctx, s, idx); err != nil {
		t.Fatalf("Reindex: %v", err)
	}

	got, err := s.GetPoint(ctx, "compat-point")
	if err != nil {
		t.Fatalf("GetPoint: %v", err)
	}
	if got == nil {
		t.Fatal("compat-point not found")
	}

	// FoundationCompat must be the list of compatible foundation names.
	if got.FoundationCompat == "" {
		t.Fatal("FoundationCompat is empty")
	}
	if !json.Valid([]byte(got.FoundationCompat)) {
		t.Errorf("FoundationCompat is not valid JSON: %q", got.FoundationCompat)
	}
}
