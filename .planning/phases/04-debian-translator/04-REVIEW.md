---
phase: 04-debian-translator
reviewed: 2026-06-13T00:00:00Z
depth: standard
files_reviewed: 35
fixes_applied: true
fixed_at: 2026-06-13T00:00:00Z
resolved_ids:
  - CR-01
  - CR-02
  - CR-03
  - WR-01
  - WR-02
  - WR-03
  - WR-04
  - WR-05
  - IN-01
  - IN-02
  - IN-03
files_reviewed_list:
  - translators/common/contract.py
  - translators/common/manifest.py
  - translators/common/firstrun.py
  - translators/arch/contract.py
  - translators/arch/manifest.py
  - translators/arch/firstrun.py
  - translators/debian/capabilities.py
  - translators/debian/capabilities.json
  - translators/debian/variant.py
  - translators/debian/profile.py
  - translators/debian/generator.py
  - translators/debian/translate
  - translators/debian/contract.py
  - translators/debian/manifest.py
  - translators/debian/profiles/debian.yaml
  - translators/debian/templates/preseed.cfg.tpl
  - translators/debian/templates/chroot-install.hook.tpl
  - translators/debian/Dockerfile
  - translators/debian/README.md
  - translators/debian/tests/test_capability_gate.py
  - translators/debian/tests/test_profile.py
  - translators/debian/tests/test_variant.py
  - translators/debian/tests/test_generator.py
  - translators/debian/tests/fixtures/df_resolved.json
  - translators/debian/tests/fixtures/df_opinions.json
  - cli/build/build.go
  - cli/build/build_test.go
  - examples/dual-foundation/speech.yaml
  - examples/dual-foundation/opinions/DF-001.yaml
  - examples/dual-foundation/opinions/DF-002.yaml
  - examples/dual-foundation/opinions/DF-003.yaml
  - examples/dual-foundation/opinions/DF-004.yaml
  - examples/dual-foundation/opinions/DF-005.yaml
  - examples/dual-foundation/points/DF-base-cli.yaml
  - examples/dual-foundation/points/DF-system-tuning.yaml
  - scripts/dual-foundation-check.sh
  - scripts/debian-build-iso.sh
  - scripts/debian-validate-iso.sh
  - docs/arch-leak-audit.md
  - docs/ownership-model.md
findings:
  critical: 3
  warning: 5
  info: 3
  total: 11
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-06-13
**Depth:** standard
**Files Reviewed:** 35 (core implementation + tests + examples + scripts + docs)
**Status:** issues_found

## Summary

Phase 4 delivers the Debian translator (translators/debian/), the common/ extraction (translators/common/), the foundationRegistry refactor in build.go, the dual-foundation proof (examples/dual-foundation/), and related scripts and docs. The architecture is sound and the overall design is correct: invariant 1 is real, the common/ extraction is behavior-identical, the foundationRegistry dispatch is clean, and the dual-foundation test structure is correct.

However, three blockers require attention before this ships:

1. **Shell injection through opinion data embedded in the chroot hook** (CR-01) — the code claims T-04-06 protection ("all opinion data via build-manifest.json; %%SENTINEL%% replacement only") but directly embeds unvalidated sysctl key/value, kernel params, service names, group names, and file_asset mode into generated shell code via Python f-strings.

2. **Preseed user account sentinels are never replaced** (CR-02) — `_write_preseed` calls `content.replace("%%USERNAME%%", "%%USERNAME%%")` and similarly for `%%USER_FULLNAME%%` and `%%HASHED_PASSWORD%%`. These are identity no-ops. The resulting preseed.cfg will cause d-i to set `username="%%USERNAME%%"` and fail to parse `%%HASHED_PASSWORD%%` as a valid crypt hash, making the fully-unattended installation claim false.

3. **Chroot hook breaks when there are no packages** (CR-03) — when `manifest.target_packages` is empty, `pkg_lines = ":"` is substituted into the template, producing `apt-get install -y --no-install-recommends \<newline>:`. The `:` is not a valid apt package name; with `set -euo pipefail` active in the hook, this terminates the hook at the first line.

---

## Critical Issues

### CR-01: Shell injection via raw opinion data in chroot hook generation

**File:** `translators/debian/profile.py:259-301`

**Issue:** The `_write_chroot_hook` function's docstring and the chroot hook template both claim T-04-06 protection: "all opinion data via build-manifest.json; %%SENTINEL%% replacement only." This claim is false. Raw opinion data is embedded in the generated shell script at generation time via Python f-strings, with no validation or escaping:

- **Service names** (line 259): `f"systemctl enable {svc['name']}"` — a service name containing `; rm -rf /` or a newline would break the shell command.
- **Sysctl key/value** (line 271): `f"printf '{key} = {value}\\n' >> /etc/sysctl.d/{drop_in}"` — a key containing `'` (single-quote) breaks the `printf` argument; a `drop_in_file` containing `;` or `/dev/null` injects commands or redirects.
- **Kernel params** (line 286): `f"  sed -i 's/^GRUB_CMDLINE_LINUX=.*/GRUB_CMDLINE_LINUX=\"{params_str}\"/' /etc/default/grub"` — a kernel param key or value containing `'`, `*`, or `/` breaks the sed expression; a value like `' /etc/default/grub; reboot;#` allows sed expression injection.
- **Group names** (line 299): `f"getent group {group} >/dev/null 2>&1 || groupadd {group}"` — a group name with `;` injects commands after `groupadd`.
- **File asset mode** (line 250): `f"  install -Dm{mode} /debateos-assets/{dst} /{dst}"` — a mode like `0644; rm -rf /` injects a shell command.
- **Remove packages** (line 232-234): `"DEBIAN_FRONTEND=noninteractive apt-get remove -y " + " ".join(manifest.remove_packages)` — a package name with a shell metacharacter injects commands.

The Arch translator explicitly avoids this by using the T-02-09 pattern: all opinion data flows into `build-manifest.json` and is read at install time via `jq`. The Debian chroot hook breaks from this pattern without documenting the deviation or implementing equivalent protection.

The sentinel replacement (`%%STANZA%%` → stanza string) is the outer wrapper, but the stanza values themselves contain raw, unvalidated opinion data. This is shell injection.

**Fix:**

Option A (recommended — matches Arch's T-02-09 pattern): Write the sysctl params, kernel params, service names, group memberships, and remove packages into `build-manifest.json` and have the chroot hook read them via `jq`, never embedding values directly in shell code:
```bash
# In the hook template, read from the co-located build-manifest.json:
MANIFEST="/debateos-assets/build-manifest.json"
# Services:
jq -r '.system_services[] | select(.enable==true) | .name' "$MANIFEST" | while read -r svc; do
    systemctl enable "$svc"
done
# Sysctl:
jq -r '.sysctl_params[] | .key + " = " + .value' "$MANIFEST" >> /etc/sysctl.d/50-debateos.conf
```

Option B (lightweight input validation): Add a validator that rejects sysctl keys, kernel param keys/values, service names, group names, drop_in_file values, and mode values that contain characters outside a safe allowlist (`[a-zA-Z0-9._\-/]` for paths; `[a-zA-Z0-9._\-]` for identifiers; `[0-7]{3,4}` for modes) before any code generation:
```python
import re
_SAFE_SYSCTL_KEY = re.compile(r'^[a-zA-Z0-9._-]+$')
_SAFE_MODE = re.compile(r'^[0-7]{3,4}$')

def _validate_sysctl_key(key: str) -> str:
    if not _SAFE_SYSCTL_KEY.match(key):
        raise ValueError(f"sysctl key {key!r} contains unsafe characters")
    return key
```

---

### CR-02: Preseed user account sentinels are never replaced — unattended install fails

**File:** `translators/debian/profile.py:206-208`

**Issue:** The `_write_preseed` function contains three identity no-ops:
```python
content = content.replace("%%USERNAME%%", "%%USERNAME%%")
content = content.replace("%%USER_FULLNAME%%", "%%USER_FULLNAME%%")
content = content.replace("%%HASHED_PASSWORD%%", "%%HASHED_PASSWORD%%")
```

These are literal no-ops — the right-hand side is the same as the left-hand side. The emitted `preseed.cfg` will contain the literal strings `%%USERNAME%%`, `%%USER_FULLNAME%%`, and `%%HASHED_PASSWORD%%` as values for `d-i passwd/*` directives.

At install time, `debian-installer` will attempt to create a user with the literal username `%%USERNAME%%` and set the password hash to the string `%%HASHED_PASSWORD%%`. The latter is not a valid crypt hash (crypt hashes begin with `$`); depending on the d-i version, this either fails with an error (breaking unattended install) or sets an invalid password, making the installed system inaccessible.

The README says `debian-build-iso.sh` documents the replacement step, but that script contains no `sed` or replacement logic. There is no mechanism in the build pipeline to perform this replacement.

The claim that the generated profile tree is usable for "fully-unattended install" is false as long as these sentinels are in the preseed.

**Fix (two-part):**

1. Either supply real values at generation time (e.g., from a `--user` flag or a config section) or use a documented pre-build `sed` step in `debian-build-iso.sh`:
```bash
# In debian-build-iso.sh, before lb build:
PRESEED="${CONFIG_ABS}/config/includes.installer/preseed.cfg"
sed -i \
    "s/%%USERNAME%%/${DEBATEOS_USERNAME:-debian}/g" \
    "s/%%USER_FULLNAME%%/${DEBATEOS_USER_FULLNAME:-DebateOS User}/g" \
    "s/%%HASHED_PASSWORD%%/${DEBATEOS_HASHED_PASSWORD:?must set DEBATEOS_HASHED_PASSWORD}/g" \
    "$PRESEED"
```

2. At minimum, remove the three no-op lines and add a prominent comment explaining that these sentinels MUST be replaced before `lb build`:
```python
# NOTE: %%USERNAME%%, %%USER_FULLNAME%%, %%HASHED_PASSWORD%% are left as sentinels.
# The operator MUST replace them before running lb build:
#   sed -i "s/%%USERNAME%%/myuser/g" preseed.cfg
#   sed -i "s/%%USER_FULLNAME%%/My Name/g" preseed.cfg
#   sed -i "s/%%HASHED_PASSWORD%%/$(openssl passwd -6)/g" preseed.cfg
```

---

### CR-03: Chroot hook fails when `target_packages` is empty

**File:** `translators/debian/profile.py:225-227` and `translators/debian/templates/chroot-install.hook.tpl:16-17`

**Issue:** When `manifest.target_packages` is empty, `_write_chroot_hook` assigns:
```python
pkg_lines = ":"
```

The template expands to:
```bash
apt-get install -y --no-install-recommends \
:
```

In bash, the backslash-newline is a line continuation, so the actual command executed is:
```bash
apt-get install -y --no-install-recommends :
```

`apt-get` does not accept `:` as a valid package name and exits non-zero. The hook runs with `set -euo pipefail`, so the hook terminates immediately, causing `lb build` to fail. Any speech that results in zero target packages (e.g., a speech that only configures sysctl or group memberships, with all packages coming from `first-run` opinions) would produce a broken build.

The Python no-op `apt-get install :` was presumably intended to be the bash null command `:` (true), but it is not positioned correctly for that use.

**Fix:** Wrap the package installation in a conditional:
```bash
# In chroot-install.hook.tpl:
%%PACKAGES_STANZA%%
```

And in profile.py:
```python
if packages:
    pkg_lines_joined = " \\\n  ".join(packages)
    packages_stanza = (
        f"apt-get install -y --no-install-recommends \\\n  {pkg_lines_joined}"
    )
else:
    packages_stanza = "# No target packages (opinion set has no packaging-phase packages)"
content = content.replace("%%PACKAGES_STANZA%%", packages_stanza)
```

Alternatively, keep the current template and change the no-package fallback to output a blank section instead of `apt-get install :`.

---

## Warnings

### WR-01: `_write_preseed` replaces `%%PKGSEL_PACKAGES%%` with a hardcoded value

**File:** `translators/debian/profile.py:209`

**Issue:** The preseed d-i package selection is hardcoded to `"ssh openssh-server"` regardless of the manifest:
```python
content = content.replace("%%PKGSEL_PACKAGES%%", "ssh openssh-server")
```

This means every generated preseed installs `ssh` and `openssh-server` at d-i time, unconditionally. A speech whose opinions do not include SSH, or which actively wants to exclude it (e.g., an air-gapped desktop image), will have SSH installed against its intent. The `pkgsel` stanza is the d-i-level package selection — it should either be empty (all packages come from the chroot hook) or reflect what the manifest declares for the installer phase.

**Fix:** Either omit the `%%PKGSEL_PACKAGES%%` sentinel entirely and hardcode an empty `pkgsel/include string` in the template (since packages come from the chroot hook), or compute it from the manifest's `target_packages` filtered to `install_phase == "packaging"`:
```python
pkgsel_packages = " ".join(manifest.target_packages) if manifest.target_packages else ""
content = content.replace("%%PKGSEL_PACKAGES%%", pkgsel_packages)
```

---

### WR-02: `dual-foundation-check.sh` runs the Arch translator on a `foundation: debian` speech

**File:** `scripts/dual-foundation-check.sh:134-146` and `examples/dual-foundation/speech.yaml:4`

**Issue:** The dual-foundation proof speech declares `foundation: debian`. When resolved, `resolved.json` carries `"foundation": "debian"`. The check script then runs the **Arch translator** on this same resolved.json (Step 2). The Arch translator does not validate that the foundation field matches its expected value; it silently produces an `arch-profile/build-manifest.json` with `"foundation": "debian"`.

This creates a semantic inconsistency: the arch-profile output purports to be a "debian" foundation install while being an Arch archiso tree. While the equivalence check (Step 4f) correctly tests package sets, the `foundation` field mismatch in the arch-profile could mislead downstream consumers of `build-manifest.json` and represents an untrue claim in the build artifact.

The intent — to prove the same resolved.json can drive both translators — is valid, but the proof is better structured using a speech whose `foundation` is explicitly neutral or by noting in the check that the arch translator ignores the field.

**Fix (recommended):** Document the known inconsistency at the top of the check script, and add a Step 4g assertion that the arch manifest `foundation` field differs from the speech foundation (as a known expected behavior, not a bug), or give the speech `foundation: debian` and note that the Arch translator is invoked explicitly out-of-band for the proof only. At minimum, add a comment:
```bash
# Step 2: We intentionally run the Arch translator on a debian-foundation resolved.json
# to prove the abstraction holds. The Arch translator ignores foundation for translation
# but the emitted build-manifest.json will carry foundation="debian" -- this is expected
# and documented as part of the dual-foundation proof design.
```

---

### WR-03: `go test` pass detection in `dual-foundation-check.sh` false-passes when test binary fails to build

**File:** `scripts/dual-foundation-check.sh:286-292`

**Issue:**
```bash
go test ./... -count=1 >"${WORK_DIR}/go-test-all.log" 2>&1 || true
if grep -Eq '^ok ' "${WORK_DIR}/go-test-all.log" && ! grep -Eq '^FAIL' "${WORK_DIR}/go-test-all.log"; then
    pass "go test ./... -count=1: all packages GREEN"
```

When a Go package fails to compile, `go test` outputs `FAIL <package> [build failed]` to stderr and exits non-zero. The `|| true` suppresses the non-zero exit. The `FAIL` line goes to the log file via `2>&1`, so `grep -Eq '^FAIL'` catches it — the gate correctly fires.

However, there is a subtler case: if `go test ./...` encounters zero testable packages (no `_test.go` files in any package), it outputs nothing and exits 0. The `^ok` grep then fails → the gate fires with `fail "go test ./... failed"` even though nothing is wrong. This is an unlikely scenario in a non-trivial repo but still a correctness gap.

More importantly: `FAIL\t<pkg>\t[build failed]` lines use a tab separator, and `'^FAIL'` matches the prefix correctly — this part is OK.

**Fix:** Add a `[no test files]` exclusion and check for zero-test scenario:
```bash
# After go test, check log is non-empty and contains results:
if grep -Eq '^(ok|FAIL)\b' "${WORK_DIR}/go-test-all.log" && \
   ! grep -Eq '^FAIL\b' "${WORK_DIR}/go-test-all.log"; then
    pass "go test ./... -count=1: all packages GREEN"
elif grep -q 'no test files' "${WORK_DIR}/go-test-all.log"; then
    pass "go test ./... -count=1: no test files (skipped)"
else
    fail "go test ./... failed"
fi
```

---

### WR-04: Arch re-export shims expose private names that aren't in `__all__`

**File:** `translators/arch/contract.py:13-18`, `translators/arch/manifest.py:18-22`, `translators/arch/firstrun.py:18-26`

**Issue:** The Arch shim for `contract.py` re-exports private names:
```python
from common.contract import (
    load_resolved_speech,
    load_opinion_bodies,
    _load_opinions_from_json_file,   # ← private
    _load_opinions_from_directory,   # ← private
    _RESOLVED_SPEECH_KEYS,           # ← private
)
```

Similarly for `manifest.py` (`_MIN_EPOCH`, `_MAX_EPOCH`) and `firstrun.py` (`_DEFAULT_TEMPLATES_DIR`, `_flag_file_path`, `_load_unit_template`). These private names are not declared in `__all__`. Any test or module that currently imports them from `translators.arch.contract` (by name, not via `from contract import *`) will continue to work, but the shim is simultaneously over-exposing private internals while under-documenting that this exposure is intentional.

More practically: if a future refactor renames `_RESOLVED_SPEECH_KEYS` to `_SPEECH_KEYS` in common/, the shim import breaks silently (or raises ImportError). The private re-exports are not tested by the existing test suite.

**Fix:** Either add the private names to `__all__` of the shim (making the intent explicit) or remove them from the shim imports and let any code that needs them import from `translators.common.contract` directly. Prefer removal — private names should not be part of the shim contract:
```python
# arch/contract.py — simpler, correct shim
from common.contract import load_resolved_speech, load_opinion_bodies  # noqa: F401
__all__ = ["load_resolved_speech", "load_opinion_bodies"]
```

---

### WR-05: `profile.py` calls `_sanitize_dst` twice per file asset

**File:** `translators/debian/profile.py:162-163` and `246`

**Issue:** `emit_profile_tree` validates all file_asset dst paths in a pre-flight loop (lines 162-163):
```python
for fa in manifest.file_assets:
    _sanitize_dst(fa.get("dst", ""))
```

Then `_write_chroot_hook` (called immediately after) calls `_sanitize_dst` again for each asset (line 242):
```python
dst = _sanitize_dst(fa.get("dst", ""))
```

The double validation is redundant. More importantly, the pre-flight loop uses `fa.get("dst", "")` which would call `_sanitize_dst("")` and raise `ValueError` (empty path rejection) before the outer function raises. This is the intended fail-fast behavior, but the duplication means the same error could appear to originate from two different call sites depending on which assertion fires first.

**Fix:** Accept the redundancy as belt-and-suspenders (the pre-flight guard is the important one for CR-04 compliance), but add a comment:
```python
# Pre-flight T-04-05 gate: validate ALL dst paths before any I/O.
# _write_chroot_hook also calls _sanitize_dst when building the stanza — intentional
# double-check that ensures the stanza builder always gets the normalized path.
```

Or, cache the sanitized paths from the pre-flight loop and pass them through rather than re-sanitizing.

---

## Info

### IN-01: `_preseed_V1` header may not be required for d-i preseed

**File:** `translators/debian/templates/preseed.cfg.tpl:1`

**Issue:** The preseed template begins with `_preseed_V1`, which is the format identifier used by `debconf-set-selections`. The `d-i` format used in the rest of the file (lines starting with `d-i owner/question type value`) is the standard preseed syntax accepted by the Debian Installer directly. The `_preseed_V1` header is appropriate when loading via `debconf-set-selections` but is not part of the `d-i` preseed specification.

Most live-build configurations load `preseed.cfg` via `debconf-set-selections` at boot time before d-i starts, so the header is likely benign. However, if the preseed is loaded via a different mechanism (e.g., `auto=true` kernel parameter pointing to the file), the header may cause a parse warning or be treated as an unknown directive.

**Fix:** Test with live-build on a capable host to confirm `_preseed_V1` is accepted in the live-build preseed context. If not needed, remove it:
```
# config/includes.installer/preseed.cfg
# Generated by translators/debian/profile.py
```

---

### IN-02: `dual-foundation-check.sh` uses `ls *.iso` which is not robust

**File:** `scripts/dual-foundation-check.sh:359`

**Issue:**
```bash
ISO_FILE="$(ls "${ISO_DIR}"/*.iso 2>/dev/null | head -1 || true)"
if [[ -f "${ISO_FILE:-}" ]]; then
```

`ls` output is not suitable for programmatic use (filenames with spaces, symlinks). The safer pattern is `find ... | head -1`. Additionally, `ISO_FILE:-` with an empty default results in `[[ -f "" ]]` which is always false — this is correct behavior but masks whether the file glob was attempted.

**Fix:**
```bash
ISO_FILE="$(find "${ISO_DIR}" -maxdepth 1 -name '*.iso' | head -1 || true)"
```

---

### IN-03: `debian-build-iso.sh` uses `bash -c "..."` heredoc (T-03-DKARG equivalent concern)

**File:** `scripts/debian-build-iso.sh:218-236`

**Issue:** The Docker run command uses `bash -c "..."` with a multi-line string containing literal variable references (`${SOURCE_DATE_EPOCH}`, paths). This is a shell-within-shell pattern — the inner bash -c script is a string passed to the container's shell. While `SOURCE_DATE_EPOCH` is validated (it's an integer derived from a hash), the `CONFIG_ABS` and `OUT_ABS` paths are derived from user-provided positional arguments and could contain shell-special characters (spaces, semicolons) if the paths are unusual.

The T-03-DKARG pattern in `build.go` uses variadic args specifically to avoid this. The scripts directory is exempt from the T-03 constraint (these are shell scripts, not Go code), but the risk is worth noting.

**Fix:** For `CONFIG_ABS` and `OUT_ABS`, validate that they contain no shell metacharacters before interpolation, or use Docker volume mounting to avoid needing to interpolate them inside the container command:
```bash
# Alternatively: pass paths via environment variables that Docker -e can inject,
# then reference them inside the container as $VAR (not ${}):
docker run ... -e "CONFIG_PATH=/debateos-config" -e "OUT_PATH=/out" ...
```

---

## Summary of Key Findings by Category

| # | ID | Category | Severity | File | Line |
|---|-----|----------|----------|------|------|
| 1 | CR-01 | Shell injection via raw opinion data in chroot hook | BLOCKER | profile.py | 259-301 |
| 2 | CR-02 | Preseed user account sentinels never replaced | BLOCKER | profile.py | 206-208 |
| 3 | CR-03 | Chroot hook fails on empty target_packages | BLOCKER | profile.py | 225-227 |
| 4 | WR-01 | PKGSEL_PACKAGES hardcoded to ssh/openssh-server | WARNING | profile.py | 209 |
| 5 | WR-02 | Arch translator invoked on debian-foundation speech | WARNING | dual-foundation-check.sh | 134 |
| 6 | WR-03 | go test pass detection false-passes on build failure edge case | WARNING | dual-foundation-check.sh | 286-292 |
| 7 | WR-04 | Shims re-export private names not in __all__ | WARNING | arch/contract.py, manifest.py, firstrun.py | 13-26 |
| 8 | WR-05 | _sanitize_dst called twice per file asset | WARNING | profile.py | 162, 246 |
| 9 | IN-01 | _preseed_V1 header may be unnecessary | INFO | preseed.cfg.tpl | 1 |
| 10 | IN-02 | ls *.iso glob not robust | INFO | dual-foundation-check.sh | 359 |
| 11 | IN-03 | Docker bash -c with interpolated paths | INFO | debian-build-iso.sh | 219 |

---

## What Is Correct

For completeness, the following areas were verified and found correct:

- **sig_level → apt mapping**: `variant.py apply_variant()` correctly maps `Required`/`RequiredDatabaseOptional` to `[signed-by=...]` and `OptionalTrustAll`/`Never` to `[trusted=yes]` with trust warnings. The `Never` case correctly emits a LOUD WARNING comment in the archive file AND a trust_warning entry (T-04-07). No case silently trusts without warning.
- **_sanitize_dst**: Correctly rejects absolute paths, `..` traversal, empty strings, and `.`. Uses the correct sentinel prefix pattern. The fail-fast gate (pre-flight validation before any I/O) correctly fires before any file is created.
- **foundationRegistry dispatch**: `build.go` correctly dispatches by `rs.Foundation`, returns an error for unknown foundations, resolves `--profile ""` to the foundation default, and is backward-compatible for `arch` + explicit `--profile vanilla-arch`.
- **common/ extraction**: The Arch shims re-export the correct public symbols; the generator modules add both the translator dir and the translators/ dir to sys.path, making bare-name imports and `import common` work.
- **capabilities.json invariant**: Debian capabilities.json correctly omits `mkinitcpio`, `limine`, and `pacman-AUR` tokens. All five dual-foundation capability tokens are declared.
- **chroot_install suffix**: Package list correctly uses `.list.chroot_install` (not `.list.chroot`) for both-environment installation.
- **Hook executable bit**: `emit_profile_tree` correctly sets chmod 0755 on the hook file.
- **Dockerfile digest pinning**: Base image is pinned by sha256 digest (T-04-13). The comment correctly documents when to re-verify.
- **dual-foundation-check.sh PASS/FAIL counting**: `((PASS++)) || true` is correct — the `|| true` is necessary because `((expr))` returns non-zero when the result is zero (when PASS is 0 before increment).
- **arch-leak-audit.md accuracy**: The audit correctly identifies `build.go` as the only genuine infrastructure leak. The claims that `sig_level`, `install_phase`, `mkinitcpio`, and `limine` tokens are either intentional abstractions or correctly isolated are all verifiable from the code.
- **Dual-foundation opinions invariant-1 compliance**: All five DF opinions (DF-001 through DF-005) contain only `intent` fields with no apt/pacman/dpkg references. The `category` and `translator_capabilities` fields are genuinely foundation-neutral.
- **Preseed partman/LVM configuration**: Includes all required `partman` confirmation directives (`partman/confirm`, `partman/confirm_nooverwrite`, `partman-lvm/confirm`, `partman-lvm/confirm_nooverwrite`, `partman-md/confirm`, `partman-md/confirm_nooverwrite`), covering the common failure modes for unattended partitioning.

---

_Reviewed: 2026-06-13_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
