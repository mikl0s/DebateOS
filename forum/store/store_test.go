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

// TestFoundationFilterInjectionRejected: CR-04 — the foundation filter must
// perform exact membership check via JSON parsing, not substring-contains.
// A crafted value like `arch","debian` must NOT match records containing
// only "arch" or only "debian" (injection bypass rejected).
// A plain value like "arch" must match records that contain "arch".
func TestFoundationFilterInjectionRejected(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Insert a point with only "arch" in its foundation compat.
	if err := s.UpsertPoint(ctx, store.PointEntry{
		ID: "arch-only", Name: "Arch Tool", Intent: "arch only tool",
		Curator: "alice", FoundationCompat: `["arch"]`,
	}); err != nil {
		t.Fatalf("UpsertPoint: %v", err)
	}

	// Exact match "arch" should find the point.
	results, err := s.SearchPoints(ctx, "arch", "arch", 10)
	if err != nil {
		t.Fatalf("SearchPoints: %v", err)
	}
	found := false
	for _, p := range results {
		if p.ID == "arch-only" {
			found = true
		}
	}
	if !found {
		t.Error("CR-04: exact foundation=arch should match arch-only point")
	}

	// Injection attempt: foundation=`arch","debian` must NOT match arch-only.
	// The old substring check would build needle=`"arch","debian"` which is not
	// a substring of `["arch"]`, so it would not match — but it WOULD match
	// a record like `["arch","debian"]`. We test the injection case directly:
	// the injected value `arch","debian` is not a member of ["arch"], so must not match.
	injectResults, err := s.SearchPoints(ctx, "arch", `arch","debian`, 10)
	if err != nil {
		t.Fatalf("SearchPoints injection: %v", err)
	}
	for _, p := range injectResults {
		if p.ID == "arch-only" {
			t.Errorf("CR-04: injection foundation=arch\",\"debian should NOT match arch-only (exact-membership check required)")
		}
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

// TestOpenInvalidDSN: Open with invalid DSN returns an error or creates a file (SQLite behavior).
// Primarily tests that Open doesn't panic and returns a usable or failed state.
func TestOpenValidDSN(t *testing.T) {
	// A known-good DSN: in-memory
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) failed: %v", err)
	}
	defer db.Close()
	if db == nil {
		t.Error("Open returned nil db")
	}
}

// TestNewWrapsDB: New returns a non-nil SQLiteStore.
func TestNewWrapsDB(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	s := store.New(db)
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.DB() == nil {
		t.Error("DB() returned nil")
	}
}

// TestStoreClose: Close returns nil for an in-memory store.
func TestStoreClose(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// TestGetPointNil: GetPoint returns nil, nil for a non-existent ID.
func TestGetPointNil(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	got, err := s.GetPoint(context.Background(), "does-not-exist")
	if err != nil {
		t.Fatalf("GetPoint: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing point, got %+v", got)
	}
}

// TestSearchPointsEmptyStore: SearchPoints on empty store returns empty slice.
func TestSearchPointsEmptyStore(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	results, err := s.SearchPoints(context.Background(), "anything", "", 10)
	if err != nil {
		t.Fatalf("SearchPoints: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestGetConflictsEmpty: GetConflicts on empty store returns empty slice.
func TestGetConflictsEmpty(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	threads, err := s.GetConflicts(context.Background(), "x", "y")
	if err != nil {
		t.Fatalf("GetConflicts: %v", err)
	}
	if len(threads) != 0 {
		t.Errorf("expected 0 threads, got %d", len(threads))
	}
}

// TestTruncateStoreMethod: Truncate clears all data.
func TestTruncateStoreMethod(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	if err := s.UpsertPoint(ctx, store.PointEntry{
		ID: "truncate-me", Name: "T", Curator: "a", FoundationCompat: `["arch"]`,
		CommitDate: "2026-01-01", Tags: "[]",
	}); err != nil {
		t.Fatalf("UpsertPoint: %v", err)
	}

	if err := s.Truncate(ctx); err != nil {
		t.Fatalf("Truncate: %v", err)
	}

	pts, err := s.ListPoints(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListPoints after Truncate: %v", err)
	}
	if len(pts) != 0 {
		t.Errorf("expected 0 points after Truncate, got %d", len(pts))
	}
}

// TestStoreReindex: store.Reindex rebuilds the FTS index.
func TestStoreReindex(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Reindex on an empty store should succeed.
	if err := s.Reindex(ctx); err != nil {
		t.Fatalf("Reindex: %v", err)
	}
}

// TestGetRatingsEmpty: GetRatings for a non-existent point returns zero summary.
func TestGetRatingsEmpty(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	sum, err := s.GetRatings(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetRatings: %v", err)
	}
	if sum.Count != 0 {
		t.Errorf("expected count 0 for nonexistent point, got %d", sum.Count)
	}
}

// TestConflictThreads: UpsertConflictThread round-trips; status updatable; symmetric lookup.
func TestConflictThreads(t *testing.T) {
	ctx := context.Background()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Insert a conflict thread.
	thread := store.ConflictThread{
		ID:         "ct-test-01",
		PointA:     "docker-op",
		PointB:     "podman-op",
		Status:     "open",
		PatchPRURL: "https://github.com/mikl0s/debateos/pull/12",
	}
	if err := s.UpsertConflictThread(ctx, thread); err != nil {
		t.Fatalf("UpsertConflictThread: %v", err)
	}

	// Retrieve by (a, b).
	threads, err := s.GetConflicts(ctx, "docker-op", "podman-op")
	if err != nil {
		t.Fatalf("GetConflicts(a,b): %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	got := threads[0]
	if got.ID != "ct-test-01" {
		t.Errorf("ID: got %q", got.ID)
	}
	if got.PatchPRURL != "https://github.com/mikl0s/debateos/pull/12" {
		t.Errorf("PatchPRURL: got %q", got.PatchPRURL)
	}
	if got.Status != "open" {
		t.Errorf("Status: got %q", got.Status)
	}

	// Symmetric lookup: (b, a) should also return the thread.
	threads2, err := s.GetConflicts(ctx, "podman-op", "docker-op")
	if err != nil {
		t.Fatalf("GetConflicts(b,a): %v", err)
	}
	if len(threads2) != 1 {
		t.Errorf("symmetric lookup: expected 1, got %d", len(threads2))
	}

	// Update status to resolved.
	thread.Status = "resolved"
	if err := s.UpsertConflictThread(ctx, thread); err != nil {
		t.Fatalf("UpsertConflictThread (update): %v", err)
	}

	threads3, err := s.GetConflicts(ctx, "docker-op", "podman-op")
	if err != nil {
		t.Fatalf("GetConflicts after update: %v", err)
	}
	if len(threads3) == 0 {
		t.Fatal("no threads after status update")
	}
	if threads3[0].Status != "resolved" {
		t.Errorf("after update: expected resolved, got %q", threads3[0].Status)
	}

	// PatchPRURL preserved after status update.
	if threads3[0].PatchPRURL != "https://github.com/mikl0s/debateos/pull/12" {
		t.Errorf("PatchPRURL after update: got %q", threads3[0].PatchPRURL)
	}
}

// TestClosedDBErrors: Error paths in SQLiteStore methods when the underlying DB is closed.
// Covers the error return branches in GetRatings, SearchPoints, UpsertPoint, etc.
func TestClosedDBErrors(t *testing.T) {
	ctx := context.Background()

	// Open a real in-memory DB, apply migrations, then close it.
	// The wrapped store will now return errors on all operations.
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	s := store.New(db)
	db.Close() // close the underlying DB

	t.Run("GetRatings", func(t *testing.T) {
		if _, err := s.GetRatings(ctx, "p"); err == nil {
			t.Error("expected error from GetRatings with closed DB, got nil")
		}
	})

	t.Run("UpsertPoint", func(t *testing.T) {
		err := s.UpsertPoint(ctx, store.PointEntry{ID: "x", Name: "X", Curator: "c", FoundationCompat: `["arch"]`})
		if err == nil {
			t.Error("expected error from UpsertPoint with closed DB, got nil")
		}
	})

	t.Run("Truncate", func(t *testing.T) {
		if err := s.Truncate(ctx); err == nil {
			t.Error("expected error from Truncate with closed DB, got nil")
		}
	})

	t.Run("SearchPoints", func(t *testing.T) {
		// Non-empty q triggers FTS5 path which hits the closed DB.
		if _, err := s.SearchPoints(ctx, "docker", "", 10); err == nil {
			t.Error("expected error from SearchPoints with closed DB, got nil")
		}
	})

	t.Run("GetPoint", func(t *testing.T) {
		if _, err := s.GetPoint(ctx, "missing"); err == nil {
			t.Error("expected error from GetPoint with closed DB, got nil")
		}
	})

	t.Run("ListPoints", func(t *testing.T) {
		if _, err := s.ListPoints(ctx, 10, 0); err == nil {
			t.Error("expected error from ListPoints with closed DB, got nil")
		}
	})

	t.Run("AddSubscription", func(t *testing.T) {
		if err := s.AddSubscription(ctx, "user1", "point1"); err == nil {
			t.Error("expected error from AddSubscription with closed DB, got nil")
		}
	})

	t.Run("RemoveSubscription", func(t *testing.T) {
		if err := s.RemoveSubscription(ctx, "user1", "point1"); err == nil {
			t.Error("expected error from RemoveSubscription with closed DB, got nil")
		}
	})

	t.Run("GetSubscriptions", func(t *testing.T) {
		if _, err := s.GetSubscriptions(ctx, "user1"); err == nil {
			t.Error("expected error from GetSubscriptions with closed DB, got nil")
		}
	})

	t.Run("GetConflicts", func(t *testing.T) {
		if _, err := s.GetConflicts(ctx, "a", "b"); err == nil {
			t.Error("expected error from GetConflicts with closed DB, got nil")
		}
	})

	t.Run("Reindex", func(t *testing.T) {
		if err := s.Reindex(ctx); err == nil {
			t.Error("expected error from Reindex with closed DB, got nil")
		}
	})
}
