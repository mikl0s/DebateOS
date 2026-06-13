---
phase: 3
slug: cli-build-channels
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-06-13
---

# Phase 3 — Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (CLI + resolver regression); bash gate scripts; pytest regression (translator) |
| **Config file** | none |
| **Quick run command** | `go test ./cli/... ./cmd/... -count=1` |
| **Full suite command** | `go test ./... -count=1 && pytest translators/arch/tests -q && bash scripts/determinism-test.sh && bash scripts/check-coverage.sh` |
| **Estimated runtime** | quick < 10s; full < 3 min (determinism runs resolve+translate twice) |

## Sampling Rate

- **After every task commit:** `go test ./cli/... ./cmd/... -count=1`
- **After every plan wave:** full suite
- **Before `/gsd-verify-work`:** full suite + secret-free artifact check + workflow YAML validation
- **Max feedback latency:** 60s (determinism gate exempt, wave-level)

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-XX | TBD | 1+ | CLI-01 | — | no os.Exit in Run(); DEBATEOS_DIR honored | unit | `go test ./cli/... -run 'TestCompose\|TestValidate' -count=1` | ❌ W0 | ⬜ pending |
| 03-XX | TBD | 1+ | CLI-02 | — | pane.yaml + identity created 0600 | unit | `go test ./cli/... -run TestPane -count=1` (perm asserts) | ❌ W0 | ⬜ pending |
| 03-XX | TBD | 1+ | PRIV-01 | — | age round-trip; no secrets in profile/ISO listing | unit+script | `go test ./cli/... -run TestPaneBackup` + `bash scripts/secret-free-check.sh` | ❌ W0 | ⬜ pending |
| 03-XX | TBD | 1+ | BLD-01 | — | docker invocation via Runner; --skip-iso works on this host | unit | `go test ./cli/... -run TestBuild -count=1` (FakeRunner asserts) | ❌ W0 | ⬜ pending |
| 03-XX | TBD | 2+ | BLD-02 | — | same image both channels | script | workflow YAML validation (actionlint@v1.6.27 or PyYAML) + image ref grep equality | ❌ W0 | ⬜ pending |
| 03-XX | TBD | 2+ | BLD-03 | — | N/A | script | `bash scripts/determinism-test.sh` (two runs, byte-identical tars) | ❌ W0 | ⬜ pending |
| 03-XX | TBD | 2+ | BLD-04 | — | no central service in path | unit+doc | grep gate: no non-GitHub/user-host URLs in build path; README zero-cost walkthrough | ❌ W0 | ⬜ pending |
| 03-XX | TBD | all | regression | — | N/A | unit | `go test ./... -count=1 && pytest translators/arch/tests -q` | ✅ | ⬜ pending |

## Wave 0 Requirements

- [ ] cli/ package layout + cmd/debateos/main.go skeleton with RED tests
- [ ] scripts/determinism-test.sh, scripts/secret-free-check.sh
- [ ] FakeRunner test helper

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Live GitHub Actions run on a fork | BLD-02 | Requires CI minutes + fork; deferred verification item | Fork template, push speech, observe ISO artifact |
| Full ISO build + ISO-level determinism | BLD-01/03 | Host cannot run mkarchiso (Proxmox devtmpfs) | Run scripts on standard Linux host |

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity OK; no watch-mode flags; latency < 60s
- [x] `nyquist_compliant: true`

**Approval:** approved 2026-06-13 (autonomous run)
