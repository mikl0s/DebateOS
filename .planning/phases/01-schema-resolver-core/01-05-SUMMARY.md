---
phase: 01-schema-resolver-core
plan: 05
status: complete
completed: 2026-06-12
requirements: [RSLV-05, RSLV-06]
subsystem: resolver/wasm, examples, scripts
tags: [tdd, wasm, parity, golden-files, examples, coverage-gate, phase-gate]
dependency_graph:
  requires:
    - 01-01 (resolver/types.go — shared Opinion/Speech/HardwareProfile types)
    - 01-02 (resolver/graph — BuildGraph + TopoSort)
    - 01-03 (resolver/hardware — EvalCondition; resolver/patch — FindPatch)
    - 01-04 (resolver/resolve — Resolve + CanonicalJSON; 27 EC corpus)
  provides:
    - resolver/wasm/main.go — js/wasm entrypoint exporting debateosResolve js.Func
    - resolver/resolve/testdata/golden/ — 4 committed canonical-JSON golden files (parity baseline)
    - scripts/wasm-parity-test.sh — byte-identical native vs WASM parity assertion
    - scripts/check-coverage.sh — coverage gate enforcing >=90% over ./resolver/...
    - examples/ — 4 evidence-derived compositions + end-to-end test harness
  affects:
    - Phase 3 CLI (native resolver library consumer)
    - Phase 5 UI (WASM resolver consumer — debateosResolve js.Func)
tech_stack:
  added: []
  patterns:
    - WASM entrypoint: init() registration (available to both go test runner and wasm_exec.js production runtime)
    - Golden-file parity: GOLDEN_DIR env var redirects write-mode to temp dir for diff
    - Coverage gate: go test -coverprofile + go tool cover -func + awk comparison (bc-free)
    - Parity guard: ls count >= 4 prevents diff against empty golden dir
    - Gap-closure supplemental tests: appended to existing *_test.go files under file-ownership exception
key_files:
  created:
    - resolver/wasm/main.go
    - resolver/wasm/main_test.go
    - scripts/wasm-parity-test.sh
    - scripts/check-coverage.sh
    - resolver/resolve/testdata/golden/omarchy-mini.json
    - resolver/resolve/testdata/golden/two-point-clean.json
    - resolver/resolve/testdata/golden/conflicting.json
    - resolver/resolve/testdata/golden/hardware-conditional.json
    - examples/omarchy-mini/speech.yaml
    - examples/omarchy-mini/opinions.yaml
    - examples/two-point-clean/speech.yaml
    - examples/two-point-clean/opinions.yaml
    - examples/conflicting/speech.yaml
    - examples/conflicting/opinions.yaml
    - examples/hardware-conditional/speech.yaml
    - examples/hardware-conditional/opinions.yaml
    - examples/examples_test.go
    - examples/README.md
  modified:
    - resolver/resolve/resolve_test.go (TestCanonicalGolden, loadExample, findModuleRoot, gap-closure tests)
    - resolver/graph/graph_test.go (gap-closure: ordering.before, depends_on, phantom node)
    - resolver/hardware/hardware_test.go (gap-closure: OR, NOT-error, generic values/match, cpu-model-missing)
    - resolver/parse/parse_test.go (gap-closure: ParsePoint/ParseSpeech YAML+schema error paths)
    - resolver/patch/patch_test.go (gap-closure: missing patch, wrong category, unknown conflicting ID)
    - .gitignore (/wasm generated binary)
decisions:
  - WASM function registered in init() (not only main()) so go test runner (which calls init but not main) can call debateosResolve
  - GOLDEN_DIR env var overrides golden directory in TestCanonicalGolden to enable parity script temp-dir diffing
  - Gap-closure tests appended to existing *_test.go files (plans 01-01..01-04 all complete; no parallel writer conflict)
  - Parity script uses diff -r against committed goldens (not cross-diff between native and WASM) for cleaner failure messages
  - /wasm binary added to .gitignore (GOOS=js GOARCH=wasm go build default output name)
metrics:
  duration: ~11 min
  completed_date: 2026-06-12
  tasks_completed: 3
  files_created: 19
commits:
  - 245aa0d test(01-05): RED — TestCanonicalGolden, TestWasmEntryPointSmoke, example end-to-end tests
  - 5506991 feat(01-05): GREEN — WASM entrypoint, 4 example compositions, golden parity, coverage gate
  - 74b1f48 chore(01-05): Task 3 phase gate verification — all green
---

# Phase 1 Plan 05: WASM Parity + Examples + Coverage Gate Summary

One-liner: WASM entrypoint (debateosResolve js.Func) proven byte-identical to native via golden-file parity script, four evidence-derived example compositions (omarchy-mini, two-point-clean, conflicting, hardware-conditional) exercise parse→resolve end-to-end, coverage gate passes at 92.9% — Phase 1 gate satisfied.

## What Was Built

### Package `resolver/wasm` — WASM Entrypoint

**`main.go`** (build tag `//go:build js && wasm`):

```go
func init() {
    // Registered in init() for both production wasm_exec.js and go test runner.
    js.Global().Set("debateosResolve", js.FuncOf(debateosResolveFunc))
}

func main() {
    select {} // keep runtime alive for JS callbacks
}
```

- Input: single JSON string `{"speech": {...}, "opinions": [...], "hardware": {...}}`
- JSON decode with YAML fallback for flexibility
- Calls `resolve.Resolve` + `resolve.CanonicalJSON`; returns `{"result": "<canonical JSON>"}` on success
- Always returns partial result on hard conflict with `"error"` field populated alongside `"result"`
- Returns `{"error": "..."}` on malformed input; never panics (T-01-16)
- No committed wasm_exec.js copy — script references `$(go env GOROOT)/lib/wasm/go_js_wasm_exec` (T-01-15)

**`main_test.go`** (`//go:build js && wasm`):
- `TestWasmEntryPointSmoke`: verifies debateosResolve is registered, returns non-empty valid JSON, with `"result"` key for clean input

### Golden Files — `resolver/resolve/testdata/golden/`

4 committed canonical-JSON golden files (one per example composition):

| File | Size | Description |
|------|------|-------------|
| `omarchy-mini.json` | 778 bytes | All 4 opinions applied; topological order OM-001→OM-006→OM-007→OM-015 |
| `two-point-clean.json` | 335 bytes | Both opinions applied; no-conflict explanations |
| `conflicting.json` | 246 bytes | Partial result — hard conflict explanation for SDDM vs greetd |
| `hardware-conditional.json` | 398 bytes | OM-006 applied; OM-068 skipped (no NVIDIA PCI ID in speech hardware) |

Generated by: `GOLDEN_UPDATE=1 go test ./resolver/resolve/ -run TestCanonicalGolden`

### `scripts/wasm-parity-test.sh`

1. **GUARD**: `[ "$(ls golden | wc -l)" -ge 4 ]` — exits non-zero with descriptive message if golden dir empty/missing
2. **NATIVE run**: `GOLDEN_UPDATE=1 GOLDEN_DIR=$TMP go test ./resolver/resolve/ -run TestCanonicalGolden`; `diff -r golden $TMP`
3. **WASM run**: same test under `GOOS=js GOARCH=wasm go test -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec"`; `diff -r golden $TMP`
4. Exits 0 printing `=== PARITY OK ===` only when both diffs clean

### `scripts/check-coverage.sh`

- Runs `go test -coverprofile ./resolver/... -count=1`
- Parses `go tool cover -func` total line
- Compares against threshold (90%) using `awk` (bc-free)
- Exits 0 with `=== COVERAGE OK ===` or non-zero with `=== COVERAGE FAIL ===`

### Four Example Compositions — `examples/`

All reference real OM-NNN opinion IDs from `research/omarchy-opinion-inventory.md` (D17).

#### `omarchy-mini`

Coherent subset of Omarchy's Hyprland desktop stack:
- OM-001 (custom-repo / required): Omarchy package repository
- OM-006 (package-install / required): Wayland compositor stack — depends on OM-001
- OM-007 (package-install / required): Terminal and shell tools — depends on OM-001
- OM-015 (package-install / required): Desktop shell components — depends on OM-006

Resolution: all 4 applied; install order enforced by depends_on chain.

#### `two-point-clean`

Two non-conflicting opinions:
- OM-007 (package-install / required): Terminal tools
- OM-064 (service-enable / nice-to-have): Bluetooth service

Resolution: both applied; no conflict rules fired.

#### `conflicting`

Deliberately hard conflict (Rule 2 — required vs required):
- OM-015 (required): SDDM display manager stack
- OM-015-greetd (required): greetd/tuigreet alternative — mutual `conflicts:` declarations

Resolution: hard conflict error surfaced; explanation contains "Hard conflict"; partial ResolvedSpeech returned with the conflict text for display.

#### `hardware-conditional`

NVIDIA driver gated on PCI-ID set membership (OM-068 shape from inventory):
- OM-006 (required): always applied
- OM-068 (nice-to-have): `hardware_condition.type=leaf predicate=pci-id values=[10de:2204, ...]`

Resolution with matching hardware: OM-068 Applied, Rule="hardware-apply"
Resolution with empty hardware: OM-068 Skipped, Rule="hardware-skip"

### `examples/examples_test.go`

4 end-to-end tests exercising parse→resolve:
- `TestExampleOmarchyMini`: clean resolve; non-empty InstallOrder + Applied; all IDs have OM- prefix
- `TestExampleTwoPointClean`: both opinions Applied; no Dropped; no rule1/2/3/4 firings
- `TestExampleConflicting`: hard conflict surfaced (error.Error() or Explanation.Text contains "Hard conflict")
- `TestExampleHardwareConditional/matching-hardware`: gated opinion Applied
- `TestExampleHardwareConditional/non-matching-hardware`: gated opinion Skipped with hardware explanation

### Coverage Gap-Closure Tests

Added under the plan's file-ownership exception (all of plans 01-01..01-04 complete; no parallel writers):

| Package | Tests Added | Coverage Gain |
|---------|-------------|---------------|
| resolver/graph | ordering.before, depends_on, phantom ensureNode | 88% → 95.2% |
| resolver/hardware | OR expression, NOT-error, generic-fact-values/match, cpu-model-missing | 73.6% → 92.5% |
| resolver/parse | ParsePoint/ParseSpeech YAML error + schema violation error paths | 76.6% → 79.7% |
| resolver/patch | missing patch opinion ID, wrong category, unknown conflicting ID | 82.1% → 92.9% |
| resolver/resolve | CanonicalJSON nil, malformed hw condition, Rule1-B-drops-A, Rule3, ordering explanation | 92.3% → 96.1% |

**Final coverage: 92.9% total (threshold: 90%)** ✓

## Phase Gate Results

| Gate | Result |
|------|--------|
| `go test ./... -count=1` | PASS — all 7 packages (examples + 5 resolver + wasm) |
| `GOOS=js GOARCH=wasm go test -exec go_js_wasm_exec ./resolver/...` | PASS — 6 packages incl. wasm smoke test |
| `bash scripts/wasm-parity-test.sh` | PARITY OK — 4 goldens; native and WASM byte-identical |
| `bash scripts/check-coverage.sh` | PASS — 92.9% >= 90% |
| `go test ./examples/ -count=1` | PASS — 4 examples + 5 subtests |
| `GOOS=js GOARCH=wasm go build ./resolver/wasm/` | PASS — builds cleanly |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Design] WASM function must be registered in init(), not only main()**
- **Found during:** Task 1 GREEN verification (WASM smoke test)
- **Issue:** `go test` calls `init()` but NOT `main()` when running WASM tests under `go_js_wasm_exec`. The `debateosResolve` js.Func was only registered in `main()`, so the test runner could not find it.
- **Fix:** Moved `js.Global().Set("debateosResolve", ...)` to `init()` so it runs for both the test runner (init only) and the production wasm_exec.js runtime (init + main).
- **Files modified:** `resolver/wasm/main.go`

**2. [Rule 2 - Design] Generated WASM binary needed .gitignore entry**
- **Found during:** Task 1 GREEN (post-commit check)
- **Issue:** `GOOS=js GOARCH=wasm go build ./resolver/wasm/` outputs a file named `wasm` at the module root. The existing `.gitignore` covers `*.wasm` but not the extension-less `wasm` binary.
- **Fix:** Added `/wasm` to `.gitignore`.
- **Files modified:** `.gitignore`

## TDD Gate Compliance

| Gate | Commit | Verified |
|------|--------|---------|
| RED | 245aa0d | `go test ./examples/` → `FAIL` (fixtures missing); `go test ./resolver/resolve/ -run TestCanonicalGolden` → `FAIL` (goldens missing) |
| GREEN | 5506991 | all tests PASS; `wasm-parity-test.sh` → PARITY OK; `check-coverage.sh` → 92.9% |
| Verification | 74b1f48 | Full Phase 1 suite green |

RED commit `245aa0d` precedes GREEN commit `5506991` in git history (D19 satisfied).

## Threat Surface Scan

No new network endpoints or auth paths introduced.

Threat model items addressed:
- **T-01-14 (Tampering — WASM output)**: Parity script asserts byte-identical canonical JSON native vs WASM against committed goldens. Guard prevents comparison against empty dir.
- **T-01-15 (DoS — wasm_exec.js version mismatch)**: Scripts reference `$(go env GOROOT)/lib/wasm/go_js_wasm_exec` exclusively; no committed copy.
- **T-01-16 (Tampering — WASM panic on bad input)**: Entrypoint returns `{"error": "..."}` JSON on all error paths; never panics. Tested by `TestWasmEntryPointSmoke`.

## Known Stubs

None — all four example compositions produce stable ResolvedSpeech outputs; WASM entrypoint is fully functional; parity and coverage scripts are production-ready.

## RSLV Requirements Satisfied

| Req | Description | Status |
|-----|-------------|--------|
| RSLV-05 | Native and WASM produce identical results (automated parity) | SATISFIED — parity script exits 0, byte-identical canonical JSON |
| RSLV-06 (remainder) | Coverage >=90%; 3-4 example compositions incl. conflicting | SATISFIED — 92.9% coverage; 4 examples incl. conflicting + hardware-conditional |

## Self-Check: PASSED

- `resolver/wasm/main.go` exists: FOUND
- `scripts/wasm-parity-test.sh` exists: FOUND
- `scripts/check-coverage.sh` exists: FOUND
- `resolver/resolve/testdata/golden/` holds >= 4 files: FOUND (4)
- All 4 example directories exist: FOUND
- `examples/examples_test.go` exists: FOUND
- RED commit `245aa0d` exists in git log: VERIFIED
- GREEN commit `5506991` exists after RED: VERIFIED
- `go test ./... -count=1` → PASS
- WASM tests under go_js_wasm_exec → PASS (6 packages)
- `bash scripts/wasm-parity-test.sh` → PARITY OK
- `bash scripts/check-coverage.sh` → 92.9% PASS
