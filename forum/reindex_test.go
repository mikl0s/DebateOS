package forum_test

// Reindex tests (FORM-05 — total DB loss recoverable by re-reading the registry index).
// RED phase: Reindex function does not exist yet.
// TestReindex: starting from empty store, Reindex upserts all points from a RegistryIndex;
//              second Reindex is idempotent (no duplicates).

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mikl0s/debateos/forum"
	"github.com/mikl0s/debateos/forum/store"
	"github.com/mikl0s/debateos/registry/index"
)

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
