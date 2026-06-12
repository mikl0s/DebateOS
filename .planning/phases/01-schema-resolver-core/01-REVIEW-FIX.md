---
phase: 01-schema-resolver-core
fixed_at: 2026-06-12T00:00:00Z
review_path: .planning/phases/01-schema-resolver-core/01-REVIEW.md
iteration: 1
findings_in_scope: 8
fixed: 8
skipped: 0
status: all_fixed
---

# Phase 01: Code Review Fix Report

**Fixed at:** 2026-06-12T00:00:00Z
**Source review:** `.planning/phases/01-schema-resolver-core/01-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 8 (CR-01, CR-02, WR-01, WR-02, WR-03, WR-04, IN-03, IN-04; IN-01/IN-02 fall out of their parent CR fixes)
- Fixed: 8
- Skipped: 0

---

## Fixed Issues

### CR-01: Phase tie-breaking dead code — implemented

**Files modified:** `resolver/graph/toposort.go`, `resolver/graph/graph_test.go`
**Commit:** `4d2162a`
**Applied fix:** Replaced the lexicographic-only `opinionIDHeap` with a `heapEntry` struct heap that carries both the opinion ID and its `phaseOrder` weight. The `Less` function now orders by phase weight first (preflight=1 < packaging=2 < config=3 < login=4 < post-install=5 < first-run=6), then lexicographic ID within the same phase. Unspecified phases (weight 0) sort last using MaxInt32 promotion. All recursive successor pushes pass the successor's phase weight. The existing `phaseOrder` map and `g.phase` field (previously populated but never read) are now fully active.

**Test added:** `TestTopoSortPhaseTieBreak` — two independent opinions with no explicit ordering edges: `"config/aa-dotfiles"` (config phase) and `"pkg/zz-mesa"` (packaging phase). Lexicographically `"config/aa"` < `"pkg/zz"`, but packaging (weight 2) must precede config (weight 3). Test failed before fix, passes after.

IN-01 (dead code note) resolves as a side effect — `phaseOrder` and `g.phase` are now live code.

---

### CR-02: pci_ids absent from resolver.HardwareProfile and speech schema

**Files modified:** `resolver/types.go`, `schemas/speech.schema.json`, `resolver/resolve/resolve_test.go`, `resolver/resolve/testdata/ec038b-pci-ids-in-speech.yaml` (new)
**Commit:** `cbe7595`
**Applied fix:**
- Added `PCIIDs []string` (json:`pci_ids` yaml:`pci_ids`) to `resolver.HardwareProfile` in `types.go`.
- Added `pci_ids` array property to the `hardware` block in `schemas/speech.schema.json`.
- Updated `loadFixture` in `resolve_test.go` to propagate `PCIIDs` directly: `hw.PCIIDs = doc.Speech.Hardware.PCIIDs`. Removed the dead marshal/unmarshal workaround block (lines 60-70 of the old code) — this was always a no-op since `resolver.HardwareProfile` lacked the field. The `hardware_override` path is retained for the legacy EC-038 fixture.
- Updated `loadExample` similarly to propagate `PCIIDs`.

**Test added:** `ec038b-pci-ids-in-speech.yaml` fixture with `pci_ids: [106b:1801]` declared directly in `speech.hardware` (no `hardware_override`). `EC-038b` subtest in `TestResolveEC` verifies the apple-t2-direct opinion is applied via the normal parse path.

IN-02 (dead workaround code) resolves as a side effect — the block is deleted.

---

### WR-01: Multiple required-vs-required hard conflicts — only last pair reported

**Files modified:** `resolver/resolve/resolve.go`, `resolver/resolve/resolve_test.go`
**Commit:** `b29a8d5`
**Applied fix:** Replaced `var hardErr error` / `hardErr = err` with `var hardConflicts []string` / `hardConflicts = append(hardConflicts, err.Error())`. After the pairwise loop, joined all messages with `"; "` and returned the combined error. All conflict pair IDs now appear in `err.Error()`.

**Test added:** `TestResolveMultipleHardConflictsAllReported` — four required opinions in two conflict pairs (A/B and C/D); error message must contain all four opinion IDs. Test failed before fix (only C/D appeared), passes after.

---

### WR-02: SigLevel=Never trust warning only emitted in hardware-apply path

**Files modified:** `resolver/resolve/resolve.go`, `resolver/resolve/resolve_test.go`
**Commit:** `f1a06cf`
**Applied fix:** In the Step 5 "No conflict" loop, call `collectTrustWarning(*op)` for each unexplained opinion and attach the result to `TrustWarning` and `Text` in the emitted `Explanation`. Previously this was only done in the Step 1 hardware-apply branch.

**Test added:** `TestResolveTrustWarningNonHardwareOpinion` — plain required opinion with a `sig_level: Never` custom repo and no `hardware_condition`; its "no-conflict" `Explanation` must have a non-empty `TrustWarning`. Test failed before fix, passes after.

---

### WR-03: HardwareExpr recursion depth unbounded

**Files modified:** `resolver/hardware/eval.go`, `resolver/hardware/hardware_test.go`, `schemas/opinion.schema.json`
**Commit:** `cfe3714`
**Applied fix:**
- Added `const maxHardwareExprDepth = 32` to `eval.go`.
- Refactored `EvalCondition` to delegate to `evalConditionDepth(expr, hw, 0)`.
- `evalConditionDepth` checks `depth > maxHardwareExprDepth` at entry and returns `fmt.Errorf("hardware: expression depth exceeds maximum (%d)", maxHardwareExprDepth)`.
- All recursive calls pass `depth+1`.
- Updated package doc to correct the T-01-07 security note (it incorrectly claimed schema validation bounded depth).
- Added `maxItems: 16` to `hardwareExpr.operands` in `opinion.schema.json` to limit validator recursion fan-out.

**Test added:** `TestHardwareEvalDepthLimit` — builds a `NOT(NOT(...))` chain 100 levels deep; `EvalCondition` must return an error mentioning "depth" rather than overflowing the stack.

---

### WR-04: detectSysctlCollisions early-return prevents pairwise conflict detection

**Files modified:** `resolver/resolve/resolve.go`, `resolver/resolve/resolve_test.go`
**Commit:** `3ca3a34`
**Applied fix:** Changed `if err := detectSysctlCollisions(...); err != nil { return rs, err }` to capture `sysctlErr := detectSysctlCollisions(...)` and continue. At the end of Step 4, combine `sysctlErr` with `hardConflicts` into a single joined error message before returning. Both sysctl collision errors and required-vs-required hard conflicts are now visible in a single pass.

**Test added:** `TestResolveSysctlAndHardConflictBothReported` — two required opinions sharing a sysctl key AND declaring mutual conflicts; returned error must contain both "Sysctl" and "Hard conflict".

---

### IN-03: Misleading Rule 3 comment corrected

**Files modified:** `resolver/resolve/resolve.go`
**Commit:** `19f2cee` (combined with IN-04)
**Applied fix:** Replaced the comment "Pick the first-listed (opA, as the lower-ID in canonical sorted pair since we iterate in sorted order)" with an accurate description: "opA wins because the outer loop iterates in input order (sortedActiveIDs preserves the opinions slice order) and opA's Conflicts list was processed first."

---

### IN-04: KnownFields(true) in WASM YAML fallback

**Files modified:** `resolver/wasm/main.go`
**Commit:** `19f2cee` (combined with IN-03)
**Applied fix:** Replaced `yaml.Unmarshal` with `yaml.NewDecoder(...).KnownFields(true)` in the YAML fallback path. Also replaced `json.Unmarshal` with `json.NewDecoder(...).DisallowUnknownFields()` for the JSON primary path. Both paths now reject unknown fields and emit helpful error messages (e.g., "unknown field 'speeech'") instead of silently ignoring typos and returning "speech field is required".

---

## Skipped Issues

None.

---

## Gate Results

All gates passed after fixes:

- `go test ./...`: PASS (all 7 packages, 0 failures)
- `bash scripts/wasm-parity-test.sh`: PARITY OK (native and WASM outputs byte-identical to committed goldens)
- `bash scripts/check-coverage.sh`: COVERAGE OK (93.4% >= 90% threshold)

Goldens were NOT regenerated — no output behavior changed. The fixes corrected dead code paths, missing fields, and error accumulation logic; existing golden outputs remain valid.

---

_Fixed: 2026-06-12T00:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
