package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mikl0s/debateos/forum/store"
)

// searchPoints handles GET /api/search?q=...&foundation=...&limit=...
// Returns JSON: {"points": [...]}
// Content served as JSON text — Svelte auto-escapes on render (T-05-08 stored-XSS mitigation).
func (h *handlers) searchPoints(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	foundation := r.URL.Query().Get("foundation")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	results, err := h.store.SearchPoints(r.Context(), q, foundation, limit)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []store.PointEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"points": results})
}

// listPoints handles GET /api/points?limit=...&offset=...
func (h *handlers) listPoints(w http.ResponseWriter, r *http.Request) {
	limit, offset := 20, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	points, err := h.store.ListPoints(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []store.PointEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"points": points})
}

// getPoint handles GET /api/points/{id}
func (h *handlers) getPoint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		// chi path param fallback
		id = r.URL.Path[len("/api/points/"):]
	}

	point, err := h.store.GetPoint(r.Context(), id)
	if err != nil {
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if point == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(point)
}
