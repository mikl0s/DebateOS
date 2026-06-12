---
phase: 1
slug: schema-resolver-core
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-06-12
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.24.1 stdlib) |
| **Config file** | none — Wave 0 bootstraps go.mod |
| **Quick run command** | `go test ./resolver/...` |
| **Full suite command** | `go test ./resolver/... && GOOS=js GOARCH=wasm go test -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./resolver/... && bash scripts/wasm-parity-test.sh && bash scripts/check-coverage.sh` |
| **Estimated runtime** | quick < 5s; full < 30s (WASM compile ~10–15s) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./resolver/...`
- **After every plan wave:** Run full suite (incl. WASM parity + coverage check)
- **Before `/gsd-verify-work`:** Full suite green, coverage ≥ 90%, WASM parity PASS
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-XX | TBD | 0 | SCHM-01 | — | strict parse (KnownFields), alias-bomb safe | unit | `go test ./resolver/parse/...` | ❌ W0 | ⬜ pending |
| 01-XX | TBD | 0 | SCHM-01 | — | N/A | unit | `go test ./resolver/parse/ -run TestParseOpinionSR` (SR-001..SR-022 expressible) | ❌ W0 | ⬜ pending |
| 01-XX | TBD | 0 | SCHM-02 | — | N/A | unit | `go test ./resolver/parse/ -run TestSchemaOSAgnostic` | ❌ W0 | ⬜ pending |
| 01-XX | TBD | 1+ | RSLV-01 | — | N/A | unit | `go test ./resolver/resolve/ -run TestResolveEC` (27 EC-NNN, IDs in test names) | ❌ W0 | ⬜ pending |
| 01-XX | TBD | 1+ | RSLV-02 | — | N/A | unit | `go test ./resolver/patch/ -run TestPatchDiscovery` | ❌ W0 | ⬜ pending |
| 01-XX | TBD | 1+ | RSLV-03 | — | deterministic tie-breaking; cycle error names opinions | unit | `go test ./resolver/graph/ -run TestTopoSort` | ❌ W0 | ⬜ pending |
| 01-XX | TBD | 1+ | RSLV-04 | — | N/A | unit | `go test ./resolver/hardware/ -run TestHardwareEval` | ❌ W0 | ⬜ pending |
| 01-XX | TBD | final | RSLV-05 | — | N/A | parity | `bash scripts/wasm-parity-test.sh` (byte-identical canonical JSON) | ❌ W0 | ⬜ pending |
| 01-XX | TBD | final | RSLV-06 | — | N/A | coverage | `bash scripts/check-coverage.sh` (≥90% resolver pkgs, 100% docs/04 rule branches) | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` / module bootstrap (`go 1.24`)
- [ ] `resolver/parse/parse_test.go` — SCHM-01, SCHM-02 stubs
- [ ] `resolver/graph/graph_test.go` — RSLV-03 stubs
- [ ] `resolver/resolve/resolve_test.go` — RSLV-01/06 EC table stubs
- [ ] `resolver/patch/patch_test.go` — RSLV-02 stubs
- [ ] `resolver/hardware/hardware_test.go` — RSLV-04 stubs
- [ ] EC fixture YAML files under `resolver/resolve/testdata/`
- [ ] `scripts/wasm-parity-test.sh`, `scripts/check-coverage.sh`
- [ ] `schemas/*.schema.json` + `schemas/README.md` traceability table

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| "Understandable from YAML alone" (invariant 3) | SCHM-01 | Semantic readability judgment | Read examples/*.yaml cold; confirm composition + resolution comprehensible without code |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-06-12 (autonomous run)
