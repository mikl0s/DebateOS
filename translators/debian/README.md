# DebateOS Debian Translator

**translators/debian/** — converts a resolved speech to a bootable, fully-unattended Debian installer ISO via live-build and preseed.

This translator is the Phase 4 deliverable of [DebateOS](../../README.md). It consumes the same `ResolvedSpeech` contract as the Arch translator (the Phase 1 resolver's canonical JSON output) and emits a live-build `config/` tree which is built into a Debian ISO via `lb build` inside a privileged Docker container.

**License:** AGPL-3.0-only (root `LICENSE` covers all translator code). Example content in `examples/dual-foundation/` is CC0-1.0.

---

## Table of Contents

1. [Input Contract](#input-contract)
2. [Argv-Stable Entrypoint](#argv-stable-entrypoint)
3. [capabilities.json](#capabilitiesjson)
4. [Variant Profiles](#variant-profiles)
5. [Config/ Tree Layout](#config-tree-layout)
6. [Slow Gates](#slow-gates)
7. [Full Build Status](#full-build-status)
8. [Architecture](#architecture)
9. [Developer Guide](#developer-guide)

---

## Input Contract

The Debian translator takes two inputs:

### 1. ResolvedSpeech JSON (`resolved.json`)

The canonical output of `resolve.Resolve` — a typed `ResolvedSpeech` struct serialized to JSON by `resolve.CanonicalJSON`. This is the **Phase 1 resolver's output** and the defined translator input contract per `docs/11`.

Key fields the translator consumes:

| Field | Description |
|-------|-------------|
| `schema` | Version (currently `1`) |
| `foundation` | Target foundation (must be `"debian"` for this translator) |
| `applied` | Opinion IDs accepted by the resolver; form the install set |
| `install_order` | Deterministic topological install order (from `applied`) |
| `skipped` | Hardware-gated opinions that did not match the declared hardware |
| `dropped` | Nice-to-have opinions dropped by the resolution rules |
| `explanations` | Per-decision explanation records (surfaced in the build manifest) |

Generate `resolved.json` from a speech directory:
```bash
go run ./cmd/resolve-json examples/dual-foundation > resolved.json
```

### 2. Opinion Bodies (`--opinions <dir>`)

A directory of YAML opinion files (e.g. `examples/dual-foundation/opinions/`), one per opinion ID. These provide the payload fields (packages, file_assets, services, sysctl_params, etc.) that the generator reads when building the live-build `config/` tree.

The translator reads **only opinions that appear in `resolved.json`'s `applied` list**. Opinions in `skipped` or `dropped` are ignored.

---

## Argv-Stable Entrypoint

```
translate <resolved.json> --opinions <dir> [--profile <name>] [--out <dir>]
```

**This argv is FROZEN for Phase 3 CLI subprocess invocation (04-CONTEXT.md Integration Points).**
Do not change argument names or ordering without updating the Phase 3 CLI contract.

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `<resolved.json>` | yes | — | Path to the ResolvedSpeech JSON file |
| `--opinions <dir>` | yes | — | Directory of opinion YAML files |
| `--profile <name>` | no | `debian` | Variant profile name (see `profiles/`) |
| `--out <dir>` | no | `./debian-profile` | Output directory for the live-build config/ tree |

**Examples:**

```bash
# Dual-foundation proof: translate the DF speech on Debian
translate resolved.json --opinions examples/dual-foundation/opinions --profile debian --out ./debian-profile

# Ubuntu variant (post-v1.0; profile structure accommodates it)
translate resolved.json --opinions examples/dual-foundation/opinions --profile ubuntu --out ./ubuntu-profile
```

**Exit codes:**
- `0` — config/ tree generated successfully; output in `<out>`.
- `1` — error (capability gate failure, missing file, profile validation error).

The entrypoint is a thin bash wrapper (`translators/debian/translate`) that calls the Python generator module:
```bash
exec python3 -m translators.debian.generator <resolved.json> <opinions-path> <profile> <out-dir>
```

---

## capabilities.json

`translators/debian/capabilities.json` declares every `translator_capability` token this translator supports. The generator checks this list against the capabilities declared by each opinion in the resolved speech.

### Capability Gate Behavior (DEB-01 / SC-3)

The check runs **before any file I/O** — it is a composition-time gate, never a silent install-time failure.

| Opinion Status | Unsupported Capability | Behavior |
|----------------|------------------------|----------|
| `required` | any | `CapabilityError` raised with the opinion ID, capability name, and the phrase `"composition time"` — generator exits 1 |
| `nice-to-have` | any | Opinion is silently dropped; logged in the build manifest's `dropped_capabilities` list |

**Example error (required opinion requires a pacman-specific capability):**
```
CapabilityError: Opinion OM-023 requires capability 'configure-mkinitcpio-hooks-and-modules'
which is not declared by the Debian translator at composition time.
Add it to translators/debian/capabilities.json when implemented.
```

### Arch-Only Capabilities (Intentionally Absent)

The following Arch-specific capability tokens are **correctly absent** from `translators/debian/capabilities.json` (DEB-03 audit — see `docs/arch-leak-audit.md`):

| Token | Why Absent |
|-------|-----------|
| `configure-mkinitcpio-hooks-and-modules` | mkinitcpio is Arch-specific; Debian uses initramfs-tools |
| `write-mkinitcpio-config-drop-in` | mkinitcpio is Arch-specific |
| `write-mkinitcpio-module-list` | mkinitcpio is Arch-specific |
| `manage-limine-bootloader-installation` | Limine is used by Omarchy on Arch; Debian uses GRUB2 |
| `write-bootloader-entry-tool-drop-in` | Arch limine tooling; not applicable to GRUB2 |

Opinions that require these capabilities will `CapabilityError` at composition time when targeting Debian — which is correct behavior. The dual-foundation proof speech (`examples/dual-foundation/`) uses only the 5 tokens declared by BOTH translators.

### Adding a Capability

1. Implement the handler in `generator.py` or the relevant module.
2. Add the token to `capabilities.json`.
3. Add a test in `tests/test_capability_gate.py` and the relevant implementation test.
4. Run `python3 -m pytest translators/debian/tests/ -x -q` to verify.

---

## Variant Profiles

Declarative YAML profiles in `profiles/` parameterize the generator without forking code.
**One generator, no per-variant branches (ARCH-04 invariant applies to Debian too).**

| Profile | File | Description |
|---------|------|-------------|
| `debian` | `profiles/debian.yaml` | Debian stable — standard apt repos, GRUB2 bootloader, initramfs-tools, ext4. The reference implementation. |

**Post-v1.0 (structure ready):** The profile schema accommodates Ubuntu and other Debian-family variants (same key schema as `debian.yaml`, different repo/kernel/bootloader values). Add `profiles/ubuntu.yaml` to activate.

Each profile declares:
- **repos**: Custom apt repositories (with `sig_level` → `signed-by`/`trusted=yes` mapping per DEB-03)
- **kernel**: Package names (`linux-image-amd64`, `linux-headers-amd64`)
- **defaults**: initramfs tool, bootloader, filesystem
- **pre_seeded_opinions**: What the variant's base packages already express (conflict detection)

### Sig-Level → apt Mapping (DEB-03)

The schema `sig_level` enum maps to apt source options in `variant.py`:

| Schema `sig_level` | apt option | Trust warning |
|-------------------|-----------|---------------|
| `Required` | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | None |
| `RequiredDatabaseOptional` | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | None |
| `OptionalTrustAll` | `[trusted=yes]` | Yes (T-04-07) |
| `Never` | `[trusted=yes]` + LOUD `# WARNING:` comment | Yes + louder (T-04-07) |

---

## Config/ Tree Layout

The Debian translator emits a live-build `config/` tree:

```
<out-dir>/
  config/
    includes.installer/
      preseed.cfg                         — d-i automation (locale/disk/user/pkgsel)
    hooks/live/
      9000-debateos-apply.hook.chroot     — opinion effectuation at lb chroot time (0755)
    package-lists/
      debateos.list.chroot_install        — target packages (both live + installed env)
    archives/
      debateos-<name>.list.chroot_install — custom apt repos (with signed-by/trusted=yes)
      debateos-<name>.key.chroot          — keyring stubs for signed-by repos
    includes.chroot/
      etc/systemd/user/
        debateos-firstrun-<id>.service    — first-run systemd user units
  build-manifest.json                     — full manifest dict (target_packages, file_assets, ...)
```

**Key differences from the Arch profile tree:**
- No `packages.x86_64` (packages are in `config/package-lists/debateos.list.chroot_install`)
- No `profiledef.sh` or `pacman.conf` (live-build uses `config/` conventions, not archiso)
- `preseed.cfg` drives d-i automation; the chroot hook applies all opinion packages/config
- The chroot hook runs at `lb build` time (inside the build chroot), not at install time

**Note on %%SENTINEL%% placeholders in preseed.cfg:**
The emitted `preseed.cfg` contains `%%USERNAME%%`, `%%USER_FULLNAME%%`, `%%HASHED_PASSWORD%%`, and `%%PKGSEL_PACKAGES%%` sentinels. These are intentional build-time placeholders — replace them with actual values before running `lb build`. The `debian-build-iso.sh` script documents this step.

---

## Slow Gates

These scripts run at wave/phase verification, **not per-commit** (they are 20-40 minute operations):

### `scripts/dual-foundation-check.sh` — DEB-02 Dual-Foundation Proof Gate

```bash
bash scripts/dual-foundation-check.sh --skip-iso  # Fast path (equivalence + structural)
bash scripts/dual-foundation-check.sh             # Full path (includes lb build; capable host)
```

The **DEB-02 gate**: resolves `examples/dual-foundation` once, then runs BOTH the Arch and Debian translators on the **same** `resolved.json`:

1. **Resolve** (`go run ./cmd/resolve-json examples/dual-foundation`) — one resolve, shared input
2. **Arch translate** (`translators/arch/translate`) — Arch profile tree
3. **Debian translate** (`translators/debian/translate`) — Debian config/ tree
4. **Equivalence checks** (per-foundation + cross-foundation):
   - Both exit 0 (no CapabilityError on either foundation)
   - Arch: `target_packages` contains git/curl/vim; `file_assets` includes `etc/motd`
   - Debian: `debateos.list.chroot_install` contains git/curl/vim; chroot hook exists + executable; `preseed.cfg` exists
   - Cross-foundation: package set from the same resolved.json is identical on both foundations
5. **Structural validation** (`debian-validate-iso.sh`) — validates the Debian config/ tree

**`--skip-iso` PASSES on this host: 20/20 checks GREEN (VERIFIED 2026-06-13)**

### `scripts/debian-build-iso.sh` — Docker lb-build ISO build

```bash
SOURCE_DATE_EPOCH=<epoch> bash scripts/debian-build-iso.sh <config-tree-dir> <out-dir>
bash scripts/debian-build-iso.sh <config-tree-dir> <out-dir> --skip-iso
```

- Runs `lb build` inside a **privileged** Docker container (required for chroot bind mounts — T-04-12).
- Docker image: `debian:stable` pinned by sha256 digest (T-04-13 stale-image threat mitigation).
- `SOURCE_DATE_EPOCH` is derived from the resolved-speech SHA-256 hash for build determinism (BLD-03 groundwork).
- `--skip-iso`: skips `lb build` and runs `debian-validate-iso.sh` only (suitable for this host).

**Security note:** `--privileged` is scoped to this script only. No other DebateOS operation uses a privileged container (T-04-12).

### `scripts/debian-validate-iso.sh` — config/ tree structural validation

```bash
bash scripts/debian-validate-iso.sh <config-tree-dir>
```

Host-runnable structural validation of the live-build `config/` tree. Checks:
1. `preseed.cfg` present + contains `d-i` directives
2. Chroot hook `9000-debateos-apply.hook.chroot` present + executable (0755)
3. `package-lists/debateos.list.chroot_install` present + non-empty
4. `build-manifest.json` present + valid JSON

Exit 0 = structural validation passed. Exit 1 = one or more checks failed.

---

## Full Build Status

**Profile emission + structural validation:** GREEN on this host (verified).

**Full `lb build` ISO:** DEFERRED-TO-CAPABLE-HOST.

The Docker `lb build` (chroot + squashfs + ISO assembly) requires `devtmpfs` filesystem mounting and unrestricted loop device access. This environment (Proxmox VE kernel 6.17.4-2-pve) restricts devtmpfs inside containers — the same restriction that blocks mkarchiso in the Arch translator (see `translators/arch/README.md §Full Build Status`).

The build command and structural validation tooling are implemented and correct. Full execution requires a host that allows devtmpfs in Docker (standard Linux with Docker, or a VM with unrestricted kernel capabilities).

**To run the full build on a compatible host:**
```bash
bash scripts/dual-foundation-check.sh  # without --skip-iso
```

---

## Architecture

```
ResolvedSpeech JSON (Phase 1 resolver output)
          │
          ▼
  translators/debian/generator.py
    1. load_resolved_speech()    — read the ResolvedSpeech JSON
    2. load_opinion_bodies()     — load opinion YAML files from --opinions dir
    3. load_capabilities()       — read capabilities.json (Debian subset)
    4. check_capabilities()      — capability gate BEFORE any file I/O (SC-3)
    5. BuildManifest.from_resolved() — aggregate packages/file_assets/services/first_run
    6. load_variant_profile()    — read profiles/debian.yaml
    7. emit_profile_tree()       — write complete live-build config/ tree
          │
          ▼ (live-build config/ dir)
  scripts/debian-build-iso.sh
    docker run --privileged debian:stable@<digest>
      apt-get install live-build debootstrap squashfs-tools xorriso
      copy config/ tree into /build/
      lb build
          │
          ▼ (.iso file — DEFERRED-TO-CAPABLE-HOST)
  scripts/debian-validate-iso.sh
    preseed.cfg d-i lines check
    chroot hook 0755 check
    package-lists non-empty check
    build-manifest.json valid JSON check
          │
          ▼ (structural validation PASSED — on this host)
  scripts/dual-foundation-check.sh
    go run ./cmd/resolve-json examples/dual-foundation (resolve ONCE)
    + translators/arch/translate (same resolved.json)
    + translators/debian/translate (same resolved.json)
    + per-foundation + cross-foundation equivalence checks
    + regression tests (go test ./... + pytest suites)
```

**Shared code with Arch translator (`translators/common/`):**

```
translators/common/
  contract.py    — load_resolved_speech(), load_opinion_bodies() (foundation-neutral)
  manifest.py    — BuildManifest, derive_source_date_epoch() (foundation-neutral)
  firstrun.py    — render_firstrun_unit() (same flag-file pattern as Arch)
```

Both translators re-export from `translators/common/` via thin shims (e.g. `translators/debian/manifest.py` re-exports `from translators.common.manifest import *`). The shims preserve bare-name import compatibility for tests running under each translator's `pytest.ini`.

See `docs/arch-leak-audit.md` for the full DEB-03 audit: which capabilities are correctly translator-owned vs. schema-owned, and the single genuine leak that was fixed (build.go dispatch in cli/build/build.go via `foundationRegistry`).

---

## Developer Guide

### Running the Python tests

```bash
cd /path/to/DebateOS
python3 -m pytest translators/debian/tests/ -x -q
# 75 tests; all GREEN
```

### Running the Go tests (includes TestExampleDualFoundation)

```bash
go test ./... -count=1
```

### Full dual-foundation gate (equivalence only, fast)

```bash
bash scripts/dual-foundation-check.sh --skip-iso
# 20/20 PASS; resolves DF speech + translates on both foundations + equivalence checks
```

### Module layout

| File | Responsibility |
|------|----------------|
| `generator.py` | Public entrypoint: `generate(resolved, opinions, profile, out)` |
| `manifest.py` | Shim re-export from `translators/common/manifest` |
| `contract.py` | Shim re-export from `translators/common/contract` |
| `profile.py` | `emit_profile_tree()` — writes full live-build config/ tree; `_sanitize_dst()` (T-04-05) |
| `variant.py` | `load_variant_profile()`, `apply_variant()` — sig_level → apt mapping (DEB-03) |
| `capabilities.py` | `load_capabilities()`, `check_capabilities()` — SC-3 gate |
| `translate` | Thin bash wrapper: frozen argv for Phase 3 CLI subprocess invocation |
| `capabilities.json` | Declared translator capability tokens (45 tokens; Arch-only tokens absent) |
| `profiles/` | Declarative variant YAML profiles (`debian.yaml`; Ubuntu post-v1.0) |
| `templates/` | `preseed.cfg.tpl`, `chroot-install.hook.tpl` — %%SENTINEL%% replacement |
| `tests/` | pytest suite: 75 tests covering all modules + fixtures |

### Security properties

| Threat | Mitigation |
|--------|-----------|
| T-04-05: file_asset path traversal | `_sanitize_dst()` rejects absolute paths and `..` escapes before any write (mirrors T-02-08 in Arch translator) |
| T-04-06: shell injection in chroot hook | All opinion data via build-manifest.json; chroot hook reads it at lb build time; %%SENTINEL%% replacement in templates (never `str.format` with raw opinion data) |
| T-04-07: unsigned apt repos | `OptionalTrustAll`/`Never` → `[trusted=yes]` + LOUD WARNING comment in archives/*.list; `Required` → `[signed-by=...]` |
| T-04-08: plaintext password in preseed | `%%HASHED_PASSWORD%%` sentinel; never plaintext — replace with `openssl passwd -6` hash before `lb build` |
| T-04-12: privileged Docker container | Scoped to `debian-build-iso.sh` only; documented |
| T-04-13: stale Docker base image | Image pinned by sha256 digest; quarterly re-verification reminder in Dockerfile |
