---
phase: 4
slug: debian-translator
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-06-13
---

# Phase 4 — Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | pytest (debian generator + shared common); go test (resolver/cli regression incl. foundation-aware build); bash gate scripts |
| **Config file** | translators/debian/pytest.ini (or shared) |
| **Quick run command** | `pytest translators/debian/tests -q && go test ./cli/... -count=1` |
| **Full suite command** | `go test ./... -count=1 && pytest translators/arch/tests translators/debian/tests -q && bash scripts/dual-foundation-check.sh && bash scripts/check-coverage.sh` |
| **Estimated runtime** | quick < 15s; full < 3 min; ISO builds (slow gate) deferred-to-capable-host |

## Sampling Rate

- **After every task commit:** `pytest translators/debian/tests -q` (+ `go test ./cli/...` when build.go changes)
- **After every plan wave:** full suite (incl. dual-foundation-check.sh, Arch regression, examples)
- **Before `/gsd-verify-work`:** full suite + dual-foundation equivalence; ISO builds documented-deferred
- **Max feedback latency:** 60s (ISO build gates exempt)

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-XX | TBD | 1+ | DEB-01 | — | preseed/hook dst sanitized (no traversal) | unit | `pytest translators/debian/tests/test_profile.py -x` | ❌ W0 | ⬜ pending |
| 04-XX | TBD | 1+ | DEB-02 | — | required+unsupported → CapabilityError names opinion+token | unit | `pytest translators/debian/tests/test_capability_gate.py -x` | ❌ W0 | ⬜ pending |
| 04-XX | TBD | 1+ | DEB-03 | — | apt signed-by/trusted maps sig_level verbatim | unit | `pytest translators/debian/tests/test_variant.py -x` | ❌ W0 | ⬜ pending |
| 04-XX | TBD | 2+ | DEB-01 | — | foundation→translator data-driven (no hardcode) | unit | `go test ./cli/build/... -run TestBuildFoundation -count=1` | ❌ W0 | ⬜ pending |
| 04-XX | TBD | 2+ | DEB-01 | — | N/A | script | `bash scripts/dual-foundation-check.sh` (one resolve → both translators → equivalence) | ❌ W0 | ⬜ pending |
| 04-XX | TBD | 1+ | DEB-03 | — | schema change → no Arch/examples regression | unit | `go test ./... -count=1 && pytest translators/arch/tests -q` | ✅ | ⬜ pending |
| 04-XX | TBD | final | COMM-01 | — | N/A | doc | `grep -q "own their translators" docs/ownership-model.md` + new-translator contract section | ❌ W0 | ⬜ pending |

## Wave 0 Requirements

- [ ] translators/debian/tests/ package + fixtures (RED stubs)
- [ ] examples/dual-foundation/ representative speech (neutral opinions/points)
- [ ] scripts/dual-foundation-check.sh, scripts/debian-build-iso.sh, scripts/debian-validate-iso.sh

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Full Debian ISO build + boot | DEB-01 | Host devtmpfs/loop blocked (Proxmox) — confirmed in research | `docker run --privileged ... lb build` on bare-metal Linux host |
| Full Arch ISO of the same speech | DEB-01 | Same host limitation | run on capable host; compare both installers |

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity OK; no watch-mode; latency < 60s (ISO gates exempt)
- [x] `nyquist_compliant: true`

**Approval:** approved 2026-06-13 (autonomous run)
