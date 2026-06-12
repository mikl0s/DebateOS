// Package graph constructs the dependency and ordering graph from parsed opinions
// and provides a deterministic Kahn's algorithm topological sort.
//
// Dependency edges come from Opinion.DependsOn (SR-002).
// Ordering edges come from Opinion.Ordering.Before/After (SR-006).
// Phase enum (InstallPhase) is an additional tie-breaking key only — explicit
// before/after edges (including cross-phase, per OM-023) take precedence and
// are never overridden by phase enum order.
package graph

import (
	"fmt"
	"sort"

	"github.com/mikkelraglan/debateos/resolver"
)

// phaseOrder assigns a numeric weight to install_phase values for tie-breaking.
// Lower value = earlier in install sequence. Zero means unspecified (sorts last).
// Phase enum is NOT a hard partition — explicit ordering edges take precedence.
var phaseOrder = map[string]int{
	"preflight":    1,
	"packaging":    2,
	"config":       3,
	"login":        4,
	"post-install": 5,
	"first-run":    6,
}

// Graph is a directed graph of OpinionIDs representing dependency and ordering
// constraints between opinions. Edges represent "predecessor must come before
// successor" relationships.
type Graph struct {
	// nodes is the complete set of opinion IDs known to the graph.
	nodes map[resolver.OpinionID]struct{}

	// edges maps each node to the set of nodes that must come AFTER it.
	// edges[A] = {B, C} means A must be installed before B and C.
	edges map[resolver.OpinionID][]resolver.OpinionID

	// phase is the install_phase for each node (used only for tie-breaking).
	phase map[resolver.OpinionID]string
}

// BuildGraph assembles a dependency + ordering graph from a slice of opinions.
// Every opinion in the slice becomes a node. Edges are added for:
//   - depends_on (SR-002): dependency must be installed before the dependent.
//   - ordering.after (SR-006): predecessor must be installed before the node.
//   - ordering.before (SR-006): node must be installed before the named opinion.
//
// Cross-phase before/after edges are accepted without restriction (OM-023 case).
// InstallPhase is stored for tie-breaking in TopoSort but does not create edges.
func BuildGraph(opinions []resolver.Opinion) (*Graph, error) {
	g := &Graph{
		nodes: make(map[resolver.OpinionID]struct{}, len(opinions)),
		edges: make(map[resolver.OpinionID][]resolver.OpinionID, len(opinions)),
		phase: make(map[resolver.OpinionID]string, len(opinions)),
	}

	// Register all nodes first so ordering constraints referencing any ID are valid.
	for _, op := range opinions {
		g.nodes[op.ID] = struct{}{}
		g.edges[op.ID] = nil
		g.phase[op.ID] = op.InstallPhase
	}

	// Add edges from each opinion's constraints.
	for _, op := range opinions {
		// depends_on: predecessor (dep.ID) must come before op.ID
		for _, dep := range op.DependsOn {
			if err := g.ensureNode(dep.ID, fmt.Sprintf("depends_on of %s", op.ID)); err != nil {
				return nil, err
			}
			g.addEdge(dep.ID, op.ID)
		}

		if op.Ordering == nil {
			continue
		}

		// ordering.after: named predecessor must come before op.ID
		for _, pred := range op.Ordering.After {
			if err := g.ensureNode(pred.ID, fmt.Sprintf("ordering.after of %s", op.ID)); err != nil {
				return nil, err
			}
			g.addEdge(pred.ID, op.ID)
		}

		// ordering.before: op.ID must come before each named successor
		for _, succ := range op.Ordering.Before {
			if err := g.ensureNode(succ.ID, fmt.Sprintf("ordering.before of %s", op.ID)); err != nil {
				return nil, err
			}
			g.addEdge(op.ID, succ.ID)
		}
	}

	// Deduplicate adjacency lists to keep in-degree calculations correct.
	for id := range g.edges {
		g.edges[id] = dedupIDs(g.edges[id])
	}

	return g, nil
}

// ensureNode registers an implicitly referenced node (e.g., from ordering edges
// that reference an opinion not in the input slice). In the current design,
// all referenced nodes should already be in the graph (the caller is responsible
// for passing a complete set), but this allows referencing unresolved IDs without
// panicking — they are simply added as phantom nodes.
func (g *Graph) ensureNode(id resolver.OpinionID, context string) error {
	_ = context // reserved for future error messages
	if _, ok := g.nodes[id]; !ok {
		// Phantom node: referenced but not in the input opinions.
		// Add it so in-degree calculations don't miss it.
		g.nodes[id] = struct{}{}
		g.edges[id] = nil
	}
	return nil
}

// addEdge adds a directed edge from -> to (meaning `from` must come before `to`).
func (g *Graph) addEdge(from, to resolver.OpinionID) {
	g.edges[from] = append(g.edges[from], to)
}

// dedupIDs removes duplicate OpinionIDs from a slice, preserving order.
func dedupIDs(ids []resolver.OpinionID) []resolver.OpinionID {
	seen := make(map[resolver.OpinionID]bool, len(ids))
	out := ids[:0]
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// sortedNodes returns the graph's node IDs in a stable, deterministic order.
// Used internally to avoid map-iteration-order non-determinism (Pitfall 1).
func (g *Graph) sortedNodes() []resolver.OpinionID {
	ids := make([]resolver.OpinionID, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
