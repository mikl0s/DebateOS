package resolve

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mikkelraglan/debateos/resolver"
	"github.com/mikkelraglan/debateos/resolver/graph"
	"github.com/mikkelraglan/debateos/resolver/hardware"
	"github.com/mikkelraglan/debateos/resolver/patch"
)

// Resolve applies the docs/04 four-rule resolution hierarchy to the given
// opinions (assembled from a parsed Speech) against a declared HardwareProfile.
// It returns a *ResolvedSpeech with InstallOrder, Applied/Skipped/Dropped slices
// and one Explanation per resolution decision, or a non-nil error on hard
// conflict or ordering cycle.
//
// Resolution steps (in order):
//  1. Hardware-conditional evaluation (EvalCondition): skip/apply each
//     hardware-gated opinion; emit "Skipped (hardware condition false)" or
//     "Applied (hardware condition true)" explanation.
//  2. Pairwise conflict detection: for each declared conflict pair (opinion.Conflicts)
//     and per-key sysctl collision (SR-016): apply docs/04 rules.
//  3. docs/04 hierarchy per conflict:
//     Rule 1: required beats nice-to-have → drop nice-to-have visibly.
//     Rule 2: required-vs-required → hard conflict UNLESS patch exists.
//     Rule 3: nice-vs-nice → pick first-listed, drop second visibly.
//     Rule 4: patch opinion explicitly present overrides 1-3.
//  4. Topological sort via graph.BuildGraph + graph.TopoSort: cycle = hard error.
//  5. Assemble ResolvedSpeech (Applied/Skipped/Dropped all sorted deterministically).
//
// On hard conflict or cycle error, Resolve returns a partial *ResolvedSpeech
// (non-nil) with the error Explanation attached so callers can display the
// conflict text, AND a non-nil error.
func Resolve(speech *resolver.Speech, opinions []resolver.Opinion, hw hardware.HardwareProfile) (*ResolvedSpeech, error) {
	rs := &ResolvedSpeech{
		Schema:     1,
		Foundation: speech.Foundation,
	}

	// ── Step 0: Build opinion index ─────────────────────────────────────────
	// index maps OpinionID → *Opinion for O(1) lookups. Not iterated
	// directly in output paths (Pitfall 1 guard).
	index := make(map[resolver.OpinionID]*resolver.Opinion, len(opinions))
	for i := range opinions {
		index[opinions[i].ID] = &opinions[i]
	}

	// ── Step 1: Hardware-conditional evaluation ─────────────────────────────
	// Iterate in stable input order (not map order) for determinism.
	skipped := make(map[resolver.OpinionID]bool)
	for _, op := range opinions {
		if op.HardwareCondition == nil {
			continue
		}
		ok, err := hardware.EvalCondition(*op.HardwareCondition, hw)
		if err != nil {
			// Malformed hardware condition — treat as skip with explanation.
			skipped[op.ID] = true
			rs.Explanations = append(rs.Explanations, Explanation{
				Text:             fmt.Sprintf("Skipped (hardware condition error): %s — %v", op.ID, err),
				Rule:             "hardware-skip",
				OpinionsInvolved: []resolver.OpinionID{op.ID},
			})
			continue
		}
		if !ok {
			skipped[op.ID] = true
			rs.Explanations = append(rs.Explanations, Explanation{
				Text:             fmt.Sprintf("Skipped (hardware condition false): %s — hardware condition not met for declared hardware profile.", op.ID),
				Rule:             "hardware-skip",
				OpinionsInvolved: []resolver.OpinionID{op.ID},
			})
		} else {
			// Hardware condition true — emit Apply explanation.
			// Check for sig_level=Never custom repos (T-01-10).
			trustWarn := collectTrustWarning(op)
			text := fmt.Sprintf("Applied (hardware condition true): %s — hardware condition matched declared hardware profile.", op.ID)
			if trustWarn != "" {
				text += " " + trustWarn
			}
			rs.Explanations = append(rs.Explanations, Explanation{
				Text:             text,
				Rule:             "hardware-apply",
				OpinionsInvolved: []resolver.OpinionID{op.ID},
				Kept:             []resolver.OpinionID{op.ID},
				TrustWarning:     trustWarn,
			})
		}
	}

	// ── Step 2: Build active opinion set (non-skipped) ─────────────────────
	// active is the set of opinion IDs eligible for conflict resolution.
	active := make(map[resolver.OpinionID]bool, len(opinions))
	for _, op := range opinions {
		if !skipped[op.ID] {
			active[op.ID] = true
		}
	}

	// dropped tracks opinions removed by Rule 1 or Rule 3.
	dropped := make(map[resolver.OpinionID]bool)

	// ── Step 3: Sysctl key collision detection (SR-016) ────────────────────
	// For any two active opinions that both declare the same sysctl key, emit
	// a collision error. This is separate from the Conflicts declaration.
	if err := detectSysctlCollisions(opinions, active, dropped, rs, index); err != nil {
		return rs, err
	}

	// ── Step 4: Pairwise conflict resolution ───────────────────────────────
	// Process each active opinion's Conflicts list.
	// We collect all (a,b) pairs and resolve each once (canonical sorted pair).
	processed := make(map[[2]resolver.OpinionID]bool)
	// sortedIDs gives stable iteration order for the outer loop (Pitfall 1).
	sortedIDs := sortedActiveIDs(opinions, active, dropped)

	var hardErr error
	for _, aID := range sortedIDs {
		if dropped[aID] {
			continue
		}
		op := index[aID]
		if op == nil {
			continue
		}
		for _, conflictRef := range op.Conflicts {
			bID := conflictRef.ID
			if !active[bID] || dropped[bID] {
				continue
			}
			// Canonical pair for deduplication.
			pair := canonicalPair(aID, bID)
			if processed[pair] {
				continue
			}
			processed[pair] = true

			opA := index[aID]
			opB := index[bID]
			if opA == nil || opB == nil {
				continue
			}

			if err := resolveConflict(opA, opB, opinions, active, dropped, rs, index); err != nil {
				hardErr = err
				// Continue to collect more conflicts before returning.
			}
		}
	}

	if hardErr != nil {
		return rs, hardErr
	}

	// ── Step 4b: Repo ordering explanations (EC-010/EC-011) ────────────────
	// When multiple active custom-repo opinions are present (and no conflict
	// was declared between them), emit a repo ordering/priority explanation
	// noting the resolved order.
	emitRepoOrderingExplanations(opinions, active, dropped, rs)

	// ── Step 5: Emit "no conflict" explanations for clean opinions ──────────
	// Every active, non-dropped opinion that received no conflict explanation
	// gets a "No conflict" note (satisfies EC-020..EC-023, EC-041..EC-052 etc.)
	explainedOps := make(map[resolver.OpinionID]bool)
	for _, ex := range rs.Explanations {
		for _, id := range ex.OpinionsInvolved {
			explainedOps[id] = true
		}
		for _, id := range ex.Kept {
			explainedOps[id] = true
		}
		for _, id := range ex.Dropped {
			explainedOps[id] = true
		}
	}
	// Emit "No conflict" for any applied opinion not already explained.
	noConflictIDs := sortedActiveIDs(opinions, active, dropped)
	for _, id := range noConflictIDs {
		if !explainedOps[id] {
			rs.Explanations = append(rs.Explanations, Explanation{
				Text:             fmt.Sprintf("No conflict: %s applied.", id),
				Rule:             "no-conflict",
				OpinionsInvolved: []resolver.OpinionID{id},
				Kept:             []resolver.OpinionID{id},
			})
		}
	}

	// ── Step 6: Topological sort ────────────────────────────────────────────
	// Build the graph only over active, non-dropped opinions.
	var appliedOpinions []resolver.Opinion
	for _, op := range opinions {
		if active[op.ID] && !dropped[op.ID] {
			appliedOpinions = append(appliedOpinions, op)
		}
	}

	g, err := graph.BuildGraph(appliedOpinions)
	if err != nil {
		return rs, fmt.Errorf("Resolve: build graph: %w", err)
	}
	order, cycleIDs, err := graph.TopoSort(g)
	if err != nil {
		// Cycle detected — hard error naming the offending opinions.
		cycleText := fmt.Sprintf("Cycle detected in install ordering: %v — remove one of the ordering constraints or introduce an intermediate opinion to break the cycle", cycleIDs)
		rs.Explanations = append(rs.Explanations, Explanation{
			Text:             cycleText,
			Rule:             "cycle",
			OpinionsInvolved: cycleIDs,
		})
		return rs, fmt.Errorf("Cycle detected: %v", cycleIDs)
	}

	// Filter install order to only include opinions from the original input slice
	// (phantom nodes from the graph builder should not appear in output).
	var filteredOrder []resolver.OpinionID
	for _, id := range order {
		if active[id] && !dropped[id] {
			filteredOrder = append(filteredOrder, id)
		}
	}
	rs.InstallOrder = filteredOrder

	// Emit install order explanation if ordering constraints were present.
	if len(filteredOrder) > 1 {
		hasOrdering := false
		for _, op := range appliedOpinions {
			if op.Ordering != nil || len(op.DependsOn) > 0 {
				hasOrdering = true
				break
			}
		}
		if hasOrdering {
			rs.Explanations = append(rs.Explanations, Explanation{
				Text:             fmt.Sprintf("Install order (topological sort): %v — deterministic and reproducible.", filteredOrder),
				Rule:             "ordering",
				OpinionsInvolved: filteredOrder,
			})
		}
	}

	// ── Step 7: Assemble output slices (sorted deterministically) ───────────
	var applied, skippedSlice, droppedSlice []resolver.OpinionID
	for _, op := range opinions {
		switch {
		case skipped[op.ID]:
			skippedSlice = append(skippedSlice, op.ID)
		case dropped[op.ID]:
			droppedSlice = append(droppedSlice, op.ID)
		default:
			applied = append(applied, op.ID)
		}
	}
	// Use input order for Applied (same order as filteredOrder / opinions slice);
	// sort Dropped and Skipped for determinism.
	sort.Slice(droppedSlice, func(i, j int) bool { return droppedSlice[i] < droppedSlice[j] })
	sort.Slice(skippedSlice, func(i, j int) bool { return skippedSlice[i] < skippedSlice[j] })

	rs.Applied = applied
	rs.Skipped = skippedSlice
	rs.Dropped = droppedSlice

	return rs, nil
}

// ─── conflict resolution ───────────────────────────────────────────────────

// resolveConflict applies the docs/04 four-rule hierarchy to a single pair (opA, opB).
// It mutates dropped, active, and appends to rs.Explanations.
// Returns a non-nil error only for Rule 2 hard conflicts (no patch available).
func resolveConflict(
	opA, opB *resolver.Opinion,
	allOpinions []resolver.Opinion,
	active map[resolver.OpinionID]bool,
	dropped map[resolver.OpinionID]bool,
	rs *ResolvedSpeech,
	index map[resolver.OpinionID]*resolver.Opinion,
) error {
	aReq := opA.Status == resolver.StatusRequired
	bReq := opB.Status == resolver.StatusRequired

	// ── Rule 4 (checked first): patch overrides hierarchy ─────────────────
	// If a patch opinion is present in the active set for this pair, apply it
	// and skip Rules 1-3.
	patchOffer := patch.FindPatch(opA.ID, opB.ID, allOpinions)
	if patchOffer != nil && active[patchOffer.PatchID] && !dropped[patchOffer.PatchID] {
		rs.Explanations = append(rs.Explanations, Explanation{
			Text: fmt.Sprintf(
				"Required-vs-required conflict on %s and %s resolved by patch: %s combines both opinions.",
				opA.ID, opB.ID, patchOffer.PatchID,
			),
			Rule:             "rule4",
			OpinionsInvolved: []resolver.OpinionID{opA.ID, opB.ID, patchOffer.PatchID},
			Kept:             []resolver.OpinionID{opA.ID, opB.ID, patchOffer.PatchID},
			PatchOffered:     patchOffer.PatchID,
		})
		return nil
	}

	// ── Rule 1: required beats nice-to-have ────────────────────────────────
	if aReq && !bReq {
		dropped[opB.ID] = true
		rs.Explanations = append(rs.Explanations, Explanation{
			Text: fmt.Sprintf(
				"Required beats nice-to-have: %s (required) wins over %s (nice-to-have). %s has been dropped.",
				opA.ID, opB.ID, opB.ID,
			),
			Rule:             "rule1",
			OpinionsInvolved: []resolver.OpinionID{opA.ID, opB.ID},
			Dropped:          []resolver.OpinionID{opB.ID},
			Kept:             []resolver.OpinionID{opA.ID},
		})
		return nil
	}
	if bReq && !aReq {
		dropped[opA.ID] = true
		rs.Explanations = append(rs.Explanations, Explanation{
			Text: fmt.Sprintf(
				"Required beats nice-to-have: %s (required) wins over %s (nice-to-have). %s has been dropped.",
				opB.ID, opA.ID, opA.ID,
			),
			Rule:             "rule1",
			OpinionsInvolved: []resolver.OpinionID{opA.ID, opB.ID},
			Dropped:          []resolver.OpinionID{opA.ID},
			Kept:             []resolver.OpinionID{opB.ID},
		})
		return nil
	}

	// ── Rule 2: required-vs-required hard conflict ─────────────────────────
	if aReq && bReq {
		// Check if a patch opinion exists (even if not in the active set — the
		// patch may not have been added to the speech yet).
		patchExist := patch.FindPatch(opA.ID, opB.ID, allOpinions)
		var patchText string
		var patchID resolver.OpinionID
		if patchExist != nil {
			patchText = fmt.Sprintf(" A patch opinion is available: %s.", patchExist.PatchID)
			patchID = patchExist.PatchID
		}
		text := fmt.Sprintf(
			"Hard conflict: %s (required) and %s (required) cannot coexist. You must drop one or provide a patch opinion.%s",
			opA.ID, opB.ID, patchText,
		)
		rs.Explanations = append(rs.Explanations, Explanation{
			Text:             text,
			Rule:             "rule2",
			OpinionsInvolved: []resolver.OpinionID{opA.ID, opB.ID},
			PatchOffered:     patchID,
		})
		return fmt.Errorf("Hard conflict: %s and %s are both required and cannot coexist", opA.ID, opB.ID)
	}

	// ── Rule 3: nice-to-have vs nice-to-have ──────────────────────────────
	// Pick the first-listed (opA, as the lower-ID in canonical sorted pair
	// since we iterate in sorted order). opA is the "winner" / "default".
	dropped[opB.ID] = true
	rs.Explanations = append(rs.Explanations, Explanation{
		Text: fmt.Sprintf(
			"Nice-to-have conflict: %s and %s are both nice-to-have and declare mutual conflicts. Default selected: %s (first-listed / Omarchy default). %s dropped.",
			opA.ID, opB.ID, opA.ID, opB.ID,
		),
		Rule:             "rule3",
		OpinionsInvolved: []resolver.OpinionID{opA.ID, opB.ID},
		Dropped:          []resolver.OpinionID{opB.ID},
		Kept:             []resolver.OpinionID{opA.ID},
	})
	return nil
}

// ─── repo ordering explanations ───────────────────────────────────────────

// emitRepoOrderingExplanations detects when two or more active custom-repo
// opinions are present without a declared conflict between them. It emits
// a "Repo ordering" or "Repo priority" explanation noting the resolved order.
// This satisfies EC-010 ("Repo ordering decision needed") and EC-011 ("Repo
// priority undeclared"). No conflict is raised — these are informational notes.
func emitRepoOrderingExplanations(
	opinions []resolver.Opinion,
	active map[resolver.OpinionID]bool,
	dropped map[resolver.OpinionID]bool,
	rs *ResolvedSpeech,
) {
	// Collect active custom-repo opinions.
	var repoOps []resolver.Opinion
	for _, op := range opinions {
		if active[op.ID] && !dropped[op.ID] && op.Category == "custom-repo" && len(op.CustomRepos) > 0 {
			repoOps = append(repoOps, op)
		}
	}
	if len(repoOps) < 2 {
		return
	}
	// Check that these opinions were not already paired in a conflict explanation.
	// (If two custom-repo opinions conflict AND one is required-beats-nice, that's
	// Rule 1 and already explained. EC-010/011 are the non-conflicting case.)
	var ids []resolver.OpinionID
	for _, op := range repoOps {
		ids = append(ids, op.ID)
	}

	// Determine whether ALL repos have explicit priority declarations.
	// If any repo lacks a priority (Priority == 0), the relative ordering is undeclared.
	allHavePriority := true
	for _, op := range repoOps {
		for _, repo := range op.CustomRepos {
			if repo.Priority == 0 {
				allHavePriority = false
				break
			}
		}
		if !allHavePriority {
			break
		}
	}

	// Collect repo names for the explanation.
	var repoNames []string
	for _, op := range repoOps {
		for _, repo := range op.CustomRepos {
			repoNames = append(repoNames, repo.Name)
		}
	}

	var text, rule string
	if allHavePriority {
		text = fmt.Sprintf("Repo ordering decision: multiple custom-repo opinions present (%s). Resolved order by priority declaration: %s. Verify the order before building.",
			strings.Join(ids2strings(ids), ", "), strings.Join(repoNames, " > "))
		rule = "ordering"
	} else {
		text = fmt.Sprintf("Repo priority undeclared: multiple custom-repo opinions present (%s); at least one repo lacks an explicit priority. Resolved with default ordering: %s.",
			strings.Join(ids2strings(ids), ", "), strings.Join(repoNames, " > "))
		rule = "no-conflict"
	}

	rs.Explanations = append(rs.Explanations, Explanation{
		Text:             text,
		Rule:             rule,
		OpinionsInvolved: ids,
		Kept:             ids,
	})
}

func ids2strings(ids []resolver.OpinionID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	return out
}

// ─── sysctl collision detection ────────────────────────────────────────────

// detectSysctlCollisions scans active opinions for per-key sysctl collisions
// (two opinions declaring the same sysctl key — SR-016). Appends collision
// explanations and returns the first collision error found.
func detectSysctlCollisions(
	opinions []resolver.Opinion,
	active map[resolver.OpinionID]bool,
	dropped map[resolver.OpinionID]bool,
	rs *ResolvedSpeech,
	_ map[resolver.OpinionID]*resolver.Opinion,
) error {
	// keyOwners maps sysctl key → slice of (opinionID, value) that claim it.
	type owner struct {
		id  resolver.OpinionID
		val string
	}
	keyOwners := make(map[string][]owner)

	for _, op := range opinions {
		if !active[op.ID] || dropped[op.ID] {
			continue
		}
		for _, sp := range op.SysctlParams {
			keyOwners[sp.Key] = append(keyOwners[sp.Key], owner{id: op.ID, val: sp.Value})
		}
	}

	// Sort keys for deterministic error ordering.
	var collisionKeys []string
	for k, owners := range keyOwners {
		if len(owners) > 1 {
			collisionKeys = append(collisionKeys, k)
		}
	}
	sort.Strings(collisionKeys)

	var firstErr error
	for _, k := range collisionKeys {
		owners := keyOwners[k]
		var ids []resolver.OpinionID
		var parts []string
		for _, o := range owners {
			ids = append(ids, o.id)
			parts = append(parts, fmt.Sprintf("%s (value=%s)", o.id, o.val))
		}
		text := fmt.Sprintf(
			"Sysctl key collision: %q is written by multiple opinions in this speech — %s. The effective value depends on drop-in prefix order, which is non-deterministic. A patch opinion is needed to merge these.",
			k, strings.Join(parts, " and "),
		)
		rs.Explanations = append(rs.Explanations, Explanation{
			Text:             text,
			Rule:             "sysctl-collision",
			OpinionsInvolved: ids,
		})
		if firstErr == nil {
			firstErr = fmt.Errorf("Sysctl key collision on %q: %s", k, strings.Join(parts, ", "))
		}
	}
	return firstErr
}

// ─── helpers ───────────────────────────────────────────────────────────────

// sortedActiveIDs returns the IDs of opinions in the opinions slice that are
// active and not dropped, in their original input order (not sorted).
// This preserves the "first-listed wins" semantic for Rule 3.
func sortedActiveIDs(opinions []resolver.Opinion, active, dropped map[resolver.OpinionID]bool) []resolver.OpinionID {
	out := make([]resolver.OpinionID, 0, len(opinions))
	for _, op := range opinions {
		if active[op.ID] && !dropped[op.ID] {
			out = append(out, op.ID)
		}
	}
	return out
}

// canonicalPair returns the two IDs as a sorted [2]resolver.OpinionID key
// so that the same pair is represented the same way regardless of argument order.
func canonicalPair(a, b resolver.OpinionID) [2]resolver.OpinionID {
	if a <= b {
		return [2]resolver.OpinionID{a, b}
	}
	return [2]resolver.OpinionID{b, a}
}

// collectTrustWarning returns a non-empty warning string when any of the
// opinion's custom repos has SigLevel == Never (T-01-10 mitigation).
func collectTrustWarning(op resolver.Opinion) string {
	for _, repo := range op.CustomRepos {
		if repo.SigLevel == resolver.SigLevelNever {
			return fmt.Sprintf("Note: this opinion adds repo %q with SigLevel=Never — review this trust level before building.", repo.Name)
		}
	}
	return ""
}
