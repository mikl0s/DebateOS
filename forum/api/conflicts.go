// Package api — Conflict thread endpoints (FORM-04).
// GET /api/conflicts?a=X&b=Y  — public, returns conflict threads for a pair.
// POST /api/conflicts           — identity-gated, creates/updates a conflict thread.
//
// Security: patch opinion content lives in Git (only PR URL stored here, FORM-04).
// T-05-15: writes require OAuth session. T-05-16: no code-exec/upload surface.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/mikl0s/debateos/forum/store"
)

// getConflicts handles GET /api/conflicts?a=X&b=Y
// Returns conflict threads for the given point pair (symmetric lookup).
func (h *handlers) getConflicts(w http.ResponseWriter, r *http.Request) {
	a := r.URL.Query().Get("a")
	b := r.URL.Query().Get("b")
	if a == "" || b == "" {
		http.Error(w, "query params a and b are required", http.StatusBadRequest)
		return
	}

	threads, err := h.store.GetConflicts(r.Context(), a, b)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if threads == nil {
		threads = []store.ConflictThread{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conflicts": threads,
	})
}

// postConflictRequest is the JSON body for POST /api/conflicts.
type postConflictRequest struct {
	ID         string `json:"id"`
	PointA     string `json:"point_a"`
	PointB     string `json:"point_b"`
	Status     string `json:"status"`
	PatchPRURL string `json:"patch_pr_url"`
}

// postConflict handles POST /api/conflicts — identity-gated.
// Validates input and delegates to store.UpsertConflictThread.
// Only a GitHub PR URL may be supplied for patch_pr_url (patches live in Git, FORM-04).
func (h *handlers) postConflict(w http.ResponseWriter, r *http.Request) {
	// Require GitHub OAuth identity (T-05-15).
	_, ok := h.requireIdentity(w, r)
	if !ok {
		return
	}

	var req postConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.PointA == "" || req.PointB == "" {
		http.Error(w, "point_a and point_b are required", http.StatusBadRequest)
		return
	}
	if req.Status == "" {
		req.Status = "open"
	}
	if req.Status != "open" && req.Status != "resolved" {
		http.Error(w, "status must be 'open' or 'resolved'", http.StatusBadRequest)
		return
	}

	thread := store.ConflictThread{
		ID:         req.ID,
		PointA:     req.PointA,
		PointB:     req.PointB,
		Status:     req.Status,
		PatchPRURL: req.PatchPRURL,
	}

	if err := h.store.UpsertConflictThread(r.Context(), thread); err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thread)
}
