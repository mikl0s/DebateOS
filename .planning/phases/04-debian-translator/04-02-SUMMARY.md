---
phase: 04-debian-translator
plan: "02"
subsystem: examples
tags: [dual-foundation, deb-02, example-speech, go-test, foundation-neutral]
dependency_graph:
  requires: [04-01-SUMMARY.md]
  provides: [examples/dual-foundation/, examples/dual_foundation_test.go]
  affects: [04-03-PLAN.md, 04-05-PLAN.md]
tech_stack:
  added: []
  patterns: [omarchy_test.go harness mirror, assembleOpinionsFromSpeech, findRoot reuse]
key_files:
  created:
    - examples/dual-foundation/speech.yaml
    - examples/dual-foundation/points/DF-base-cli.yaml
    - examples/dual-foundation/points/DF-system-tuning.yaml
    - examples/dual-foundation/opinions/DF-001.yaml
    - examples/dual-foundation/opinions/DF-002.yaml
    - examples/dual-foundation/opinions/DF-003.yaml
    - examples/dual-foundation/opinions/DF-004.yaml
    - examples/dual-foundation/opinions/DF-005.yaml
    - examples/dual-foundation/assets/motd
    - examples/dual-foundation/README.md
    - examples/dual_foundation_test.go
  modified: []
decisions:
  - "[04-02] foundation: debian chosen for speech.yaml so dual-foundation-check can route to debian by default; Arch translator is invoked on the same resolved.json explicitly in 04-05"
  - "[04-02] No hardware_condition on any DF opinion — ensures all 5 apply on empty baseline profile, keeping clean-resolve assertion simple and unambiguous"
  - "[04-02] assembleOpinionsFromSpeech and findRoot reused from examples_test package (same package) — no redeclaration, mirrors omarchy_test.go exactly"
  - "[04-02] dst path in DF-002 is relative (etc/motd, no leading slash) — complies with T-04-03 threat mitigation and the _sanitize_dst gate documented in 02-05-SUMMARY deviation #2"
metrics:
  duration: "~8 min"
  completed: "2026-06-13T13:45:00Z"
  tasks_completed: 2
  files_changed: 11
---

# Phase 4 Plan 02: Foundation-Neutral Dual-Foundation Example Speech Summary

**One-liner:** Foundation-neutral dual-foundation speech (5 opinions, 2 points, foundation: debian) with TestExampleDualFoundation clean-resolve gate proving DEB-02.

## What Was Built

### Task 1 — 5 foundation-neutral opinions + 2 points + speech + asset (commit 691ae12)

Created `examples/dual-foundation/` from scratch (not Omarchy). Every artifact uses
only OS-agnostic capability tokens that BOTH the Arch and (forthcoming Debian) translators
can declare:

**Opinions:**

| File | ID | Category | install_phase | Capability Token | Why OS-agnostic |
|------|----|----------|--------------|------------------|-----------------|
| DF-001.yaml | DF-001 | package-install | packaging | `install-packages` | `git`, `curl`, `vim` ship under identical upstream names on Arch and Debian |
| DF-002.yaml | DF-002 | config-file | config | `deploy-config-file-tree` | File copy to `etc/motd`; translator owns the mechanic, schema carries only src/dst |
| DF-003.yaml | DF-003 | service | config | `enable-systemd-service` | `systemd-timesyncd.service` unit name is identical on both foundations |
| DF-004.yaml | DF-004 | sysctl | config | `write-sysctl-drop-in` | `/etc/sysctl.d/` drop-in path exists on both Arch and Debian |
| DF-005.yaml | DF-005 | user-group | post-install | `add-user-to-group` | `video` group exists on both; `usermod`/`gpasswd` mechanic is translator-owned |

**Points:**
- `points/DF-base-cli.yaml` (id: DF-base-cli) — members: [DF-001, DF-002]
- `points/DF-system-tuning.yaml` (id: DF-system-tuning) — members: [DF-003, DF-004, DF-005]

**Speech:**
- `speech.yaml`: `foundation: debian`, 2 point refs (DF-base-cli, DF-system-tuning)

**Asset:** `assets/motd` — static CC0 banner text (referenced by DF-002 file_asset)

**All 5 capability tokens verified present in `translators/arch/capabilities.json`** — the
exact same tokens Plan 04-03 must declare in Debian's capabilities.json.

### Task 2 — TestExampleDualFoundation Go gate (commit 376673b)

Created `examples/dual_foundation_test.go` (package `examples_test`) mirroring
`omarchy_test.go`:

- Loads speech, points, opinions via the same helper pattern
- Reuses `findRoot` and `assembleOpinionsFromSpeech` from the existing package (no redeclaration)
- Resolves on an empty `hardware.HardwareProfile` (baseline — no hardware gates)
- **Assertions all pass:**
  - No `resolve.Resolve` error
  - No "Hard conflict" in Explanations
  - `len(Applied) == 5`, `len(InstallOrder) == 5`
  - `len(Dropped) == 0`, `len(Skipped) == 0`
  - Every applied ID has `DF-` prefix

```
--- PASS: TestExampleDualFoundation (0.00s)
    TestExampleDualFoundation: Applied=5 Skipped=0 Dropped=0 InstallOrder=5
```

Full `go test ./... -count=1`: **all packages green, zero regressions**.

## Capability Token Set (for Plan 04-03)

**Plan 04-03 MUST declare ALL of the following tokens in `translators/debian/capabilities.json`:**

| Token | Used by opinion |
|-------|----------------|
| `install-packages` | DF-001 |
| `deploy-config-file-tree` | DF-002 |
| `enable-systemd-service` | DF-003 |
| `write-sysctl-drop-in` | DF-004 |
| `add-user-to-group` | DF-005 |

These 5 tokens are the minimum required for the dual-foundation proof. Plan 04-03 may declare
additional tokens beyond this minimum as needed for Debian-specific mechanics.

## Deviations from Plan

None — plan executed exactly as written.

The threat mitigation T-04-03 was honoured by design: `DF-002`'s `file_assets` dst is
`etc/motd` (relative, no leading `/`, no `..` components) — the translator's `_sanitize_dst`
gate will accept it on both foundations.

## Known Stubs

None — all 5 opinions carry real payload data; no placeholder text; motd banner is
substantive and referenced by DF-002.

## Threat Flags

No new security-relevant surface introduced. Opinion YAML files are curator-authored
example content (CC0 public, T-04-04 accepted). No network endpoints, no auth paths,
no schema trust-boundary changes in this plan.

## Self-Check: PASSED

Files exist:
- examples/dual-foundation/speech.yaml: FOUND
- examples/dual-foundation/opinions/DF-005.yaml: FOUND
- examples/dual-foundation/assets/motd: FOUND
- examples/dual_foundation_test.go: FOUND
- .planning/phases/04-debian-translator/04-02-SUMMARY.md: FOUND (this file)

Commits verified:
- 691ae12: feat(04-02) — Task 1 (10 files)
- 376673b: test(04-02) — Task 2 (1 file)
