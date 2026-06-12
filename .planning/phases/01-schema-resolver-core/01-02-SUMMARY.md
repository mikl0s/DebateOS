---
phase: 01-schema-resolver-core
plan: 02
status: complete
completed: 2026-06-12
requirements: [RSLV-03]
subsystem: resolver/graph
tags: [graph, toposort, kahn, tdd, determinism, cycle-detection, ordering]
dependency_graph:
  requires:
    - 01-01 (resolver/types.go — Opinion/OpinionID/Ordering/OpinionRef shared types)
  provides:
    - resolver/graph.BuildGraph (consumes []resolver.Opinion, produces *Graph)
    - resolver/graph.TopoSort (consumes *Graph, produces order []OpinionID + cycle []OpinionID + err)
  affects:
    - 01-04 (resolver/resolve — consumes TopoSort output for install ordering)
tech_stack:
  added:
    - container/heap (stdlib min-heap for O(E log V) lexicographic toposort tie-breaking)
  patterns:
    - Kahn's algorithm with heap-based lexicographic tie-breaking (Pattern 2 from 01-RESEARCH.md)
    - sortedNodes() iteration guard — no range-over-map in output-producing paths (Pitfall 1)
    - External test package (graph_test) importing graph — clean API boundary
key_files:
  created:
    - resolver/graph/graph.go
    - resolver/graph/toposort.go
    - resolver/graph/graph_test.go
    - resolver/graph/testdata/three-hop.yaml
    - resolver/graph/testdata/cycle.yaml
    - resolver/graph/testdata/cross-phase.yaml
decisions:
  - TopoSort is a free function (not a method on Graph) — cleaner call site for 01-04
  - opinionIDHeap is unexported — implementation detail of toposort.go
  - Phase enum stored in Graph.phase but NOT converted to edges — tie-breaking key only, per SR-006/OM-023
  - External test package (graph_test) used for graph_test.go — tests the public API, not internals
  - Phantom nodes (ordering references to opinions outside the input slice) are accepted silently to avoid early errors at graph-build time; resolver/resolve can enforce completeness
metrics:
  duration: ~2 min
  completed_date: 2026-06-12
  tasks_completed: 2
  files_created: 6
commits:
  - 276566f test(01-02): RED — graph/toposort tests (EC-035, EC-036, cross-phase, determinism)
  - 2950b54 feat(01-02): GREEN — BuildGraph + deterministic Kahn TopoSort with cycle naming
---

# Phase 1 Plan 02: Graph + TopoSort Summary

**One-liner:** Kahn toposort with container/heap lexicographic tie-breaking, BuildGraph from Opinion ordering edges, and cycle detection naming offending opinions — all built test-first against EC-035/EC-036/cross-phase/determinism gates.

## What Was Built

### resolver/graph/graph.go — BuildGraph

`BuildGraph(opinions []resolver.Opinion) (*Graph, error)` assembles a directed graph where each node is an `OpinionID` and edges represent "must come before" relationships:

- `DependsOn` edges (SR-002): `dep.ID → op.ID` for each `dep` in `op.DependsOn`.
- `Ordering.After` edges (SR-006): `pred.ID → op.ID` for each `pred` in `op.Ordering.After`.
- `Ordering.Before` edges (SR-006): `op.ID → succ.ID` for each `succ` in `op.Ordering.Before`.
- `InstallPhase` stored in `Graph.phase` map — for tie-breaking only, never as an edge. Cross-phase edges are accepted without restriction (OM-023 case: packaging-phase opinion with explicit `after: [config-phase-opinion]`).
- Adjacency lists are deduplicated after construction.

### resolver/graph/toposort.go — TopoSort

`TopoSort(g *Graph) (order []OpinionID, cycle []OpinionID, err error)` implements Kahn's algorithm:

- **In-degree computation:** Iterates nodes via `sortedNodes()` (not range-over-map) to seed deterministically.
- **Min-heap tie-breaking:** `opinionIDHeap` implements `heap.Interface`; `Less(i,j)` uses `h[i] < h[j]` (lexicographic string comparison). At each step the smallest-ID eligible node is processed first — stable regardless of input order.
- **Cycle detection:** After the main loop, any node with `inDegree > 0` is in a cycle. These are collected, sorted with `sort.Slice`, and returned as `cycle` with a non-nil error naming them. This satisfies T-01-06 (deterministic cycle message, no map-order leak).
- **No range-over-map in output path:** `sortedNodes()` collects keys, sorts, then iterates; successor lists are sorted before processing.

### Test Coverage

| Test | Scenario | Result |
|------|----------|--------|
| `TestTopoSort/EC-035` | OM-009 → OM-041 → OM-023 three-hop chain | PASS — exact order `[OM-009, OM-041, OM-023]` |
| `TestTopoSort/EC-036` | docker-service ↔ docker-dns mutual `after` cycle | PASS — non-nil error, both IDs in cycle slice |
| `TestTopoSort/CrossPhase` | packaging-op with `after: [config-op]` overrides phase enum order | PASS — config-op precedes packaging-op |
| `TestTopoSortDeterministic` | 3 shuffled orderings of independent alpha/beta/gamma | PASS across 5 runs — lexicographic output stable |

## Public API for Downstream Plans (01-04)

```go
// package resolver/graph

type Graph struct { /* opaque */ }

// BuildGraph assembles dependency + ordering edges from a slice of parsed opinions.
// Phase enum stored for tie-breaking only; explicit before/after edges take precedence.
func BuildGraph(opinions []resolver.Opinion) (*Graph, error)

// TopoSort performs deterministic Kahn toposort with lexicographic heap tie-breaking.
// Returns the full install order, or (nil, cycleIDs, error) when a cycle is detected.
func TopoSort(g *Graph) (order []resolver.OpinionID, cycle []resolver.OpinionID, err error)
```

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] External test package required package prefix for BuildGraph/TopoSort**
- **Found during:** Task 2 GREEN verification (first `go test` run)
- **Issue:** `graph_test.go` used `package graph_test` (external) but called `BuildGraph` / `TopoSort` without the `graph.` qualifier — compilation failed.
- **Fix:** Added `"github.com/mikl0s/debateos/resolver/graph"` import and prefixed all calls with `graph.`. Standard external-package test pattern — not a logic change, no TDD gate impact (RED was still RED; fix applied before GREEN commit).
- **Files modified:** `resolver/graph/graph_test.go`
- **Commit:** 2950b54 (GREEN commit includes both implementation and corrected test import)

No other deviations — plan executed as written.

## TDD Gate Compliance

- RED commit `276566f` precedes GREEN commit `2950b54` in git history.
- RED: `go test ./resolver/graph/` → `FAIL [build failed]` (BuildGraph/TopoSort undefined).
- GREEN: `go test ./resolver/graph/ -count=1` → PASS; `-count=5` determinism stable; `go vet` clean.

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes introduced. `resolver/graph` operates entirely on in-memory typed structs (validated by 01-01 parse layer before reaching this package). T-01-06 mitigation (deterministic cycle message) implemented as designed.

## Self-Check: PASSED

- `resolver/graph/graph.go` — exists, builds
- `resolver/graph/toposort.go` — exists, builds, contains `container/heap`
- `resolver/graph/graph_test.go` — exists, 4 tests, all PASS
- RED commit `276566f` — confirmed in `git log --oneline`
- GREEN commit `2950b54` — confirmed in `git log --oneline`
- `go test ./resolver/graph/ -count=1` — PASS
- `go test ./resolver/graph/ -run TestTopoSortDeterministic -count=5` — PASS
- `go vet ./resolver/graph/` — clean (exit 0)
