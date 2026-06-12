package graph

import (
	"container/heap"
	"fmt"
	"sort"

	"github.com/mikl0s/debateos/resolver"
)

// heapEntry is the element type for the topoSort min-heap. Tie-breaking uses
// phase weight first (lower weight = earlier install phase), then lexicographic
// OpinionID for same-phase or unspecified-phase opinions.
type heapEntry struct {
	id    resolver.OpinionID
	phase int // from phaseOrder[g.phase[id]]; 0 means unspecified
}

// heapEntries is a min-heap of heapEntry values for deterministic toposort
// tie-breaking: phase order first, lexicographic ID second.
type heapEntries []heapEntry

func (h heapEntries) Len() int { return len(h) }

// Less orders by phase weight first (lower weight = earlier). Unspecified
// phase (weight 0) sorts last — treated as MaxInt for the comparison.
// Within the same phase, ordering is lexicographic on OpinionID.
func (h heapEntries) Less(i, j int) bool {
	pi, pj := h[i].phase, h[j].phase
	const maxPhase = 1<<31 - 1
	if pi == 0 {
		pi = maxPhase
	}
	if pj == 0 {
		pj = maxPhase
	}
	if pi != pj {
		return pi < pj
	}
	return h[i].id < h[j].id
}

func (h heapEntries) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *heapEntries) Push(x any) {
	*h = append(*h, x.(heapEntry))
}

func (h *heapEntries) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TopoSort performs a deterministic Kahn's algorithm topological sort on the
// given Graph. Tie-breaking among equally-eligible nodes uses a min-heap keyed
// on install-phase order first (preflight < packaging < config < login <
// post-install < first-run), then lexicographic OpinionID within the same
// phase. Unspecified phases sort last, with lexicographic tie-breaking among
// them. This guarantees reproducible output regardless of input order
// (Pitfall 1 guard) and correct cross-phase ordering when no explicit edges
// are declared.
//
// Returns:
//   - order: the full topological install order when no cycle exists.
//   - cycle: a sorted slice of OpinionIDs involved in the cycle (non-nil when err != nil).
//   - err: non-nil when a cycle is detected; message names the offending opinions.
func TopoSort(g *Graph) (order []resolver.OpinionID, cycle []resolver.OpinionID, err error) {
	// Compute in-degree for every node.
	// We must iterate nodes in a fixed order to avoid map non-determinism
	// in the inDegree initialisation (even though the values are the same,
	// the heap must be seeded deterministically).
	nodes := g.sortedNodes()

	inDegree := make(map[resolver.OpinionID]int, len(nodes))
	for _, id := range nodes {
		if _, ok := inDegree[id]; !ok {
			inDegree[id] = 0
		}
		// Each edge from→to increments to's in-degree.
		for _, to := range g.edges[id] {
			inDegree[to]++
		}
	}

	// Seed the heap with all zero-in-degree nodes, attaching each node's
	// phase weight so the heap can order by phase first.
	h := &heapEntries{}
	heap.Init(h)
	for _, id := range nodes {
		if inDegree[id] == 0 {
			heap.Push(h, heapEntry{id: id, phase: phaseOrder[g.phase[id]]})
		}
	}

	order = make([]resolver.OpinionID, 0, len(nodes))

	for h.Len() > 0 {
		// Pop the node with the smallest (phase, id) key.
		entry := heap.Pop(h).(heapEntry)
		order = append(order, entry.id)

		// Sort successors before iterating to avoid map/slice non-determinism.
		succs := make([]resolver.OpinionID, len(g.edges[entry.id]))
		copy(succs, g.edges[entry.id])
		sort.Slice(succs, func(i, j int) bool { return succs[i] < succs[j] })

		for _, succ := range succs {
			inDegree[succ]--
			if inDegree[succ] == 0 {
				heap.Push(h, heapEntry{id: succ, phase: phaseOrder[g.phase[succ]]})
			}
		}
	}

	// If not all nodes were processed, there is a cycle.
	if len(order) != len(nodes) {
		// Collect all nodes with remaining in-degree > 0 — these are in the cycle.
		// Sort for deterministic error message (T-01-06 mitigation).
		for _, id := range nodes {
			if inDegree[id] > 0 {
				cycle = append(cycle, id)
			}
		}
		sort.Slice(cycle, func(i, j int) bool { return cycle[i] < cycle[j] })
		return nil, cycle, fmt.Errorf("cycle detected in install ordering: %v — remove one of the ordering constraints or introduce an intermediate opinion to break the cycle", cycle)
	}

	return order, nil, nil
}
