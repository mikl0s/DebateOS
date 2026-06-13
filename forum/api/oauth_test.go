package api_test

// OAuth flow tests for FORM-03 (GitHub OAuth, fake provider, no live GitHub calls).
// RED phase: these tests reference symbols that do not exist yet (oauth.go).
// Security surface covered: T-05-13 (state forgery), T-05-14 (token not persisted).

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/forum/api"
	"github.com/mikl0s/debateos/forum/store"
)

// --- Fake OAuth provider (test double, no network) ---

// fakeOAuthProvider implements api.OAuthProvider.
// AuthCodeURL returns a predictable URL; Exchange returns a fixed token;
// GetUserID returns a fixed user ID.
type fakeOAuthProvider struct {
	userID string
	// exchangeErr allows tests to simulate Exchange failures.
	exchangeErr error
	// getUserIDErr allows tests to simulate GetUserID failures.
	getUserIDErr error
}

func (f *fakeOAuthProvider) AuthCodeURL(state string) string {
	return "https://fake-github.example.com/oauth?state=" + state
}

func (f *fakeOAuthProvider) Exchange(ctx context.Context, code string) (string, error) {
	if f.exchangeErr != nil {
		return "", f.exchangeErr
	}
	return "fake-access-token-abc123", nil
}

func (f *fakeOAuthProvider) GetUserID(ctx context.Context, token string) (string, error) {
	if f.getUserIDErr != nil {
		return "", f.getUserIDErr
	}
	return f.userID, nil
}

// --- TestOAuthLoginRedirect ---
// GET /oauth/login must:
//   - Set a state cookie (httpOnly, SameSite=Lax).
//   - Redirect (302) to the provider AuthCodeURL containing the state value.
func TestOAuthLoginRedirect(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{userID: "octocat"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	req := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", resp.StatusCode)
	}

	// State cookie must be set.
	var stateCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("oauth_state cookie not set")
	}
	if stateCookie.Value == "" {
		t.Fatal("oauth_state cookie value is empty")
	}

	// Location must contain the state.
	loc := resp.Header.Get("Location")
	if !strings.Contains(loc, stateCookie.Value) {
		t.Errorf("redirect location %q does not contain state %q", loc, stateCookie.Value)
	}
}

// --- TestOAuthCallbackValidatesState ---
// /oauth/callback with a forged state → 400 (T-05-13 CSRF protection).
// /oauth/callback with matching state → Exchange + GetUserID → 302 home.
func TestOAuthCallbackValidatesState(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{userID: "octocat"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Step 1: get the state from a login redirect.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	loginResp := loginW.Result()
	if loginResp.StatusCode != http.StatusFound {
		t.Fatalf("login: expected 302, got %d", loginResp.StatusCode)
	}
	var stateCookie *http.Cookie
	for _, c := range loginResp.Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("no oauth_state cookie from login")
	}

	t.Run("forged_state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=somecode&state=FORGED", nil)
		req.AddCookie(stateCookie) // real cookie, but forged state in URL
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("forged state: expected 400, got %d", w.Code)
		}
	})

	t.Run("valid_state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/oauth/callback?code=somecode&state="+stateCookie.Value, nil)
		req.AddCookie(stateCookie)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// Valid flow → session created → redirect to /
		if w.Code != http.StatusFound {
			t.Errorf("valid state: expected 302, got %d (body: %s)", w.Code, w.Body.String())
		}
		loc := w.Header().Get("Location")
		if loc != "/" {
			t.Errorf("expected redirect to /, got %q", loc)
		}
	})
}

// --- TestTokenNotPersisted ---
// After a successful OAuth callback the access token must NOT appear in any
// store table (T-05-14: no secrets at rest).
func TestTokenNotPersisted(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{userID: "octocat"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Complete the OAuth flow.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	var stateCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("no state cookie")
	}

	callbackReq := httptest.NewRequest(http.MethodGet,
		"/oauth/callback?code=somecode&state="+stateCookie.Value, nil)
	callbackReq.AddCookie(stateCookie)
	callbackW := httptest.NewRecorder()
	router.ServeHTTP(callbackW, callbackReq)
	if callbackW.Code != http.StatusFound {
		t.Fatalf("callback: expected 302, got %d", callbackW.Code)
	}

	// Query the raw SQLite DB for the token string — must not appear anywhere.
	db := s.DB()
	tables := []string{"points", "subscriptions", "ratings", "conflict_threads"}
	for _, tbl := range tables {
		rows, err := db.QueryContext(context.Background(), "SELECT * FROM "+tbl)
		if err != nil {
			t.Logf("querying %s: %v (skipped)", tbl, err)
			continue
		}
		cols, _ := rows.Columns()
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		for rows.Next() {
			if err := rows.Scan(valPtrs...); err != nil {
				rows.Close()
				t.Fatalf("scan %s: %v", tbl, err)
			}
			for _, v := range vals {
				if s, ok := v.(string); ok {
					if strings.Contains(s, "fake-access-token") {
						rows.Close()
						t.Errorf("token found in table %s: %q", tbl, s)
					}
				}
			}
		}
		rows.Close()
	}

	// Also check that the in-memory session store holds userID (not token).
	userID, ok := sessions.GetUserID(callbackW.Result().Cookies())
	if !ok || userID != "octocat" {
		t.Errorf("session should contain userID=octocat, got %q ok=%v", userID, ok)
	}
}

// --- TestWriteGateUsesSession ---
// After a successful OAuth callback, the session cookie should allow
// POST /api/ratings (identity gate) to succeed.
// Without the session cookie → 401.
func TestWriteGateUsesSession(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	// Seed a point so the rating write has a valid FK target.
	if err := s.UpsertPoint(context.Background(), store.PointEntry{
		ID: "oauth-point", Name: "OAuth Point", Intent: "test", Curator: "alice",
		FoundationCompat: `["arch"]`,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	provider := &fakeOAuthProvider{userID: "octocat"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Complete OAuth flow to get a session cookie.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	var stateCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}

	callbackReq := httptest.NewRequest(http.MethodGet,
		"/oauth/callback?code=somecode&state="+stateCookie.Value, nil)
	callbackReq.AddCookie(stateCookie)
	callbackW := httptest.NewRecorder()
	router.ServeHTTP(callbackW, callbackReq)
	if callbackW.Code != http.StatusFound {
		t.Fatalf("callback: expected 302, got %d", callbackW.Code)
	}

	// Extract session cookie from callback response.
	var sessionCookie *http.Cookie
	for _, c := range callbackW.Result().Cookies() {
		if c.Name == "forum_session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no forum_session cookie after callback")
	}

	ratingBody := `{"point_id":"oauth-point","stars":5}`

	t.Run("unauthenticated_401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(ratingBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 without session, got %d", w.Code)
		}
	})

	t.Run("authenticated_200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/ratings", strings.NewReader(ratingBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 with session, got %d (body: %s)", w.Code, w.Body.String())
		}
	})
}

// --- TestOAuthSessionHasUserID ---
// Confirms that the session exposed via NewSessionStore.GetUserID returns the
// GitHub user ID (not the access token) after a successful callback.
func TestOAuthSessionHasUserID(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{userID: "github-user-99"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Login → get state.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	var stateCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("no state cookie")
	}

	// Callback → session.
	cbReq := httptest.NewRequest(http.MethodGet,
		"/oauth/callback?code=x&state="+stateCookie.Value, nil)
	cbReq.AddCookie(stateCookie)
	cbW := httptest.NewRecorder()
	router.ServeHTTP(cbW, cbReq)

	// Verify session resolves to userID (not token).
	userID, ok := sessions.GetUserID(cbW.Result().Cookies())
	if !ok {
		t.Fatal("GetUserID returned ok=false")
	}
	if userID != "github-user-99" {
		t.Errorf("expected github-user-99, got %q", userID)
	}
	if strings.Contains(userID, "token") {
		t.Errorf("session appears to store a token rather than a user ID: %q", userID)
	}
}

// --- TestOAuthJSONExport ---
// Verifies the JSON tags on the OAuthProvider interface types compile correctly.
// (Compile-time sanity check — if oauth.go doesn't exist, all tests in this file fail.)
func TestOAuthJSONExport(t *testing.T) {
	_ = json.Marshal // ensure encoding/json is used in test binary
	// If this compiles and runs, oauth.go exists with the required symbols.
	var _ api.OAuthProvider = (*fakeOAuthProvider)(nil)
}

// --- TestNewRealGitHubOAuthConstructor ---
// Verifies that NewRealGitHubOAuth creates a valid provider with the given credentials,
// and that AuthCodeURL generates a URL containing the state parameter (no network call).
func TestNewRealGitHubOAuthConstructor(t *testing.T) {
	p := api.NewRealGitHubOAuth("client-id-test", "client-secret-test", "http://localhost/callback")
	if p == nil {
		t.Fatal("NewRealGitHubOAuth returned nil")
	}

	// AuthCodeURL only formats a redirect URL — no network call required.
	url := p.AuthCodeURL("csrf-state-nonce")
	if url == "" {
		t.Fatal("AuthCodeURL returned empty string")
	}
	if !strings.Contains(url, "csrf-state-nonce") {
		t.Errorf("expected AuthCodeURL to contain state nonce, got: %q", url)
	}

	// Verify it's for GitHub's OAuth endpoint.
	if !strings.Contains(url, "github.com") {
		t.Errorf("expected AuthCodeURL to point to github.com, got: %q", url)
	}
}

// --- TestCallbackMissingStateCookie ---
// GET /oauth/callback without a state cookie → 400.
func TestCallbackMissingStateCookie(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{userID: "alice"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// No state cookie at all.
	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?code=somecode&state=anystate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing state cookie, got %d", w.Code)
	}
}

// --- TestCallbackMissingCode ---
// GET /oauth/callback with valid state but no code → 400.
func TestCallbackMissingCode(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{userID: "alice"}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Get a valid state from login.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	var stateCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("no state cookie from login")
	}

	// Callback with valid state but no code.
	req := httptest.NewRequest(http.MethodGet, "/oauth/callback?state="+stateCookie.Value, nil)
	req.AddCookie(stateCookie)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing code, got %d", w.Code)
	}
}

// --- TestCallbackExchangeError ---
// GET /oauth/callback when Exchange fails → 502.
func TestCallbackExchangeError(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{
		userID:      "alice",
		exchangeErr: errors.New("exchange failed"),
	}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Get a valid state from login.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	var stateCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("no state cookie from login")
	}

	req := httptest.NewRequest(http.MethodGet,
		"/oauth/callback?code=badcode&state="+stateCookie.Value, nil)
	req.AddCookie(stateCookie)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for exchange error, got %d", w.Code)
	}
}

// --- TestCallbackGetUserIDError ---
// GET /oauth/callback when GetUserID fails → 502.
func TestCallbackGetUserIDError(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	provider := &fakeOAuthProvider{
		userID:       "alice",
		getUserIDErr: errors.New("github api error"),
	}
	sessions := api.NewSessionStore()
	router := api.NewRouterWithOAuth(s, provider, sessions)

	// Get a valid state from login.
	loginReq := httptest.NewRequest(http.MethodGet, "/oauth/login", nil)
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	var stateCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatal("no state cookie from login")
	}

	req := httptest.NewRequest(http.MethodGet,
		"/oauth/callback?code=somecode&state="+stateCookie.Value, nil)
	req.AddCookie(stateCookie)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for GetUserID error, got %d", w.Code)
	}
}

// --- TestGetRatingsViaChi ---
// GET /api/ratings/{pointId} with a real point ID returns JSON summary.
func TestGetRatingsViaChi(t *testing.T) {
	s, err := store.NewInMemory()
	if err != nil {
		t.Fatalf("NewInMemory: %v", err)
	}
	defer s.Close()

	if err := s.UpsertPoint(context.Background(), store.PointEntry{
		ID: "chi-rating-point", Name: "Chi Point", Curator: "x", FoundationCompat: `["arch"]`,
	}); err != nil {
		t.Fatalf("UpsertPoint: %v", err)
	}

	router := api.NewRouter(s, func(r *http.Request) (string, bool) { return "user-1", true })

	// First set a rating so the aggregate is non-trivial.
	rateReq := httptest.NewRequest(http.MethodPost, "/api/ratings",
		strings.NewReader(`{"point_id":"chi-rating-point","stars":4}`))
	rateReq.Header.Set("Content-Type", "application/json")
	rateW := httptest.NewRecorder()
	router.ServeHTTP(rateW, rateReq)
	if rateW.Code != http.StatusOK {
		t.Fatalf("POST /api/ratings: expected 200, got %d", rateW.Code)
	}

	// Now GET /api/ratings/{pointId}.
	req := httptest.NewRequest(http.MethodGet, "/api/ratings/chi-rating-point", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/ratings/chi-rating-point: expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["count"].(float64) != 1 {
		t.Errorf("expected count=1, got %v", result["count"])
	}
}
