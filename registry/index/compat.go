package index

import (
	"sort"

	resolver "github.com/mikl0s/debateos/resolver"
)

// FoundationCompat records whether a given foundation (e.g. "arch", "debian")
// can effectuate all translator_capabilities required by a point's member
// opinions, as derived from that translator's capabilities.json.
//
// Missing lists the unsatisfied tokens when Compatible is false; the slice is
// sorted lexically for deterministic JSON output.
type FoundationCompat struct {
	Foundation string   `json:"foundation"`
	Compatible bool     `json:"compatible"`
	Missing    []string `json:"missing_capabilities,omitempty"`
}

// ComputeCompat evaluates foundation compatibility for a single point by
// collecting the union of translator_capabilities across all member opinions
// that appear in the opinions map, then checking each foundation in caps.
//
// Algorithm (Pattern 6, 05-RESEARCH.md):
//  1. For each opinion in point.Members (by ID), look it up in opinions and
//     collect its TranslatorCapabilities tokens into a required set.
//  2. For each foundation in caps: a token is satisfied iff it is in that
//     foundation's capability slice.
//  3. Compatible = all required tokens are satisfied.
//  4. Missing = sorted slice of unsatisfied tokens.
//
// The returned slice is sorted by Foundation for deterministic output.
// Foundations in caps are always iterated in sorted order.
func ComputeCompat(
	point resolver.Point,
	opinions map[resolver.OpinionID]resolver.Opinion,
	caps map[string][]string,
) []FoundationCompat {
	// Step 1: collect required capability tokens (union over all member opinions).
	required := make(map[string]struct{})
	for _, member := range point.Members {
		op, ok := opinions[member.ID]
		if !ok {
			continue
		}
		for _, cap := range op.TranslatorCapabilities {
			required[cap] = struct{}{}
		}
	}

	// Step 2: sort foundations for deterministic iteration order.
	foundations := make([]string, 0, len(caps))
	for f := range caps {
		foundations = append(foundations, f)
	}
	sort.Strings(foundations)

	result := make([]FoundationCompat, 0, len(foundations))
	for _, foundation := range foundations {
		capSet := make(map[string]struct{}, len(caps[foundation]))
		for _, c := range caps[foundation] {
			capSet[c] = struct{}{}
		}

		var missing []string
		for tok := range required {
			if _, ok := capSet[tok]; !ok {
				missing = append(missing, tok)
			}
		}
		sort.Strings(missing)

		fc := FoundationCompat{
			Foundation: foundation,
			Compatible: len(missing) == 0,
		}
		if len(missing) > 0 {
			fc.Missing = missing
		}
		result = append(result, fc)
	}

	return result
}
