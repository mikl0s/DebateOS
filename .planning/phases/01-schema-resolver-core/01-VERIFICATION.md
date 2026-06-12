---
phase: 01-schema-resolver-core
verified: 2026-06-12T14:00:00Z
status: passed
score: 5/5
overrides_applied: 0
---

# Phase 1: Schema & Resolver Core — Verification Report

**Phase Goal:** A composition can be parsed, validated, and resolved — every conflict handled
per the docs/04 hierarchy with a human-readable explanation, identically in native and WASM

**Verified:** 2026-06-12T14:00:00Z
**Status:** passed
**Re-verification:** No — initial verification + RSLV-04 gap closure

---

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Opinion/Point/Speech YAML schemas in schemas/ (CC0), cover full SR-001..SR-022 floor, understandable from YAML alone | VERIFIED | schemas/README.md has 25 SR-0 refs covering all 22 SRs; all three schemas declare $schema 2020-12; CC0 LICENSE confirmed; no OS-specific tokens in schemas |
| 2 | Resolver resolves every harness scenario per docs/04 rules: visible drops, required-vs-required hard conflict, nice-vs-nice default, ordering cycles name offending opinions | VERIFIED | `go test ./resolver/resolve/ -run TestResolveEC` — all 28 EC subtests pass (EC-001 through EC-052 set incl. EC-037b); WR-01 (multiple hard conflicts all reported), WR-04 (sysctl + conflict combined) confirmed fixed |
| 3 | Hardware-conditional opinions resolve against declared hardware with swap suggestions surfaced at composition time | VERIFIED | EvalCondition works; hardware-skip/apply explanations emitted; AlternativeSuggestion field added to Explanation (commit 1b1d046). When a same-category hw-true alternative exists in the composition, the skip explanation carries "You declared <category>: consider '<Name>' (<ID>) instead." |
| 4 | WASM/native identical results proven by automated parity tests; near-total coverage per D19; EC corpus encoded as tests BEFORE implementation | VERIFIED | `bash scripts/wasm-parity-test.sh` → PARITY OK (4 goldens, byte-identical native+WASM); `bash scripts/check-coverage.sh` → 93.5% >= 90%; 6 RED commits precede GREEN commits in git history; 28 unique EC IDs in resolve_test.go |
| 5 | 3–4 example files (incl. one deliberately conflicting) exercising the harness end-to-end | VERIFIED | 4 example directories confirmed (omarchy-mini, two-point-clean, conflicting, hardware-conditional); `go test ./examples/` → all 5 subtests pass; examples reference real OM-NNN IDs |

**Score:** 5/5 truths verified

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `go.mod` | module github.com/mikkelraglan/debateos, go 1.24, two external deps | VERIFIED | module path, go 1.24, exactly santhosh-tekuri/jsonschema/v6 and go.yaml.in/yaml/v3 |
| `schemas/opinion.schema.json` | JSON Schema 2020-12 covering SR-001..SR-020 | VERIFIED | $schema = https://json-schema.org/draft/2020-12/schema |
| `schemas/point.schema.json` | JSON Schema 2020-12 for Point (SR-021) | VERIFIED | $schema = https://json-schema.org/draft/2020-12/schema |
| `schemas/speech.schema.json` | JSON Schema 2020-12 for Speech (SR-022) incl. pci_ids | VERIFIED | $schema declared; pci_ids added in CR-02 fix |
| `schemas/README.md` | SR-001..SR-022 traceability table | VERIFIED | 25 SR-0 references; all 22 SRs covered |
| `resolver/types.go` | Shared Go types, no float64 | VERIFIED | `grep -c float64 types.go` = 0; Opinion/Point/Speech/HardwareProfile (with PCIIDs after CR-02) all present |
| `resolver/parse/parse.go` | ParseOpinion/ParsePoint/ParseSpeech with KnownFields | VERIFIED | KnownFields(true) confirmed in decodeStrict |
| `resolver/parse/schemas_embed.go` | Delegates to schemas.FS | VERIFIED | Uses schemas.FS.ReadFile — no filesystem access at runtime |
| `resolver/graph/graph.go` | BuildGraph with phase tie-breaking | VERIFIED | phaseOrder and g.phase are live code after CR-01 fix |
| `resolver/graph/toposort.go` | TopoSort with heapEntry phase+lexicographic tie-break | VERIFIED | heapEntries.Less orders by phase weight first; CR-01 fix confirmed |
| `resolver/hardware/eval.go` | EvalCondition with depth limit | VERIFIED | maxHardwareExprDepth=32; evalConditionDepth depth guard confirmed (WR-03 fix) |
| `resolver/patch/patch.go` | FindPatch discovering patches for conflict pairs | VERIFIED | Implementation complete; RSLV-02 functionality implemented |
| `resolver/resolve/resolve.go` | Resolve with full docs/04 hierarchy | VERIFIED | All 4 rules + hardware + sysctl + cycle; WR-01/WR-02/WR-04 fixes confirmed; RSLV-04 AlternativeSuggestion implemented |
| `resolver/resolve/explanation.go` | Explanation type with AlternativeSuggestion | VERIFIED | AlternativeSuggestion string field added (omitempty, no golden impact) |
| `resolver/resolve/canonical.go` | CanonicalJSON serialization | VERIFIED | Used in parity tests; byte-identical across native and WASM |
| `resolver/wasm/main.go` | WASM entrypoint, js && wasm build tag | VERIFIED | `//go:build js && wasm` confirmed; debateosResolve registered in init(); JSON DisallowUnknownFields + YAML KnownFields(true) (IN-04 fix) |
| `resolver/resolve/testdata/golden/` | >= 4 committed golden JSON files | VERIFIED | 4 files: omarchy-mini.json, two-point-clean.json, conflicting.json, hardware-conditional.json |
| `scripts/wasm-parity-test.sh` | Parity script with golden guard | VERIFIED | Guard checks >= 4 goldens; references $(go env GOROOT)/lib/wasm/go_js_wasm_exec |
| `scripts/check-coverage.sh` | Coverage gate at 90% threshold | VERIFIED | threshold=90; enforces total across resolver packages |
| `examples/` (4 dirs) | omarchy-mini, two-point-clean, conflicting, hardware-conditional | VERIFIED | All 4 present with speech.yaml + opinions.yaml |
| `examples/examples_test.go` | End-to-end parse→resolve test | VERIFIED | 4 tests + 5 subtests all pass |
| `LICENSE` | AGPL-3.0 | VERIFIED | "GNU AFFERO GENERAL PUBLIC LICENSE" confirmed |
| `schemas/LICENSE` | CC0-1.0 | VERIFIED | "CC0 1.0 Universal" confirmed |
| `examples/LICENSE` | CC0-1.0 | VERIFIED | "CC0 1.0 Universal" confirmed |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| resolver/parse/parse.go | resolver/types.go | decode into Opinion/Point/Speech structs | WIRED | decodeStrict → yaml.NewDecoder.KnownFields(true) → typed struct |
| resolver/parse/schemas_embed.go | schemas/*.schema.json | schemas.FS.ReadFile | WIRED | schemas.FS (embed.FS) in schemas/embed.go |
| resolver/resolve/resolve.go | resolver/graph, hardware, patch | BuildGraph+TopoSort+EvalCondition+FindPatch | WIRED | Imports confirmed; all four sub-packages called |
| resolver/wasm/main.go | resolver/resolve | resolve.Resolve + resolve.CanonicalJSON | WIRED | Both called in debateosResolveFunc |
| scripts/wasm-parity-test.sh | go_js_wasm_exec | $(go env GOROOT)/lib/wasm/go_js_wasm_exec | WIRED | Variable WASM_EXEC confirmed; no committed copy |
| examples/examples_test.go | resolver/parse + resolver/resolve | yaml.Unmarshal → resolve.Resolve | WIRED | loadExample + Resolve calls confirmed |

---

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| resolver/resolve/resolve.go Resolve() | opinions []Opinion | Caller-provided (assembled from speech) | Yes — test harness provides 28 EC fixtures | FLOWING |
| resolver/resolve/canonical.go CanonicalJSON() | rs *ResolvedSpeech | Return value of Resolve() | Yes — real resolution results | FLOWING |
| resolver/wasm/main.go debateosResolveFunc | input resolveInput | JS string arg (JSON/YAML) | Yes — WASM smoke test + parity | FLOWING |
| examples/examples_test.go | opinions, speech | os.ReadFile from examples/ YAML | Yes — real YAML files with OM-NNN IDs | FLOWING |

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full test suite passes | `go test ./... -count=1` | ok 5 packages (examples, graph, hardware, parse, patch, resolve); 2 packages no test files | PASS |
| WASM build succeeds | `GOOS=js GOARCH=wasm go build ./resolver/wasm/` | WASM BUILD OK | PASS |
| WASM/native parity | `bash scripts/wasm-parity-test.sh` | PARITY OK — 4 goldens, both native and WASM match | PASS |
| Coverage gate | `bash scripts/check-coverage.sh` | COVERAGE OK: 93.5% >= 90% | PASS |
| 28 EC subtests all pass | `go test ./resolver/resolve/ -run TestResolveEC -v` | All 28 PASS (EC-001..EC-052 set, incl. EC-037b, EC-038b) | PASS |
| TDD RED before GREEN | `git log --oneline` | 6 RED commits (test(01-01) through test(01-05)) precede GREEN commits | PASS |
| EC corpus count vs research | diff of EC IDs in research vs test file | All 27 research EC IDs encoded; EC-037b and EC-100 are scheme additions | PASS |
| Phase tie-breaking (CR-01) | `go test ./resolver/graph/ -run TestTopoSortPhaseTieBreak -v` | PASS | PASS |
| Multiple hard conflicts reported (WR-01) | `go test ./resolver/resolve/ -run TestResolveMultipleHardConflictsAllReported -v` | PASS | PASS |
| Trust warning non-hardware (WR-02) | `go test ./resolver/resolve/ -run TestResolveTrustWarningNonHardwareOpinion -v` | PASS | PASS |
| Depth limit enforced (WR-03) | `go test ./resolver/hardware/ -run TestHardwareEvalDepthLimit -v` | PASS | PASS |
| Sysctl+conflict both reported (WR-04) | `go test ./resolver/resolve/ -run TestResolveSysctlAndHardConflictBothReported -v` | PASS | PASS |
| Hardware swap suggestion (RSLV-04) | `go test ./resolver/resolve/ -run TestResolveHardwareSwapSuggestion -v` | PASS — AlternativeSuggestion = "You declared hardware-conditional: consider 'AMD GPU Driver' (hardware-conditional/amd-driver) instead." | PASS |
| No OS-specific tokens in schemas | `grep -rni "pacman\|mkarchiso\|aur\|dpkg\|apt\|debian" schemas/*.schema.json` | empty | PASS |
| No float64 in types.go | `grep -c float64 resolver/types.go` | 0 | PASS |
| No TBD/FIXME/XXX debt markers | `grep -rn "TBD\|FIXME\|XXX" resolver/ schemas/ examples/ scripts/` | Only XXXXXX in mktemp pattern (not a debt marker) | PASS |
| Toposort determinism | `go test ./resolver/graph/ -run TestTopoSortDeterministic -count=5` | all 5 runs pass | PASS |
| Golden files unchanged | `bash scripts/wasm-parity-test.sh` | All 4 goldens byte-identical (AlternativeSuggestion is omitempty) | PASS |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SCHM-01 | 01-01 | SR-001..SR-022 expressible in schemas | SATISFIED | 22 SR rows in schemas/README.md; all 3 schemas use 2020-12; parse tests green |
| SCHM-02 | 01-01 | Human-readable, OS-agnostic | SATISFIED | No OS tokens in schemas; TestSchemaOSAgnostic passes; YAML readability verified by orchestrator cold-read (see Resolution section) |
| RSLV-01 | 01-04 | Parse+validate+apply docs/04 hierarchy with explanations | SATISFIED | Resolve() with all 4 rules, Explanation per decision; 28 EC corpus green |
| RSLV-02 | 01-03 | Patch opinions first-class, discovered automatically | SATISFIED | FindPatch() implemented; Rule 4 in resolveConflict; EC-032/EC-034 pass |
| RSLV-03 | 01-02 | Ordering → toposort; cycles name opinions | SATISFIED | TopoSort Kahn + heapEntry phase tie-break; EC-035/EC-036 pass; CR-01 fixed |
| RSLV-04 | 01-03 | Hardware-conditional evaluation with swap suggestions | SATISFIED | EvalCondition works; AlternativeSuggestion added to Explanation; EC-037b passes; in-composition only (registry-wide deferred to Phase 5) |
| RSLV-05 | 01-05 | WASM/native identical results (automated parity) | SATISFIED | wasm-parity-test.sh PARITY OK; 4 goldens committed; AlternativeSuggestion is omitempty so existing goldens unchanged |
| RSLV-06 | 01-04/01-05 | TDD corpus + near-total coverage + examples | SATISFIED | 28 EC tests (27 from research + EC-037b); RED commits precede GREEN; 93.5% coverage; 4 examples |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None found | — | No TBD/FIXME/XXX, no stub returns, no empty handlers | — | Clean |

---

## Resolution

### 1. YAML Readability (SC-1 / SCHM-02 invariant 3)

**Verified by orchestrator cold-read:** The conflicting example is comprehensible from YAML alone. The `examples/conflicting/opinions.yaml` declares two opinions with `conflicts:` entries pointing at each other; the intent, conflict nature, and resolution path (drop one, add a patch, or change status) are self-evident from the YAML structure without reading Go source. The hardware-conditional example shows `hardware_condition: type: leaf, predicate: hw-nvidia-gpu` which makes the skip condition obvious to any reader. SC-1 and SCHM-02 invariant 3 are satisfied.

### 2. Swap Suggestion Scope (SC-3 / RSLV-04) — Option B Implemented

**Decision:** Option B implemented in commit `1b1d046`.

**What was added:** An `AlternativeSuggestion string` field (JSON/YAML tag `alternative_suggestion,omitempty`) was added to the `Explanation` type. When a hardware-conditional opinion is skipped (condition false) and the same composition contains one or more same-category opinions whose hardware condition evaluates TRUE, the skip `Explanation` carries:

```
"You declared <category>: consider '<Name>' (<ID>) instead."
```

Multiple alternatives are listed sorted lexicographically by ID for determinism. The field is `omitempty` so all existing golden JSON files are byte-identical — no golden regeneration was needed.

**Scope:** In-composition only — the resolver scans the opinions slice passed to `Resolve()`, not a registry. Registry-wide swap suggestions (e.g., "there's an AMD driver opinion in the public registry not in your speech") are deferred to Phase 5 as originally scoped.

**Test:** `TestResolveHardwareSwapSuggestion` (EC-037b fixture) — NVIDIA driver skipped (hw-amd-gpu declared), AMD driver in same category with true condition → AlternativeSuggestion = `"You declared hardware-conditional: consider 'AMD GPU Driver' (hardware-conditional/amd-driver) instead."` PASS.

---

## Gaps Summary

No open gaps. All 5 truths verified. All automated checks pass:

- `go test ./... -count=1`: all 5 testable packages green
- `bash scripts/wasm-parity-test.sh`: PARITY OK (4 goldens, native = WASM)
- `bash scripts/check-coverage.sh`: 93.5% >= 90%
- 28 EC corpus tests: all pass
- RSLV-04 AlternativeSuggestion: implemented and tested
- All 8 code review findings (CR-01, CR-02, WR-01 through WR-04, IN-03, IN-04) confirmed fixed

---

_Verified: 2026-06-12T14:00:00Z_
_Updated: 2026-06-12 (RSLV-04 gap closure)_
_Verifier: Claude (gsd-verifier + gsd-code-fixer)_
