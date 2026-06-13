// Package api provides the Forum's chi HTTP router.
// It mounts read and write endpoints for search, points, subscriptions, ratings,
// OAuth flow (FORM-03), and conflict threads (FORM-04).
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mikl0s/debateos/forum/store"
)

// IdentityFn is a function that extracts a GitHub user ID from the request.
// It returns (userID, true) if an authenticated identity is present, ("", false) otherwise.
// In production, NewRouterWithOAuth wires a SessionStore-backed IdentityFn.
// In tests, a fake function supplies a fixed userID.
type IdentityFn func(r *http.Request) (userID string, ok bool)

// NewRouter builds and returns the chi router with all forum endpoints mounted.
// store is the backing store; identityFn is used by write endpoints to gate on identity.
// This constructor is retained for test compatibility (05-03 tests inject fakeIdentity directly).
func NewRouter(s store.Store, identityFn IdentityFn) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	h := &handlers{store: s, identity: identityFn}
	mountRoutes(r, h, nil)
	return r
}

// NewRouterWithOAuth builds the chi router wired to a real OAuthProvider and
// SessionStore. The identityFn reads the forum_session cookie via sessions.
// OAuth routes /oauth/login and /oauth/callback are mounted.
// Conflict-thread routes are also mounted here.
func NewRouterWithOAuth(s store.Store, provider OAuthProvider, sessions *SessionStore) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	h := &handlers{store: s, identity: identityFnFromSessions(sessions)}
	oh := &oauthHandlers{provider: provider, sessions: sessions}
	mountRoutes(r, h, oh)
	return r
}

// mountRoutes attaches all routes to mux. oh may be nil (legacy NewRouter path).
func mountRoutes(r *chi.Mux, h *handlers, oh *oauthHandlers) {
	// Public read endpoints
	r.Get("/api/search", h.searchPoints)
	r.Get("/api/points", h.listPoints)
	r.Get("/api/points/{id}", h.getPoint)
	r.Get("/api/ratings/{pointId}", h.getRatings)

	// Write endpoints gated on identity
	r.Post("/api/ratings", h.postRating)
	r.Post("/api/subscriptions", h.addSubscription)
	r.Delete("/api/subscriptions", h.removeSubscription)

	// Conflict thread endpoints (FORM-04)
	r.Get("/api/conflicts", h.getConflicts)
	r.Post("/api/conflicts", h.postConflict)

	// OAuth routes (mounted only when a provider is configured)
	if oh != nil {
		r.Get("/oauth/login", oh.loginHandler)
		r.Get("/oauth/callback", oh.callbackHandler)
	}
}

// handlers holds shared dependencies for all HTTP handlers.
type handlers struct {
	store    store.Store
	identity IdentityFn
}

// requireIdentity extracts the user identity from the request context.
// Returns (userID, true) on success; writes 401 and returns ("", false) if not authenticated.
func (h *handlers) requireIdentity(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := h.identity(r)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: GitHub identity required", http.StatusUnauthorized)
		return "", false
	}
	return userID, true
}
