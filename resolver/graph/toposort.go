package graph

import (
	"container/heap"
	"fmt"
	"sort"

	"github.com/mikkelraglan/debateos/resolver"
)

// opinionIDHeap is a min-heap of OpinionID strings for lexicographic tie-breaking.
// container/heap requires the heap.Interface: Len, Less, Swap, Push, Pop.
type opinionIDHeap []resolver.OpinionID

func (h opinionIDHeap) Len() int           { return len(h) }
func (h opinionIDHeap) Less(i, j int) bool { return h[i] < h[j] } // lexicographic
func (h opinionIDHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *opinionIDHeap) Push(x any) {
	*h = append(*h, x.(resolver.OpinionID))
}

func (h *opinionIDHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TopoSort performs a deterministic Kahn's algorithm topological sort on the
// given Graph. Tie-breaking among equally-eligible nodes uses a min-heap keyed
// on OpinionID string (lexicographic order), guaranteeing reproducible output
// regardless of input order (Pitfall 1 guard).
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

	// Seed the heap with all zero-in-degree nodes.
	h := &opinionIDHeap{}
	heap.Init(h)
	for _, id := range nodes {
		if inDegree[id] == 0 {
			heap.Push(h, id)
		}
	}

	order = make([]resolver.OpinionID, 0, len(nodes))

	for h.Len() > 0 {
		// Pop the lexicographically smallest eligible node.
		id := heap.Pop(h).(resolver.OpinionID)
		order = append(order, id)

		// Sort successors before iterating to avoid map/slice non-determinism.
		succs := make([]resolver.OpinionID, len(g.edges[id]))
		copy(succs, g.edges[id])
		sort.Slice(succs, func(i, j int) bool { return succs[i] < succs[j] })

		for _, succ := range succs {
			inDegree[succ]--
			if inDegree[succ] == 0 {
				heap.Push(h, succ)
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
