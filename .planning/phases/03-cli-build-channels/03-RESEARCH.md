# Phase 3: CLI & Build Channels - Research

**Researched:** 2026-06-13
**Domain:** Go CLI design, age encryption, GitHub Actions reusable workflows, deterministic builds, multi-stage Docker
**Confidence:** HIGH (core findings verified via live tool calls against this host and the Go module proxy)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**CLI Design (CLI-01/CLI-02, D7)**
- stdlib `flag` + small subcommand dispatch in `cli/` (package main at `cmd/debateos/`, logic in `cli/` packages) — module stays minimal-deps (cobra rejected).
- Speech home: XDG `~/.config/debateos/` (override with `--dir` and `DEBATEOS_DIR` for tests): `speech.yaml` (public panes) + `pane.yaml` (private pane, created 0600).
- `debateos compose` — assemble/edit the speech and print a resolution preview with full explanations (uses resolve.Resolve + Explanation).
- `debateos validate` — parse + schema-validate the speech and all referenced documents; clean-resolve check; non-zero exit on any failure (CI-friendly).
- `debateos build` — resolve → write canonical resolved.json → invoke Arch translator → invoke Docker build (channel 1). `--dry-run` stops after emitting build plan. `--skip-iso` stops after profile emission.
- `debateos pane` — `get/set/list/backup/restore` for the private pane. All external invocations (docker, git, translator subprocess) go through a Runner interface with a recording fake for tests.
- cmd/resolve-json (Phase 2 seed) is absorbed/kept as the plumbing the CLI reuses; keep it working.

**Private Pane & Secrets (PRIV-01, D16)**
- `pane.yaml` lives only in the XDG dir, 0600, never copied into shared artifacts; resolution merges it locally.
- Backup: `debateos pane backup` encrypts pane.yaml with age (filippo.io/age Go library — one new dep) to `pane.yaml.age` and commits/pushes to user's own private Git repo; `pane restore` decrypts.
- Age identity stored at `~/.config/debateos/identity.age` (0600), generated on first use; recovery documented (lose the identity = lose the backup, by design).
- First-boot injection: `debateos build` emits `private-injection.tar` LOCALLY (next to ISO, never inside it); installer's first-boot unit looks for artifact on removable media at first boot.
- Key-management: age X25519 identities, local-only, no escrow, no central service.

**Build Channels (BLD-01..04, D11)**
- ONE image `build/docker/Dockerfile`: digest-pinned archlinux:base-devel + archiso + debateos CLI binary + translators.
- Actions channel: `build/actions/build-speech.yml` reusable workflow (workflow_call) + `.github/workflows/build-speech.yml` thin caller; uses the SAME image; artifact upload of ISO.
- Determinism (BLD-03): SOURCE_DATE_EPOCH from resolved-speech sha256. Automated gate `scripts/determinism-test.sh`: run resolve+translate TWICE → deterministic tar → sha256 compare → MUST be identical.

**TDD (D19)**
- Go table-driven tests per subcommand: exit codes, golden stdout, pane perms (assert 0600), dry-run build plan content, fake-Runner command assertions. RED before GREEN.
- Coverage: extend scripts/check-coverage.sh — cli packages ≥85%, resolver packages stay ≥90%.
- Determinism + secret-free-ISO checks are automated scripts.

### Claude's Discretion
- Internal package layout under cli/, flag spelling, exact golden formats, age identity file naming.

### Deferred Ideas (OUT OF SCOPE)
- Embedded Debate UI serving (`debateos compose` web mode) → Phase 5.
- Live GitHub Actions run validation → deferred verification item (document exact steps, don't block phase).
- Registry fetch of remote points → Phase 5; v1 compose uses local paths.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLI-01 | `debateos compose | validate | build | pane` work, wrapping native resolver and invoking translators as subprocesses | Go flag.FlagSet subcommand dispatch verified; Runner interface pattern verified; resolve.Resolve + CanonicalJSON already importable |
| CLI-02 | CLI manages user's speech including private pane in $HOME, with optional backup to user's own private Git repo | os.UserConfigDir() verified; age v1.3.1 API verified; git subprocess via Runner interface |
| BLD-01 | Published Docker image bundles resolver + translators + ISO builders; docker run with speech YAML mounted → ISO locally | Multi-stage Dockerfile pattern documented; CGO_ENABLED=0 static binary verified; existing archlinux digest-pinned image reused |
| BLD-02 | Published reusable GitHub Actions workflow (using SAME image) for fork → commit → build ISO on free-tier CI | workflow_call syntax verified; container: directive documented; actions/upload-artifact@v4 current |
| BLD-03 | Builds are deterministic: identical inputs → identical ISO, SOURCE_DATE_EPOCH from resolved-speech hash, verified by automated tests | GNU tar --sort=name --mtime=@EPOCH --owner=0 --group=0 --numeric-owner flags verified bit-identical; gzip -n verified |
| BLD-04 | End-to-end compose → resolve → build runs at zero hosting cost on both channels with no central service | Architectural pattern clear; user's own Docker + GitHub minutes; no DebateOS infrastructure required |
| PRIV-01 | Secrets and private pane never enter shared/public artifacts; inject at first boot; key-management design finalized | filippo.io/age v1.3.1 X25519 API verified; 0600 permissions testable; private-injection.tar pattern documented |
</phase_requirements>

---

## Summary

Phase 3 builds the `debateos` Go CLI and two zero-cost build channels (local Docker + reusable GitHub Actions) that share one Docker image. The CLI absorbs the Phase 2 `cmd/resolve-json` seed, wraps the existing resolver packages behind a subcommand dispatch, and adds private-pane management with age-based encryption for backup. The Docker image extends the Phase 2 digest-pinned `archlinux:base-devel` image with a multi-stage Go builder stage that produces a static `CGO_ENABLED=0` binary that runs cleanly inside the archlinux runtime layer. Deterministic builds use `SOURCE_DATE_EPOCH` already produced by the Phase 2 translator pipeline, with GNU tar deterministic flags verified bit-identical on this host.

The three key technical risks are: (1) `os.UserConfigDir()` returns an error when `HOME` and `XDG_CONFIG_HOME` are both unset in CI — the CLI must handle this and fall back to `DEBATEOS_DIR` env override; (2) the `debateos pane backup` git subprocess must use the Runner interface (not `os/exec` directly) so tests can record and assert on external calls without network; (3) the secret-free ISO property must be verified by automated grep/listing tests rather than inspection, since the private-injection.tar pattern is easy to accidentally break.

**Primary recommendation:** Implement the CLI as `cmd/debateos/main.go` dispatching to `cli/{compose,validate,build,pane}/` packages, each exposing a `Run(args []string, stdout, stderr io.Writer) int` function that never calls `os.Exit` directly — making table-driven tests trivial. All external subprocess calls route through a `cli/runner.Runner` interface with an `ExecRunner` production impl and a `FakeRunner` test impl.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Subcommand dispatch | CLI binary (cmd/debateos) | — | Entry point; delegates to cli/ packages |
| Speech parsing + resolution | cli/build package → resolver packages | — | Existing resolver.Resolve + CanonicalJSON already handle this; CLI just wires them |
| Private pane management | cli/pane package | XDG filesystem layer | PRIV-01: pane.yaml lives in $HOME only, 0600 |
| age encryption/decryption | cli/pane package | filippo.io/age library | Identity stored at ~/.config/debateos/identity.age |
| External subprocess invocation | cli/runner.Runner interface | ExecRunner (prod) / FakeRunner (test) | All docker/git/translate calls via interface for testability |
| Docker image (channel 1) | build/docker/Dockerfile | Multi-stage Go builder | Static binary baked into archlinux runtime stage |
| GitHub Actions (channel 2) | build/actions/build-speech.yml | .github/workflows/build-speech.yml | workflow_call reusable; thin caller in this repo |
| Determinism gate | scripts/determinism-test.sh | SOURCE_DATE_EPOCH from manifest.py | sha256 of resolved-speech → EPOCH; double-run tar compare |
| Secret-free ISO assertion | scripts/determinism-test.sh (or separate) | Automated grep/listing test | PRIV-01 verification: grep for pane.yaml in ISO profile tree |
| Coverage gate | scripts/check-coverage.sh | go test -coverprofile | Extend threshold to cli/ ≥85% on top of resolver/ ≥90% |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| filippo.io/age | v1.3.1 | X25519 identity generation, encrypt/decrypt pane backup | Only dep added this phase; BSD-3-Clause; audited; Go module proxy verified; no CVE on X25519 path |
| filippo.io/hpke | v0.4.0 | Transitive dep of age | Auto-pulled |
| golang.org/x/crypto | v0.45.0 | Transitive dep of age (X25519 primitives) | Auto-pulled |
| stdlib flag | go 1.24 built-in | Subcommand dispatch (flag.FlagSet per subcommand) | Locked decision D7; no cobra |
| stdlib os | go 1.24 built-in | os.UserConfigDir(), os.OpenFile 0600, os.MkdirAll | Standard XDG + file permission handling |
| stdlib os/exec | go 1.24 built-in | ExecRunner production implementation | Behind Runner interface; never called directly in cli/ logic |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| go.yaml.in/yaml/v3 | v3.0.4 | Already in go.mod; YAML parse for pane.yaml/speech.yaml | Already used by resolver; no new dep |
| github.com/santhosh-tekuri/jsonschema/v6 | v6.0.2 | Already in go.mod; used by validate subcommand | Already used by resolver/parse |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib flag | cobra/urfave-cli | Cobra is rejected (D7); minimal deps is a project invariant |
| filippo.io/age | gpg, age-go alternatives | age is the canonical Go age implementation by the age spec author; GPG is heavy; locked decision D16 |
| os/exec directly | Runner interface + FakeRunner | Direct exec makes tests require real binaries; Runner enables recording fakes — TDD requirement D19 |

**Installation:**
```bash
go get filippo.io/age@v1.3.1
```
(All other deps already in go.mod.)

**Version verification (completed):**
- `filippo.io/age v1.3.1` — confirmed via Go module proxy `proxy.golang.org/filippo.io/age/@latest`: `{"Version":"v1.3.1","Time":"2025-12-28T12:00:23Z"}` [VERIFIED: Go module proxy]
- `go run /tmp/test-age.go` — GenerateX25519Identity, Encrypt, Decrypt round-trip confirmed working [VERIFIED: live execution]

---

## Package Legitimacy Audit

> Go ecosystem; `filippo.io/age` is not an npm package. Legitimacy checked via Go module proxy (the authoritative source for Go modules) and live API execution.

| Package | Registry | Age | Source Repo | Verdict | Disposition |
|---------|----------|-----|-------------|---------|-------------|
| filippo.io/age | Go module proxy (pkg.go.dev) | ~5 yrs (v1.0.0 2021) | github.com/FiloSottile/age | OK | Approved |
| filippo.io/hpke | Go module proxy | Transitive dep | github.com/FiloSottile/age (same author) | OK | Approved — auto-pulled |

**Verification method:** `curl -s https://proxy.golang.org/filippo.io/age/@latest` returned canonical version info; `go run /tmp/test-age.go` demonstrated full encrypt/decrypt cycle. Package authored by Filippo Valsorda (Google/Go security team). [VERIFIED: Go module proxy]

**CVE status:** GHSA-32gq-x56h-299c (plugin path traversal, fixed v1.2.1) does NOT affect X25519 encrypt/decrypt API. DebateOS uses no plugin system. v1.3.1 is clean for the DebateOS use case. [CITED: advisories.gitlab.com/pkg/golang/filippo.io/age/GHSA-32gq-x56h-299c]

**Packages removed due to SLOP verdict:** none
**Packages flagged as suspicious SUS:** none

---

## Architecture Patterns

### System Architecture Diagram

```
User invokes: debateos <subcommand>
         │
         ▼
cmd/debateos/main.go         ← subcommand dispatch (flag.NewFlagSet per subcommand)
         │
    ┌────┴────────────────────────────────────┐
    │                                         │
    ▼                                         ▼
cli/compose/, cli/validate/           cli/build/, cli/pane/
    │                                         │
    ▼                                         ▼
resolver packages                    cli/runner.Runner interface
(resolve.Resolve, CanonicalJSON,         │           │
 parse.ParseSpeech, etc.)           ExecRunner   FakeRunner (tests)
                                         │
                              ┌──────────┴──────────────┐
                              ▼                          ▼
                   translators/arch/translate        docker run
                   (subprocess: argv-stable           debateos:latest
                    contract from Phase 2)
                              │
                              ▼
                   Arch ISO profile tree
                   (--skip-iso stops here)
                              │
                              ▼
                   docker build (channel 1)
                   OR
                   GitHub Actions (channel 2)
                   ┌─────────────────────────────┐
                   │ build/actions/build-speech.yml │
                   │ workflow_call reusable         │
                   │ container: debateos:latest     │
                   │ artifact: ISO uploaded v4      │
                   └─────────────────────────────────┘

Private pane flow (PRIV-01):
  ~/.config/debateos/pane.yaml  ──►  resolution (local merge only)
  pane backup: age encrypt → pane.yaml.age → git push (user's private repo)
  build: emits private-injection.tar (LOCAL, never in ISO)
         └── first-boot unit reads from USB at install time
```

### Recommended Project Structure

```
cmd/
├── debateos/
│   └── main.go              # subcommand dispatch only; no business logic
├── resolve-json/
│   └── main.go              # existing Phase 2 seed — keep working
cli/
├── runner/
│   ├── runner.go            # Runner interface + ExecRunner
│   └── fake.go              # FakeRunner for tests
├── compose/
│   ├── compose.go           # Run(args, stdout, stderr) int
│   └── compose_test.go
├── validate/
│   ├── validate.go
│   └── validate_test.go
├── build/
│   ├── build.go             # --dry-run, --skip-iso flags
│   └── build_test.go
├── pane/
│   ├── pane.go              # get/set/list/backup/restore
│   ├── age.go               # identity generation + encrypt/decrypt
│   └── pane_test.go
└── config/
    └── config.go            # XDG dir resolution + DEBATEOS_DIR override
build/
├── docker/
│   ├── Dockerfile           # multi-stage: golang builder + archlinux runtime
│   └── entrypoint.sh        # /speech mounted, /out for ISO
└── actions/
    ├── build-speech.yml     # reusable workflow (workflow_call)
    └── README.md            # template-repo usage documentation
.github/
└── workflows/
    └── build-speech.yml     # thin caller workflow for this repo
scripts/
├── check-coverage.sh        # extend: cli ≥85% + resolver ≥90%
└── determinism-test.sh      # NEW: double-run → deterministic tar → sha256 compare
```

### Pattern 1: Subcommand Dispatch with Testable Exit Codes

**What:** Each subcommand is a `Run(args []string, stdout, stderr io.Writer) int` function that returns an exit code. The `main()` function calls `os.Exit(Run(...))`. Tests call `Run()` directly — no `os.Exit` in test paths.

**When to use:** All four subcommands (compose, validate, build, pane).

**Example:**
```go
// Source: verified live execution /tmp/test-flagset.go + /tmp/test-exitcode.go

// cli/validate/validate.go
package validate

import (
    "flag"
    "fmt"
    "io"
)

func Run(args []string, stdout, stderr io.Writer) int {
    fs := flag.NewFlagSet("validate", flag.ContinueOnError)
    fs.SetOutput(stderr)
    dir := fs.String("dir", "", "speech directory (overrides DEBATEOS_DIR)")
    if err := fs.Parse(args); err != nil {
        return 1
    }
    _ = dir
    // ... validation logic ...
    fmt.Fprintln(stdout, "validate: OK")
    return 0
}

// cmd/debateos/main.go
package main

import (
    "os"
    "github.com/mikl0s/debateos/cli/validate"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "usage: debateos <command>")
        os.Exit(1)
    }
    switch os.Args[1] {
    case "validate":
        os.Exit(validate.Run(os.Args[2:], os.Stdout, os.Stderr))
    // ... other subcommands ...
    }
}

// cli/validate/validate_test.go
func TestValidateNoArgs(t *testing.T) {
    var out, errOut bytes.Buffer
    code := validate.Run([]string{}, &out, &errOut)
    if code == 0 {
        t.Fatal("expected non-zero exit")
    }
}
```

### Pattern 2: Runner Interface for External Subprocess Calls

**What:** All invocations of `docker`, `git`, `translators/arch/translate`, and `sha256sum` go through a `Runner` interface. Production uses `ExecRunner` (os/exec). Tests use `FakeRunner` that records calls.

**When to use:** Any CLI package that invokes external binaries.

**Example:**
```go
// Source: verified live execution /tmp/test-runner.go

// cli/runner/runner.go
package runner

import "os/exec"

type Runner interface {
    Run(name string, args ...string) error
    Output(name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(name string, args ...string) error {
    return exec.Command(name, args...).Run()
}

func (ExecRunner) Output(name string, args ...string) ([]byte, error) {
    return exec.Command(name, args...).Output()
}

// cli/runner/fake.go
type FakeRunner struct {
    Calls  []string
    Err    error
    Outputs map[string][]byte
}

func (f *FakeRunner) Run(name string, args ...string) error {
    f.Calls = append(f.Calls, name+" "+strings.Join(args, " "))
    return f.Err
}

func (f *FakeRunner) Output(name string, args ...string) ([]byte, error) {
    key := name + " " + strings.Join(args, " ")
    f.Calls = append(f.Calls, key)
    return f.Outputs[key], f.Err
}
```

### Pattern 3: age X25519 Identity Generation and Encrypt/Decrypt

**What:** Generate an X25519 identity once (stored 0600 at `~/.config/debateos/identity.age`), use it to encrypt `pane.yaml` → `pane.yaml.age` for backup, decrypt on restore.

**When to use:** `debateos pane backup` and `debateos pane restore`.

**Example:**
```go
// Source: verified live execution /tmp/test-age.go + /tmp/test-age-parse.go
// filippo.io/age v1.3.1 [VERIFIED: Go module proxy + live execution]

import "filippo.io/age"

// Generate identity (first use)
identity, err := age.GenerateX25519Identity()
// identity.String() → "AGE-SECRET-KEY-1..." (save this 0600 to disk)
// identity.Recipient().String() → "age1..." (public key, safe to log)

// Save identity to disk
os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0600)

// Load identity from disk
data, _ := os.ReadFile(identityPath)
identity, err := age.ParseX25519Identity(strings.TrimSpace(string(data)))

// Encrypt pane.yaml → pane.yaml.age
var encrypted bytes.Buffer
w, err := age.Encrypt(&encrypted, identity.Recipient())
// write pane content to w
w.Close()
os.WriteFile(backupPath, encrypted.Bytes(), 0600)

// Decrypt pane.yaml.age → pane.yaml
f, _ := os.Open(backupPath)
r, err := age.Decrypt(f, identity)
plaintext, _ := io.ReadAll(r)
```

### Pattern 4: GitHub Actions Reusable Workflow (workflow_call)

**What:** Define `build/actions/build-speech.yml` as the reusable workflow. Users call it from their own repo with `uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main`. The SAME Docker image that powers the local channel is specified via `container:` directive on the build job.

**When to use:** BLD-02 implementation; template-repo pattern for user forks.

**Example:**
```yaml
# Source: docs.github.com/actions/sharing-automations/reusing-workflows [CITED]
# build/actions/build-speech.yml  (this lives in the DebateOS repo)
on:
  workflow_call:
    inputs:
      speech-dir:
        required: true
        type: string
        description: Path to the speech directory in the caller's repo
      profile:
        required: false
        type: string
        default: vanilla-arch
    secrets:
      # No secrets needed for public-pane builds; reserved for future private-pane CI
      PANE_AGE_KEY:
        required: false

jobs:
  build:
    runs-on: ubuntu-latest
    container:
      image: ghcr.io/mikl0s/debateos:latest
    steps:
      - uses: actions/checkout@v4
      - name: Build ISO
        run: |
          debateos build --speech-dir "${{ inputs.speech-dir }}" \
                         --profile "${{ inputs.profile }}" \
                         --skip-iso  # or full build if mkarchiso available
      - uses: actions/upload-artifact@v4
        with:
          name: debateos-iso
          path: out/*.iso

# .github/workflows/build-speech.yml  (thin caller in this repo for CI testing)
on:
  push:
    paths:
      - 'examples/**'
      - 'build/**'
jobs:
  build-example:
    uses: ./.github/workflows/build-speech.yml  # same-repo call
    with:
      speech-dir: examples/omarchy
      profile: vanilla-arch
```

### Pattern 5: Deterministic Tar Gate

**What:** `scripts/determinism-test.sh` runs the full resolve+translate pipeline twice into independent output directories, then produces deterministic tarballs from each and compares sha256 checksums.

**When to use:** Wave/phase verification gate.

**Example:**
```bash
# Source: verified live execution on this host (GNU tar 1.35, gzip 1.12) [VERIFIED: live execution]

EPOCH="${SOURCE_DATE_EPOCH:-$(date +%s)}"

tar \
  --sort=name \
  --mtime="@${EPOCH}" \
  --owner=0 --group=0 --numeric-owner \
  --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
  -czf /tmp/profile-run1.tar.gz -C "${OUT_DIR_1}" .

tar \
  --sort=name \
  --mtime="@${EPOCH}" \
  --owner=0 --group=0 --numeric-owner \
  --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
  -czf /tmp/profile-run2.tar.gz -C "${OUT_DIR_2}" .

SHA1=$(sha256sum /tmp/profile-run1.tar.gz | cut -d' ' -f1)
SHA2=$(sha256sum /tmp/profile-run2.tar.gz | cut -d' ' -f1)

if [ "$SHA1" = "$SHA2" ]; then
  echo "DETERMINISM OK: $SHA1"
else
  echo "DETERMINISM FAIL: $SHA1 != $SHA2"
  exit 1
fi
```

### Pattern 6: Multi-Stage Docker Build (Go builder + archlinux runtime)

**What:** A two-stage Dockerfile: stage 1 compiles the static `debateos` binary using `golang:1.24-alpine` (or `FROM golang:1.24` on linux/amd64); stage 2 copies the binary into the existing Phase 2 digest-pinned `archlinux:base-devel` image. The archlinux image already has archiso, python3, and python-yaml.

**When to use:** `build/docker/Dockerfile`.

**Example:**
```dockerfile
# Source: verified via ldd /tmp/test-static (statically linked) [VERIFIED: live execution]
# Stage 1: Build the static debateos binary
FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/debateos ./cmd/debateos

# Stage 2: archlinux runtime (same digest as Phase 2 translator Dockerfile)
FROM archlinux:base-devel@sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722
# archiso already installed from Phase 2 layer
RUN pacman -Sy --noconfirm archiso python-yaml && pacman -Scc --noconfirm
COPY --from=builder /out/debateos /usr/local/bin/debateos
COPY translators/ /debateos/translators/
COPY schemas/ /debateos/schemas/
WORKDIR /build
ENTRYPOINT ["/debateos/build/docker/entrypoint.sh"]
```

**Key property:** `CGO_ENABLED=0` produces a fully statically linked ELF binary (verified: `ldd /tmp/test-static` → "not a dynamic executable"). This runs cleanly in archlinux without glibc version mismatches. [VERIFIED: live execution on this host]

### Anti-Patterns to Avoid

- **Calling `os.Exit` inside subcommand packages:** Prevents table-driven testing. Instead, return int exit code and call `os.Exit` only in `main()`.
- **Calling `os/exec.Command` directly in cli/ logic:** Bypasses the Runner interface, making tests require real external binaries. Route all external calls through `runner.Runner`.
- **Storing pane.yaml world-readable:** Must be created with `os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)` — never `os.WriteFile(path, data, 0644)`. Test: `assert(stat.Mode().Perm() == 0600)`.
- **Embedding private-injection.tar inside the ISO:** The tar must be written next to the ISO on the local filesystem only. Automated grep test must verify it is absent from the ISO profile tree.
- **Using `gzip` without `-n` for deterministic output:** gzip by default embeds current timestamp in the header, breaking bit-identical output. Always use `gzip -n` or rely on `tar -czf` with `--mtime` (which internally sets gzip header correctly when combined with pax options). [VERIFIED: live execution — `gzip -n` produces identical sha256 across runs]
- **Not testing with `HOME` unset:** `os.UserConfigDir()` returns an error when both `$HOME` and `$XDG_CONFIG_HOME` are unset. The CLI must check for `DEBATEOS_DIR` env override first, and fail with a clear message if neither is set. [VERIFIED: live go test on this host]

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| X25519 key generation | Custom ECDH key gen | `filippo.io/age.GenerateX25519Identity()` | Age spec compliance; stream format; correct random sourcing |
| Age file encryption | Custom XChaCha20/Poly1305 | `age.Encrypt()` / `age.Decrypt()` | Correct chunk size, HMAC, header format; interoperable with age CLI |
| Deterministic archive | Custom tar writer | GNU tar with `--sort=name --mtime=@EPOCH --owner=0 --group=0 --numeric-owner --pax-option=...` | PAX header timestamp fields (atime, ctime) are the most common reproducibility bug |
| XDG directory resolution | Custom HOME parsing | `os.UserConfigDir()` (stdlib) | Handles XDG_CONFIG_HOME, fallback to ~/.config on Linux, correct on macOS |

**Key insight:** The two hardest correctness problems in this phase are age file format compliance (wrong chunk size or HMAC = unrecoverable backup) and reproducible tar archives (stale PAX metadata fields silently break determinism). Neither is worth hand-rolling.

---

## Common Pitfalls

### Pitfall 1: os.UserConfigDir Fails in CI

**What goes wrong:** When CI runner does not set `HOME` (some minimal containers), `os.UserConfigDir()` returns `("", error)`. The CLI panics or uses an empty path.

**Why it happens:** Go's stdlib implementation: on Linux, checks `$XDG_CONFIG_HOME` then `$HOME/.config`. If neither is set, returns error.

**How to avoid:** Check `DEBATEOS_DIR` env var first. If set, use it. If unset, call `os.UserConfigDir()`. If that errors, fail with a clear message: `"set DEBATEOS_DIR or HOME environment variable"`. Tests always pass `DEBATEOS_DIR` via `t.Setenv("DEBATEOS_DIR", t.TempDir())`.

**Warning signs:** Silent empty path used for config → writes to wrong location or panics. Test with `HOME=""` `XDG_CONFIG_HOME=""` to expose. [VERIFIED: live execution]

### Pitfall 2: PAX Header Timestamps Break Determinism

**What goes wrong:** Two identical archive runs produce different sha256 because atime/ctime PAX headers differ.

**Why it happens:** GNU tar includes file access timestamps in PAX extended headers by default, even when `--mtime` is set for the main header.

**How to avoid:** Use `--pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime` to strip those fields. Verified bit-identical output on this host. [VERIFIED: live execution]

**Warning signs:** Determinism script fails intermittently; sha256 of two identical-content archives differs.

### Pitfall 3: private-injection.tar Accidentally Inside the ISO

**What goes wrong:** The first-boot injection tar is accidentally included in the archiso profile airootfs, making secrets visible in the shared ISO.

**Why it happens:** Build script writes private-injection.tar to a directory that gets swept into the airootfs overlay by the translator.

**How to avoid:** Write private-injection.tar to the same directory as the output ISO (not to any path under the arch-profile/ tree). Automated test: `grep -r "pane.yaml" <arch-profile-dir>` must return non-zero (no match).

**Warning signs:** The arch profile airootfs contains any of: `pane.yaml`, `identity.age`, `private-injection.tar`.

### Pitfall 4: age Identity File Not 0600

**What goes wrong:** `identity.age` is created with default permissions (0644), exposing the private key to other users on the system.

**Why it happens:** Using `os.WriteFile(path, data, 0644)` or creating the file via a temp file with wrong permissions.

**How to avoid:** Use `os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)`. Test: `stat, _ := os.Stat(identityPath); assert(stat.Mode().Perm() == 0600)`. [VERIFIED: pattern in Phase 2 `pane.yaml` design]

**Warning signs:** `ls -la ~/.config/debateos/identity.age` shows `-rw-r--r--` instead of `-rw-------`.

### Pitfall 5: Translator Argv Contract Drift

**What goes wrong:** The CLI calls `translators/arch/translate` with wrong argument order, silently breaking the build pipeline.

**Why it happens:** The translate script has a frozen argv contract (Phase 2): `translate <resolved.json> --opinions <path> --profile <name> --out <dir>`. The CLI must call it exactly.

**How to avoid:** Use the FakeRunner in tests to assert the exact argv: `assert(fake.Calls[0] == "translators/arch/translate resolved.json --opinions opinions/ --profile vanilla-arch --out arch-profile/")`. [CITED: translators/arch/translate — FROZEN ARGV CONTRACT comment]

**Warning signs:** Exit code 1 from translate with "usage:" on stderr; profile output directory empty.

### Pitfall 6: actionlint Not Available on This Host

**What goes wrong:** The plan requires actionlint for YAML validation of GitHub Actions workflows, but `actionlint v1.7.x` requires `go >= 1.25.0` which is not the project's go version. `go install github.com/rhysd/actionlint/cmd/actionlint@latest` downloads go 1.25.11 toolchain automatically.

**Why it happens:** actionlint v1.7.12 raised its Go minimum to 1.25. The project uses go 1.24.0.

**How to avoid:** Use `actionlint@v1.6.27` (last version supporting go 1.24; installs via `go install github.com/rhysd/actionlint/cmd/actionlint@v1.6.27`). Alternative: use `python3 -c "import yaml; yaml.safe_load(open('workflow.yml'))"` for basic YAML syntax validation. `python3` with PyYAML is confirmed available on this host.

**Warning signs:** `go install actionlint@latest` triggers `go: switching to go1.25.11` download mid-install (observed in session). [VERIFIED: live execution on this host]

---

## Code Examples

### age Identity Round-Trip (Verified)

```go
// Source: verified live execution filippo.io/age v1.3.1 [VERIFIED: Go module proxy + live execution]
import "filippo.io/age"

// Generate
id, err := age.GenerateX25519Identity()  // returns *age.X25519Identity
privKeyStr := id.String()                 // "AGE-SECRET-KEY-1..."
pubKeyStr := id.Recipient().String()      // "age1..."

// Save (0600)
os.WriteFile(path, []byte(privKeyStr+"\n"), 0600)

// Load
raw, _ := os.ReadFile(path)
id, err = age.ParseX25519Identity(strings.TrimSpace(string(raw)))

// Encrypt stream
var buf bytes.Buffer
w, _ := age.Encrypt(&buf, id.Recipient())
io.WriteString(w, plaintext)
w.Close()

// Decrypt stream
r, _ := age.Decrypt(&buf, id)
plaintext, _ := io.ReadAll(r)
```

### XDG Dir Resolution with DEBATEOS_DIR Override (Verified)

```go
// Source: verified via os.UserConfigDir() live test [VERIFIED: live execution]
func DebateOSDir() (string, error) {
    if d := os.Getenv("DEBATEOS_DIR"); d != "" {
        return d, nil
    }
    base, err := os.UserConfigDir()
    if err != nil {
        return "", fmt.Errorf("cannot determine config dir: %w\n(set DEBATEOS_DIR or HOME)", err)
    }
    return filepath.Join(base, "debateos"), nil
}

// In tests:
t.Setenv("DEBATEOS_DIR", t.TempDir())
```

### Deterministic Tar (Verified Flags)

```bash
# Source: verified bit-identical output on this host (GNU tar 1.35) [VERIFIED: live execution]
EPOCH=1000000
tar \
  --sort=name \
  --mtime="@${EPOCH}" \
  --owner=0 --group=0 --numeric-owner \
  --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
  -czf profile.tar.gz -C "${PROFILE_DIR}" .
```

### workflow_call Reusable Workflow Skeleton (Cited)

```yaml
# Source: docs.github.com/en/actions/sharing-automations/reusing-workflows [CITED]
# build/actions/build-speech.yml

on:
  workflow_call:
    inputs:
      speech-dir:
        required: true
        type: string
      profile:
        required: false
        type: string
        default: vanilla-arch

jobs:
  build:
    runs-on: ubuntu-latest
    container:
      image: ghcr.io/mikl0s/debateos:latest
    steps:
      - uses: actions/checkout@v4
      - run: debateos build --speech-dir "${{ inputs.speech-dir }}" --skip-iso
      - uses: actions/upload-artifact@v4
        with:
          name: debateos-profile
          path: out/
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| GPG for backup encryption | age (filippo.io/age) | ~2021 | Simpler API; no key-ring; X25519 by default; Go-native |
| `actions/upload-artifact@v3` | `actions/upload-artifact@v4` | 2024 | v3 deprecated; v4 is current stable (v7 exists but very new as of 2026) |
| `tar -czf` without sort/mtime flags | `tar --sort=name --mtime=@EPOCH --owner=0 --group=0 --numeric-owner --pax-option=...` | Reproducible Builds project | PAX header atime/ctime were the last remaining non-determinism source |
| cobra for Go CLI | stdlib `flag` + manual dispatch | Project decision D7 | Minimal deps; easier testing; sufficient for 4 subcommands |

**Deprecated/outdated:**
- `age.GenerateHybridIdentity()`: post-quantum hybrid keys exist in v1.3.1 but are experimental; X25519 is the locked choice for DebateOS.
- `actions/upload-artifact@v2,v3`: deprecated by GitHub; use v4.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `actions/upload-artifact@v4` is current stable for GitHub Actions artifact upload | Standard Stack / GHA Pattern | Minor: v4 may be superseded; update version pin in workflow |
| A2 | The reusable workflow in `mikl0s/DebateOS` (public repo) can be called cross-repo by any user's fork | GHA Pattern | Medium: if GitHub changes cross-repo workflow_call access rules for public repos, the template-repo pattern breaks; verify with a real fork test (deferred) |
| A3 | `ghcr.io/mikl0s/debateos:latest` will be the published image path | Architecture Diagram | Low: image name is a decision for plan execution; any valid GHCR path works |

**If this table is empty:** All claims in this research were verified or cited — no user confirmation needed. (Not empty: 3 low-risk assumptions documented above.)

---

## Open Questions

1. **Source of SOURCE_DATE_EPOCH in debateos build**
   - What we know: `manifest.py` in the translator derives `SOURCE_DATE_EPOCH` from the resolved-speech sha256. The CLI calls the translator as a subprocess.
   - What's unclear: Does the CLI need to independently compute and export `SOURCE_DATE_EPOCH` before calling the translator, or does the translator always derive it internally from the resolved.json it receives?
   - Recommendation: The translator already derives it from resolved.json via `manifest.py`. The CLI should pass it as an env var to the docker/translate subprocess call for the outer docker build layer. The `scripts/determinism-test.sh` should export the same epoch computed from sha256 of the resolved.json, consistent with how manifest.py computes it.

2. **private-injection.tar format**
   - What we know: `debateos build` emits this file locally. The first-boot unit (generated by the translator in Phase 2) looks for it on mounted removable media.
   - What's unclear: The exact expected format (directory layout inside the tar) is not fully specified in CONTEXT.md. The first-boot unit must know where to find files within the tar.
   - Recommendation: Define the layout in this phase: `private-injection.tar` contains `pane.yaml` at the root, plus any `assets/` files referenced in the private pane. The first-boot unit extracts to `~/.config/debateos/` on first boot.

3. **Container registry authentication for the GitHub Actions workflow**
   - What we know: The `container:` directive in GHA can use `credentials:` for private registries. GHCR public images need no credentials.
   - What's unclear: Will `ghcr.io/mikl0s/debateos:latest` be a public GHCR image (no credentials needed in the workflow)?
   - Recommendation: Plan for public GHCR image (no secrets in the workflow for pulling the image). The workflow only needs `GITHUB_TOKEN` for artifact upload, which is auto-provided.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Building debateos binary | Yes | go 1.24.1 | — |
| Docker | BLD-01 local channel, multi-stage build | Yes | 29.5.3 | — |
| GNU tar | BLD-03 determinism gate | Yes | 1.35 | — |
| gzip | BLD-03 determinism gate | Yes | 1.12 | — |
| git | pane backup/restore subprocess | Yes | 2.43.0 | — |
| sha256sum | determinism gate script | Yes | coreutils 9.4 | — |
| python3 + PyYAML | GHA workflow YAML syntax validation (fallback for actionlint) | Yes | Python 3.12.3 + PyYAML | Primary fallback |
| actionlint | GHA workflow validation (preferred) | No (v1.7.x needs go>=1.25) | — | `actionlint@v1.6.27` via `go install` installs OK; OR python3 yaml.safe_load |
| mkarchiso | Full ISO build (BLD-01 full path) | No (not in PATH; host devtmpfs restriction) | — | --skip-iso path; document full path requires capable host |
| gh CLI | GitHub Actions publishing (deferred) | No | — | Manual push; deferred verification |

**Missing dependencies with no fallback:**
- mkarchiso: documented limitation from Phase 2; `--skip-iso` flag covers this host; full ISO build path is correct and tested in Docker.

**Missing dependencies with fallback:**
- actionlint: install `@v1.6.27` or use python3 PyYAML fallback; both confirmed available.

---

## Validation Architecture

> `workflow.nyquist_validation` not set to false in `.planning/config.json` — validation section required.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (go test) |
| Config file | none — standard go test discovery |
| Quick run command | `go test ./cli/... -count=1` |
| Full suite command | `go test ./... -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CLI-01 | `validate` exits non-zero on bad speech | unit | `go test ./cli/validate/... -run TestValidate -count=1` | No — Wave 0 gap |
| CLI-01 | `validate` exits 0 on omarchy speech | unit | `go test ./cli/validate/... -run TestValidateOmarchy -count=1` | No — Wave 0 gap |
| CLI-01 | `build --dry-run` emits plan, no docker call | unit | `go test ./cli/build/... -run TestBuildDryRun -count=1` | No — Wave 0 gap |
| CLI-01 | `build --skip-iso` calls translate, no docker build | unit | `go test ./cli/build/... -run TestBuildSkipISO -count=1` | No — Wave 0 gap |
| CLI-01 | `compose` prints resolution preview | unit | `go test ./cli/compose/... -run TestCompose -count=1` | No — Wave 0 gap |
| CLI-02 | `pane set/get` round-trip | unit | `go test ./cli/pane/... -run TestPaneSetGet -count=1` | No — Wave 0 gap |
| CLI-02 | `pane backup` creates 0600 pane.yaml | unit | `go test ./cli/pane/... -run TestPanePermissions -count=1` | No — Wave 0 gap |
| CLI-02 | `pane backup/restore` age round-trip | unit | `go test ./cli/pane/... -run TestPaneBackupRestore -count=1` | No — Wave 0 gap |
| CLI-02 | FakeRunner records git push command | unit | `go test ./cli/pane/... -run TestPaneBackupGitPush -count=1` | No — Wave 0 gap |
| PRIV-01 | pane.yaml not in arch-profile tree | script | `bash scripts/determinism-test.sh` | No — Wave 0 gap |
| PRIV-01 | identity.age created 0600 | unit | `go test ./cli/pane/... -run TestIdentityPermissions -count=1` | No — Wave 0 gap |
| BLD-03 | Two resolve+translate runs produce identical tar sha256 | script | `bash scripts/determinism-test.sh` | No — Wave 0 gap |
| BLD-01,BLD-02 | GHA workflow YAML is syntactically valid | script | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/build-speech.yml'))"` | No — Wave 0 gap |

### Coverage Extension

Current threshold for `scripts/check-coverage.sh`:
- `./resolver/...` ≥ 90% (existing; all passing)

Phase 3 extension:
- Add `./cli/...` ≥ 85% to the coverage gate script
- Resolver packages must stay at ≥ 90%

### Sampling Rate

- **Per task commit:** `go test ./cli/... -count=1` (cli packages for the current task)
- **Per wave merge:** `bash scripts/check-coverage.sh` (full coverage gate including cli)
- **Phase gate:** `bash scripts/determinism-test.sh` + `go test ./... -count=1` + python3 GHA YAML validation

### Wave 0 Gaps

- [ ] `cli/runner/runner.go` + `cli/runner/fake.go` — Runner interface and FakeRunner
- [ ] `cli/config/config.go` — XDG dir resolution with DEBATEOS_DIR override
- [ ] `cli/compose/compose.go` + `cli/compose/compose_test.go` — compose subcommand
- [ ] `cli/validate/validate.go` + `cli/validate/validate_test.go` — validate subcommand
- [ ] `cli/build/build.go` + `cli/build/build_test.go` — build subcommand
- [ ] `cli/pane/pane.go` + `cli/pane/age.go` + `cli/pane/pane_test.go` — pane subcommand
- [ ] `cmd/debateos/main.go` — top-level binary entry point
- [ ] `scripts/determinism-test.sh` — double-run sha256 gate
- [ ] `build/docker/Dockerfile` — multi-stage Go builder + archlinux runtime
- [ ] `build/actions/build-speech.yml` — reusable workflow (workflow_call)
- [ ] `.github/workflows/build-speech.yml` — thin caller for this repo
- [ ] `build/actions/README.md` — template-repo usage documentation
- [ ] Framework install: `go get filippo.io/age@v1.3.1` — already done in this session; update go.mod to mark direct dependency

---

## Security Domain

> `security_enforcement` not set to false — required.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No user login; key management is local only |
| V3 Session Management | No | CLI is stateless between invocations |
| V4 Access Control | Partial | File permissions: 0600 for pane.yaml, identity.age; OS enforces |
| V5 Input Validation | Yes | Speech YAML validated by existing jsonschema (santhosh-tekuri/jsonschema/v6); path traversal must be checked for private-injection.tar output path |
| V6 Cryptography | Yes | filippo.io/age X25519 — never hand-roll; identity.age 0600; no plaintext backup |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| private-injection.tar path traversal (output path controlled by user input) | Tampering | Validate output path is absolute and within expected dir before write |
| identity.age world-readable | Information Disclosure | `os.OpenFile` with 0600; test asserts `stat.Mode().Perm() == 0600` |
| pane.yaml included in Docker image build context | Information Disclosure | `.dockerignore` must exclude `~/.config/debateos/`; build context is the repo root, not HOME |
| Subprocess arg injection (translator path, docker run args) | Tampering | Use `exec.Command(binary, args...)` not `sh -c` string interpolation; Runner.Run() takes variadic args not a string |
| Stale archlinux base image | Elevation of Privilege | Same digest-pin policy as Phase 2; re-verify quarterly (documented in Dockerfile comments) |

---

## Sources

### Primary (VERIFIED via live execution)

- Go module proxy `proxy.golang.org/filippo.io/age/@latest` — version v1.3.1 confirmed, publish date 2025-12-28
- Live execution on this host: `go run /tmp/test-age.go` — age GenerateX25519Identity, Encrypt, Decrypt API verified
- Live execution on this host: `go run /tmp/test-age-parse.go` — ParseX25519Identity round-trip verified
- Live execution on this host: GNU tar 1.35 deterministic flags — sha256 bit-identical output verified
- Live execution on this host: `gzip -n` reproducibility — sha256 identical across runs verified
- Live execution on this host: `CGO_ENABLED=0 go build` → `ldd` → "not a dynamic executable" — static binary confirmed
- Live execution on this host: `go run /tmp/test-xdg.go` — `os.UserConfigDir()` returns error when HOME+XDG_CONFIG_HOME both unset
- Live execution on this host: `go run /tmp/test-runner.go` — Runner interface + FakeRunner pattern verified
- Live execution on this host: `go run /tmp/test-flagset.go` + `go run /tmp/test-exitcode.go` — flag.FlagSet subcommand dispatch verified

### Secondary (CITED from official documentation)

- `docs.github.com/en/actions/sharing-automations/reusing-workflows` — workflow_call inputs/secrets syntax, cross-repo calling
- `docs.github.com/en/actions/writing-workflows/choosing-where-your-workflow-runs/running-jobs-in-a-container` — container: directive syntax
- `docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/storing-workflow-data-as-artifacts` — actions/upload-artifact@v4 syntax
- `pkg.go.dev/filippo.io/age` — API surface: GenerateX25519Identity, Encrypt, Decrypt, ParseX25519Identity, ParseRecipients

### Tertiary (CITED from security advisory)

- `advisories.gitlab.com/pkg/golang/filippo.io/age/GHSA-32gq-x56h-299c` — CVE affects plugin API only, not X25519 path; fixed in v1.2.1; v1.3.1 is clean

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — filippo.io/age API verified via live Go execution and module proxy; stdlib flag patterns verified
- Architecture: HIGH — all patterns verified via live execution; existing resolver/translate contracts confirmed from source
- Pitfalls: HIGH — all pitfalls discovered via live verification (os.UserConfigDir test, tar determinism test, actionlint version conflict)
- GitHub Actions: MEDIUM — syntax verified from official docs; cross-repo access behavior tagged ASSUMED pending live CI test

**Research date:** 2026-06-13
**Valid until:** 2026-07-13 (stable domain; filippo.io/age releases infrequently; GHA syntax rarely changes incompatibly)
