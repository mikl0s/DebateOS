---
phase: 04-debian-translator
verified: 2026-06-13T14:46:07Z
status: human_needed
score: 4/4 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Run dual-foundation-check.sh WITHOUT --skip-iso on a capable host (bare-metal Linux or KVM VM without Proxmox devtmpfs restrictions)"
    expected: "Both Arch and Debian ISO builds complete from the same resolved.json; debian-validate-iso.sh passes on the produced .iso; total 20+ checks PASS"
    why_human: "Host is Proxmox VE with loop/devtmpfs restrictions — lb build fails with 'mount: permission denied'. Deferred-to-capable-host per VALIDATION.md policy."
---

# Phase 4: Debian Translator Verification Report

**Phase Goal:** The opinion/translator abstraction is proven real, not Arch-shaped — one resolved speech yields installers for two foundations
**Verified:** 2026-06-13T14:46:07Z
**Status:** human_needed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | DUAL-FOUNDATION PROOF: a representative speech builds installers for BOTH Arch and Debian from the same resolved input | VERIFIED (code + --skip-iso) | `dual-foundation-check.sh --skip-iso` 20/20 PASS: resolve once → Arch translate (exit 0) → Debian translate (exit 0) → equivalence assert [curl,git,vim] identical → structural validation PASS; full ISO build deferred-to-capable-host per VALIDATION.md |
| 2 | Debian translator wraps live-build/preseed, declares capabilities, unsupported required opinions break visibly at composition time | VERIFIED | `translators/debian/capabilities.py` `check_capabilities()` raises `CapabilityError` naming opinion+token at composition time; 20 capability gate tests pass; generator hard-fails before any I/O; `capabilities.json` has 45 tokens excluding mkinitcpio/limine/pacman-AUR |
| 3 | Arch assumptions that leaked into schema/resolver/examples identified and fixed, adjustments documented | VERIFIED | `docs/arch-leak-audit.md` documents 6 findings; only genuine leak (build.go hardcoded Arch) fixed via `foundationRegistry`; schema/resolver/examples unchanged and backward-compatible; `TestBuildFoundationDispatch`, `TestBuildArchUnchanged`, `TestBuildUnknownFoundation` all PASS |
| 4 | Translator ownership model documented (distros own translators, curators own points/speeches, PRs welcome) | VERIFIED | `docs/ownership-model.md` present; `grep "own their translators"` matches line 11 and 35; COMM-01 requirement satisfied |

**Score:** 4/4 truths verified (code-level; ISO build deferred per host policy)

---

### Security Fix Verification: T-04-08 (DEBATEOS_HASHED_PASSWORD)

**Requirement from verification context:** `translators/debian/profile.py` has NO baked default password and HARD-FAILS without `DEBATEOS_HASHED_PASSWORD`.

| Check | Result |
|-------|--------|
| `profile.py` has no baked default password | VERIFIED — lines 269-279: `hashed_password = os.environ.get("DEBATEOS_HASHED_PASSWORD", "")` then `if not hashed_password.startswith("$"): raise ValueError(...)` |
| Unset env → `ValueError` raised | VERIFIED — manual probe: `unset DEBATEOS_HASHED_PASSWORD; python3 -c "from profile import _write_preseed; _write_preseed('/tmp')"` → `ValueError: DEBATEOS_HASHED_PASSWORD must be set to a valid crypt hash...` |
| Tests assert hard-fail | VERIFIED — `TestPreseedPasswordRequired::test_emit_raises_without_password` (monkeypatch.delenv) and `test_emit_raises_on_invalid_hash` both pass (88 total debian tests pass) |
| Valid hash → preseed has `$6$` and no `%%` sentinels | VERIFIED — `test_preseed_has_no_sentinels_and_real_hash` asserts both; confirmed `%%` not in preseed and `$6$` present |

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `translators/common/contract.py` | load_resolved_speech, load_opinion_bodies | VERIFIED | Contains `def load_opinion_bodies`; 36 common tests pass |
| `translators/common/manifest.py` | BuildManifest, derive_source_date_epoch | VERIFIED | Contains `class BuildManifest`; shared by both translators |
| `translators/common/firstrun.py` | render_firstrun_unit, firstrun_unit_name | VERIFIED | Parameterized template path; used by debian/profile.py |
| `translators/debian/capabilities.py` | CapabilityError, check_capabilities | VERIFIED | Required+unsupported raises CapabilityError naming opinion+token |
| `translators/debian/capabilities.json` | 45 tokens, no mkinitcpio/limine | VERIFIED | 45 tokens confirmed; mkinitcpio/limine/pacman-AUR absent; grep confirms |
| `translators/debian/profile.py` | emit_profile_tree, _sanitize_dst, T-04-08 | VERIFIED | Hard-fails without DEBATEOS_HASHED_PASSWORD; CR-01/02/03 all fixed |
| `translators/debian/generator.py` | generate() entrypoint | VERIFIED | 14 generator tests pass including end-to-end translate test |
| `translators/debian/translate` | Frozen argv shell wrapper | VERIFIED | `-rwxr-xr-x`; `test_translate_end_to_end` passes |
| `translators/debian/templates/preseed.cfg.tpl` | d-i preseed; no _preseed_V1 | VERIFIED | IN-01 fixed: `_preseed_V1` removed; standard comment block instead |
| `translators/debian/templates/chroot-install.hook.tpl` | jq-driven safe pattern (CR-01) | VERIFIED | All opinion data via build-manifest.json jq path; no raw data in shell command position |
| `cli/build/build.go` | foundationRegistry dispatch | VERIFIED | Lines 70-73: `foundationRegistry` with arch+debian; unknown foundation returns error; `TestBuildFoundationDispatch` PASS |
| `examples/dual-foundation/speech.yaml` | Foundation-neutral speech | VERIFIED | 5 opinions applied; dual-foundation-check.sh equivalence asserts [curl,git,vim] match |
| `scripts/dual-foundation-check.sh` | Genuine assertions, 20 checks | VERIFIED | 20/20 PASS with `--skip-iso`; equivalence Step 4f compares pkg sets between Arch+Debian manifests |
| `docs/arch-leak-audit.md` | 6-finding audit, DEB-03 | VERIFIED | All 6 findings documented; build.go (Finding 5) the only genuine leak, fixed |
| `docs/ownership-model.md` | COMM-01 ownership doc | VERIFIED | "own their translators" phrase present; entrypoint contract, capabilities.json, PR guidelines documented |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `translators/arch/generator.py` | `translators/common` | shim re-export in `translators/arch/contract.py` | VERIFIED | WR-04 fixed: only public names in `__all__`; private names removed |
| `translators/debian/profile.py` | `translators/common/firstrun.py` | `from common.firstrun import render_firstrun_unit` | VERIFIED | Import present at line 56; used in `_write_firstrun_units()` |
| `cli/build/build.go` | `translators/debian/translate` | `foundationRegistry["debian"].TranslateBin` | VERIFIED | Lines 70-73; `TestBuildFoundationDispatch` confirms debian path dispatches to `translators/debian/translate` |
| `dual-foundation-check.sh` | Both translators on same resolved.json | Step 1 → Step 2 (Arch) + Step 3 (Debian) | VERIFIED | WR-02 fix: intentional out-of-band invocation documented at line 131-136 |
| `preseed.cfg.tpl` `%%HASHED_PASSWORD%%` sentinel | real `$6$` crypt hash | `profile.py _write_preseed()` env var substitution | VERIFIED | Hard-fails without env var; replaces sentinel with `$`-prefixed hash |

---

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `chroot-install.hook.tpl` → emitted hook | `jq .system_services`, `.sysctl_params`, `.group_memberships` | `build-manifest.json` (written by `_write_build_manifest`) | Yes — manifest JSON contains verbatim opinion data | FLOWING |
| `preseed.cfg.tpl` → `preseed.cfg` | `%%HASHED_PASSWORD%%` | `DEBATEOS_HASHED_PASSWORD` env var | Yes — hard-fails if absent; no hardcoded fallback | FLOWING |
| `debateos.list.chroot_install` | `manifest.target_packages` | `BuildManifest.from_resolved()` | Yes — from opinion `packages` field | FLOWING |
| `build-manifest.json` | All opinion data | `BuildManifest.to_dict()` | Yes — complete serialization; jq reads at chroot time | FLOWING |

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| go test ./... all pass | `go test ./... -count=1` | 16 packages OK (7 with tests; 9 no test files — no FAIL) | PASS |
| common tests pass | `pytest translators/common/tests -q` | 36 passed | PASS |
| arch tests pass (regression) | `pytest translators/arch/tests -q` | 135 passed | PASS |
| debian tests pass | `pytest translators/debian/tests -q` | 88 passed | PASS |
| dual-foundation-check --skip-iso | `bash scripts/dual-foundation-check.sh --skip-iso` | 20/20 PASS | PASS |
| coverage gates | `bash scripts/check-coverage.sh` | resolver/ 93.5% >= 90%; cli/ 85.6% >= 85% | PASS |
| T-04-08: unset → ValueError | `unset DEBATEOS_HASHED_PASSWORD; python3 -c "...` | ValueError raised: "DEBATEOS_HASHED_PASSWORD must be set..." | PASS |
| T-04-08: valid hash → no sentinels | `pytest -k preseed_no_percent_sentinels` (via 88-test suite) | PASS (no %% in emitted preseed.cfg) | PASS |
| Capability gate: required+unsupported raises | `pytest translators/debian/tests -k capability` | 20 passed including `test_required_unsupported_raises_capability_error` | PASS |
| `grep "own their translators" docs/ownership-model.md` | direct grep | Lines 11 and 35 match | PASS |
| Arch backward compat: `TestBuildArchUnchanged` | `go test ./cli/build/... -run TestBuildArchUnchanged` | PASS | PASS |
| omarchy still resolves | `go test ./examples/... -count=1` | TestExampleOmarchy: Applied=99, PASS | PASS |

---

### Probe Execution

No conventional `scripts/*/tests/probe-*.sh` probes defined for this phase. The dual-foundation-check.sh script serves as the phase's proof gate and was run above.

---

### Requirements Coverage

| Requirement | Description | Status | Evidence |
|-------------|-------------|--------|----------|
| DEB-01 | Debian translator wraps live-build/preseed, declares capabilities, CapabilityError on unsupported required opinions | SATISFIED | `capabilities.py` + `generator.py` gate; `emit_profile_tree` produces valid config/ tree; 88 debian tests pass; full ISO deferred-to-capable-host per VALIDATION.md |
| DEB-02 | Dual-foundation proof: same resolved input → both Arch and Debian installers | SATISFIED (code) | `dual-foundation-check.sh --skip-iso` 20/20 PASS; full ISO build human_needed |
| DEB-03 | Arch leaks identified, fixed, documented | SATISFIED | `docs/arch-leak-audit.md` 6 findings; build.go fixed via foundationRegistry; 135 arch tests pass (regression clean) |
| COMM-01 | Translator ownership model documented | SATISFIED | `docs/ownership-model.md` present with required phrase; entrypoint contract, capabilities.json, profiles, PR model all documented |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `translators/debian/README.md` | 206, 381 | Stale documentation: says `%%HASHED_PASSWORD%%` is a manual-replacement sentinel, but code now hard-fails without `DEBATEOS_HASHED_PASSWORD` | WARNING | Documentation inaccuracy — misleads operators about actual behavior. Code path is correct (profile.py hard-fails); README describes the old CR-02 behavior |
| `translators/debian/tests/test_generator.py` | 252-253, 263-264 | `pytest.skip("translate not yet implemented")` guards | INFO | Guards are dead code (translate exists and is executable); tests skipped only if `TRANSLATE` file is absent. Since translate IS present, all 14 generator tests run and pass — not a real issue |

No `TBD`, `FIXME`, or `XXX` markers found in any phase-modified file. The `XXXXXX` in dual-foundation-check.sh line 86 is a `mktemp` template pattern, not a debt marker.

---

### Human Verification Required

#### 1. Full Dual-Foundation ISO Build on Capable Host

**Test:** On a bare-metal Linux host or KVM VM (not Proxmox VE), with `DEBATEOS_HASHED_PASSWORD` set to a valid crypt hash, run:

```bash
bash scripts/dual-foundation-check.sh
```

(without `--skip-iso`)

**Expected:** Both the Arch ISO and the Debian ISO build successfully from the same `examples/dual-foundation/` speech. The script reports 20+ total PASS with no FAIL. `debian-validate-iso.sh` validates the produced `.iso`.

**Why human:** This host (Proxmox VE kernel 6.17.4-2-pve) cannot run `lb build` or `mkarchiso` — both require `devtmpfs` and loop device access blocked by Proxmox container policy. Confirmed in `04-RESEARCH.md` and documented in `scripts/debian-build-iso.sh` header. All code-level gates pass; this is the only item deferred by host capability, per the VALIDATION.md policy ("ISO builds (slow gate) deferred-to-capable-host").

---

### Gaps Summary

No gaps found. All four success criteria are verified at the code level:

1. **Dual-foundation proof** — `dual-foundation-check.sh --skip-iso` 20/20 PASS: one `resolved.json`, both translators exit 0, equivalence assertion [curl,git,vim] identical between Arch and Debian manifests. Full ISO build is deferred-to-capable-host per VALIDATION.md policy (not a gap).

2. **Debian translator** — wraps live-build/preseed via `emit_profile_tree()`; declares 45 capabilities in `capabilities.json`; required+unsupported opinions raise `CapabilityError` naming opinion+token at composition time before any file I/O.

3. **Arch leak audit** — 6 findings in `docs/arch-leak-audit.md`; only genuine leak (build.go) fixed via `foundationRegistry`; schema/resolver unchanged; 135 arch tests and omarchy example pass clean.

4. **Ownership model** — `docs/ownership-model.md` present with "distributions own their translators" principle, entrypoint contract, capabilities.json guidance, and PR model.

**Security fix T-04-08 confirmed:** `profile.py` hard-fails without `DEBATEOS_HASHED_PASSWORD`; no baked default; tests assert both fail-fast and valid-hash paths.

**Review fixes confirmed:** All 3 CR (critical) blockers and 5 WR (warning) and 3 IN (info) findings from `04-REVIEW.md` verified fixed in code and covered by tests.

---

_Verified: 2026-06-13T14:46:07Z_
_Verifier: Claude (gsd-verifier)_
