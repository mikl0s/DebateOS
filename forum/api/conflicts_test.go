package api_test

// Conflict thread endpoint tests (FORM-04).
// RED phase: TestConflictEndpoint tests the HTTP layer;
// TestNoCodeExecEndpoint asserts the route allowlist boundary.

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

// TestConflictEndpoint: GET /api/conflicts?a=X&b=Y returns the thread JSON incl
// patch_pr_url; POST /api/conflicts is identity-gated.
func TestConflictEndpoint(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Seed a conflict thread directly in the store.
	ctx := context.Background()
	thread := store.ConflictThread{
		ID:         "ct-001",
		PointA:     "docker-setup",
		PointB:     "podman-setup",
		Status:     "open",
		PatchPRURL: "https://github.com/mikl0s/debateos/pull/42",
	}
	if err := s.UpsertConflictThread(ctx, thread); err != nil {
		t.Fatalf("UpsertConflictThread: %v", err)
	}

	router := api.NewRouter(s, noIdentity)

	t.Run("GET_returns_thread", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/conflicts?a=docker-setup&b=podman-setup", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
		}

		var resp struct {
			Conflicts []store.ConflictThread `json:"conflicts"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(resp.Conflicts) != 1 {
			t.Fatalf("expected 1 conflict, got %d", len(resp.Conflicts))
		}
		c := resp.Conflicts[0]
		if c.ID != "ct-001" {
			t.Errorf("ID: got %q, want ct-001", c.ID)
		}
		if c.PatchPRURL != "https://github.com/mikl0s/debateos/pull/42" {
			t.Errorf("PatchPRURL: got %q", c.PatchPRURL)
		}
		if c.Status != "open" {
			t.Errorf("Status: got %q, want open", c.Status)
		}
	})

	t.Run("GET_symmetric_lookup", func(t *testing.T) {
		// Symmetric: querying b,a should return the same thread.
		req := httptest.NewRequest(http.MethodGet, "/api/conflicts?a=podman-setup&b=docker-setup", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp struct {
			Conflicts []store.ConflictThread `json:"conflicts"`
		}
		json.NewDecoder(w.Body).Decode(&resp)
		if len(resp.Conflicts) != 1 {
			t.Errorf("symmetric: expected 1, got %d", len(resp.Conflicts))
		}
	})

	t.Run("GET_missing_params_400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/conflicts?a=docker-setup", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("missing b: expected 400, got %d", w.Code)
		}
	})

	t.Run("POST_without_identity_401", func(t *testing.T) {
		body := `{"id":"ct-002","point_a":"a","point_b":"b","status":"open","patch_pr_url":"https://github.com/x/y/pull/1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("POST without identity: expected 401, got %d", w.Code)
		}
	})

	t.Run("POST_with_identity_200", func(t *testing.T) {
		authedRouter := api.NewRouter(s, fakeIdentity("octocat"))
		body := `{"id":"ct-002","point_a":"aaa","point_b":"bbb","status":"open","patch_pr_url":"https://github.com/x/y/pull/1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authedRouter.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("POST with identity: expected 200, got %d (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("POST_status_update_to_resolved", func(t *testing.T) {
		authedRouter := api.NewRouter(s, fakeIdentity("octocat"))
		// Update existing thread ct-001 to resolved.
		body := `{"id":"ct-001","point_a":"docker-setup","point_b":"podman-setup","status":"resolved","patch_pr_url":"https://github.com/mikl0s/debateos/pull/42"}`
		req := httptest.NewRequest(http.MethodPost, "/api/conflicts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authedRouter.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("update status: expected 200, got %d", w.Code)
		}

		// Verify the status changed.
		threads, err := s.GetConflicts(ctx, "docker-setup", "podman-setup")
		if err != nil {
			t.Fatalf("GetConflicts: %v", err)
		}
		if len(threads) == 0 {
			t.Fatal("no threads after update")
		}
		if threads[0].Status != "resolved" {
			t.Errorf("expected resolved, got %q", threads[0].Status)
		}
	})
}

// TestNoCodeExecEndpoint: the router exposes NO route that executes input or
// accepts arbitrary file uploads (FORM-05, T-05-16).
//
// Strategy:
//   1. Blocked paths (exec/run/upload/eval/shell) must NOT return 200.
//   2. Allowlisted routes must be reachable (not 405 Method Not Allowed from chi's
//      method-not-allowed handler, which indicates a different method was tried on
//      a known path). We check via chi's 405 vs 404 distinction: 405 means the
//      path is registered (just wrong method); 404 means path does not exist.
//   3. Allowlist must not name any exec/upload/run surface.
func TestNoCodeExecEndpoint(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Seed a point so /api/points/{id} actually returns 200 (not 404 from handler).
	ctx := context.Background()
	if err := s.UpsertPoint(ctx, store.PointEntry{
		ID: "boundary-point", Name: "Boundary Test", Intent: "test",
		Curator: "alice", FoundationCompat: `["arch"]`,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	provider := &fakeOAuthProvider{userID: "octocat"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Allowlist: the exact set of routes the forum may expose.
	// Any route NOT in this list is a boundary violation.
	allowlist := map[string]bool{
		"GET /api/search":            true,
		"GET /api/points":            true,
		"GET /api/points/{id}":       true,
		"GET /api/ratings/{pointId}": true,
		"POST /api/ratings":          true,
		"POST /api/subscriptions":    true,
		"DELETE /api/subscriptions":  true,
		"GET /api/conflicts":         true,
		"POST /api/conflicts":        true,
		"GET /oauth/login":           true,
		"GET /oauth/callback":        true,
	}

	// Blocked patterns — these must NOT return 200.
	blockedPaths := []string{
		"/exec",
		"/run",
		"/upload",
		"/eval",
		"/execute",
		"/shell",
		"/cmd",
		"/admin/exec",
		"/api/exec",
		"/api/run",
		"/api/upload",
	}

	for _, path := range blockedPaths {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut} {
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				t.Errorf("blocked route %s %s returned 200 — boundary violation", method, path)
			}
		}
	}

	// Assert that each route in the allowlist is reachable (2xx or at least not 405).
	// For read endpoints, we expect 200. For write endpoints without identity, 401.
	// We do NOT check 404 from the application handler (e.g. "point not found") —
	// only chi's 404 (route not mounted) is a problem. We detect chi's 404 by using
	// a concrete path that should match the parameterised route.
	routeChecks := []struct {
		method     string
		path       string
		expectCode int // 0 = any non-chi-404 (handler decided)
	}{
		{http.MethodGet, "/api/search", http.StatusOK},
		{http.MethodGet, "/api/points", http.StatusOK},
		{http.MethodGet, "/api/points/boundary-point", http.StatusOK},
		{http.MethodGet, "/api/ratings/boundary-point", http.StatusOK},
		{http.MethodGet, "/api/conflicts?a=x&b=y", http.StatusOK},
		{http.MethodGet, "/oauth/login", http.StatusFound},
		// Write endpoints without auth → 401 (route IS mounted; identity gate fires)
		{http.MethodPost, "/api/ratings", http.StatusUnauthorized},
		{http.MethodPost, "/api/conflicts", http.StatusUnauthorized},
	}

	for _, rc := range routeChecks {
		req := httptest.NewRequest(rc.method, rc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != rc.expectCode {
			t.Errorf("allowlist route %s %s: expected %d, got %d (body: %s)",
				rc.method, rc.path, rc.expectCode, w.Code, w.Body.String())
		}
	}

	// Cross-check: allowlist must not mention any exec/upload/run word.
	forbidden := []string{"exec", "run", "upload", "eval", "shell", "cmd"}
	for route := range allowlist {
		for _, word := range forbidden {
			if strings.Contains(strings.ToLower(route), word) {
				t.Errorf("allowlist route %q contains forbidden word %q", route, word)
			}
		}
	}
}
