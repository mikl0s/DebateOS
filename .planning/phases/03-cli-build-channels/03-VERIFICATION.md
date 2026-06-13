---
phase: 03-cli-build-channels
verified: 2026-06-13T13:10:00Z
status: human_needed
score: 10/10 must-haves verified (code level)
overrides_applied: 0
human_verification:
  - test: "Full ISO build via docker run (without --skip-iso) on a capable host"
    expected: "out/*.iso produced, bootable unattended Arch installer matching Omarchy north star"
    why_human: "Requires kernel devtmpfs access unavailable on Proxmox VE; mkarchiso will not run"
  - test: "Live cross-repo GitHub Actions run from a fork"
    expected: >
      Fork calls uses: mikl0s/DebateOS/.github/workflows/reusable-build-speech.yml@main;
      job appears in the fork's runner (not upstream); artifact uploaded to fork's run.
      Both DEFERRED per VALIDATION.md Manual-Only Verifications table.
    why_human: "Requires GitHub Actions CI minutes on a fork; cannot be tested programmatically on this host"
---

# Phase 3: CLI & Build Channels Verification Report

**Phase Goal:** Anyone can go compose → resolve → build to an ISO at zero cost, via local Docker or their own GitHub Actions minutes, deterministically, with their private pane never leaving their control

**Verified:** 2026-06-13T13:10:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Environment Notes (Policy-Required Limitations)

Per the orchestrator VERIFICATION POLICY and 03-VALIDATION.md Manual-Only Verifications:
- **This host CANNOT run mkarchiso**: Proxmox VE devtmpfs restriction — ISO build deferred to human
- **Live GitHub Actions run CANNOT be tested**: requires fork + CI minutes — deferred to human
- All code-level criteria are fully verified below; the two human items are environment-blocked, not code gaps

---

## Goal Achievement

### Observable Truths

| #  | Truth                                                                                                                                                    | Status     | Evidence                                                                                                                   |
|----|----------------------------------------------------------------------------------------------------------------------------------------------------------|------------|----------------------------------------------------------------------------------------------------------------------------|
| 1  | `debateos compose/validate/build/pane` all dispatch correctly and return exit codes without calling os.Exit                                              | ✓ VERIFIED | All four subcommands wired in cmd/debateos/main.go; all tests pass including TestCompose, TestValidate, TestBuildSkipISO, TestPane* |
| 2  | Speech + private pane managed in $HOME (XDG); optional private-git backup                                                                               | ✓ VERIFIED | config.DebateOSDir() resolves DEBATEOS_DIR → os.UserConfigDir()/debateos; backup/restore via git through FakeRunner |
| 3  | pane.yaml and identity.age created 0600 (CLI-02)                                                                                                         | ✓ VERIFIED | TestPanePermissions asserts 0600 after multiple sets; age.go L65 uses O_CREATE|O_WRONLY|O_TRUNC, 0600; WR-03 re-check on existing files verified |
| 4  | Private pane entries reach private-injection.tar; tar is in out dir NOT in arch-profile/; no plaintext secret in arch-profile                            | ✓ VERIFIED | Live run with DEBATEOS_DIR set to temp dir containing pane.yaml: etc/debateos/ssh_authorized_key + etc/debateos/secret_token in tar; pane.yaml/identity.age NOT in arch-profile/; grep for secret string in profile returns empty |
| 5  | same Docker image powers both build channels (local Docker + GitHub Actions)                                                                             | ✓ VERIFIED | build/docker/Dockerfile builds ghcr.io/mikl0s/debateos:latest; both .github/workflows/reusable-build-speech.yml and build/actions/build-speech.yml use `container: image: ghcr.io/mikl0s/debateos:latest` (substantively identical) |
| 6  | Deterministic builds: identical inputs → byte-identical tars; SOURCE_DATE_EPOCH exported; gzip -n used (BLD-03 / CR-02 fix verified)                    | ✓ VERIFIED | determinism-test.sh: both runs produce sha256 `99db28900f05b1230baaec16b3a729144fe8582f9a39b50c1835f2cc6d73b069`; script exports SOURCE_DATE_EPOCH L138 and pipes through `gzip -n` L172+179 |
| 7  | secret-free-check.sh gate passes: no pane.yaml/identity.age/private-injection.tar in profile tree                                                       | ✓ VERIFIED | `bash scripts/secret-free-check.sh` exit 0; all three patterns absent from arch-profile/ |
| 8  | check-coverage.sh: resolver/ ≥ 90%, cli/ ≥ 85%                                                                                                          | ✓ VERIFIED | resolver/ 93.5% >= 90% OK; cli/ 85.4% >= 85% OK |
| 9  | CR-01 fixed: .github/workflows/build-speech.yml calls .github/workflows/reusable-build-speech.yml (a distinct, real file — NOT itself)                  | ✓ VERIFIED | build-speech.yml L32: `uses: ./.github/workflows/reusable-build-speech.yml`; reusable-build-speech.yml exists under .github/workflows/ with `on: workflow_call` |
| 10 | Zero cost, no central service in build path; archlinux digest identical in build/docker/Dockerfile and translators/arch/Dockerfile (same-image property) | ✓ VERIFIED | Both Dockerfiles use `archlinux:base-devel@sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722`; build/actions/README.md documents zero-cost invariants |

**Score:** 10/10 truths verified (code level)

---

### Required Artifacts

| Artifact                                        | Expected                                        | Status     | Details                                                                 |
|-------------------------------------------------|-------------------------------------------------|------------|-------------------------------------------------------------------------|
| `cli/config/config.go`                          | XDG dir resolution + DEBATEOS_DIR override      | ✓ VERIFIED | func DebateOSDir() present; DEBATEOS_DIR checked first                  |
| `cli/runner/runner.go`                          | Runner interface + ExecRunner                   | ✓ VERIFIED | type Runner interface + ExecRunner struct                               |
| `cli/runner/fake.go`                            | FakeRunner test double                          | ✓ VERIFIED | type FakeRunner with Calls []string                                     |
| `cli/compose/compose.go`                        | compose subcommand                              | ✓ VERIFIED | func Run exported; connects to resolver                                 |
| `cli/validate/validate.go`                      | validate subcommand                             | ✓ VERIFIED | func Run exported; schema validate + clean-resolve                      |
| `cli/pane/pane.go`                              | pane set/get/list/backup/restore                | ✓ VERIFIED | All 5 verbs; savePane atomic via temp+rename (WR-04 fix)                |
| `cli/pane/age.go`                               | age identity management                         | ✓ VERIFIED | LoadOrCreateIdentity 0600; EncryptFile/DecryptFile via filippo.io/age   |
| `cli/build/build.go`                            | build subcommand + pane merge (CR-03 fix)       | ✓ VERIFIED | loadPaneAssets() reads from config.DebateOSDir(); WR-05, WR-07 fixed   |
| `cli/build/inject.go`                           | WriteInjectionTar + CR-04 fix                   | ✓ VERIFIED | tw.Close() explicit with error check; time.Unix(0,0) deterministic Created |
| `cmd/debateos/main.go`                          | subcommand dispatch                             | ✓ VERIFIED | switch on os.Args[1]; compose/validate/build/pane all wired             |
| `build/docker/Dockerfile`                       | multi-stage; golang pinned (WR-01 fix); arch pinned | ✓ VERIFIED | golang:1.24@sha256:d2d2bc1... builder; archlinux:base-devel@sha256:dd60... runtime |
| `build/docker/entrypoint.sh`                    | array SKIP_ARGS (WR-02 fix); /speech:ro        | ✓ VERIFIED | SKIP_ARGS=() array pattern; SKIP_ISO env var                            |
| `.github/workflows/build-speech.yml`            | thin caller → reusable-build-speech.yml        | ✓ VERIFIED | uses: ./.github/workflows/reusable-build-speech.yml                    |
| `.github/workflows/reusable-build-speech.yml`  | reusable workflow with workflow_call            | ✓ VERIFIED | on: workflow_call; container: ghcr.io/mikl0s/debateos:latest; array SKIP_ARGS (WR-02 fix) |
| `build/actions/build-speech.yml`                | documentation copy of reusable workflow         | ✓ VERIFIED | Substantively identical to reusable-build-speech.yml                   |
| `scripts/determinism-test.sh`                   | SOURCE_DATE_EPOCH + gzip -n (CR-02 fix)         | ✓ VERIFIED | export SOURCE_DATE_EPOCH L138; gzip -n both tar calls L172+179          |
| `scripts/secret-free-check.sh`                  | asserts private files absent from profile       | ✓ VERIFIED | Checks pane.yaml/identity.age/private-injection.tar; genuinely asserts (not echo-only) |
| `scripts/check-coverage.sh`                     | resolver/ ≥ 90%, cli/ ≥ 85%                    | ✓ VERIFIED | Two gates with numeric comparison via awk; hard fail on either gate     |
| `build/actions/README.md`                       | zero-cost walkthrough; deferred items documented | ✓ VERIFIED | Step-by-step fork-and-build; deferred ISO + cross-repo run documented   |

---

### Key Link Verification

| From                                  | To                                              | Via                                                        | Status     | Details                                                   |
|---------------------------------------|-------------------------------------------------|------------------------------------------------------------|------------|-----------------------------------------------------------|
| `cmd/debateos/main.go`                | cli/compose, cli/validate, cli/build, cli/pane  | switch on os.Args[1] calling pkg.Run                       | ✓ WIRED    | All 4 subcommands present                                 |
| `cli/build/build.go`                  | cli/internal/loader                             | loader.ResolveDir(speechDir)                               | ✓ WIRED    | Step 1 of build pipeline                                  |
| `cli/build/build.go`                  | config.DebateOSDir() → loadPaneAssets           | reads pane.yaml from config dir                            | ✓ WIRED    | CR-03 fix: loadPaneAssets() on L187; tested live          |
| `cli/build/build.go`                  | cli/build/inject.go WriteInjectionTar           | outDir, paneAssets passed to WriteInjectionTar L192        | ✓ WIRED    | tar written to outDir, not arch-profile                   |
| `cli/build/build.go`                  | docker via Runner (WR-07: :ro mount)            | runner.Run("docker", dockerArgs...) with :/speech:ro       | ✓ WIRED    | L160 `speechDir + ":/speech:ro"`                          |
| `.github/workflows/build-speech.yml`  | .github/workflows/reusable-build-speech.yml     | uses: ./.github/workflows/reusable-build-speech.yml        | ✓ WIRED    | CR-01 fix confirmed; not self-referential                 |
| `reusable-build-speech.yml`           | ghcr.io/mikl0s/debateos:latest                 | container: image:                                          | ✓ WIRED    | Same image as build/docker/Dockerfile output              |
| `determinism-test.sh`                 | SOURCE_DATE_EPOCH → gzip -n                    | export SOURCE_DATE_EPOCH + pipe gzip -n                    | ✓ WIRED    | CR-02 fix verified in script L138 + L172,179              |
| `archlinux digest`                    | build/docker/Dockerfile ↔ translators/arch/Dockerfile | same sha256:dd60dfcca...                           | ✓ WIRED    | Both use identical digest                                 |

---

### Data-Flow Trace (Level 4)

| Artifact              | Data Variable  | Source                                     | Produces Real Data      | Status      |
|-----------------------|----------------|--------------------------------------------|-------------------------|-------------|
| `cli/build/build.go`  | paneAssets     | loadPaneAssets() → os.ReadFile(pane.yaml)  | Yes — live test confirms | ✓ FLOWING  |
| `cli/build/inject.go` | assets[]       | Passed from loadPaneAssets                 | Yes — tar entries verified | ✓ FLOWING |
| `determinism-test.sh` | arch-profile/  | debateos build --skip-iso (two runs)       | Yes — sha256 identical  | ✓ FLOWING   |

---

### Behavioral Spot-Checks

| Behavior                                     | Command                                                                 | Result                                          | Status   |
|----------------------------------------------|-------------------------------------------------------------------------|-------------------------------------------------|----------|
| go build succeeds                            | `go build ./...`                                                        | exit 0, no output                               | ✓ PASS   |
| All Go tests pass                            | `go test ./... -count=1`                                                | 11 packages pass, 0 failures                    | ✓ PASS   |
| Determinism gate                             | `bash scripts/determinism-test.sh`                                      | sha256 identical: 99db28900f05b123...           | ✓ PASS   |
| Secret-free gate                             | `bash scripts/secret-free-check.sh`                                     | exit 0: all 3 patterns absent                   | ✓ PASS   |
| Coverage gate                                | `bash scripts/check-coverage.sh`                                        | resolver/ 93.5% ≥ 90%; cli/ 85.4% ≥ 85%        | ✓ PASS   |
| --dry-run makes no subprocess calls          | `debateos build --dry-run` + TestBuildDryRun                            | 0 Runner calls; plan printed with epoch + argvs | ✓ PASS   |
| --skip-iso stops before docker               | `debateos build --skip-iso`                                             | profile emitted; no docker call                 | ✓ PASS   |
| Private pane → injection tar (CR-03)         | DEBATEOS_DIR test run with pane.yaml containing 2 entries               | etc/debateos/ssh_authorized_key + secret_token in tar | ✓ PASS |
| private-injection.tar NOT in arch-profile    | find arch-profile/ -name private-injection.tar                          | empty (PRIV-01 invariant holds)                 | ✓ PASS   |
| pane.yaml NOT in arch-profile                | find arch-profile/ -name pane.yaml                                      | empty (PRIV-01 invariant holds)                 | ✓ PASS   |
| Full ISO via docker run                      | N/A — mkarchiso requires devtmpfs (Proxmox restriction)                 | DEFERRED                                        | ? SKIP   |
| Live cross-repo GitHub Actions run           | N/A — requires fork + CI minutes                                        | DEFERRED                                        | ? SKIP   |

---

### Probe Execution

No probe scripts found under `scripts/tests/probe-*.sh` — not applicable for this phase (gate scripts used instead; all gate scripts verified above).

---

### Requirements Coverage

| Requirement | Source Plan     | Description                                                                 | Status         | Evidence                                                                      |
|-------------|-----------------|-----------------------------------------------------------------------------|----------------|-------------------------------------------------------------------------------|
| CLI-01      | 03-01-PLAN.md   | compose/validate/build/pane subcommands, DEBATEOS_DIR, Runner interface     | ✓ SATISFIED    | All 4 subcommands work; config.DebateOSDir() XDG+env; FakeRunner wired        |
| CLI-02      | 03-02-PLAN.md   | pane.yaml + identity.age created 0600; backup/restore via age               | ✓ SATISFIED    | TestPanePermissions passes; 0600 on create + re-check on existing (WR-03)    |
| BLD-01      | 03-03/04-PLAN.md | Omarchy speech builds via local Docker (speech mount → ISO out)             | PARTIAL-DEFERRED | --skip-iso path fully verified; full docker run to ISO deferred (devtmpfs)  |
| BLD-02      | 03-04-PLAN.md   | Published reusable GitHub Actions workflow; same image both channels        | ✓ SATISFIED    | reusable-build-speech.yml in .github/workflows/; identical image ref         |
| BLD-03      | 03-04-PLAN.md   | Deterministic builds verified by automated tests                            | ✓ SATISFIED    | determinism-test.sh passes; SOURCE_DATE_EPOCH + gzip -n both present         |
| BLD-04      | 03-04-PLAN.md   | Zero cost, no central service in build path                                 | ✓ SATISFIED    | build/actions/README.md documents invariants; reusable workflow runs on caller's minutes |
| PRIV-01     | 03-02/04-PLAN.md | Secrets/private pane never in shared artifacts; first-boot injection; key mgmt | ✓ SATISFIED | Live test: private entries in tar only, not profile; inject.go T-03-TRAV guard |

---

### Anti-Patterns Found

| File                        | Line    | Pattern                                   | Severity  | Impact                                           |
|-----------------------------|---------|-------------------------------------------|-----------|--------------------------------------------------|
| `scripts/determinism-test.sh` | 72    | `XXXXXX` in mktemp template               | ℹ️ Info   | Shell-required mktemp syntax; not a debt marker  |
| `scripts/secret-free-check.sh` | 69   | `XXXXXX` in mktemp template               | ℹ️ Info   | Shell-required mktemp syntax; not a debt marker  |
| `scripts/check-coverage.sh`   | 35    | `XXXXXX` in mktemp template               | ℹ️ Info   | Shell-required mktemp syntax; not a debt marker  |

No TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers found in any phase-modified file. No stub implementations. No empty handlers. No hardcoded empty returns that would suppress real data.

**One notable observation (not a gap):** The `debateos` binary shipped in the repo root was compiled before the CR-03/CR-04/IN-03 fixes landed, so it showed stale behavior (wall-clock Created timestamp, empty pane assets) in the first test pass. After `go build -o debateos ./cmd/debateos`, the rebuilt binary passed all checks. The gate scripts (determinism-test.sh, secret-free-check.sh) auto-rebuild the binary when the repo root binary is absent; they will use the stale one if present. This is not a code gap — the source code is correct — but the stale binary in the repo root (if committed) could mislead developers running scripts without first rebuilding.

---

### Critical Fix Verification (CR-01 through CR-04)

| Fix  | Description                                                     | Expected Evidence                                                | Status     |
|------|-----------------------------------------------------------------|------------------------------------------------------------------|------------|
| CR-01 | build-speech.yml must NOT be self-referential                  | calls `.github/workflows/reusable-build-speech.yml` (distinct file) | ✓ CONFIRMED |
| CR-02 | determinism-test.sh: export SOURCE_DATE_EPOCH + gzip -n        | L138 `export SOURCE_DATE_EPOCH="${EPOCH}"` + L172/179 `\| gzip -n` | ✓ CONFIRMED |
| CR-03 | `debateos build` merges private pane from config dir            | loadPaneAssets() in build.go L187; live test shows entries in tar | ✓ CONFIRMED |
| CR-04 | tw.Close() error not swallowed                                  | inject.go L187: explicit `if err := tw.Close()` with error return | ✓ CONFIRMED |

---

### Human Verification Required

#### 1. Full ISO Build (BLD-01 full path)

**Test:** On a bare-metal Linux host or VM with devtmpfs access:
```bash
docker run --rm \
  -v "$(pwd)/examples/omarchy:/speech:ro" \
  -v "$(pwd)/out:/out" \
  ghcr.io/mikl0s/debateos:latest
```

**Expected:** `out/*.iso` produced; image boots unattended into an Omarchy-equivalent Arch system with zero install-time prompts

**Why human:** mkarchiso requires kernel devtmpfs, blocked on this Proxmox VE host

---

#### 2. Live Cross-Repo GitHub Actions Run (BLD-02 full path)

**Test:**
1. Fork `mikl0s/DebateOS` to your GitHub account
2. Create `.github/workflows/test-cross-repo.yml` with:
   ```yaml
   jobs:
     build:
       uses: mikl0s/DebateOS/.github/workflows/reusable-build-speech.yml@main
       with:
         speech-dir: examples/omarchy
         profile: vanilla-arch
         skip-iso: true
   ```
3. Push and observe the Actions tab

**Expected:** Job appears and runs on the fork's runner (not in mikl0s/DebateOS); artifact uploaded to fork's run summary with arch-profile/ tree and resolved.json

**Why human:** Requires GitHub Actions CI minutes on a fork; not testable programmatically on this host

---

### Gaps Summary

No code gaps found. All 10 must-have truths are verified at the code level. The two human verification items are environment-blocked deferrals documented in 03-VALIDATION.md Manual-Only Verifications table — they represent host/infrastructure limitations, not missing implementation.

The phase goal is achieved in the codebase: the CLI subcommands work, the private pane reaches the injection tar and never enters the profile tree, builds are deterministic (gate scripts genuinely assert sha256 equality), secrets never appear in shared artifacts, and both build channels reference the same Docker image. The CR-01 through CR-04 critical fixes from the code review are all confirmed in the source.

---

_Verified: 2026-06-13T13:10:00Z_
_Verifier: Claude (gsd-verifier)_
