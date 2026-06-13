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

// TestSubscriptions: AddSubscription, GetSubscriptions, RemoveSubscription, idempotency.
func TestSubscriptions(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Seed a point first (FK constraint)
	if err := s.UpsertPoint(ctx, store.PointEntry{
		ID: "sub-point", Name: "Sub Point", Intent: "test", Curator: "alice",
		FoundationCompat: `["arch"]`,
	}); err != nil {
		t.Fatalf("UpsertPoint: %v", err)
	}

	// Add subscription
	if err := s.AddSubscription(ctx, "user1", "sub-point"); err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}

	// GetSubscriptions should return the point
	subs, err := s.GetSubscriptions(ctx, "user1")
	if err != nil {
		t.Fatalf("GetSubscriptions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].ID != "sub-point" {
		t.Errorf("expected sub-point, got %q", subs[0].ID)
	}

	// Duplicate Add is idempotent (PK conflict DO NOTHING)
	if err := s.AddSubscription(ctx, "user1", "sub-point"); err != nil {
		t.Fatalf("duplicate AddSubscription should be idempotent: %v", err)
	}
	subs, _ = s.GetSubscriptions(ctx, "user1")
	if len(subs) != 1 {
		t.Errorf("after duplicate add, expected 1 subscription, got %d", len(subs))
	}

	// RemoveSubscription
	if err := s.RemoveSubscription(ctx, "user1", "sub-point"); err != nil {
		t.Fatalf("RemoveSubscription: %v", err)
	}
	subs, _ = s.GetSubscriptions(ctx, "user1")
	if len(subs) != 0 {
		t.Errorf("after remove, expected 0 subscriptions, got %d", len(subs))
	}
}

// TestRatings: SetRating, GetRatings aggregate, out-of-range rejection.
func TestRatings(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Seed a point
	if err := s.UpsertPoint(ctx, store.PointEntry{
		ID: "rate-point", Name: "Rate Point", Intent: "test", Curator: "bob",
		FoundationCompat: `["arch"]`,
	}); err != nil {
		t.Fatalf("UpsertPoint: %v", err)
	}

	// SetRating: user1 rates 4
	if err := s.SetRating(ctx, "user1", "rate-point", 4); err != nil {
		t.Fatalf("SetRating 4: %v", err)
	}
	sum, err := s.GetRatings(ctx, "rate-point")
	if err != nil {
		t.Fatalf("GetRatings: %v", err)
	}
	if sum.Count != 1 {
		t.Errorf("expected count 1, got %d", sum.Count)
	}
	if sum.Avg != 4.0 {
		t.Errorf("expected avg 4.0, got %f", sum.Avg)
	}

	// Second user rates 2 → avg 3.0 count 2
	if err := s.SetRating(ctx, "user2", "rate-point", 2); err != nil {
		t.Fatalf("SetRating 2: %v", err)
	}
	sum, _ = s.GetRatings(ctx, "rate-point")
	if sum.Count != 2 {
		t.Errorf("expected count 2, got %d", sum.Count)
	}
	if sum.Avg != 3.0 {
		t.Errorf("expected avg 3.0, got %f", sum.Avg)
	}

	// Out-of-range stars=6 must error (V5 validation)
	if err := s.SetRating(ctx, "user3", "rate-point", 6); err == nil {
		t.Error("SetRating with stars=6 should return an error")
	}

	// stars=0 must also error
	if err := s.SetRating(ctx, "user3", "rate-point", 0); err == nil {
		t.Error("SetRating with stars=0 should return an error")
	}
}
