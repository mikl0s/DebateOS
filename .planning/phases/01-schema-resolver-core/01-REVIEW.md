---
phase: 01-schema-resolver-core
reviewed: 2026-06-12T00:00:00Z
depth: standard
files_reviewed: 18
files_reviewed_list:
  - resolver/types.go
  - resolver/parse/parse.go
  - resolver/parse/validate.go
  - resolver/parse/schemas_embed.go
  - resolver/parse/parse_test.go
  - resolver/graph/graph.go
  - resolver/graph/toposort.go
  - resolver/graph/graph_test.go
  - resolver/hardware/eval.go
  - resolver/hardware/hardware_test.go
  - resolver/patch/patch.go
  - resolver/patch/patch_test.go
  - resolver/resolve/resolve.go
  - resolver/resolve/explanation.go
  - resolver/resolve/canonical.go
  - resolver/resolve/resolve_test.go
  - resolver/wasm/main.go
  - schemas/opinion.schema.json
  - schemas/point.schema.json
  - schemas/speech.schema.json
  - schemas/embed.go
  - scripts/wasm-parity-test.sh
  - scripts/check-coverage.sh
  - examples/examples_test.go
findings:
  critical: 2
  warning: 4
  info: 4
  total: 10
status: issues_found
fixes_applied: true
resolved_ids:
  - CR-01
  - CR-02
  - WR-01
  - WR-02
  - WR-03
  - WR-04
  - IN-02
  - IN-03
  - IN-04
fix_report: 01-REVIEW-FIX.md
---

# Phase 01: Code Review Report

**Reviewed:** 2026-06-12T00:00:00Z
**Depth:** standard
**Files Reviewed:** 24
**Status:** issues_found

## Summary

Phase 01 delivers a well-structured, rule-based resolver that correctly implements the docs/04
four-rule hierarchy, topological sort (Kahn's algorithm with deterministic heap tie-breaking),
hardware-conditional evaluation, patch discovery, schema validation, and WASM entrypoint. The
type system is clean, invariant 1 (no distro mechanics) is respected throughout, and D6
(no SAT solver) holds. The 27-case EC corpus is fully wired to fixtures and the tests are
substantive rather than trivially non-error checks.

Two critical findings stand out. First, the documented phase-based tie-breaking in the
topological sort is entirely dead code: `phaseOrder` and `g.phase` are populated but never
read; toposort uses only lexicographic string ordering. This can produce incorrect install
order for independent opinions in different phases when no explicit ordering edges are declared.
Second, the speech schema and `resolver.HardwareProfile` type both lack a `pci_ids` field,
meaning PCI-ID hardware conditions cannot be expressed in a speech document via the production
parse path. The EC-038 test works only because of a test-only `hardware_override` fixture block
that bypasses `ParseSpeech` entirely.

---

## Critical Issues

### CR-01: Phase tie-breaking is documented but completely unimplemented — dead code in production sort path

**File:** `resolver/graph/graph.go:21-42` and `resolver/graph/toposort.go:40-101`

**Issue:** `graph.go` declares `phaseOrder` (a `map[string]int` mapping install-phase names to
priority weights) and `g.phase` (a per-node phase map), and the package doc states "Phase enum
is an additional tie-breaking key." The `BuildGraph` function populates `g.phase[op.ID]` for
every opinion (line 64). However, `toposort.go` never reads `g.phase` or `phaseOrder` — it
tie-breaks exclusively via a `container/heap` min-heap on lexicographic `OpinionID` string
order. The `CrossPhase` test in `graph_test.go` passes only because it uses an explicit
`Ordering.After` edge, not phase-based tie-breaking.

Consequence: two independent opinions with no explicit ordering edges, one in `"packaging"` phase
and one in `"config"` phase, can be sorted in the wrong order if the config-phase opinion's ID
is lexicographically smaller. For example, opinion `"config/aa-dotfiles"` (config phase) would
be placed before `"pkg/zz-mesa"` (packaging phase) despite packaging preceding config in the
install pipeline. The documented invariant — "phase enum is a tie-breaking key" — is silently
violated.

**Fix:** Either implement phase-based tie-breaking in `TopoSort` using the existing `g.phase`
data, or remove the dead `phaseOrder`/`g.phase` structures and correct the package doc to say
"tie-breaking is lexicographic on OpinionID only." Implementing phase tie-breaking is the
correct fix:

```go
// In toposort.go, replace the opinionIDHeap Less function:
// Current: lexicographic only
func (h opinionIDHeap) Less(i, j int) bool { return h[i] < h[j] }

// Correct: phase-order first, then lexicographic for same-phase tie-breaking
// Pass g into TopoSort so it can consult g.phase[id] + phaseOrder[phase].
// Or: embed the phase-weight alongside the ID in a struct heap.

type heapEntry struct {
    id    resolver.OpinionID
    phase int // from phaseOrder[g.phase[id]], 0 if unspecified
}

// Less: lower phase weight = earlier; same phase → lexicographic on ID
func (h heapEntries) Less(i, j int) bool {
    if h[i].phase != h[j].phase {
        // unspecified (0) sorts last, so treat 0 as MaxInt for comparison
        pi, pj := h[i].phase, h[j].phase
        if pi == 0 { pi = 1<<31 - 1 }
        if pj == 0 { pj = 1<<31 - 1 }
        return pi < pj
    }
    return h[i].id < h[j].id
}
```

---

### CR-02: `pci_ids` field is absent from both `resolver.HardwareProfile` (types.go) and `speech.schema.json` — PCI-ID hardware conditions are inexpressible in production speech documents

**File:** `resolver/types.go:225-228`, `schemas/speech.schema.json:40-49`, `resolver/resolve/resolve_test.go:55-70`

**Issue:** `hardware.HardwareProfile` (in `resolver/hardware/eval.go`) carries a `PCIIDs
[]string` field that is essential for PCI-ID set-membership predicates (used by EC-038, the
Apple T2 case, and the `examples/hardware-conditional` example). However, `resolver.HardwareProfile`
in `types.go` — the type that holds speech-level hardware declarations and is the only type
reachable through `ParseSpeech` — has **no `PCIIDs` / `pci_ids` field**:

```go
// types.go — no PCIIDs
type HardwareProfile struct {
    Predicates []string          `json:"predicates,omitempty" yaml:"predicates,omitempty"`
    Facts      map[string]string `json:"facts,omitempty" yaml:"facts,omitempty"`
}
```

The `speech.schema.json` hardware block also lacks `pci_ids` and has `additionalProperties:
false`, so a YAML speech file that includes `hardware.pci_ids` would be **rejected by
`ParseSpeech`** schema validation. The only working path is the test-only `hardware_override`
fixture block in `resolve_test.go` (line 24), which bypasses the speech parse entirely.

The code comment at lines 55-70 of `resolve_test.go` acknowledges the gap: "resolver.HardwareProfile
does NOT carry PCIIDs. We need to re-parse the hardware block raw." The round-trip workaround
(lines 62-69) is itself dead code: `yaml.Marshal(doc.Speech.Hardware)` serializes a
`resolver.HardwareProfile` value, which has no `PCIIDs` field, so the re-decode struct
`struct{ PCIIDS []string \`yaml:"pci_ids"\` }` always produces an empty slice. The actual PCI
IDs come only from `hardware_override`.

Consequence: no production speech YAML can declare PCI IDs today. Any user trying to use a
`pci-id` hardware condition in a real speech file will have hardware evaluation silently return
false for all PCI predicates, with no error, causing hardware-conditional opinions to be
incorrectly skipped.

**Fix:** Add `PCIIDs []string` to `resolver.HardwareProfile` in `types.go` (with proper JSON/YAML
tags), add `pci_ids` to the `speech.schema.json` hardware block, and propagate it through
`Resolve`'s hardware profile construction:

```go
// types.go
type HardwareProfile struct {
    Predicates []string          `json:"predicates,omitempty" yaml:"predicates,omitempty"`
    Facts      map[string]string `json:"facts,omitempty" yaml:"facts,omitempty"`
    PCIIDs     []string          `json:"pci_ids,omitempty"   yaml:"pci_ids,omitempty"`
}
```

```json
// speech.schema.json — inside hardware.properties:
"pci_ids": { "type": "array", "items": { "type": "string" },
             "description": "PCI device IDs present on the machine (vendor:device hex, e.g. 106b:1801)." }
```

In `resolve.go` and `resolve_test.go`, propagate `speech.Hardware.PCIIDs` into
`hardware.HardwareProfile.PCIIDs` directly (removing the dead `hardware_override` workaround
once the type carries the field). Delete the dead workaround block at lines 60-70 of
`resolve_test.go`.

---

## Warnings

### WR-01: Multiple required-vs-required hard conflicts — the returned error names only the last pair, not all

**File:** `resolver/resolve/resolve.go:120,148`

**Issue:** When multiple required-vs-required conflicts exist in a single speech, each call to
`resolveConflict` appends an explanation to `rs.Explanations` (correct) but also overwrites
`hardErr` (line 148: `hardErr = err`). After the loop, only the **last** conflict pair's error
message survives in the returned error value. With three hard conflicts, the returned
`err.Error()` will say only `"Hard conflict: E and F are both required..."` — the pairs A/B and
C/D are silenced in the error value even though their explanations are in `rs.Explanations`.

Callers (e.g., CLI, WASM JS caller) that surface the error string directly will show only the
last conflict. This is misleading when the user needs to resolve all conflicts simultaneously.

**Fix:** Accumulate all hard-conflict error messages rather than overwriting:

```go
// Replace:
var hardErr error
// ...
hardErr = err

// With:
var hardConflicts []string
// ...
hardConflicts = append(hardConflicts, err.Error())

// After the loop:
if len(hardConflicts) > 0 {
    return rs, fmt.Errorf("%s", strings.Join(hardConflicts, "; "))
}
```

---

### WR-02: `collectTrustWarning` only emits `SigLevel=Never` warnings in the hardware-apply path — non-hardware-conditional opinions with `Never` repos get no warning

**File:** `resolver/resolve/resolve.go:79,543-549`

**Issue:** The T-01-10 mitigation comment claims `SigLevel=Never` repos surface a visible trust
warning. However, `collectTrustWarning` is called **only** from the `hardware-apply` explanation
branch (Step 1 of Resolve, line 79). A non-hardware-conditional opinion (e.g., a plain
`package-install` opinion) that declares a `custom_repo` with `sig_level: "Never"` will be
applied silently with no trust warning — the `TrustWarning` field in its explanation will be
empty.

The EC-038 test verifies the warning for the T2 opinion specifically, but only because that
opinion happens to be hardware-conditional. An opinion like:

```yaml
id: "custom-repo/myrepo"
category: "custom-repo"
status: "required"
custom_repos:
  - name: "myrepo"
    url: "https://example.com/repo"
    sig_level: "Never"
```

...will be applied with no `TrustWarning` in any explanation.

**Fix:** Call `collectTrustWarning` (or an equivalent check) during the "No conflict" or
"Applied" explanation emission in Step 5, and attach it to the explanation for any applied
opinion that carries a `Never` sig-level repo:

```go
// In the Step 5 "no-conflict" loop:
for _, id := range noConflictIDs {
    if !explainedOps[id] {
        op := index[id]
        trustWarn := ""
        if op != nil {
            trustWarn = collectTrustWarning(*op)
        }
        text := fmt.Sprintf("No conflict: %s applied.", id)
        if trustWarn != "" {
            text += " " + trustWarn
        }
        rs.Explanations = append(rs.Explanations, Explanation{
            Text:             text,
            Rule:             "no-conflict",
            OpinionsInvolved: []resolver.OpinionID{id},
            Kept:             []resolver.OpinionID{id},
            TrustWarning:     trustWarn,
        })
    }
}
```

---

### WR-03: `HardwareExpr` recursion depth is unbounded — no schema-level or evaluator-level depth limit

**File:** `resolver/hardware/eval.go:7-10` (security note), `schemas/opinion.schema.json:73-87` (hardwareExpr def)

**Issue:** The security note in `hardware/eval.go` states "expression depth is bounded by the
validated document depth (JSON Schema validates the input before it reaches here)." This is
incorrect: the `hardwareExpr` definition in `opinion.schema.json` uses a recursive `$ref` for
`operands` items with **no `maxItems` constraint and no nesting depth limit**. A hostile YAML
opinion with a HardwareExpr tree nested 10,000 levels deep would:

1. Pass schema validation (no depth constraint).
2. Trigger recursive calls to `EvalCondition` 10,000 frames deep, potentially causing a goroutine
   stack overflow.

Similarly, the JSON schema validator itself (`santhosh-tekuri/jsonschema/v6`) recurses into nested
`hardwareExpr` nodes during validation, which could also overflow for extreme nesting.

**Fix:** Add a `maxItems` constraint on `operands` in the schema (limiting fan-out) and add a
depth parameter to `EvalCondition` to enforce a maximum recursion depth:

```json
// opinion.schema.json — in hardwareExpr.properties:
"operands": {
    "type": "array",
    "maxItems": 64,
    "items": { "$ref": "#/$defs/hardwareExpr" }
}
```

```go
// hardware/eval.go — add depth guard:
const maxHardwareExprDepth = 32

func EvalCondition(expr resolver.HardwareExpr, hw HardwareProfile) (bool, error) {
    return evalConditionDepth(expr, hw, 0)
}

func evalConditionDepth(expr resolver.HardwareExpr, hw HardwareProfile, depth int) (bool, error) {
    if depth > maxHardwareExprDepth {
        return false, fmt.Errorf("hardware: expression depth exceeds maximum (%d)", maxHardwareExprDepth)
    }
    // ... existing switch, replacing recursive EvalCondition calls with evalConditionDepth(..., depth+1)
}
```

---

### WR-04: `detectSysctlCollisions` returns early, preventing pairwise conflict detection when a sysctl collision and a `required`-vs-`required` conflict coexist

**File:** `resolver/resolve/resolve.go:109-111`

**Issue:**

```go
if err := detectSysctlCollisions(opinions, active, dropped, rs, index); err != nil {
    return rs, err
}
```

When a speech has both a sysctl key collision (SR-016) AND a required-vs-required conflict
(Rule 2), the early return on sysctl collision means the Rule 2 conflict is never detected or
explained. The user sees only the sysctl error and must fix it, then re-run, only to discover
the hard conflict. This violates the "collect all conflicts before returning" philosophy applied
in the pairwise conflict loop (lines 148-150).

**Fix:** Collect the sysctl error without early-returning, then combine it with any hard-conflict
error found in Step 4:

```go
sysctlErr := detectSysctlCollisions(opinions, active, dropped, rs, index)
// ... continue to pairwise conflict detection ...
// At the end, combine errors:
if sysctlErr != nil || hardErr != nil {
    msgs := []string{}
    if sysctlErr != nil { msgs = append(msgs, sysctlErr.Error()) }
    if hardErr != nil   { msgs = append(msgs, hardErr.Error()) }
    return rs, fmt.Errorf("%s", strings.Join(msgs, "; "))
}
```

---

## Info

### IN-01: Dead code — `phaseOrder` map and `g.phase` field are populated but never read

**File:** `resolver/graph/graph.go:21-28,41-42,64`

**Issue:** Beyond the correctness implication described in CR-01, if the decision is made to
use lexicographic-only tie-breaking (acceptable if all opinions use explicit ordering edges),
then `phaseOrder`, `g.phase`, and all code that writes to them should be deleted. Currently
they are dead weight that misleads readers into thinking phase-based ordering is active.

**Fix:** Either implement phase tie-breaking (preferred, see CR-01 fix) or delete the dead
fields and correct the package-level comment to say "tie-breaking is lexicographic on OpinionID
only."

---

### IN-02: Dead code — `loadFixture` workaround block (lines 60-70 of `resolve_test.go`) can never populate `PCIIDs` from `speech.Hardware`

**File:** `resolver/resolve/resolve_test.go:55-70`

**Issue:** The block at lines 60-70 marshals `doc.Speech.Hardware` (a `*resolver.HardwareProfile`)
through `yaml.Marshal`, then unmarshals into `struct{ PCIIDS []string \`yaml:"pci_ids"\` }`.
Since `resolver.HardwareProfile` has no `PCIIDs` field, `yaml.Marshal` produces YAML with only
`predicates` and `facts`. The `rawHWBlock.PCIIDS` is always empty. The block is dead code.
The comment at lines 55-59 describing this as a "workaround" is also factually incorrect: the
workaround cannot work by design.

**Fix:** Delete lines 60-70 of `loadFixture`. After CR-02 is resolved (adding `pci_ids` to
`resolver.HardwareProfile`), the propagation path for PCI IDs will be straightforward and no
workaround is needed.

---

### IN-03: Misleading comment in `resolveConflict` — "iterate in sorted order" is false; `sortedActiveIDs` returns input order

**File:** `resolver/resolve/resolve.go:358`

**Issue:** The comment at line 358 reads:

> `// Pick the first-listed (opA, as the lower-ID in canonical sorted pair`
> `// since we iterate in sorted order). opA is the "winner" / "default".`

`sortedActiveIDs` (called at line 118) preserves **input order**, not lexicographic order. The
winning opinion in Rule 3 is the one listed **first in the `opinions` slice passed to `Resolve`**,
which is input order. The comment's claim that opA is "the lower-ID in canonical sorted pair"
happens to be true in the EC-033 test fixture (foot < ghostty lexicographically and foot is
listed first), but would be false if a fixture listed ghostty before foot.

**Fix:** Correct the comment to accurately describe the "first in input order wins" semantic:

```go
// Rule 3: nice-to-have vs nice-to-have — first-listed in the input opinions slice wins.
// opA is the winner because the outer loop iterates in input order and opA's Conflicts
// list was processed first.
```

---

### IN-04: WASM YAML fallback decode does not reject unknown fields — silent typo acceptance

**File:** `resolver/wasm/main.go:65-67`

**Issue:** The WASM entrypoint's JSON-then-YAML fallback decode:

```go
if yamlErr := yaml.Unmarshal([]byte(inputStr), &input); yamlErr != nil { ... }
```

uses `yaml.Unmarshal`, which does NOT enable `KnownFields(true)`. An unknown field (e.g., a
typo like `"speeech"` instead of `"speech"`) in YAML input to the WASM endpoint is silently
ignored, with the speech field left nil. The endpoint then returns `{"error": "speech field is
required"}` rather than the more helpful "unknown field 'speeech'". This makes WASM debugging
harder.

For JSON input (the primary production path), `json.Unmarshal` does not reject unknown fields
either; there is no `json.Decoder.DisallowUnknownFields()` applied.

**Fix:** For YAML: use a decoder with `KnownFields(true)`. For JSON: use `json.NewDecoder` with
`DisallowUnknownFields()`:

```go
// JSON path:
dec := json.NewDecoder(strings.NewReader(inputStr))
dec.DisallowUnknownFields()
if err := dec.Decode(&input); err != nil {
    // fall through to YAML
}

// YAML fallback:
dec := yaml.NewDecoder(strings.NewReader(inputStr))
dec.KnownFields(true)
if yamlErr := dec.Decode(&input); yamlErr != nil { ... }
```

---

_Reviewed: 2026-06-12T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
