package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ratingRequest struct {
	PointID string `json:"point_id"`
	Stars   int    `json:"stars"`
}

// postRating handles POST /api/ratings
// Identity gate: returns 401 if no authenticated user identity (T-05-07 elevation-of-privilege mitigation).
// Validates stars in [1,5] (V5 input validation).
func (h *handlers) postRating(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireIdentity(w, r)
	if !ok {
		return
	}

	var req ratingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.PointID == "" {
		http.Error(w, "point_id required", http.StatusBadRequest)
		return
	}
	if req.Stars < 1 || req.Stars > 5 {
		http.Error(w, "stars must be between 1 and 5", http.StatusBadRequest)
		return
	}

	if err := h.store.SetRating(r.Context(), userID, req.PointID, req.Stars); err != nil {
		http.Error(w, "rating failed", http.StatusInternalServerError)
		return
	}

	summary, err := h.store.GetRatings(r.Context(), req.PointID)
	if err != nil {
		http.Error(w, "failed to read aggregate", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"point_id": req.PointID,
		"avg":      summary.Avg,
		"count":    summary.Count,
	})
}

// getRatings handles GET /api/ratings/{pointId}
func (h *handlers) getRatings(w http.ResponseWriter, r *http.Request) {
	pointID := chi.URLParam(r, "pointId")
	if pointID == "" {
		http.Error(w, "pointId required", http.StatusBadRequest)
		return
	}

	summary, err := h.store.GetRatings(r.Context(), pointID)
	if err != nil {
		http.Error(w, "failed to read ratings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"point_id": pointID,
		"avg":      summary.Avg,
		"count":    summary.Count,
	})
}
