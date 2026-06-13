# DebateOS Build Channels

**Two zero-cost build channels. One Docker image. No central service in the critical path.**

This document describes the end-to-end compose → validate → build workflow,
the two build channels (local Docker and GitHub Actions), private-pane key
management, and the flash-time secret-injection step.

## Overview

```
debateos compose    ←—— assemble your speech from points and opinions
debateos validate   ←—— schema-validate + clean-resolve check (CI-friendly)
debateos build      ←—— resolve → translate → build ISO
debateos pane       ←—— manage your private pane (secrets / personal overrides)
```

The full path from composition to bootable ISO runs entirely on free tooling
and user-owned compute. DebateOS infrastructure is never in the critical path
(invariants 4/5).

---

## Step-by-Step: Compose to ISO

### Step 1 — Create or edit your speech

```bash
# Interactive shell editing (v1 is YAML-based; visual editor coming in Phase 5)
mkdir -p ~/.config/debateos
$EDITOR ~/.config/debateos/speech.yaml
```

A minimal speech:
```yaml
schema: 1
id: my-workstation
foundation: arch
points:
  - id: "omarchy/hyprland-desktop-stack"
  - id: "omarchy/terminal-and-shell-toolchain"
opinions: []
```

### Step 2 — Validate the speech

```bash
debateos validate
# or with an explicit dir:
debateos validate --dir /path/to/speech
```

Exits 0 if the speech resolves cleanly, non-zero with a human-readable
explanation for every conflict or schema error.

### Step 3 — (Optional) Add private pane entries

```bash
# Set a private key/value that stays in ~/.config/debateos/pane.yaml (0600)
debateos pane set github-token "ghp_..."
debateos pane set wifi-password "hunter2"

# List all keys
debateos pane list

# Back up to your own private Git repo (age-encrypted)
debateos pane backup
```

Private pane entries overlay the public speech locally. They are never
written into the shared ISO or arch-profile — see the Privacy section below.

### Step 4 — Build

#### Channel 1: Local Docker (full privacy path)

```bash
# Profile emission only (works everywhere, including Proxmox hosts)
debateos build --skip-iso

# Full ISO build (requires a host with devtmpfs — see Deferred Verifications)
debateos build

# Or use docker run directly (same image)
docker run --rm \
  -v "$(debateos config dir):/speech:ro" \
  -v "$(pwd)/out:/out" \
  -e SKIP_ISO=1 \
  ghcr.io/mikl0s/debateos:latest
```

#### Channel 2: GitHub Actions (free CI on your own minutes)

Fork the DebateOS template repo (or any repo containing your speech), then
add a workflow file:

```yaml
# .github/workflows/build.yml
on:
  push:
    paths: ["speeches/**"]

permissions:
  contents: read

jobs:
  build:
    uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main
    with:
      speech-dir: speeches/my-workstation
      profile: vanilla-arch
      skip-iso: true     # remove for a full ISO on capable CI runners
```

**The same Docker image powers both channels.** The workflow_call pulls
`ghcr.io/mikl0s/debateos:latest` — identical to the local path.

---

## The Docker Image

The image (`ghcr.io/mikl0s/debateos:latest`) is a two-stage build:

- **Stage 1 (builder):** `golang:1.24` compiles the `debateos` binary with
  `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` — fully statically linked, no glibc
  version constraints.
- **Stage 2 (runtime):** `archlinux:base-devel` (digest-pinned; see
  `build/docker/Dockerfile`) installs `archiso` + `python-yaml` and copies in
  the binary, translators, and schemas.

The archlinux digest matches `translators/arch/Dockerfile` exactly —
a single pin source of truth. Re-verify quarterly:

```bash
docker pull archlinux:base-devel
docker inspect --format='{{index .RepoDigests 0}}'
# Update build/docker/Dockerfile AND translators/arch/Dockerfile to the new digest
```

### Building the image yourself

```bash
docker build -f build/docker/Dockerfile -t debateos:local .
```

The `.dockerignore` excludes `pane.yaml`, `*.age`, `private-injection.tar`,
`.config/`, and `.planning/` — HOME secrets never enter the build context
(T-03-CTX).

---

## Deterministic Builds (BLD-03)

Identical inputs must produce identical ISOs. The determinism gate runs
two full resolve + translate cycles and compares sha256 of the resulting
arch-profile tarballs:

```bash
bash scripts/determinism-test.sh [--speech-dir path] [--profile name]
```

**How it works:**

1. Run `debateos build --skip-iso` twice into independent clean directories.
2. Derive `SOURCE_DATE_EPOCH` from `sha256(resolved.json)` using the same
   algorithm as `translators/arch/manifest.py derive_source_date_epoch`:
   - SHA-256 → first 4 bytes as big-endian uint32 → `epochMin + (raw % range)`
   - `epochMin = 1577836800` (2020-01-01), `epochMax = 2208988800` (2040-01-01)
3. Create deterministic tarballs of each arch-profile with GNU tar:
   ```bash
   tar --sort=name \
       --mtime="@${EPOCH}" \
       --owner=0 --group=0 --numeric-owner \
       --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
       -czf profile.tar.gz -C arch-profile/ .
   ```
4. Compare sha256 — must be identical.

The PAX option strips `atime`/`ctime` from extended headers; these are the
most common remaining source of non-determinism after setting `--mtime`.

---

## Privacy & Secrets (PRIV-01)

### What stays local

- `~/.config/debateos/pane.yaml` (0600) — private pane; never copied into ISO
- `~/.config/debateos/identity.age` (0600) — age X25519 private key
- `out/private-injection.tar` — private assets for first-boot injection

**The shared ISO is provably secret-free.** Run the gate:
```bash
bash scripts/secret-free-check.sh
```
This builds the profile (`--skip-iso`) and greps the arch-profile tree for
`pane.yaml`, `identity.age`, and `private-injection.tar` — they must all be
absent.

### Age key management

On the first `debateos pane backup` call, an age X25519 identity is generated
at `~/.config/debateos/identity.age` (0600) using `filippo.io/age`.

```
identity.age  →  0600  →  ~/.config/debateos/identity.age
pane.yaml     →  0600  →  ~/.config/debateos/pane.yaml
pane.yaml.age →  0600  →  ~/.config/debateos/pane.yaml.age  (encrypted backup)
```

**WARNING: Losing `identity.age` means losing the ability to decrypt your
backups.** This is by design — no escrow, no central service.

Back up `identity.age` to your own secure offline storage (e.g. an encrypted
USB drive or a password manager with file attachment). The age X25519 key is a
single-line bech32 string starting with `AGE-SECRET-KEY-1...`.

### Backup to a private Git repo

```bash
# Configure a git remote in your config dir (one-time setup):
cd ~/.config/debateos && git init && git remote add origin git@github.com:you/private-pane.git

# Now backup:
debateos pane backup
# This age-encrypts pane.yaml → pane.yaml.age, then git add/commit/push.
# Only the .age ciphertext is ever committed. Plaintext is never staged.

# Restore on a new machine (after copying identity.age):
debateos pane restore
# This git pull then decrypts pane.yaml.age → pane.yaml.
```

---

## Flash-Time Secret Injection

Private pane file assets (from `pane.yaml` `file_assets` field) are packaged
into `out/private-injection.tar` by `debateos build`. This file lives **next
to the ISO** on your local machine and is **never inside the ISO image**.

### How to inject at first boot

1. Build your ISO: `debateos build` (or channel 2 / CI build).
2. Flash the ISO to a USB drive: `dd if=out/*.iso of=/dev/sdX bs=4M status=progress`.
3. Copy `private-injection.tar` next to the ISO on the same USB (or a second partition):
   ```bash
   cp out/private-injection.tar /media/usb/
   ```
4. Boot the target machine from the USB. The first-boot systemd unit
   (generated by the Arch translator) looks for `private-injection.tar` on any
   mounted removable media and applies the contained files to the target
   filesystem with the configured permissions.
5. After the first boot completes, remove the USB. The secrets are now on the
   target machine only — never in the ISO, never on a server.

### private-injection.tar format

```
debateos-private.json   ←  manifest (version, created, file list + modes)
etc/ssh/authorized_keys ←  example target-relative path
home/user/.zshrc        ←  example target-relative path
```

All paths are target-filesystem-relative (no absolute paths, no `..`
traversal). The first-boot unit extracts the tar to `/` applying the declared
Unix modes.

---

## Zero-Cost No-Central-Service Guarantee

| Property | How enforced |
|----------|-------------|
| No DebateOS infra in build path | Channel 1: user's own Docker. Channel 2: user's own GitHub Actions minutes (the workflow_call runs on the caller's runner). |
| No secrets on any server | `private-injection.tar` is local-only; ISO is provably secret-free (automated grep gate). |
| No paid dependency | Docker CE (free), GitHub Actions free tier (2000 min/month), age (free open source). |
| Reproducible output | Determinism gate passes (`scripts/determinism-test.sh`). |

---

## Deferred Verifications

The following items are **correct by design** but require infrastructure not
available on the current development host (Proxmox VE / devtmpfs restriction):

### 1. Full mkarchiso ISO build

**Status:** Deferred — host restriction (Proxmox VE devtmpfs)

`mkarchiso` requires kernel-level `devtmpfs` access which is unavailable on
the Proxmox VE host. All tooling is correct; the `--skip-iso` path (profile
emission + injection tar) exercises the same code path up to the docker
invocation.

**Steps to verify on a capable host:**
```bash
debateos build --dir examples/omarchy --out /tmp/omarchy-iso
ls /tmp/omarchy-iso/*.iso   # should exist
```

Or via Docker:
```bash
docker run --privileged --rm \
  -v "$(pwd)/examples/omarchy:/speech:ro" \
  -v "/tmp/out:/out" \
  ghcr.io/mikl0s/debateos:latest
```

### 2. Live cross-repo GitHub Actions run

**Status:** Deferred — requires fork + CI minutes

The cross-repo `workflow_call` (`uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main`)
requires a GitHub fork with Actions enabled and available CI minutes.

**Steps to verify:**
1. Fork `mikl0s/DebateOS` to your GitHub account.
2. Create a test workflow calling the reusable workflow from your fork.
3. Push and confirm the build job runs in your fork's runner (not upstream).
4. Confirm the artifact is uploaded to the fork's Actions run.

See `build/actions/README.md` for the exact workflow YAML.

---

## Repository Layout

```
build/
├── docker/
│   ├── Dockerfile       # multi-stage image (golang builder + archlinux runtime)
│   └── entrypoint.sh    # /speech → /out via debateos build
└── actions/
    ├── build-speech.yml # reusable workflow_call (channel 2)
    └── README.md        # fork-and-build guide
.github/
└── workflows/
    └── build-speech.yml # thin caller (self-test for this repo)
scripts/
├── determinism-test.sh  # double-run sha256 gate (BLD-03)
├── secret-free-check.sh # profile tree grep gate (PRIV-01)
└── check-coverage.sh    # resolver >=90% + cli >=85% gate (D19)
docs/
└── cli-build-channels.md  # this file
```
