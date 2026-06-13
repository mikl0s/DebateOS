---
phase: 04-debian-translator
plan: "05"
subsystem: scripts/translators/docs
tags: [deb-01, deb-02, comm-01, dual-foundation, lb-build, ownership-model, phase-complete]

# Dependency graph
requires:
  - phase: 04-debian-translator
    provides: translators/debian/translate entrypoint (04-03) + foundationRegistry (04-04) + examples/dual-foundation/ (04-02)
provides:
  - "scripts/dual-foundation-check.sh: resolve ONCE → both translators → 20-check equivalence gate (DEB-02)"
  - "scripts/debian-validate-iso.sh: host-runnable structural validation of live-build config/ tree"
  - "scripts/debian-build-iso.sh: docker --privileged lb-build wrapper (deferred-to-capable-host)"
  - "translators/debian/Dockerfile: digest-pinned debian:stable for lb build (T-04-13)"
  - "translators/debian/README.md: frozen argv contract + capabilities + config/ layout + full-build-status"
  - "docs/ownership-model.md: COMM-01 translator ownership model + how-to-add-a-translator"
  - "DEB-01, DEB-02, DEB-03, COMM-01 all Complete in REQUIREMENTS.md; Phase 4 Complete in ROADMAP.md"
affects:
  - Phase 5 (registry/UI/Forum) — Phase 4 complete; Phase 5 can begin

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Dual-foundation equivalence gate: resolve ONCE → translate×2 → PASS/FAIL counter; --skip-iso on restricted hosts"
    - "debian-validate-iso.sh: host-runnable structural validation (preseed d-i lines, chroot hook 0755, package-lists, valid JSON)"
    - "Ownership model: distributions own translators; curators own points/speeches; 5-step how-to-add"
    - "Per-suite pytest invocation: run arch/debian/common tests separately to avoid bare-name import collision"

key-files:
  created:
    - scripts/dual-foundation-check.sh
    - scripts/debian-validate-iso.sh
    - scripts/debian-build-iso.sh
    - translators/debian/Dockerfile
    - translators/debian/README.md
    - docs/ownership-model.md
  modified:
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md

key-decisions:
  - "[04-05]: dual-foundation-check.sh runs pytest suites separately (arch, debian, common) — combined invocation causes bare-name import collision (capabilities from arch/ shadows debian/capabilities when collected together; pre-existing issue not caused by this plan)"
  - "[04-05]: Cross-foundation equivalence proof: same 3 packages (git, curl, vim) verified from both build-manifest.json target_packages fields — the dual-foundation abstraction is real"
  - "[04-05]: debian:stable pinned to sha256:fa0ca9c113cbc97c1e2eb40d7012b43bdfb70abe4218229c366de911c5b32cd2 (verified 2026-06-13)"

patterns-established:
  - "Dual-foundation gate resolves ONCE feeds both translators — single resolved.json is the shared proof artifact"
  - "per-suite pytest invocation pattern avoids bare-name module collision in multi-translator repos"

requirements-completed: [DEB-01, DEB-02, DEB-03, COMM-01]

# Metrics
duration: ~35 min
completed: 2026-06-13
---

# Phase 4 Plan 05: Dual-Foundation Proof Gate + Debian Build Scripts + Ownership Model Summary

**One-liner:** DEB-02 dual-foundation proof gate (20/20 PASS), Debian lb-build Docker wrapper + digest-pinned Dockerfile + translator README (DEB-01), and COMM-01 ownership model with how-to-add-a-translator; DEB-01/02/03/COMM-01 all Complete; Phase 4 finalized.

## Performance

- **Duration:** ~35 min
- **Started:** 2026-06-13T15:00:00Z
- **Completed:** 2026-06-13T15:35:00Z
- **Tasks:** 3 (3 commits)
- **Files created:** 6 (dual-foundation-check.sh, debian-validate-iso.sh, debian-build-iso.sh, Dockerfile, README.md, ownership-model.md)
- **Files modified:** 2 (REQUIREMENTS.md, ROADMAP.md)

## Accomplishments

### Task 1 — dual-foundation-check.sh + debian-validate-iso.sh (commit dc89584)

**`scripts/dual-foundation-check.sh`** — the DEB-02 gate:
1. Resolves `examples/dual-foundation` ONCE via `go run ./cmd/resolve-json`
2. Runs `translators/arch/translate` on the resolved.json (Arch profile)
3. Runs `translators/debian/translate` on the SAME resolved.json (Debian config/ tree)
4. Asserts 16 per-foundation + cross-foundation equivalence checks (PASS/FAIL counter)
5. Calls `debian-validate-iso.sh` for structural validation
6. Runs regression test suites (go test + 3 × pytest individually)
7. `--skip-iso` skips the Docker lb build (default on this Proxmox host)

**Result: 20/20 PASS (VERIFIED 2026-06-13)**

Equivalence checks confirmed:
- Both translators exit 0 (no CapabilityError — all 5 DF opinions effectuable on both foundations)
- Arch: `target_packages` = [git, curl, vim]; `file_assets` includes `etc/motd`
- Debian: `debateos.list.chroot_install` contains git/curl/vim; chroot hook 0755; `preseed.cfg` has `d-i` lines
- Cross-foundation: `sorted(arch.target_packages) == sorted(debian.target_packages)` → `[curl, git, vim]`

**`scripts/debian-validate-iso.sh`** — host-runnable structural validator:
- Check 1: `preseed.cfg` present + `d-i` directives + `partman`/`netcfg` lines
- Check 2: `9000-debateos-apply.hook.chroot` present + executable (0755) + references `apt-get`
- Check 3: `package-lists/*.chroot_install` present + non-empty
- Check 4: `build-manifest.json` valid JSON + `target_packages` non-empty

Both scripts pass `bash -n` and are executable (0755).

### Task 2 — debian-build-iso.sh + Dockerfile + README (commit 9391aed)

**`scripts/debian-build-iso.sh`**:
- Wraps `docker run --privileged debian:stable@<digest> lb build` (T-04-12)
- Documents deferred-to-capable-host policy prominently in header (Proxmox devtmpfs restriction VERIFIED)
- `--skip-iso` path runs `debian-validate-iso.sh` only
- `SOURCE_DATE_EPOCH` passthrough for build determinism (BLD-03)
- Full argument validation + docker availability check

**`translators/debian/Dockerfile`**:
- `FROM debian:stable@sha256:fa0ca9c...` — pinned by digest (T-04-13, quarterly-reverify comment)
- Installs: live-build, debootstrap, squashfs-tools, xorriso, python3, python3-yaml, jq
- Mirrors the Arch `translators/arch/Dockerfile` pattern exactly

**`translators/debian/README.md`**:
- Frozen argv contract, capabilities gate (DEB-01/SC-3), Arch-only tokens correctly absent
- Sig-level → apt mapping table, config/ tree layout, deferred build status
- Architecture diagram, module layout, security properties (T-04-05 through T-04-13)

### Task 3 — ownership-model.md + REQUIREMENTS/ROADMAP (commit 5fae71f)

**`docs/ownership-model.md`** (COMM-01):
- "Distributions own their translators" — Ubuntu controls `translators/ubuntu/`, etc.
- Three-tier ownership: project (schema/resolver/common), distributions (translators), curators (points/speeches)
- How-to-add-a-translator: 5 steps (argv contract + capabilities.json + profiles YAML + foundationRegistry entry + common/ reuse)
- Translator-vs-schema boundary table (what each tier owns)
- Reference implementations (Arch/Debian complete; Ubuntu post-v1.0; Fedora community)

**REQUIREMENTS.md**:
- DEB-01, DEB-02, DEB-03, COMM-01: all body checkboxes marked `[x]`
- All four added to Traceability table with one-line evidence notes
- Last-updated line updated

**ROADMAP.md**:
- Phase 4 main entry: `[x]` with `completed 2026-06-13`
- 04-05-PLAN.md entry checked
- Progress table: `4. Debian Translator | 5/5 | Complete | 2026-06-13`

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | dual-foundation-check + debian-validate-iso | dc89584 | scripts/dual-foundation-check.sh, scripts/debian-validate-iso.sh |
| 2 | Debian ISO build script + Dockerfile + README | 9391aed | scripts/debian-build-iso.sh, translators/debian/Dockerfile, translators/debian/README.md |
| 3 | ownership-model + DEB/COMM requirements status | 5fae71f | docs/ownership-model.md, .planning/REQUIREMENTS.md, .planning/ROADMAP.md |

## DEB-02 Gate Result (Phase 4 Headline Outcome)

```
=== DUAL-FOUNDATION GATE SUMMARY (DEB-02) ===
  Passed: 20
  Failed: 0

  RESULT: DUAL-FOUNDATION GATE PASSED (DEB-02)
  (Equivalence+structural run. Full lb build requires a capable host.)
```

The dual-foundation abstraction is proven real:
- One resolved speech (`examples/dual-foundation/`) → two independent foundation translators → identical package sets
- `scripts/arch-northstar-check.sh --skip-build` (Arch gate): GREEN
- `scripts/dual-foundation-check.sh --skip-iso` (Dual-foundation gate): GREEN
- `go test ./... -count=1`: ALL PASS
- `python3 -m pytest translators/arch/tests/ -q`: 135 passed
- `python3 -m pytest translators/debian/tests/ -q`: 75 passed
- `python3 -m pytest translators/common/tests/ -q`: 36 passed

## Phase 4 Requirements Summary

| Requirement | Evidence | Status |
|-------------|----------|--------|
| DEB-01 | `translators/debian/translate`, 75 pytest tests, `debian-build-iso.sh`, Dockerfile (04-03 + 04-05) | Complete |
| DEB-02 | `scripts/dual-foundation-check.sh --skip-iso`: 20/20 PASS; one resolve, both translators, equivalence (04-05) | Complete |
| DEB-03 | `docs/arch-leak-audit.md`: 6 findings, `build.go` foundationRegistry fix (04-04) | Complete |
| COMM-01 | `docs/ownership-model.md`: "own their translators", entrypoint contract, 5-step how-to (04-05) | Complete |

## Deviations from Plan

**1. [Rule 2 - Pre-existing issue] Separate pytest invocations for each suite**
- **Found during:** Task 1 Step 5b (regression tests in dual-foundation-check.sh)
- **Issue:** Running `python3 -m pytest translators/arch/tests/ translators/debian/tests/ translators/common/tests/` together causes a bare-name import collision: `capabilities` from `translators/arch/` is loaded when collecting `translators/debian/tests/`, causing 46 Debian tests to fail on `assert "configure-mkinitcpio-hooks-and-modules" not in caps` (the Arch capabilities are loaded instead of Debian's).
- **Fix:** dual-foundation-check.sh runs each pytest suite in a separate invocation (`pytest translators/arch/tests/` then `pytest translators/debian/tests/` then `pytest translators/common/tests/`). All 246 tests pass across the three separate runs.
- **Status:** Pre-existing issue; not caused by this plan. The plan's verification footnote reflects this.
- **Files modified:** scripts/dual-foundation-check.sh (3 invocations instead of 1)

No other deviations — plan executed as written.

## Known Stubs

None — all scripts are functional (not stubs). The deferred ISO build is documented as a policy deferral (host restriction), not a code stub.

## Threat Flags

No new security-relevant surface beyond the plan's threat model. All T-04-12 through T-04-14 + T-04-SC mitigations applied:
- T-04-12: `--privileged` scoped to `debian-build-iso.sh` header + documented; never in generator path
- T-04-13: `debian:stable` pinned by sha256 digest in Dockerfile + build script; quarterly-reverify comment
- T-04-14: dual-foundation gate runs against `examples/dual-foundation/` (project-authored CC0 content); accepted
- T-04-SC: all build tools from Debian official apt; no PyPI/npm/cargo installs

## Self-Check: PASSED

Files exist:
- FOUND: scripts/dual-foundation-check.sh (executable, bash -n clean)
- FOUND: scripts/debian-validate-iso.sh (executable, bash -n clean)
- FOUND: scripts/debian-build-iso.sh (executable, bash -n clean)
- FOUND: translators/debian/Dockerfile (sha256 digest pinned, live-build installed)
- FOUND: translators/debian/README.md (argv contract, capabilities, config/ layout)
- FOUND: docs/ownership-model.md ("own their translators" phrase present)
- FOUND: .planning/REQUIREMENTS.md (DEB-01/02/03/COMM-01 all Complete)
- FOUND: .planning/ROADMAP.md (Phase 4 5/5 Complete)
- FOUND: .planning/phases/04-debian-translator/04-05-SUMMARY.md

Commits verified:
- dc89584: feat(04-05) — Task 1: dual-foundation-check + debian-validate-iso gates
- 9391aed: feat(04-05) — Task 2: Debian ISO build script + Dockerfile + README
- 5fae71f: docs(04-05) — Task 3: ownership-model + DEB/COMM requirements status

Gate result verified:
- `bash scripts/dual-foundation-check.sh --skip-iso`: 20/20 PASS (DEB-02 GREEN)
- `go test ./... -count=1`: ALL PASS
- `python3 -m pytest translators/arch/tests/ -q`: 135 passed
- `python3 -m pytest translators/debian/tests/ -q`: 75 passed
- `python3 -m pytest translators/common/tests/ -q`: 36 passed

---
*Phase: 04-debian-translator*
*Completed: 2026-06-13*
*This is the LAST plan of Phase 4. Phase 4 is COMPLETE.*
