package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/forum/api"
	"github.com/mikl0s/debateos/forum/store"
)

// fakeIdentity always returns the same userID — used as the identityFn in tests.
func fakeIdentity(userID string) api.IdentityFn {
	return func(r *http.Request) (string, bool) {
		return userID, true
	}
}

// noIdentity always returns no userID — simulates unauthenticated request.
func noIdentity(r *http.Request) (string, bool) {
	return "", false
}

func seedStore(t *testing.T, s *store.SQLiteStore) {
	t.Helper()
	ctx := context.Background()
	points := []store.PointEntry{
		{ID: "docker-setup", Name: "Docker Setup", Intent: "install docker engine", Curator: "alice", FoundationCompat: `["arch"]`},
		{ID: "vim-config", Name: "Vim Config", Intent: "configure vim editor", Curator: "bob", FoundationCompat: `["arch","debian"]`},
	}
	for _, p := range points {
		if err := s.UpsertPoint(ctx, p); err != nil {
			t.Fatalf("seed UpsertPoint %q: %v", p.ID, err)
		}
	}
}

// TestRatingRequiresIdentity: POST /api/ratings returns 401 without identity, 200 with.
func TestRatingRequiresIdentity(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()
	seedStore(t, s)

	// Without identity — expect 401
	router := api.NewRouter(s, noIdentity)
	body := `{"point_id":"docker-setup","stars":4}`
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("without identity: expected 401, got %d", w.Code)
	}

	// With identity — expect 200
	router2 := api.NewRouter(s, fakeIdentity("github-user-1"))
	req2 := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("with identity: expected 200, got %d (body: %s)", w2.Code, w2.Body.String())
	}
}

// TestSearchEndpoint: GET /api/search?q=docker returns JSON with the matching point.
func TestSearchEndpoint(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()
	seedStore(t, s)

	router := api.NewRouter(s, noIdentity)
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=docker", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var result struct {
		Points []store.PointEntry `json:"points"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result.Points) == 0 {
		t.Error("expected at least 1 point in search results")
	}
	found := false
	for _, p := range result.Points {
		if p.ID == "docker-setup" {
			found = true
		}
	}
	if !found {
		t.Errorf("docker-setup not in results: %+v", result.Points)
	}
}
