package graph_test

import (
	"testing"

	"github.com/mikkelraglan/debateos/resolver"
	"github.com/mikkelraglan/debateos/resolver/graph"
)

// buildOp is a helper that constructs a minimal resolver.Opinion for graph tests.
func buildOp(id resolver.OpinionID, after []resolver.OpinionID) resolver.Opinion {
	op := resolver.Opinion{
		Schema:   1,
		ID:       id,
		Name:     string(id),
		Category: "test",
		Status:   resolver.StatusRequired,
	}
	if len(after) > 0 {
		refs := make([]resolver.OpinionRef, len(after))
		for i, a := range after {
			refs[i] = resolver.OpinionRef{ID: a}
		}
		op.Ordering = &resolver.Ordering{After: refs}
	}
	return op
}

// TestTopoSort covers EC-035 (three-hop), EC-036 (cycle), cross-phase override,
// and determinism (lexicographic tie-break).
func TestTopoSort(t *testing.T) {
	t.Run("EC-035", func(t *testing.T) {
		// Three-hop dependency chain: OM-009 → OM-041 → OM-023
		// Expected order: [OM-009, OM-041, OM-023]
		opinions := []resolver.Opinion{
			buildOp("OM-009", nil),
			buildOp("OM-041", []resolver.OpinionID{"OM-009"}),
			buildOp("OM-023", []resolver.OpinionID{"OM-041"}),
		}
		g, err := graph.BuildGraph(opinions)
		if err != nil {
			t.Fatalf("EC-035: BuildGraph returned unexpected error: %v", err)
		}
		order, cycle, err := graph.TopoSort(g)
		if err != nil {
			t.Fatalf("EC-035: TopoSort returned unexpected error: %v (cycle: %v)", err, cycle)
		}
		want := []resolver.OpinionID{"OM-009", "OM-041", "OM-023"}
		if !equalIDSlice(order, want) {
			t.Errorf("EC-035: got order %v, want %v", order, want)
		}
	})

	t.Run("EC-036", func(t *testing.T) {
		// Circular ordering constraint: docker-service ↔ docker-dns mutual after.
		// Expected: non-nil error; cycle slice contains both opinion IDs (sorted).
		opinions := []resolver.Opinion{
			buildOp("docker-service", []resolver.OpinionID{"docker-dns"}),
			buildOp("docker-dns", []resolver.OpinionID{"docker-service"}),
		}
		g, err := graph.BuildGraph(opinions)
		if err != nil {
			t.Fatalf("EC-036: BuildGraph returned unexpected error: %v", err)
		}
		_, cycle, err := graph.TopoSort(g)
		if err == nil {
			t.Fatal("EC-036: expected a cycle error, got nil")
		}
		if len(cycle) == 0 {
			t.Error("EC-036: cycle slice is empty; expected it to name the offending opinions")
		}
		// Both IDs must appear in the cycle slice (sorted, deterministic)
		wantInCycle := map[resolver.OpinionID]bool{
			"docker-service": true,
			"docker-dns":     true,
		}
		for _, id := range cycle {
			delete(wantInCycle, id)
		}
		if len(wantInCycle) != 0 {
			t.Errorf("EC-036: cycle slice missing IDs: %v; got %v", wantInCycle, cycle)
		}
	})

	t.Run("CrossPhase", func(t *testing.T) {
		// SR-006 / OM-023 cross-phase ordering override.
		// packaging-op is in the "packaging" phase (which normally precedes "config"),
		// but has an explicit after:[config-op] constraint.
		// Expected: config-op appears BEFORE packaging-op in the output.
		configOp := resolver.Opinion{
			Schema:       1,
			ID:           "config-op",
			Name:         "config-phase-opinion",
			Category:     "config-dotfile",
			Status:       resolver.StatusRequired,
			InstallPhase: "config",
		}
		packagingOp := resolver.Opinion{
			Schema:       1,
			ID:           "packaging-op",
			Name:         "packaging-phase-opinion",
			Category:     "npm-global-install",
			Status:       resolver.StatusRequired,
			InstallPhase: "packaging",
			Ordering: &resolver.Ordering{
				After: []resolver.OpinionRef{{ID: "config-op"}},
			},
		}
		opinions := []resolver.Opinion{configOp, packagingOp}
		g, err := graph.BuildGraph(opinions)
		if err != nil {
			t.Fatalf("CrossPhase: BuildGraph returned unexpected error: %v", err)
		}
		order, cycle, err := graph.TopoSort(g)
		if err != nil {
			t.Fatalf("CrossPhase: TopoSort returned unexpected error: %v (cycle: %v)", err, cycle)
		}
		// config-op must precede packaging-op
		configIdx := -1
		packagingIdx := -1
		for i, id := range order {
			if id == "config-op" {
				configIdx = i
			}
			if id == "packaging-op" {
				packagingIdx = i
			}
		}
		if configIdx == -1 || packagingIdx == -1 {
			t.Fatalf("CrossPhase: expected both IDs in order, got %v", order)
		}
		if configIdx >= packagingIdx {
			t.Errorf("CrossPhase: config-op (idx %d) must come before packaging-op (idx %d); order: %v",
				configIdx, packagingIdx, order)
		}
	})
}

// TestTopoSortDeterministic verifies that shuffled input order yields the same
// topological output — lexicographic heap tie-breaking guards Pitfall 1.
func TestTopoSortDeterministic(t *testing.T) {
	// Three independent opinions (no ordering edges); lexicographic order expected.
	opinions1 := []resolver.Opinion{
		buildOp("beta", nil),
		buildOp("alpha", nil),
		buildOp("gamma", nil),
	}
	opinions2 := []resolver.Opinion{
		buildOp("gamma", nil),
		buildOp("alpha", nil),
		buildOp("beta", nil),
	}
	opinions3 := []resolver.Opinion{
		buildOp("alpha", nil),
		buildOp("gamma", nil),
		buildOp("beta", nil),
	}

	want := []resolver.OpinionID{"alpha", "beta", "gamma"} // lexicographic

	for i, ops := range [][]resolver.Opinion{opinions1, opinions2, opinions3} {
		g, err := graph.BuildGraph(ops)
		if err != nil {
			t.Fatalf("Deterministic[%d]: BuildGraph error: %v", i, err)
		}
		order, _, err := graph.TopoSort(g)
		if err != nil {
			t.Fatalf("Deterministic[%d]: TopoSort error: %v", i, err)
		}
		if !equalIDSlice(order, want) {
			t.Errorf("Deterministic[%d]: got %v, want %v", i, order, want)
		}
	}
}

// equalIDSlice checks two OpinionID slices for equality.
func equalIDSlice(a, b []resolver.OpinionID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
