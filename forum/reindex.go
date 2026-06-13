// Package forum provides the Forum's cross-cutting operations.
// Reindex rebuilds the Forum's SQLite store from the static registry index (FORM-05).
//
// Security invariant (D13, FORM-05):
//   - Total DB loss → re-index from registry index → recovered.
//   - Reindex is idempotent: running it twice produces the same result (UpsertPoint is ON CONFLICT DO UPDATE).
//   - Source of truth is the registry index (a Git-derived static artifact, not user input).
//   - No code is executed; no uploads are accepted; only indexed metadata is stored.
package forum

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mikl0s/debateos/forum/store"
	"github.com/mikl0s/debateos/registry/index"
)

// Reindex rebuilds the Forum's store from the given RegistryIndex.
// It iterates every PointEntry in idx, converts it to a store.PointEntry, and
// calls UpsertPoint (idempotent via ON CONFLICT DO UPDATE in SQLite).
//
// After all points are upserted, the FTS5 external-content index is rebuilt so
// that SearchPoints returns fresh results.
//
// This is the tested "rebuildable" path required by FORM-05: after a total DB
// loss, run Reindex with the latest registry index.json to recover all point
// metadata. Only ephemeral social state (ratings, subscriptions) is lost.
func Reindex(ctx context.Context, s store.Store, idx *index.RegistryIndex) error {
	if idx == nil {
		return fmt.Errorf("Reindex: registry index is nil")
	}

	// Use UpsertPointBatch (no per-insert FTS rebuild) so N points require
	// only 1 FTS rebuild at the end, not N+1 (IN-01).
	for i, p := range idx.Points {
		sp, err := registryPointToStore(p)
		if err != nil {
			return fmt.Errorf("Reindex: point[%d] %q: %w", i, p.ID, err)
		}
		if err := s.UpsertPointBatch(ctx, sp); err != nil {
			return fmt.Errorf("Reindex: UpsertPoint %q: %w", p.ID, err)
		}
	}

	// Single FTS5 rebuild after all upserts — O(1) instead of O(N).
	if err := s.Reindex(ctx); err != nil {
		return fmt.Errorf("Reindex: FTS rebuild: %w", err)
	}

	return nil
}

// registryPointToStore converts a registry/index.PointEntry to a store.PointEntry.
//
// FoundationCompat: the store stores a JSON array of foundation IDs (e.g. ["arch","debian"]).
// We encode the list of foundation names from the registry's []FoundationCompat slice
// (all entries, both compatible and incompatible, so the forum can display the full picture).
//
// Tags: the store stores a JSON array of tag strings.
func registryPointToStore(p index.PointEntry) (store.PointEntry, error) {
	// Encode all FoundationCompat structs as a JSON array for the store.
	// The store's SearchPoints does a substring filter on this field.
	// We store the full FoundationCompat array so the UI can show compatibility details.
	fcJSON, err := json.Marshal(p.FoundationCompat)
	if err != nil {
		return store.PointEntry{}, fmt.Errorf("marshal foundation_compat: %w", err)
	}

	// Encode tags as a JSON array.
	tagsJSON, err := json.Marshal(p.Tags)
	if err != nil {
		return store.PointEntry{}, fmt.Errorf("marshal tags: %w", err)
	}

	return store.PointEntry{
		ID:               p.ID,
		Name:             p.Name,
		Intent:           p.Intent,
		Curator:          p.Curator,
		FoundationCompat: string(fcJSON),
		CommitDate:       p.CommitDate,
		Tags:             string(tagsJSON),
	}, nil
}
