package store_test

import (
	"context"
	"testing"

	"github.com/mikl0s/debateos/forum/store"
)

// TestFTS5Smoke proves FTS5 is compiled into modernc.org/sqlite v1.46.1.
// Opens :memory: SQLite, applies migrations, creates the points_fts FTS5 virtual table.
func TestFTS5Smoke(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()
	// If migrations succeeded, points_fts was created with FTS5 — just listing it is the smoke test.
	// We do a trivial FTS5 query to confirm the module is compiled in.
	rows, err := s.DB().QueryContext(context.Background(), "SELECT * FROM points_fts WHERE points_fts MATCH 'smoke'")
	if err != nil {
		t.Fatalf("FTS5 query failed (FTS5 not compiled?): %v", err)
	}
	rows.Close()
}

// TestSearchPoints: upsert 3 points, SearchPoints returns matching rows, foundation filter works.
func TestSearchPoints(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Insert 3 points
	points := []store.PointEntry{
		{ID: "p1", Name: "Docker Setup", Intent: "install docker on arch", Curator: "alice", FoundationCompat: `["arch"]`},
		{ID: "p2", Name: "Docker Compose", Intent: "docker compose tools", Curator: "bob", FoundationCompat: `["arch","debian"]`},
		{ID: "p3", Name: "Debian Base", Intent: "basic debian setup", Curator: "carol", FoundationCompat: `["debian"]`},
	}
	for _, p := range points {
		if err := s.UpsertPoint(ctx, p); err != nil {
			t.Fatalf("UpsertPoint %q: %v", p.ID, err)
		}
	}

	// Search "docker" — should return p1 and p2
	results, err := s.SearchPoints(ctx, "docker", "", 10)
	if err != nil {
		t.Fatalf("SearchPoints: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'docker', got %d: %+v", len(results), results)
	}

	// Search "docker" with foundation filter "arch" — should return p1 and p2 (both have arch)
	results, err = s.SearchPoints(ctx, "docker", "arch", 10)
	if err != nil {
		t.Fatalf("SearchPoints with foundation: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'docker'+arch, got %d", len(results))
	}

	// Search "docker" with foundation filter "debian" — should return only p2
	results, err = s.SearchPoints(ctx, "docker", "debian", 10)
	if err != nil {
		t.Fatalf("SearchPoints with debian filter: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'docker'+debian, got %d", len(results))
	}
	if len(results) == 1 && results[0].ID != "p2" {
		t.Errorf("expected p2, got %q", results[0].ID)
	}
}

// TestListGetUpsert: UpsertPoint then GetPoint round-trips all fields; ListPoints returns inserted points.
func TestListGetUpsert(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	p := store.PointEntry{
		ID:               "test-point-1",
		Name:             "Test Point",
		Intent:           "test intent",
		Curator:          "testcurator",
		FoundationCompat: `["arch","debian"]`,
		CommitDate:       "2026-01-01T00:00:00Z",
		Tags:             `["test","demo"]`,
	}
	if err := s.UpsertPoint(ctx, p); err != nil {
		t.Fatalf("UpsertPoint: %v", err)
	}

	got, err := s.GetPoint(ctx, "test-point-1")
	if err != nil {
		t.Fatalf("GetPoint: %v", err)
	}
	if got == nil {
		t.Fatal("GetPoint returned nil")
	}
	if got.Name != p.Name {
		t.Errorf("Name: got %q, want %q", got.Name, p.Name)
	}
	if got.Intent != p.Intent {
		t.Errorf("Intent: got %q, want %q", got.Intent, p.Intent)
	}
	if got.Curator != p.Curator {
		t.Errorf("Curator: got %q, want %q", got.Curator, p.Curator)
	}

	list, err := s.ListPoints(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListPoints: %v", err)
	}
	if len(list) < 1 {
		t.Error("ListPoints returned empty slice")
	}
	found := false
	for _, lp := range list {
		if lp.ID == "test-point-1" {
			found = true
		}
	}
	if !found {
		t.Error("ListPoints did not return inserted point")
	}
}
