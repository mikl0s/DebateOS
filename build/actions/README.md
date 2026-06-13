# DebateOS GitHub Actions Build Channel

This directory contains the **reusable GitHub Actions workflow** that powers
build channel 2 (BLD-02).  The same Docker image used for local `docker run`
builds (channel 1 / BLD-01) is specified via `container:` in this workflow —
the same-image property is enforced by design, not by convention.

## How It Works

```
Your fork / template repo
        │
        ▼
.github/workflows/my-speech.yml
  uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main
  with:
    speech-dir: path/to/my/speech
    profile: vanilla-arch
        │
        ▼ (GitHub Actions calls back into the DebateOS repo)
build/actions/build-speech.yml  (workflow_call target)
        │  container: ghcr.io/mikl0s/debateos:latest
        │
        ▼
debateos build --dir <speech> --profile vanilla-arch --skip-iso
        │
        ▼
actions/upload-artifact@v4  →  out/ uploaded as build artifact
```

Your GitHub Actions minutes power the build.  No DebateOS infrastructure is
involved after the workflow definition is fetched from the public repo (zero
hosting cost, invariants 4/5).

## Quick Start: Fork-and-Build

### Step 1 — Fork or use the template repo

Fork `mikl0s/DebateOS` (or use it as a template repo for your own speech).

### Step 2 — Create your speech

Put your speech YAML in a directory, e.g. `speeches/my-speech/`:

```
speeches/my-speech/
├── speech.yaml
├── opinions/
│   └── my-opinion.yaml
└── points/
    └── my-point.yaml
```

### Step 3 — Add the caller workflow

Create `.github/workflows/build.yml` in your fork:

```yaml
on:
  push:
    paths:
      - "speeches/**"

permissions:
  contents: read

jobs:
  build:
    uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main
    with:
      speech-dir: speeches/my-speech
      profile: vanilla-arch
      skip-iso: true        # remove this line on a capable host for a full ISO
```

### Step 4 — Push and watch Actions

Commit and push.  GitHub will run the reusable workflow on your own runner
minutes.  The build artifact (arch-profile/ tree and resolved.json) is
uploaded via `actions/upload-artifact@v4`.

### Step 5 — Download the artifact

Download the artifact from the Actions run summary page.  For a full ISO
(without `skip-iso: true`), the ISO file is included in the artifact.

## Workflow Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `speech-dir` | Yes | — | Path to the speech directory in the caller's repo |
| `profile` | No | `vanilla-arch` | Translator variant profile (`vanilla-arch`, `cachyos`, `garuda`) |
| `skip-iso` | No | `false` | Stop after profile emission; skip full mkarchiso ISO build |

## Optional Secret

| Secret | Required | Description |
|--------|----------|-------------|
| `PANE_AGE_KEY` | No | age private key (base64) for private-pane CI builds (future) |

For public-pane builds no secrets are required — the image is pulled from
public GHCR without credentials.

## Local Build Channel (channel 1)

If you prefer to build locally:

```bash
docker run --rm \
  -v "$(pwd)/speeches/my-speech:/speech:ro" \
  -v "$(pwd)/out:/out" \
  -e SKIP_ISO=1 \
  ghcr.io/mikl0s/debateos:latest
```

Or use the CLI directly (no Docker required for profile emission):

```bash
debateos build --dir speeches/my-speech --profile vanilla-arch --skip-iso
```

## Deferred Verification Items

The following items are **documented deferred verifications** — they are
correct by design but require infrastructure not available on this host:

### 1. Live cross-repo Actions run (A2)

**What:** A user fork calling
`uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main` from their
own repo (cross-repo `workflow_call`).

**Why deferred:** Requires a fork of `mikl0s/DebateOS` plus GitHub Actions
CI minutes.  The workflow YAML is syntactically valid (PyYAML validated) and
follows the documented `workflow_call` syntax from official GHA docs.

**Steps to verify:**
1. Fork `mikl0s/DebateOS` to your GitHub account.
2. In the fork, create `.github/workflows/test-cross-repo.yml` calling
   `uses: mikl0s/DebateOS/.github/workflows/build-speech.yml@main`.
3. Push and watch the Actions tab — the job must appear and run in the fork's
   runner, not in `mikl0s/DebateOS`.
4. Confirm the artifact is uploaded to the fork's run, not the upstream repo.

### 2. Full ISO build (ARCH-01 full path)

**What:** The full mkarchiso path (without `--skip-iso`) producing a bootable
`.iso` file.

**Why deferred:** `mkarchiso` requires kernel-level devtmpfs access which is
restricted on the current Proxmox VE host.  The `--skip-iso` path (profile
emission) works on all hosts and exercises the same code path up to the
`docker run` invocation.

**Steps to verify:**
1. Use a bare-metal Linux host or a VM with devtmpfs access.
2. Run `docker run --rm -v <speech>:/speech -v <out>:/out ghcr.io/mikl0s/debateos:latest`
   (without `SKIP_ISO=1`).
3. Confirm `out/*.iso` is produced.

## Security Notes

- All workflow inputs are passed via environment variables in `run:` steps,
  never interpolated directly into shell strings (T-03-CIWF injection prevention).
- The workflow runs with `permissions: contents: read` (least privilege default).
- `private-injection.tar` is NEVER produced inside the Actions runner — it is
  a local-only artifact for the flash-time private-pane injection step.
  See `docs/cli-build-channels.md` for the flash-time injection guide.
