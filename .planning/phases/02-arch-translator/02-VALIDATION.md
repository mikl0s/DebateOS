---
phase: 2
slug: arch-translator
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-06-12
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | pytest (Arch official pkg) for generator; go test for resolver regression; bash gated scripts for slow paths |
| **Config file** | `translators/arch/pytest.ini` — Wave 0 |
| **Quick run command** | `pytest translators/arch/tests/ -x -q` |
| **Full suite command** | `pytest translators/arch/tests/ -v && go test ./... -count=1` |
| **Estimated runtime** | quick < 10s; full < 60s; slow gates (Docker ISO build) 10–40 min, run at wave/phase verification only |

---

## Sampling Rate

- **After every task commit:** `pytest translators/arch/tests/ -x -q` (plus `go test ./examples/...` when examples change)
- **After every plan wave:** full suite (pytest -v + go test ./...)
- **Before `/gsd-verify-work`:** full suite green + slow gates: `scripts/arch-build-iso.sh`, `scripts/arch-validate-iso.sh`, `scripts/arch-northstar-check.sh`
- **Max feedback latency:** 60s (fast loop); slow gates exempt by design (documented)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-XX | TBD | 1+ | ARCH-01 | — | generated paths sanitized (no traversal from file_assets dst) | unit | `pytest translators/arch/tests/test_profile.py -x` | ❌ W0 | ⬜ pending |
| 02-XX | TBD | 1+ | ARCH-01 | — | N/A | integration (slow gate) | `bash scripts/arch-build-iso.sh` + `bash scripts/arch-validate-iso.sh` | ❌ W0 | ⬜ pending |
| 02-XX | TBD | 1+ | ARCH-02 | — | N/A | unit | `go test ./examples/ -run TestExampleOmarchy` + `pytest translators/arch/tests/test_northstar.py -x` | ❌ W0 | ⬜ pending |
| 02-XX | TBD | final | ARCH-02 | — | N/A | integration (slow gate) | `bash scripts/arch-northstar-check.sh` | ❌ W0 | ⬜ pending |
| 02-XX | TBD | 1+ | ARCH-03 | — | fails loudly: CapabilityError names opinion + missing capability | unit | `pytest translators/arch/tests/test_capability_gate.py -x` | ❌ W0 | ⬜ pending |
| 02-XX | TBD | 1+ | ARCH-04 | — | [UNVERIFIED] variant data marked in profiles | unit | `pytest translators/arch/tests/test_variant.py -x` | ❌ W0 | ⬜ pending |
| 02-XX | TBD | all | regression | — | N/A | unit | `go test ./... -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `translators/arch/pytest.ini` + `translators/arch/tests/` package layout
- [ ] `translators/arch/tests/test_profile.py`, `test_capability_gate.py`, `test_variant.py`, `test_northstar.py` (RED stubs per TDD)
- [ ] `scripts/arch-build-iso.sh`, `scripts/arch-validate-iso.sh`, `scripts/arch-northstar-check.sh` (slow gates)
- [ ] Fixture ResolvedSpeech JSON files under `translators/arch/tests/fixtures/`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| QEMU boot smoke of built ISO | ARCH-01 | No QEMU on host | Optional: boot ISO in QEMU/VM elsewhere; documented in translators/arch/README.md |
| Interactive UX equivalence with real Omarchy | ARCH-02 | Requires human-driven desktop session | Out of automated scope; mechanical rootfs equivalence is the gate |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s (slow gates exempt, documented)
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-06-12 (autonomous run)
