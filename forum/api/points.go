package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type subscriptionRequest struct {
	PointID string `json:"point_id"`
}

// addSubscription handles POST /api/subscriptions
// Identity gate: requires authenticated user (T-05-07).
func (h *handlers) addSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireIdentity(w, r)
	if !ok {
		return
	}

	var req subscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.PointID == "" {
		http.Error(w, "point_id required", http.StatusBadRequest)
		return
	}

	if err := h.store.AddSubscription(r.Context(), userID, req.PointID); err != nil {
		http.Error(w, "subscription failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"subscribed": true, "point_id": req.PointID})
}

// removeSubscription handles DELETE /api/subscriptions
// Identity gate: requires authenticated user.
func (h *handlers) removeSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireIdentity(w, r)
	if !ok {
		return
	}

	pointID := chi.URLParam(r, "point_id")
	if pointID == "" {
		// Try query param or body
		pointID = r.URL.Query().Get("point_id")
	}
	if pointID == "" {
		var req subscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			pointID = req.PointID
		}
	}
	if pointID == "" {
		http.Error(w, "point_id required", http.StatusBadRequest)
		return
	}

	if err := h.store.RemoveSubscription(r.Context(), userID, pointID); err != nil {
		http.Error(w, "unsubscribe failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"subscribed": false, "point_id": pointID})
}
