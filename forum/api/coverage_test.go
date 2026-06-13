package api_test

// coverage_test.go — additional tests targeting uncovered forum/api paths.
// Goal: bring forum/api coverage from ~56% to >=85%.
// All tests use in-memory SQLite store and no live network calls.

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/forum/api"
	"github.com/mikl0s/debateos/forum/store"
)

// failingStore is a store.Store implementation that always returns errors.
// Used to test HTTP 500 error paths in API handlers.
type failingStore struct {
	getRatingsErr      error
	getPointErr        error
	listPointsErr      error
	searchPointsErr    error
	setRatingErr       error
	addSubscriptionErr error
	getConflictsErr    error
	upsertConflictErr  error
}

func (f *failingStore) SearchPoints(_ context.Context, _, _ string, _ int) ([]store.PointEntry, error) {
	return nil, f.searchPointsErr
}
func (f *failingStore) GetPoint(_ context.Context, _ string) (*store.PointEntry, error) {
	return nil, f.getPointErr
}
func (f *failingStore) ListPoints(_ context.Context, _, _ int) ([]store.PointEntry, error) {
	return nil, f.listPointsErr
}
func (f *failingStore) UpsertPoint(_ context.Context, _ store.PointEntry) error      { return nil }
func (f *failingStore) UpsertPointBatch(_ context.Context, _ store.PointEntry) error { return nil }
func (f *failingStore) AddSubscription(_ context.Context, _, _ string) error {
	return f.addSubscriptionErr
}
func (f *failingStore) RemoveSubscription(_ context.Context, _, _ string) error { return nil }
func (f *failingStore) GetSubscriptions(_ context.Context, _ string) ([]store.PointEntry, error) {
	return nil, nil
}
func (f *failingStore) SetRating(_ context.Context, _, _ string, _ int) error {
	return f.setRatingErr
}
func (f *failingStore) GetRatings(_ context.Context, _ string) (store.RatingSummary, error) {
	return store.RatingSummary{}, f.getRatingsErr
}
func (f *failingStore) GetConflicts(_ context.Context, _, _ string) ([]store.ConflictThread, error) {
	return nil, f.getConflictsErr
}
func (f *failingStore) UpsertConflictThread(_ context.Context, _ store.ConflictThread) error {
	return f.upsertConflictErr
}
func (f *failingStore) Reindex(_ context.Context) error { return nil }
func (f *failingStore) Truncate(_ context.Context) error { return nil }
func (f *failingStore) Close() error                    { return nil }
func (f *failingStore) DB() *sql.DB                     { return nil }

// seedStoreWithPoints seeds multiple points for coverage tests.
func seedStoreWithPoints(t *testing.T, s *store.SQLiteStore, entries []store.PointEntry) {
	t.Helper()
	ctx := context.Background()
	for _, p := range entries {
		if err := s.UpsertPoint(ctx, p); err != nil {
			t.Fatalf("seed UpsertPoint %q: %v", p.ID, err)
		}
	}
}

func newTestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestListPoints: GET /api/points returns JSON array.
func TestListPoints(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "p1", Name: "Point One", Intent: "first", Curator: "alice", FoundationCompat: `["arch"]`},
		{ID: "p2", Name: "Point Two", Intent: "second", Curator: "bob", FoundationCompat: `["debian"]`},
	})

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result.Points) < 2 {
		t.Errorf("expected >=2 points, got %d", len(result.Points))
	}
}

// TestListPointsWithLimit: GET /api/points?limit=1 returns at most 1 point.
func TestListPointsWithLimit(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "pa", Name: "Alpha", Curator: "x", FoundationCompat: `["arch"]`},
		{ID: "pb", Name: "Beta", Curator: "y", FoundationCompat: `["arch"]`},
		{ID: "pc", Name: "Gamma", Curator: "z", FoundationCompat: `["arch"]`},
	})

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points?limit=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result.Points) > 1 {
		t.Errorf("expected <=1 point with limit=1, got %d", len(result.Points))
	}
}

// TestGetPoint: GET /api/points/{id} returns the specific point.
func TestGetPoint(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "my-point", Name: "My Point", Curator: "alice", FoundationCompat: `["arch","debian"]`},
	})

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points/my-point", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var point store.PointEntry
	if err := json.NewDecoder(w.Body).Decode(&point); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if point.ID != "my-point" {
		t.Errorf("expected id='my-point', got %q", point.ID)
	}
}

// TestGetPointNotFound: GET /api/points/{id} returns 404 for missing point.
func TestGetPointNotFound(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points/does-not-exist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing point, got %d", w.Code)
	}
}

// TestSearchWithFoundationFilter: GET /api/search?q=Tool&foundation=arch returns only arch-compatible.
// Note: when q="" SearchPoints falls back to ListPoints (no foundation filter).
// With a non-empty q, the FTS5 search + foundation filter both apply.
func TestSearchWithFoundationFilter(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "arch-only", Name: "Arch Tool", Curator: "alice", FoundationCompat: `["arch"]`},
		{ID: "both", Name: "Cross Tool", Curator: "bob", FoundationCompat: `["arch","debian"]`},
		{ID: "deb-only", Name: "Debian Tool", Curator: "carol", FoundationCompat: `["debian"]`},
	})

	router := api.NewRouter(s, noIdentity)
	// "Tool" matches all three; foundation=arch should filter out deb-only.
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=Tool&foundation=arch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, p := range result.Points {
		if p.ID == "deb-only" {
			t.Errorf("deb-only should not appear in arch filter results")
		}
	}
}

// TestSearchWithLimit: GET /api/search?q=&limit=1 respects limit.
func TestSearchWithLimit(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		p := store.PointEntry{
			ID:               fmt.Sprintf("point-%d", i),
			Name:             fmt.Sprintf("Point %d", i),
			Curator:          "alice",
			FoundationCompat: `["arch"]`,
		}
		if err := s.UpsertPoint(context.Background(), p); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=Point&limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result.Points) > 2 {
		t.Errorf("expected <=2 points with limit=2, got %d", len(result.Points))
	}
}

// TestGetRatings: GET /api/ratings/{pointId} returns avg and count.
func TestGetRatings(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "rated-point", Name: "Rated Point", Curator: "alice", FoundationCompat: `["arch"]`},
	})
	// Set a rating via store directly to populate data.
	if err := s.SetRating(context.Background(), "user-1", "rated-point", 4); err != nil {
		t.Fatalf("SetRating: %v", err)
	}

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/ratings/rated-point", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var result struct {
		PointID string  `json:"point_id"`
		Avg     float64 `json:"avg"`
		Count   int     `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Count < 1 {
		t.Errorf("expected count >= 1, got %d", result.Count)
	}
}

// TestPostRatingInvalidStars: POST /api/ratings with stars=0 returns 400.
func TestPostRatingInvalidStars(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "test-point", Name: "Test", Curator: "x", FoundationCompat: `["arch"]`},
	})

	router := api.NewRouter(s, fakeIdentity("user-1"))
	body := `{"point_id":"test-point","stars":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for stars=0, got %d", w.Code)
	}
}

// TestPostRatingMissingPointID: POST /api/ratings without point_id returns 400.
func TestPostRatingMissingPointID(t *testing.T) {
	s := newTestStore(t)

	router := api.NewRouter(s, fakeIdentity("user-1"))
	body := `{"stars":3}`
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing point_id, got %d", w.Code)
	}
}

// TestSubscriptionRoundTrip: POST /api/subscriptions → DELETE → verify.
func TestSubscriptionRoundTrip(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "sub-point", Name: "Sub Point", Curator: "alice", FoundationCompat: `["arch"]`},
	})

	router := api.NewRouter(s, fakeIdentity("user-sub"))

	// Subscribe
	subBody := `{"point_id":"sub-point"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(subBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/subscriptions expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Unsubscribe via DELETE /api/subscriptions?point_id=sub-point
	req2 := httptest.NewRequest(http.MethodDelete, "/api/subscriptions?point_id=sub-point", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("DELETE /api/subscriptions expected 200, got %d (body: %s)", w2.Code, w2.Body.String())
	}
}

// TestSubscriptionRequiresIdentity: POST /api/subscriptions without identity returns 401.
func TestSubscriptionRequiresIdentity(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)

	body := `{"point_id":"any-point"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without identity, got %d", w.Code)
	}
}

// TestSubscriptionMissingPointID: POST /api/subscriptions with empty point_id returns 400.
func TestSubscriptionMissingPointID(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"point_id":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty point_id, got %d", w.Code)
	}
}

// TestDeleteSubscriptionMissingPointID: DELETE /api/subscriptions without point_id returns 400.
func TestDeleteSubscriptionMissingPointID(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	// Delete without point_id param (no path param, no query, no body)
	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// should be 400 (missing point_id) — not 500.
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing point_id, got %d", w.Code)
	}
}

// TestSearchEmptyResult: GET /api/search?q=xyz returns empty list (not error).
func TestSearchEmptyResult(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=zzznomatch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Empty result is fine — should be [] not nil.
	if result.Points == nil {
		t.Error("expected empty array [], got null")
	}
}

// TestJSONDecodeBodyValid: jsonDecodeBody correctly decodes valid JSON.
// Tested indirectly via POST /api/ratings with valid body.
func TestJSONDecodeBodyViaRating(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "decode-test", Name: "Decode Test", Curator: "x", FoundationCompat: `["arch"]`},
	})

	router := api.NewRouter(s, fakeIdentity("user-1"))

	// Valid JSON body — triggers jsonDecodeBody path.
	body := `{"point_id":"decode-test","stars":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestPostRatingInvalidJSON: POST /api/ratings with malformed JSON returns 400.
func TestPostRatingInvalidJSON(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

// TestPostSubscriptionInvalidJSON: POST /api/subscriptions with malformed JSON returns 400.
func TestPostSubscriptionInvalidJSON(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

// TestListPointsWithOffset: GET /api/points?offset=1 paginates correctly.
func TestListPointsWithOffset(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "first", Name: "First", Curator: "a", FoundationCompat: `["arch"]`},
		{ID: "second", Name: "Second", Curator: "b", FoundationCompat: `["arch"]`},
		{ID: "third", Name: "Third", Curator: "c", FoundationCompat: `["arch"]`},
	})

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points?limit=10&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// With offset=1 on 3 items, should get at most 2.
	if len(result.Points) > 2 {
		t.Errorf("expected <=2 points with offset=1, got %d", len(result.Points))
	}
}

// TestGetRatingsMissingPointID: GET /api/ratings/ without pointId returns 400.
func TestGetRatingsMissingPointID(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)

	// chi route /api/ratings/{pointId} requires the path param
	// accessing /api/ratings/ (no param) should be 404 from chi (route not matched).
	req := httptest.NewRequest(http.MethodGet, "/api/ratings/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// chi 404 or handler 400 — both acceptable, not 500.
	if w.Code == http.StatusInternalServerError {
		t.Errorf("expected 400 or 404 for missing pointId, got 500")
	}
}

// TestGetConflictsMissingParams: GET /api/conflicts without a and b returns 400.
func TestGetConflictsMissingParams(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/conflicts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing a/b params, got %d", w.Code)
	}
}

// TestGetConflictsWithParams: GET /api/conflicts?a=X&b=Y returns empty list.
func TestGetConflictsWithParams(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/conflicts?a=op-1&b=op-2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// TestPostConflictFull: POST /api/conflicts with valid body creates a conflict thread.
func TestPostConflictFull(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"id":"ct-full-01","point_a":"op-a","point_b":"op-b","status":"open","patch_pr_url":"https://github.com/example/pr/1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestPostConflictMissingPoints: POST /api/conflicts without point_a/point_b returns 400.
func TestPostConflictMissingPoints(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"id":"ct-bad","status":"open"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing point_a/point_b, got %d", w.Code)
	}
}

// TestPostConflictInvalidStatus: POST /api/conflicts with bad status returns 400.
func TestPostConflictInvalidStatus(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"id":"ct-bads","point_a":"op-a","point_b":"op-b","status":"pending"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status, got %d", w.Code)
	}
}

// TestPostConflictRequiresIdentity: POST /api/conflicts without identity returns 401.
func TestPostConflictRequiresIdentity(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)

	body := `{"id":"ct-auth","point_a":"op-a","point_b":"op-b"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without identity, got %d", w.Code)
	}
}

// TestPostConflictEmptyIDRejected: POST /api/conflicts with empty id returns 400 (WR-06).
// An empty string is a valid SQLite primary key value but a nonsensical thread ID;
// all threads submitted with id="" would collide on the same row.
func TestPostConflictEmptyIDRejected(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	// id is explicitly empty.
	body := `{"id":"","point_a":"op-a","point_b":"op-b","status":"open"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WR-06: expected 400 for empty id, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestPostConflictMissingIDRejected: POST /api/conflicts with no id field returns 400.
func TestPostConflictMissingIDRejected(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	// id field omitted entirely.
	body := `{"point_a":"op-a","point_b":"op-b","status":"open"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("WR-06: expected 400 for missing id, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestPostConflictDefaultStatus: POST /api/conflicts without status field defaults to open.
func TestPostConflictDefaultStatus(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"id":"ct-default","point_a":"op-a","point_b":"op-b"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with default status, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestGetRatingsStoreError: GET /api/ratings/{pointId} with store error returns 500.
func TestGetRatingsStoreError(t *testing.T) {
	s := &failingStore{getRatingsErr: errors.New("db error")}
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/ratings/some-point", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for store error, got %d", w.Code)
	}
}

// TestGetPointStoreError: GET /api/points/{id} with store error returns 500.
func TestGetPointStoreError(t *testing.T) {
	s := &failingStore{getPointErr: errors.New("db error")}
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/points/some-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for store error, got %d", w.Code)
	}
}

// TestListPointsStoreError: GET /api/points with store error returns 500.
func TestListPointsStoreError(t *testing.T) {
	s := &failingStore{listPointsErr: errors.New("db error")}
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/points", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for store error, got %d", w.Code)
	}
}

// TestSearchPointsStoreError: GET /api/search with store error returns 500.
func TestSearchPointsStoreError(t *testing.T) {
	s := &failingStore{searchPointsErr: errors.New("db error")}
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=anything", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for store error, got %d", w.Code)
	}
}

// TestGetConflictsStoreError: GET /api/conflicts with store error returns 500.
func TestGetConflictsStoreError(t *testing.T) {
	s := &failingStore{getConflictsErr: errors.New("db error")}
	router := api.NewRouter(s, noIdentity)

	req := httptest.NewRequest(http.MethodGet, "/api/conflicts?a=x&b=y", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for store error, got %d", w.Code)
	}
}

// TestPostConflictStoreError: POST /api/conflicts with store error returns 500.
func TestPostConflictStoreError(t *testing.T) {
	s := &failingStore{upsertConflictErr: errors.New("db error")}
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"id":"ct-err","point_a":"a","point_b":"b","status":"open"}`
	req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for store error, got %d", w.Code)
	}
}

// TestPostRatingStoreError: POST /api/ratings with SetRating error returns 500.
func TestPostRatingStoreError(t *testing.T) {
	s := &failingStore{setRatingErr: errors.New("db error")}
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"point_id":"some-point","stars":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for SetRating error, got %d", w.Code)
	}
}

// TestAddSubscriptionStoreError: POST /api/subscriptions with AddSubscription error returns 500.
func TestAddSubscriptionStoreError(t *testing.T) {
	s := &failingStore{addSubscriptionErr: errors.New("db error")}
	router := api.NewRouter(s, fakeIdentity("user-1"))

	body := `{"point_id":"some-point"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for AddSubscription error, got %d", w.Code)
	}
}

// TestRemoveSubscriptionViaBody: DELETE /api/subscriptions with point_id in JSON body.
// Covers the body-decode fallback path in removeSubscription.
func TestRemoveSubscriptionViaBody(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, fakeIdentity("user-1"))

	// Subscribe first
	sub := `{"point_id":"body-point"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(sub))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Delete via body (no query param, chi url param is empty for DELETE /api/subscriptions)
	body := `{"point_id":"body-point"}`
	req2 := httptest.NewRequest(http.MethodDelete, "/api/subscriptions",
		strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Should succeed (200) or at least not 500.
	if w2.Code == http.StatusInternalServerError {
		t.Errorf("unexpected 500 for removeSubscription via body: %s", w2.Body.String())
	}
}

// TestRemoveSubscriptionStoreError: DELETE /api/subscriptions with store RemoveSubscription error → 500.
func TestRemoveSubscriptionStoreError(t *testing.T) {
	// Use a closed DB store to trigger RemoveSubscription error.
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	s := store.New(db)
	db.Close()

	router := api.NewRouter(s, fakeIdentity("user-1"))

	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions?point_id=some-point", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for RemoveSubscription error, got %d", w.Code)
	}
}

// TestGetPointUsesChi: GET /api/points/{id} with chi router returns the point.
// CR-03: verifies chi.URLParam is used (not r.PathValue or URL path slicing).
func TestGetPointUsesChi(t *testing.T) {
	s := newTestStore(t)
	seedStoreWithPoints(t, s, []store.PointEntry{
		{ID: "chi-point", Name: "Chi Point", Curator: "alice", FoundationCompat: `["arch"]`},
	})

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points/chi-point", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var point store.PointEntry
	if err := json.NewDecoder(w.Body).Decode(&point); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if point.ID != "chi-point" {
		t.Errorf("expected id=chi-point, got %q", point.ID)
	}
}

// TestListPointsEmpty: GET /api/points returns [] not null when store is empty.
func TestListPointsEmpty(t *testing.T) {
	s := newTestStore(t)
	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/points", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Body should be {"points":[]} not {"points":null}.
	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), `"points":[]`) && !strings.Contains(string(body), `"points": []`) {
		t.Logf("body: %s", body)
		// Check it's at least parseable as {"points": []}
		var result struct {
			Points []store.PointEntry `json:"points"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if result.Points == nil {
			t.Error("expected empty slice [], got null")
		}
	}
}
