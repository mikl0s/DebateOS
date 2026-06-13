---
phase: 03-cli-build-channels
reviewed: 2026-06-13T12:37:21Z
depth: standard
files_reviewed: 19
files_reviewed_list:
  - cli/config/config.go
  - cli/runner/runner.go
  - cli/runner/fake.go
  - cli/internal/loader/loader.go
  - cli/compose/compose.go
  - cli/validate/validate.go
  - cli/pane/pane.go
  - cli/pane/age.go
  - cli/build/build.go
  - cli/build/inject.go
  - cmd/debateos/main.go
  - build/docker/Dockerfile
  - build/docker/entrypoint.sh
  - .dockerignore
  - build/actions/build-speech.yml
  - .github/workflows/build-speech.yml
  - scripts/determinism-test.sh
  - scripts/secret-free-check.sh
  - scripts/check-coverage.sh
findings:
  critical: 4
  warning: 7
  info: 3
  total: 14
status: issues_found
fixes_applied: true
resolved:
  - CR-01
  - CR-02
  - CR-03
  - CR-04
  - WR-01
  - WR-02
  - WR-03
  - WR-04
  - WR-05
  - WR-06
  - WR-07
  - IN-01
  - IN-02
  - IN-03
fixed_at: 2026-06-13T14:00:00Z
---

# Phase 03: Code Review Report

**Reviewed:** 2026-06-13T12:37:21Z
**Depth:** standard
**Files Reviewed:** 19
**Status:** issues_found

## Summary

The CLI skeleton, pane management, age crypto, and subcommand dispatch are structurally sound. Security-critical paths (no shell interpolation in Runner, variadic exec args, 0600 file creation) are correctly implemented. The test suite is behavior-driven and covers error paths well.

Four blockers require attention before this code ships: a self-referential GitHub Actions workflow that will never execute, a missing gzip determinism flag that silently breaks the determinism guarantee between time-separated runs, a deferred private-pane merge that is contractually promised in 03-CONTEXT.md but not implemented, and a swallowed tar-writer close error that can produce a silently corrupt artifact. Seven warnings cover security hygiene gaps (unpinned builder image, unquoted variable in shell, missing perm re-check on existing sensitive files, etc.).

---

## Critical Issues

### CR-01: GitHub Actions thin-caller workflow is self-referential and will never execute

**File:** `.github/workflows/build-speech.yml:32`
**Issue:** The thin-caller workflow (`on: push/workflow_dispatch`) calls `uses: ./.github/workflows/build-speech.yml` — which is the same file. GitHub Actions prohibits a workflow from calling itself as a reusable workflow and will reject this with a parse or cycle error. Additionally, the actual reusable workflow lives at `build/actions/build-speech.yml`, which GitHub will never discover as a callable reusable workflow because GitHub only supports reusable workflows stored under `.github/workflows/`. The CI pipeline is therefore entirely broken: no run will ever succeed.

**Fix:** Copy (or symlink, or canonically relocate) the reusable workflow into `.github/workflows/` under a distinct filename and update the caller to reference it:

```yaml
# .github/workflows/reusable-build-speech.yml  (moved from build/actions/)
on:
  workflow_call:
    # ... (content unchanged from build/actions/build-speech.yml)

# .github/workflows/build-speech.yml  (thin caller — fixed reference)
jobs:
  build-omarchy-example:
    uses: ./.github/workflows/reusable-build-speech.yml   # distinct filename
    with:
      speech-dir: examples/omarchy
      profile: vanilla-arch
      skip-iso: true
```

Cross-repo callers must then use:
```
uses: mikl0s/DebateOS/.github/workflows/reusable-build-speech.yml@main
```

---

### CR-02: Determinism gate does not export SOURCE_DATE_EPOCH; gzip header mtime is non-deterministic

**File:** `scripts/determinism-test.sh:111-164`
**Issue:** The script derives `EPOCH` from sha256(resolved.json) and correctly passes it as `--mtime=@${EPOCH}` to `tar`. However, it never exports `SOURCE_DATE_EPOCH` into the environment before the `tar` invocations. GNU tar's `-z` flag invokes gzip, which embeds the current wall-clock time in the gzip stream header unless `SOURCE_DATE_EPOCH` is set in the environment. Two runs separated by more than one second will produce gzip headers with different mtimes → different SHA-256 checksums → determinism gate returns false positive (FAIL) or, worse, masks non-determinism because both runs happen within the same second. The script's own comment on line 16 explicitly states `gzip -n` should be used, confirming the author was aware of this requirement, but the flag was never added.

**Fix:** Export `SOURCE_DATE_EPOCH` before the tar invocations, and/or pipe through `gzip -n`:

```bash
# Option A: export (GNU tar respects it for the gzip header)
export SOURCE_DATE_EPOCH="${EPOCH}"

tar \
    --sort=name \
    --mtime="@${EPOCH}" \
    --owner=0 --group=0 --numeric-owner \
    --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
    -czf "${TAR1}" -C "${PROFILE_DIR1}" .

# Option B: pipe through gzip -n explicitly (more portable)
tar \
    --sort=name \
    --mtime="@${EPOCH}" \
    --owner=0 --group=0 --numeric-owner \
    --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
    -c -C "${PROFILE_DIR1}" . | gzip -n > "${TAR1}"
```

Apply the same fix to the TAR2 invocation.

---

### CR-03: `debateos build` does not merge the private pane into resolution (contract violation)

**File:** `cli/build/build.go:104`
**Issue:** `03-CONTEXT.md` line 24 specifies: "`debateos build` — resolve **(merging private pane locally)**". Line 29 further states: "`pane.yaml` lives only in the XDG dir, 0600, never copied into shared artifacts; **resolution merges it locally** (private opinions overlay public speech)." The implementation calls `loader.ResolveDir(speechDir)` which loads only from the speech directory — not from the config dir where `pane.yaml` lives. The `WriteInjectionTar` call on line 180 explicitly notes "No private-pane assets to inject in the base implementation." While the injection tar is intentionally empty as a scaffold, the resolution step itself does not overlay private opinions. Users running `debateos build` on a speech that has a private pane in `~/.config/debateos/pane.yaml` will silently get a build that ignores their private opinions.

**Fix:** After resolving `speechDir`, load and merge the private pane from the config dir. A minimal approach:

```go
// After loader.ResolveDir(speechDir), load pane.yaml from config dir if present.
configDir, configErr := config.DebateOSDir()
if configErr == nil {
    if paneOpinions, paneErr := loader.LoadPaneOverlay(configDir); paneErr == nil && len(paneOpinions) > 0 {
        // Re-resolve with merged opinions; or pass paneOpinions to a separate merge step.
    }
}
```

The exact merge API depends on whether the resolver supports overlay opinions. If not yet implemented, document this as a known limitation and emit a warning on stderr rather than silently omitting the private pane.

---

### CR-04: Deferred `tw.Close()` error is swallowed; corrupt tar returned with nil error

**File:** `cli/build/inject.go:142`
**Issue:** `tw.Close()` is registered as a `defer` on line 142. `tw.Close()` writes the two end-of-archive 512-byte blocks and flushes the underlying writer. If this write fails (e.g., disk full), the error is silently discarded. The `tw.Flush()` call on line 176 flushes buffered data but does NOT write the end-of-archive blocks — that is exclusively `tw.Close()`'s job. `WriteInjectionTar` can therefore return `(tarPath, nil)` for an incomplete, unreadable tar archive. The first-boot unit will fail to extract the injection data, silently losing private pane secrets.

**Fix:** Call `tw.Close()` explicitly before the deferred `f.Close()` and check the error:

```go
// Remove: defer tw.Close()

// After all writes succeed, replace tw.Flush() with:
if err := tw.Close(); err != nil {
    return "", fmt.Errorf("tar close: %w", err)
}
// f.Close() is still deferred; its error loss is acceptable for a read-only output file.
```

Remove the separate `tw.Flush()` call at line 176 (it is superseded by the explicit `tw.Close()`).

---

## Warnings

### WR-01: Golang builder stage in Dockerfile is not digest-pinned

**File:** `build/docker/Dockerfile:25`
**Issue:** The runtime stage (`archlinux:base-devel`) is correctly pinned by SHA-256 digest. The builder stage uses `FROM golang:1.24` — a mutable floating tag. A compromised or silently-updated `golang:1.24` image would inject a malicious compiler into every build without any version change visible in the Dockerfile. The document header explicitly calls out the archlinux digest-pin requirement ("Re-verify the digest quarterly") but the builder stage has no equivalent.

**Fix:**
```dockerfile
# Pin the builder to a specific digest (verify and update quarterly).
FROM golang:1.24@sha256:<digest> AS builder
```

Obtain the current digest with:
```bash
docker pull golang:1.24 && docker inspect --format='{{index .RepoDigests 0}}' golang:1.24
```

---

### WR-02: Unquoted `${SKIP_ISO_FLAG}` / `${SKIP_FLAG}` variables cause word-splitting risk

**File:** `build/docker/entrypoint.sh:59`
**File:** `build/actions/build-speech.yml:82`
**Issue:** Both shell scripts use `${SKIP_ISO_FLAG}` / `${SKIP_FLAG}` unquoted in command position:

```bash
exec /usr/local/bin/debateos build \
    --dir  "${SPEECH_DIR}" \
    ...
    ${SKIP_ISO_FLAG}        # <-- unquoted
```

When `SKIP_ISO_FLAG` is empty, this expands to nothing (correct). When set to `--skip-iso`, word-splitting could theoretically produce extra empty tokens if the variable were changed to contain spaces. More critically, this is a `set -euo pipefail` context: `nounset` (`-u`) would error if the variable were accidentally unset rather than empty-string. The current initialization `SKIP_ISO_FLAG=""` prevents the `nounset` error, but the unquoted expansion is a style trap — any future addition of spaces to the value (e.g., `"--skip-iso --dry-run"`) would silently split into separate arguments. Both scripts should use arrays.

**Fix (entrypoint.sh):**
```bash
SKIP_ARGS=()
if [ "${SKIP_ISO:-0}" = "1" ]; then
    SKIP_ARGS=("--skip-iso")
fi

exec /usr/local/bin/debateos build \
    --dir  "${SPEECH_DIR}" \
    --profile "${PROFILE}" \
    --out  "${OUT_DIR}" \
    "${SKIP_ARGS[@]}"
```

Apply the same array pattern in `build-speech.yml`'s `run:` block.

---

### WR-03: `pane.yaml` and `identity.age` permissions not re-checked on existing files

**File:** `cli/pane/pane.go:99-116` (loadPane)
**File:** `cli/pane/age.go:35-46` (LoadOrCreateIdentity, existing-file path)
**Issue:** Both `loadPane` and `LoadOrCreateIdentity` correctly create files at 0600. However, when an existing file is read, neither function checks whether the current permissions are still 0600. If an external tool, editor, or misconfigured backup restored `pane.yaml` or `identity.age` with wider permissions (e.g., 0644), the CLI silently reads the looser-permission file and proceeds. 03-CONTEXT.md security requirement T-03-PERM states "perms enforced not just on create". On read, the files could have been exposed to other users without detection.

**Fix:** Add a permission check after successful `os.ReadFile` / `os.Open`:

```go
// In loadPane, after os.ReadFile succeeds:
info, statErr := os.Stat(path)
if statErr == nil && info.Mode().Perm() != 0600 {
    return nil, fmt.Errorf("%s has insecure permissions %04o (want 0600); "+
        "run: chmod 0600 %s", path, info.Mode().Perm(), path)
}
```

Apply the same check in `LoadOrCreateIdentity` after `os.ReadFile` on the existing identity path.

---

### WR-04: `savePane` is not atomic — partial write leaves pane.yaml corrupt on failure

**File:** `cli/pane/pane.go:119-138` (savePane)
**Issue:** `savePane` truncates `pane.yaml` with `O_TRUNC` and writes the new content. If the process is killed, the disk is full, or `f.Write` partially succeeds, `pane.yaml` is left in a corrupt or truncated state with no way to recover to the prior version. The `defer f.Close()` will close the file but cannot undo a partial write. For a file storing private configuration, silent data loss on write failure is unacceptable.

**Fix:** Use a write-to-temp-then-rename pattern (atomic on POSIX):

```go
func savePane(dir string, m paneData) error {
    if err := os.MkdirAll(dir, 0700); err != nil {
        return fmt.Errorf("mkdir %s: %w", dir, err)
    }
    out, err := yaml.Marshal(m)
    if err != nil {
        return fmt.Errorf("marshal pane: %w", err)
    }
    tmp, err := os.CreateTemp(dir, ".pane.yaml.tmp.*")
    if err != nil {
        return fmt.Errorf("create temp: %w", err)
    }
    tmpName := tmp.Name()
    defer func() { _ = os.Remove(tmpName) }() // clean up on failure
    if err := os.Chmod(tmpName, 0600); err != nil {
        tmp.Close()
        return fmt.Errorf("chmod temp: %w", err)
    }
    if _, err := tmp.Write(out); err != nil {
        tmp.Close()
        return fmt.Errorf("write temp: %w", err)
    }
    if err := tmp.Close(); err != nil {
        return fmt.Errorf("close temp: %w", err)
    }
    return os.Rename(tmpName, filepath.Join(dir, paneFile))
}
```

Note: `os.CreateTemp` creates with 0600 on Linux by default; the explicit `Chmod` is for clarity and cross-platform safety.

---

### WR-05: `os.Getwd()` error silently ignored; output directory becomes relative

**File:** `cli/build/build.go:99`
**Issue:**
```go
cwd, _ := os.Getwd()
outDir = filepath.Join(cwd, outDir)
```
If `os.Getwd()` fails (e.g., the working directory has been deleted), `cwd` is the empty string and `outDir` becomes a relative path `"out"`. The subsequent `os.MkdirAll(outDir, 0755)` may create a directory relative to whatever the process's cwd resolves to, potentially placing build artifacts in an unexpected location.

**Fix:**
```go
cwd, err := os.Getwd()
if err != nil {
    fmt.Fprintf(stderr, "build: getcwd: %v\n", err)
    return 1
}
outDir = filepath.Join(cwd, outDir)
```

---

### WR-06: `.dockerignore` `*.age` pattern only excludes root-level `.age` files

**File:** `.dockerignore:10`
**Issue:** Docker `.dockerignore` patterns without a path separator (`/`) match only at the root of the build context (per Docker's Go `filepath.Match` semantics). The pattern `*.age` will NOT exclude `some/nested/dir/key.age`. If an `.age` file is accidentally committed inside a subdirectory of the repository (e.g., `translators/arch/key.age` or `examples/user.age`), it would be included in the Docker build context and potentially baked into image layers. The current threat model assumes `.age` files only exist in `~/.config/debateos/` (outside the build context), but defense-in-depth suggests using `**/*.age`.

**Fix:**
```
# .dockerignore
**/*.age
identity.age
```

---

### WR-07: Speech directory mounted read-write into Docker container (PRIV-01 gap)

**File:** `cli/build/build.go:153`
**Issue:** The `entrypoint.sh` documentation (line 16) specifies the speech mount as read-only (`-v <speech-dir>:/speech:ro`). The `build.go` implementation mounts it read-write:
```go
"-v", speechDir + ":/speech",   // no :ro suffix
```
The Docker container can therefore write into the user's speech directory. While the current `debateos build` container flow does not write to `/speech`, a future entrypoint change or buggy translator could silently modify the user's speech, corrupting their data.

**Fix:**
```go
"-v", speechDir + ":/speech:ro",
```

---

## Info

### IN-01: `LoadOrCreateIdentity` uses `fmt.Sprintf` instead of `filepath.Join` for path construction

**File:** `cli/pane/age.go:32`
**Issue:**
```go
path := fmt.Sprintf("%s/%s", dir, identityFileName)
```
All other path constructions in the codebase use `filepath.Join`. The `fmt.Sprintf` form fails to handle edge cases like a trailing `/` in `dir` (produces `//identity.age`). On the target platform (Linux) this is benign but inconsistent and fragile.

**Fix:**
```go
path := filepath.Join(dir, identityFileName)
```

---

### IN-02: `EncryptFile` docstring claims streaming but implementation fully buffers in RAM

**File:** `cli/pane/age.go:70-75`
**Issue:** The function comment states "The age format is streaming so large files do not need to be fully buffered." However, line 75 calls `os.ReadFile(src)` which loads the entire plaintext into RAM. For the intended use case (`pane.yaml` is small) this is harmless, but the comment is misleading and sets false expectations for future callers passing larger files.

**Fix:** Either update the comment to accurately reflect buffered behaviour, or replace `os.ReadFile` with a streaming copy:
```go
// Remove: plaintext, err := os.ReadFile(src)
// Replace with:
sf, err := os.Open(src)
if err != nil {
    return fmt.Errorf("read %s: %w", src, err)
}
defer sf.Close()
// ... then io.Copy(w, sf) instead of w.Write(plaintext)
```

---

### IN-03: `private-injection.tar` manifest `Created` timestamp is non-deterministic

**File:** `cli/build/inject.go:121`
**Issue:**
```go
Created: time.Now().UTC().Format(time.RFC3339),
```
The `debateos-private.json` manifest embedded in `private-injection.tar` contains a wall-clock creation timestamp. This makes the tar non-deterministic between runs. The tar file headers themselves correctly use `time.Unix(0, 0)`, but the manifest JSON content varies. The `determinism-test.sh` only covers `arch-profile/` (not `private-injection.tar`), so this defect is not caught by the existing gate. If/when determinism testing is extended to cover the injection tar, this will break that gate.

**Fix:**
```go
// Use the same epoch derived from resolved.json, passed in as a parameter.
// Or use a fixed sentinel for the base implementation (empty tar case):
Created: time.Unix(0, 0).UTC().Format(time.RFC3339),
```

For the non-empty-tar case, the caller should pass `epoch` from `DeriveEpoch` into `WriteInjectionTar` and use it here.

---

_Reviewed: 2026-06-13T12:37:21Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
