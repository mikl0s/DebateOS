---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: executing
stopped_at: Completed 03-03-PLAN.md — build subcommand resolve→epoch→translate→docker + injection tar
last_updated: "2026-06-13T12:11:20Z"
last_activity: 2026-06-13 -- Phase 3 Plan 03 complete (build subcommand + private-injection.tar)
progress:
  total_phases: 6
  completed_phases: 3
  total_plans: 18
  completed_plans: 16
  percent: 50
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-06-12)

**Core value:** Compose a speech from curators' points, resolve conflicts explainably, and build a bootable unattended installer — zero cost, no central service in the critical path.
**Current focus:** Phase 2 — Arch Translator

## Current Position

Phase: 3 (CLI & Build Channels) — EXECUTING
Plan: 3 of 4
Status: Ready to execute plan 4
Last activity: 2026-06-13 -- Phase 3 Plan 03 complete (build subcommand + private-injection.tar)

Progress: [████░░░░░░] 44%

## Performance Metrics

**Velocity:**

- Total plans completed: 1
- Average duration: 10 min
- Total execution time: ~0.17 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| Phase 0 | 1/4 | ~10 min | ~10 min |

**Recent Trend:** Phase 0 Plan 01 complete in ~10 min (research/document plan)

*Updated after each plan completion*
| Phase 00-omarchy-research-arch-variant-study P03 | 10 | 2 tasks | 1 files |
| Phase 00-omarchy-research-arch-variant-study P02 | 15 | 2 tasks | 2 files |
| Phase 01-schema-resolver-core P02 | 2 min | 2 tasks | 6 files |
| Phase 01-schema-resolver-core P03 | 5 min | 2 tasks | 7 files |
| Phase 01-schema-resolver-core P04 | 12 min | 3 tasks | 31 files |
| Phase 01-schema-resolver-core P05 | 11 | 3 tasks | 19 files |
| Phase 02-arch-translator P01 | 6 min | 2 tasks (4 commits) | 14 files |
| Phase 02-arch-translator P03 | 4 min | 2 tasks | 4 files |
| Phase 02-arch-translator P02 | 11 min | 3 tasks (6 commits) | 16 files |
| Phase 02-arch-translator P04 | 25 min | 2 tasks | 170 files |
| Phase 02-arch-translator P05 | 18h | - tasks | - files |
| Phase 03-cli-build-channels P01 | 12 min | 2 tasks (4 commits) | 12 files |
| Phase 03-cli-build-channels P02 | 35 | 2 tasks | 5 files |
| Phase 03-cli-build-channels P03 | 20 min | 2 tasks (3 commits) | 5 files |

## Accumulated Context

### Decisions

All D1–D20 + D13a + 7 invariants are LOCKED (docs/09 via PROJECT.md `<decisions>` block). Do not re-open. Highlights for current work:

- D17/D20: Phase 0 gates everything — no schema drafting before the six `research/` deliverables exist
- D19: TDD everywhere — every phase plan specifies test scenarios before implementation tasks; resolver near-total coverage; determinism + WASM/native parity are automated tests
- Process: fully autonomous run to v1.0; no pausing between phases except true blockers; record new fork decisions and continue
- Roadmapper: phases numbered 0–5 to match ADR/SPEC naming exactly; Omarchy-on-variant retarget is a non-gating Phase 2 stretch criterion
- [Phase ?]: CachyOS snapper assumption A2 corrected: cachyos-snapper-support exists (optional)
- [Phase ?]: Garuda uses dracut exclusively, conflicts mkinitcpio — hard conflict with Omarchy login phase
- [Phase ?]: 32 evidence-driven candidate points for OM-001..OM-134; single-opinion points allowed where natural
- [Phase ?]: SR-009 enumerates repo trust levels: Required/Required DatabaseOptional/Optional TrustAll/Never — T-00-SIG2 mitigated
- [Phase ?]: SR-005 compound hardware conditions require AND/OR/NOT combinators; simple boolean flag insufficient
- [Phase ?]: SR-006 phase-level ordering is a discrete enum plus within-phase before/after refs; flat integer order insufficient
- [Phase ?]: 27 EC-NNN scenarios produced with full docs/04 coverage; variant-profile conflict semantics deferred to Phase 1
- [Phase ?]: migrations-as-schema-concept recorded as OQ-001 open question; deferred to Phase 1/post-v1
- [Phase ?]: 306 runtime bin/ helpers classified as translator infrastructure not opinions (OQ-008)
- [Phase ?]: TopoSort is a free function (not a method on Graph) — cleaner call site for 01-04 resolver
- [Phase ?]: Phase enum stored in Graph.phase but NOT converted to edges — tie-breaking key only per SR-006/OM-023 cross-phase override
- [01-03]: hardware.HardwareProfile is a distinct package-local struct with PCIIDs []string — richer than resolver.HardwareProfile (which has only Predicates+Facts); 01-04 will adapt at the evaluation boundary
- [01-03]: FindPatch scans known_patches on BOTH conflicting opinions to ensure symmetry; sorts candidates by ID for determinism
- [01-04]: Resolve returns (*ResolvedSpeech, error) — partial RS always returned on hard conflict so callers can display explanation text
- [01-04]: EC-038 PCIIDs via hardware_override fixture block — resolver.HardwareProfile lacks PCIIDs; hardware.HardwareProfile has it
- [01-04]: Rule4 fires when patch opinion is already active in speech; Rule2+PatchOffered fires when patch exists but not in speech
- [01-04]: sig_level=Never repos surface TrustWarning in Explanation (T-01-10); sysctl collision detection runs before conflict resolution (SR-016)
- [02-01]: pytest installed via pip --break-system-packages (Debian host restriction); version 9.0.3 matches Arch official python-pytest 1:9.0.3-1
- [02-01]: install-npm-global-packages intentionally absent from capabilities.json — leave undeclared until npm handler is implemented; gate correctly drops nice-to-haves with this token
- [02-01]: first_run opinions (execution_phase=="first-run") excluded from install-time package/service aggregation; collected as {id, script_payload} for Plan 02 systemd oneshot unit generation
- [02-01]: check_capabilities returns list[(id, reason)] for dropped nice-to-haves (empty list on clean pass); CapabilityError message always contains opinion ID + token + "composition time" (SC-3)
- [02-03]: vanilla-arch bootloader/filesystem set to null — translator/speech choice, not profile-forced; Omarchy OM-099 handles limine at speech time
- [02-03]: Garuda above_core=false for custom repos — unlike CachyOS (above_core=true), Garuda adds custom repos BELOW standard Arch repos per pacman-default.conf (VERIFIED)
- [02-03]: repos_by_arch_level extension key in cachyos.yaml allows v3/v4 ISA-optimised tiers without code fork (ARCH-04 invariant preserved)
- [02-03]: 4 Garuda hard Omarchy conflicts captured as structured data (dracut/mkinitcpio, GRUB/limine, snapper/snapper, SDDM theme) — generator surfaces via trust_warnings
- [02-04]: resolve.Resolve takes flat []resolver.Opinion; test expands speech.Points through point files — resolve.go does not read point files
- [02-04]: Status policy OQ-1: required=OM-001/006/097/099+hw-conditional; nice-to-have=themes OM-114..134+optional extras; all others required
- [02-04]: Vanilla-arch hw profile (empty predicates/pci_ids) — 35 hw-gated opinions Skipped (expected); Applied=99 Dropped=0 Hard-conflicts=0 (ARCH-02 satisfied)
- [02-02]: %%SENTINEL%% replace() for installer.sh.tpl — stdlib-only safe approach for shell-heavy templates with ${SHELL_VAR} syntax (avoids str.format KeyError)
- [02-02]: translators/__init__.py added for python -m translators.arch.generator invocation; sys.path.insert in generator.py covers both pytest and -m contexts
- [02-02]: _sanitize_dst rejects absolute and .. traversal paths (T-02-08); keyring_install_before_repos injected into build-manifest.json for Pitfall 4 ordering
- [Phase ?]: releng-baseline-overlay: arch-build-iso.sh copies releng profile then overlays generator output inside Docker to provide syslinux/ and efiboot/ directories required by mkarchiso
- [Phase ?]: capabilities.json updated to actual opinion tokens (163 tokens) extracted from examples/omarchy/opinions/*.yaml; old broad conceptual names removed
- [Phase ?]: devtmpfs restriction on Proxmox VE documented as environment limitation; all tooling is correct; full ISO build requires standard Linux host
- [Phase ?]: age X25519 identity local-only, no escrow (PRIV-01/D16): identity.age 0600 in config dir
- [Phase ?]: pane backup routes git via Runner interface — FakeRunner in tests, zero network calls
- [Phase ?]: only pane.yaml.age (ciphertext) ever staged — plaintext never in git (T-03-PLAINTEXT)

### Decisions from Plan 03-03

- [03-03]: DeriveEpoch exported from cli/build — 03-04 determinism gate imports without re-implementing (epochMin=1577836800, epochMax=2208988800 mirror manifest.py)
- [03-03]: WriteInjectionTar(outDir, nil) emits manifest-only tar when no private assets; first-boot unit always finds artifact
- [03-03]: sanitizeDst uses sentinel-root filepath.Clean containment check — identical semantic to profile.py _sanitize_dst (T-03-TRAV)
- [03-03]: .gitignore `build/` narrowed to `/web/build/` — prevents shadowing cli/build/ Go package

### Decisions from Plan 03-01

- [03-01]: Runner interface uses variadic args (exec.Command(name, args...)); never sh -c — T-03-AI mitigation
- [03-01]: cli/internal/loader.ResolveDir() extracted as shared pipeline for compose and validate — prevents drift from cmd/resolve-json
- [03-01]: os.Exit called only in cmd/debateos/main.go; all cli/ packages return int — fully testable subcommand pattern
- [03-01]: filippo.io/age v1.3.1 manually promoted to direct require in go.mod (go mod tidy prunes it until pane package imports it in 03-02)

### Decisions from Plan 00-01

- Base package list (155 packages) split into 12 logical package-install opinions — maximizes composability
- 21 themes cataloged as individual theming opinions (OM-114..OM-134) — independently selectable
- 313 migrations sampled but not inventoried — pattern documented as open question for Plan 04
- First-run scripts (13) inventoried as OM-101..OM-113 with execution-phase: first-run
- npm-global-install is a distinct category from package-install (schema surprise SR candidate)

### Pending Todos

None yet.

### Blockers/Concerns

- Phase 0 requires cloning and analyzing `https://github.com/basecamp/omarchy` (network access needed; analyze source, not summaries)
- Phases 2–4 require privileged ISO builds (mkarchiso, live-build, loop devices, Docker) — expected available on this host per docs/00; not a blocker, noted for planning

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Post-v1.0 | Phase 6 hardware-scanning installer; Fedora translator; direct-to-disk; GitLab parity; post-install reconciliation | Locked deferral (D2) | Project init |

## Session Continuity

Last session: 2026-06-13T12:11:20Z
Stopped at: Completed 03-03-PLAN.md — build subcommand + private-injection.tar
Resume file: None
Next: Phase 3 Plan 04 (03-04-PLAN.md) — Docker image + GHA workflow + determinism/coverage gates
