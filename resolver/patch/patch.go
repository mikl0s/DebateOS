// Package patch provides first-class patch opinion discovery for conflict
// pairs (SR-003/RSLV-02). The resolver calls FindPatch to check whether a
// community-contributed patch opinion exists that can make two otherwise-
// conflicting opinions coexist.
//
// Discovery model (docs/04 "Patch opinions — first-class"):
//   - Patch opinions are discoverable: attached to a conflict pair in opinion
//     metadata; the resolver offers them automatically.
//   - The resolver scans the known_patches list on BOTH conflicting opinions.
//   - For each referenced patch ID, it looks up the patch opinion in the
//     provided opinion set.
//   - Results are deterministic: candidates are sorted by PatchID before
//     returning the first match, so composition order of opinions does not
//     affect the result.
//
// Security note (T-01-09): Phase 1 trusts opinion metadata as data. Patch
// application correctness is composition-review responsibility; the explanation
// surfaces the patch for user confirmation (01-04).
package patch

import (
	"sort"

	"github.com/mikl0s/debateos/resolver"
)

// PatchOffer is returned by FindPatch when a patch opinion exists that can
// resolve the conflict between two opinions. PatchID is the ID of the patch
// opinion; Pair holds both conflicting IDs in canonical (sorted) order so the
// offer is identical regardless of argument order.
type PatchOffer struct {
	// PatchID is the OpinionID of the discovered patch opinion.
	PatchID resolver.OpinionID

	// Pair holds the two conflicting opinion IDs in sorted (canonical) order.
	// The order is deterministic: Pair[0] < Pair[1] lexicographically.
	Pair [2]resolver.OpinionID
}

// FindPatch searches the provided opinion slice for a patch opinion that
// resolves the conflict between opinion a and opinion b. It returns a
// non-nil *PatchOffer when exactly one (or more) qualifying patches exist,
// choosing the lexicographically smallest PatchID for determinism. It returns
// nil when no qualifying patch is found.
//
// The search is order-independent: FindPatch(a, b, ...) ≡ FindPatch(b, a, ...).
//
// Discovery algorithm:
//  1. Build an index from OpinionID → *Opinion for O(1) lookup.
//  2. Build the canonical pair {min(a,b), max(a,b)}.
//  3. For each of the two conflicting opinions, scan their KnownPatches lists.
//  4. For each PatchRef, look up the patch opinion by ID in the index.
//  5. Verify the looked-up opinion has category == "patch" (guards against
//     a non-patch opinion being referenced in known_patches by mistake).
//  6. Collect all qualifying PatchIDs, sort, return the first one.
func FindPatch(a, b resolver.OpinionID, opinions []resolver.Opinion) *PatchOffer {
	// Build index.
	index := make(map[resolver.OpinionID]*resolver.Opinion, len(opinions))
	for i := range opinions {
		index[opinions[i].ID] = &opinions[i]
	}

	// Canonical sorted pair for the returned PatchOffer.
	canonicalPair := sortedPair(a, b)

	// Collect qualifying patch IDs from both conflicting opinions' known_patches.
	seen := make(map[resolver.OpinionID]bool)
	var candidates []resolver.OpinionID

	for _, conflictingID := range []resolver.OpinionID{a, b} {
		op, ok := index[conflictingID]
		if !ok {
			continue
		}
		for _, ref := range op.KnownPatches {
			if seen[ref.ID] {
				continue
			}
			seen[ref.ID] = true
			// Verify the referenced opinion exists and is actually a patch.
			patchOp, exists := index[ref.ID]
			if !exists {
				continue
			}
			if patchOp.Category != "patch" {
				continue
			}
			candidates = append(candidates, ref.ID)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort candidates for deterministic selection (smallest ID wins).
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i] < candidates[j]
	})

	return &PatchOffer{
		PatchID: candidates[0],
		Pair:    canonicalPair,
	}
}

// sortedPair returns the two IDs in canonical lexicographic order so that
// the PatchOffer.Pair field is identical regardless of argument order.
func sortedPair(a, b resolver.OpinionID) [2]resolver.OpinionID {
	if a <= b {
		return [2]resolver.OpinionID{a, b}
	}
	return [2]resolver.OpinionID{b, a}
}
