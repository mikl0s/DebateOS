// Package resolve implements the docs/04 conflict resolution hierarchy for
// DebateOS speeches. It composes the Wave-2 packages (graph toposort,
// hardware eval, patch discovery) and applies the four-rule resolution
// hierarchy, attaching a first-class Explanation to every decision and
// emitting a ResolvedSpeech as deterministic canonical JSON.
//
// Rule hierarchy (docs/04, precise):
//  1. Required beats nice-to-have → required wins; nice-to-have dropped visibly.
//  2. Required-vs-required → hard conflict UNLESS a patch opinion exists.
//  3. Nice-to-have vs nice-to-have → sensible default (first-listed wins).
//  4. Patch opinions override any of the above.
//
// RSLV-01: parse → graph → docs/04 hierarchy → ResolvedSpeech + Explanation.
// RSLV-06: near-total coverage including all 27 EC-NNN scenarios.
package resolve

import (
	"github.com/mikkelraglan/debateos/resolver"
)

// Explanation is a first-class record of why a resolution decision was made.
// Every Applied, Skipped, Dropped, or conflict decision carries one.
// Fields are structured for machine consumption as well as the human-readable Text.
//
// No float64 fields anywhere — canonical-JSON parity safety (T-01-12).
type Explanation struct {
	// Text is the human-readable description of the decision.
	Text string `json:"text"`

	// Rule is the docs/04 rule number or behavioral section that triggered
	// this decision. Values: "rule1", "rule2", "rule3", "rule4",
	// "hardware-skip", "hardware-apply", "ordering", "cycle", "sysctl-collision",
	// "no-conflict".
	Rule string `json:"rule"`

	// OpinionsInvolved lists all opinion IDs that participated in this decision.
	OpinionsInvolved []resolver.OpinionID `json:"opinions_involved,omitempty"`

	// Dropped lists opinion IDs that were removed from the speech by this decision.
	Dropped []resolver.OpinionID `json:"dropped,omitempty"`

	// Kept lists opinion IDs that were retained by this decision.
	Kept []resolver.OpinionID `json:"kept,omitempty"`

	// PatchOffered is the patch opinion ID offered or applied (if any).
	PatchOffered resolver.OpinionID `json:"patch_offered,omitempty"`

	// TrustWarning is set when a sig_level=Never repo is encountered (T-01-10).
	TrustWarning string `json:"trust_warning,omitempty"`

	// AlternativeSuggestion is set on hardware-skip explanations when the
	// composition contains a same-category opinion whose hardware condition
	// evaluates TRUE. It names the best in-composition alternative (RSLV-04 / SC-3).
	// Format: "You declared <predicate>: consider '<Name>' (<ID>) instead."
	// Empty on all non-hardware-skip explanations (omitempty — existing goldens unchanged).
	AlternativeSuggestion string `json:"alternative_suggestion,omitempty"`
}

// ResolvedSpeech is the output of Resolve: the speech with all conflicts
// settled, an explicit install order, and one Explanation per decision.
//
// No maps in the output type (Pitfall 1 guard — no range-over-map in hot
// paths). All slices are sorted deterministically before output.
// No float64 fields — canonical-JSON parity (T-01-12).
type ResolvedSpeech struct {
	// Schema is the version of the resolved speech format (currently 1).
	Schema int `json:"schema"`

	// Foundation is the target foundation string from the input speech.
	Foundation string `json:"foundation"`

	// InstallOrder is the deterministic Kahn topological install order
	// across all applied opinions. Set to nil on cycle error.
	InstallOrder []resolver.OpinionID `json:"install_order,omitempty"`

	// Applied lists opinion IDs that were included in the resolved speech.
	Applied []resolver.OpinionID `json:"applied,omitempty"`

	// Skipped lists opinion IDs whose hardware condition evaluated to false.
	Skipped []resolver.OpinionID `json:"skipped,omitempty"`

	// Dropped lists opinion IDs that were removed by Rule 1 or Rule 3.
	Dropped []resolver.OpinionID `json:"dropped,omitempty"`

	// Explanations is one entry per resolution decision, in the order
	// decisions were made. Every Applied/Skipped/Dropped/conflict has one.
	Explanations []Explanation `json:"explanations"`
}
