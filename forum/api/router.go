// Package api provides the Forum's chi HTTP router.
// It mounts read and write endpoints for search, points, subscriptions, and ratings.
// OAuth flow and conflict-thread endpoints are added in plan 05-05.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mikl0s/debateos/forum/store"
)

// IdentityFn is a function that extracts a GitHub user ID from the request.
// It returns (userID, true) if an authenticated identity is present, ("", false) otherwise.
// In production (05-05) this reads from the OAuth session cookie.
// In tests, a fake function supplies a fixed userID.
type IdentityFn func(r *http.Request) (userID string, ok bool)

// NewRouter builds and returns the chi router with all forum endpoints mounted.
// store is the backing store; identityFn is used by write endpoints to gate on identity.
func NewRouter(s store.Store, identityFn IdentityFn) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	h := &handlers{store: s, identity: identityFn}

	// Public read endpoints
	r.Get("/api/search", h.searchPoints)
	r.Get("/api/points", h.listPoints)
	r.Get("/api/points/{id}", h.getPoint)
	r.Get("/api/ratings/{pointId}", h.getRatings)

	// Write endpoints gated on identity
	r.Post("/api/ratings", h.postRating)
	r.Post("/api/subscriptions", h.addSubscription)
	r.Delete("/api/subscriptions", h.removeSubscription)

	return r
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
