---
phase: 01-schema-resolver-core
plan: 03
status: complete
completed: 2026-06-12
requirements: [RSLV-02, RSLV-04]
subsystem: resolver/hardware + resolver/patch
tags: [tdd, hardware-eval, patch-discovery, ec-037, ec-038, ec-032]
dependency_graph:
  requires: [01-01]
  provides: [EvalCondition, HardwareProfile, FindPatch, PatchOffer]
  affects: [01-04]
tech_stack:
  added: []
  patterns:
    - discriminated-union recursive evaluator (HardwareExpr)
    - known_patches index-based patch discovery with deterministic sort
key_files:
  created:
    - resolver/hardware/eval.go
    - resolver/hardware/hardware_test.go
    - resolver/hardware/testdata/profile-amd.yaml
    - resolver/hardware/testdata/profile-t2.yaml
    - resolver/patch/patch.go
    - resolver/patch/patch_test.go
    - resolver/patch/testdata/dracut-bridge.yaml
  modified: []
decisions:
  - hardware.HardwareProfile defined in resolver/hardware (not alias of resolver.HardwareProfile) to carry PCIIDs []string as a first-class slice field — required for EC-038 PCI set-membership evaluation
  - Patch discovery scans known_patches on BOTH conflicting opinions (not just opinion a) — ensures patches declared on either side of a conflict are found symmetrically
  - FindPatch sorts candidate PatchIDs lexicographically and returns smallest — deterministic when multiple patches could apply
metrics:
  duration: ~5 min
  completed_date: 2026-06-12
  tasks: 2
  files: 7
commits:
  - c55a237 test(01-03): RED — hardware eval tests (EC-037, EC-038, compound, set-membership)
  - 43071dd feat(01-03): GREEN — EvalCondition hardware predicate evaluator (EC-037, EC-038)
  - 642e22c test(01-03): RED — patch discovery tests (EC-032, no-patch, symmetry)
  - dd0cd9f feat(01-03): GREEN — FindPatch patch opinion discovery (EC-032, RSLV-02)
---

# Phase 1 Plan 03: Hardware Eval + Patch Discovery Summary

One-liner: EvalCondition for compound hardware predicates (leaf/and/or/not/set-membership/string-match) and FindPatch for order-independent patch opinion discovery, both green with EC-037/EC-038/EC-032 tests.

## What Was Built

### Package `resolver/hardware`

`EvalCondition(expr resolver.HardwareExpr, hw HardwareProfile) (bool, error)` recursively evaluates the discriminated-union `HardwareExpr` tree against a declared `HardwareProfile`.

**`HardwareProfile` struct** (defined in this package, richer than `resolver.HardwareProfile`):
```go
type HardwareProfile struct {
    Predicates []string          // named boolean predicates active on this machine
    Facts      map[string]string // key→value hardware facts (gpu, cpu_model, dmi_product_name, ...)
    PCIIDs     []string          // PCI device IDs in "vendor:device" hex format (e.g. "106b:1801")
}
```

**`EvalCondition` node types:**
| Type | Behaviour |
|------|-----------|
| `leaf` (no Values, no Match) | Boolean: predicate name in `profile.Predicates` |
| `leaf` + Values + `pci-id` | PCI ID membership: any value in `profile.PCIIDs` |
| `leaf` + Values + `cpu-model-in-set` | `profile.Facts["cpu_model"]` in values set |
| `leaf` + Values (other) | `profile.Facts[pred]` in values set |
| `leaf` + Match + `dmi-product-match` | `strings.Contains(profile.Facts["dmi_product_name"], match)` |
| `and` | All operands true (short-circuit) |
| `or` | Any operand true (short-circuit) |
| `not` | Single operand; result negated |
| unknown type | Returns `(false, error)` — T-01-08 |

**Test coverage:**
- `TestHardwareEval/EC-037`: NVIDIA leaf predicate false on AMD Radeon profile → opinion skipped
- `TestHardwareEval/EC-038`: PCI OR predicate (106b:1801 OR 106b:1802) true on Apple T2 profile → opinion applied
- `TestHardwareEvalCompound`: three-predicate AND with NOT (OM-077 shape: intel AND battery AND NOT "XPS")
- `TestHardwareEvalSetMembership`: cpu-model set-membership (OM-071 shape)
- `TestHardwareEvalErrors`: unknown Type returns error (T-01-08)

### Package `resolver/patch`

`FindPatch(a, b resolver.OpinionID, opinions []resolver.Opinion) *PatchOffer` discovers patch opinions attached to a conflict pair, returning nil when none exist.

**`PatchOffer` struct:**
```go
type PatchOffer struct {
    PatchID resolver.OpinionID       // ID of the discovered patch opinion
    Pair    [2]resolver.OpinionID    // canonical sorted pair: Pair[0] <= Pair[1]
}
```

**Discovery algorithm:**
1. Build `index[OpinionID] → *Opinion` for O(1) lookup.
2. Compute canonical `{min(a,b), max(a,b)}` pair for deterministic `Pair` field.
3. Scan `known_patches` on both opinion `a` and opinion `b`.
4. For each `PatchRef`, look up the referenced opinion; verify `category == "patch"`.
5. Sort qualifying candidates by PatchID; return `PatchOffer` for lexicographically smallest.

**Test coverage:**
- `TestPatchDiscovery/EC-032`: mkinitcpio↔dracut conflict finds `patch/dracut-omarchy-bridge`
- `TestPatchDiscoveryNoPatch`: pair with no `known_patches` returns nil
- `TestPatchDiscoverySymmetric`: `FindPatch(a,b)` ≡ `FindPatch(b,a)` — pair order-independent

### Testdata Fixtures

- `resolver/hardware/testdata/profile-amd.yaml`: AMD Radeon RX 7600 workstation, no NVIDIA, no battery
- `resolver/hardware/testdata/profile-t2.yaml`: Apple MacBook Pro with T2 (PCI 106b:1801), battery present
- `resolver/patch/testdata/dracut-bridge.yaml`: `patch/dracut-omarchy-bridge` opinion (category: patch)

## Public API for 01-04 (resolve)

```go
// package hardware
type HardwareProfile struct {
    Predicates []string
    Facts      map[string]string
    PCIIDs     []string
}
func EvalCondition(expr resolver.HardwareExpr, hw HardwareProfile) (bool, error)

// package patch
type PatchOffer struct {
    PatchID resolver.OpinionID
    Pair    [2]resolver.OpinionID
}
func FindPatch(a, b resolver.OpinionID, opinions []resolver.Opinion) *PatchOffer
```

## TDD Gate Compliance

| Gate | Commit | Verified |
|------|--------|---------|
| Task 1 RED | c55a237 | tests fail before implementation |
| Task 1 GREEN | 43071dd | all hardware tests pass |
| Task 2 RED | 642e22c | tests fail before implementation |
| Task 2 GREEN | dd0cd9f | all patch tests pass |

RED commits precede GREEN commits in history for both tasks (D19 satisfied).

## Deviations from Plan

### Design Decisions

**1. `hardware.HardwareProfile` as a distinct package-local struct**
- **Reason:** `resolver.HardwareProfile` in `types.go` only has `Predicates []string` and `Facts map[string]string`. EC-038 requires PCI ID set-membership (`106b:1801`), which is naturally a `[]string` field, not a string-valued fact. Adding `PCIIDs []string` to the hardware package's profile keeps the evaluation clean.
- **Impact:** 01-04 will need to convert `resolver.HardwareProfile` (from Speech) to `hardware.HardwareProfile` at the evaluation boundary. Minimal adaptation code.
- **Files modified:** `resolver/hardware/eval.go` (defines `HardwareProfile`)

**2. Patch scan on both opinions (not just `a`)**
- **Reason:** EC-032 has `known_patches` on the mkinitcpio opinion. But in a symmetric call `FindPatch(dracut, mkinitcpio, ...)`, `a = dracut` has no `known_patches`. Scanning both ensures discovery regardless of argument order.
- **Impact:** Satisfies symmetry requirement (`TestPatchDiscoverySymmetric`).

No out-of-scope auto-fixes needed. Plan executed exactly (with the two design decisions noted above).

## Known Stubs

None — both packages are fully functional and implement their complete specified behavior.

## Threat Flags

No new threat surface beyond the threat model in the plan. `EvalCondition` returns error on unknown types (T-01-08). Recursion is bounded by schema-validated document depth (T-01-07). `FindPatch` trusts opinion metadata as data per T-01-09 (Phase 1 accept).

## Self-Check: PASSED

- `go test ./resolver/hardware/ ./resolver/patch/ -count=1` — PASS (11 tests)
- `go vet ./resolver/hardware/ ./resolver/patch/` — clean
- `go build ./...` — clean
- EC-037 in subtest name: FOUND
- EC-038 in subtest name: FOUND
- EC-032 in subtest name: FOUND
- RED commits precede GREEN: VERIFIED (git log c55a237 → 43071dd → 642e22c → dd0cd9f)
