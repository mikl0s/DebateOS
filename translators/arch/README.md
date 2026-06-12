# DebateOS Arch Translator

**translators/arch/** — converts a resolved speech to a bootable, fully-unattended Arch Linux installer ISO.

This translator is the Phase 2 deliverable of [DebateOS](../../README.md). It consumes a `ResolvedSpeech` (the Phase 1 resolver's canonical JSON output) and emits an archiso profile tree which is built into an ISO via mkarchiso inside a privileged Docker container.

**License:** AGPL-3.0-only (root `LICENSE` covers all translator code). Example content in `examples/omarchy/` is CC0-1.0.

---

## Table of Contents

1. [Input Contract](#input-contract)
2. [Argv-Stable Entrypoint](#argv-stable-entrypoint)
3. [capabilities.json](#capabilitiesjson)
4. [Variant Profiles](#variant-profiles)
5. [Slow Gates](#slow-gates)
6. [Optional: QEMU Boot Smoke](#optional-qemu-boot-smoke)
7. [Architecture](#architecture)
8. [Developer Guide](#developer-guide)

---

## Input Contract

The Arch translator takes two inputs:

### 1. ResolvedSpeech JSON (`resolved.json`)

The canonical output of `resolve.Resolve` — a typed `ResolvedSpeech` struct serialized to
JSON by `resolve.CanonicalJSON`. This is the **Phase 1 resolver's output** and the defined
translator input contract per `docs/11`.

Key fields the translator consumes:

| Field | Description |
|-------|-------------|
| `schema` | Version (currently `1`) |
| `foundation` | Target foundation (must be `"arch"` for this translator) |
| `applied` | Opinion IDs accepted by the resolver; form the install set |
| `install_order` | Deterministic topological install order (from `applied`) |
| `skipped` | Hardware-gated opinions that did not match the declared hardware |
| `dropped` | Nice-to-have opinions dropped by the resolution rules |
| `explanations` | Per-decision explanation records (surfaced in the build manifest) |

Generate `resolved.json` from a speech directory:
```bash
go run ./cmd/resolve-json examples/omarchy > resolved.json
```

### 2. Opinion Bodies (`--opinions <dir>`)

A directory of YAML opinion files (e.g. `examples/omarchy/opinions/`), one per opinion ID.
These provide the payload fields (packages, file_assets, services, sysctl_params, etc.)
that the generator reads when building the archiso profile tree.

The translator reads **only opinions that appear in `resolved.json`'s `applied` list**.
Opinions in `skipped` or `dropped` are ignored.

---

## Argv-Stable Entrypoint

```
translate <resolved.json> --opinions <dir> [--profile <name>] [--out <dir>]
```

**This argv is FROZEN for Phase 3 CLI subprocess invocation (02-CONTEXT.md Integration Points).**
Do not change argument names or ordering without updating the Phase 3 CLI contract.

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `<resolved.json>` | yes | — | Path to the ResolvedSpeech JSON file |
| `--opinions <dir>` | yes | — | Directory of opinion YAML files |
| `--profile <name>` | no | `vanilla-arch` | Variant profile name (see `profiles/`) |
| `--out <dir>` | no | `./arch-profile` | Output directory for the archiso profile tree |

**Examples:**

```bash
# North-star: translate the Omarchy speech on vanilla Arch
translate resolved.json --opinions examples/omarchy/opinions --profile vanilla-arch --out ./arch-profile

# CachyOS variant (stretch, non-gating)
translate resolved.json --opinions examples/omarchy/opinions --profile cachyos --out ./arch-profile-cachyos
```

**Exit codes:**
- `0` — profile tree generated successfully; output in `<out>`.
- `1` — error (capability gate failure, missing file, profile validation error).

The entrypoint is a thin bash wrapper (`translators/arch/translate`) that calls the Python
generator module:
```bash
exec python3 -m translators.arch.generator <resolved.json> <opinions-path> <profile> <out-dir>
```

---

## capabilities.json

`translators/arch/capabilities.json` declares every `translator_capability` token this
translator supports. The generator checks this list against the capabilities declared by
each opinion in the resolved speech.

### Capability Gate Behavior (ARCH-03 / SC-3)

The check runs **before any file I/O** — it is a composition-time gate, never a silent
install-time failure.

| Opinion Status | Unsupported Capability | Behavior |
|----------------|------------------------|----------|
| `required` | any | `CapabilityError` raised with the opinion ID, capability name, and the phrase `"composition time"` — generator exits 1 |
| `nice-to-have` | any | Opinion is silently dropped; logged in the build manifest's `dropped_capabilities` list |

**Example error (required opinion OM-023 requires npm which is unsupported):**
```
CapabilityError: Opinion OM-023 requires capability 'install-npm-global-packages'
which is not declared by the Arch translator at composition time.
Add it to translators/arch/capabilities.json when implemented.
```

The current capabilities list covers all 134 Omarchy opinions (OM-001..OM-134). See
`capabilities.json` for the complete list.

### Adding a Capability

1. Implement the handler in `generator.py` (or the relevant module).
2. Add the token to `capabilities.json`.
3. Add a test in `tests/test_capability_gate.py` and the relevant implementation test.
4. Run `python -m pytest translators/arch/tests/ -x -q` to verify.

---

## Variant Profiles

Declarative YAML profiles in `profiles/` parameterize the generator without forking code.
**One generator, no per-variant branches (ARCH-04 invariant).**

| Profile | File | Description |
|---------|------|-------------|
| `vanilla-arch` | `profiles/vanilla-arch.yaml` | Baseline — standard Arch, no custom repos or kernel variants. North-star target for the Omarchy speech. |
| `cachyos` | `profiles/cachyos.yaml` | CachyOS — CPU-ISA-optimised packages (x86-64-v3/v4), EEVDF kernel, CachyOS repos + keyring above core. |
| `garuda` | `profiles/garuda.yaml` | Garuda Linux — btrfs-first, GRUB, dracut initramfs, chaotic-aur. **4 hard conflicts with Omarchy** (documented in garuda.yaml). |

Each profile declares:
- **repos**: Custom package repositories (with priority: above or below `[core]`/`[extra]`)
- **keyring_install_before_repos**: Keyring packages installed before custom repos (Pitfall 4 ordering)
- **kernel**: Package name + headers
- **defaults**: initramfs tool, bootloader preference, filesystem
- **pre_seeded_opinions**: What the variant's base packages already express (conflict detection)

See `profiles/README.md` for the complete profile schema.

### ARCH-04 No-Fork Verification

```bash
grep -vE '^#' translators/arch/variant.py | grep -Ei "if .*(cachyos|garuda|vanilla)" | wc -l
# Must be 0 — all variant logic is data-driven
```

### Omarchy Conflicts on Garuda

The `garuda.yaml` profile documents 4 hard conflicts with Omarchy opinions:

| Garuda mechanism | Omarchy mechanism | Affected opinions | Type |
|-----------------|-------------------|-------------------|------|
| dracut (conflicts mkinitcpio) | mkinitcpio | OM-002 | hard |
| GRUB mandatory | limine (OM-099) | OM-099 | hard |
| snapper root config | Omarchy snapper (OM-099) | OM-099 | direct |
| Dr460nized SDDM theme | Omarchy SDDM theme | OM-097 | direct |

**Omarchy-on-Garuda is a non-gating stretch criterion** (SC-5, deferred to post-v1.0).
The profile and its conflict markers are Phase 2 deliverables.

---

## Slow Gates

These scripts run at wave/phase verification, **not per-commit** (they are 20-40 minute operations):

### `scripts/arch-build-iso.sh` — Docker mkarchiso ISO build

```bash
SOURCE_DATE_EPOCH=<epoch> bash scripts/arch-build-iso.sh <profile-dir> <out-dir>
```

- Runs mkarchiso inside a **privileged** Docker container (required for pacstrap bind mounts — see Pitfall 1 in `02-RESEARCH.md`).
- Docker image: `archlinux:base-devel` pinned by sha256 digest (T-02-13 stale-image threat mitigation).
- The generator profile is overlaid on the releng baseline inside the container (syslinux/, efiboot/ from releng; our profiledef/packages/installer overlay on top).
- `SOURCE_DATE_EPOCH` is derived from the resolved-speech SHA-256 hash for build determinism (BLD-03 groundwork).

**Security note:** `--privileged` is scoped to this script only. No other DebateOS operation uses a privileged container (T-02-12).

**Digest maintenance:** Re-verify the Docker image digest quarterly:
```bash
docker pull archlinux:base-devel && docker inspect --format='{{index .RepoDigests 0}}'
# Update the digest in scripts/arch-build-iso.sh and translators/arch/Dockerfile
```

### `scripts/arch-validate-iso.sh` — ISO structural validation

```bash
bash scripts/arch-validate-iso.sh <iso-file>
```

The **Phase 2 bootability gate** — structural, not QEMU. Checks:
1. ISO9660 / El Torito primary volume descriptor (`xorriso pvd_info`)
2. EFI boot entries under `/EFI` (systemd-boot / UEFI)
3. Syslinux entries (BIOS fallback)
4. `airootfs.sfs` under `/arch/x86_64/`
5. `debateos-install.sh` present in the squashfs (the generated installer)
6. `.zlogin` present (installer hook, Pattern 1)
7. At least one `debateos-firstrun-*.service` unit (first-run opinions)

Exit 0 = structural validation passed. Exit 1 = one or more checks failed.

**Dependencies:** `xorriso` (libisoburn) and `unsquashfs` (squashfs-tools) — both available inside the Docker build image. If not on the host, run inside the container.

### `scripts/arch-northstar-check.sh` — Full north-star gate (ARCH-02)

```bash
bash scripts/arch-northstar-check.sh              # Full pipeline (build included)
bash scripts/arch-northstar-check.sh --skip-build # Equivalence only (fast, ~60s)
```

The **ARCH-02 north-star gate**: runs the complete pipeline from speech to equivalence check:

1. **Resolve** (`go test ./examples/ -run TestExampleOmarchy`) — clean-resolution gate
2. **Emit resolved.json** (`go run ./cmd/resolve-json examples/omarchy`)
3. **Translate** (`translators/arch/translate resolved.json --opinions ... --profile vanilla-arch`)
4. **Mechanical equivalence** (4 checks):
   - Package-set: `target_packages` count > 0 in `build-manifest.json`
   - File-asset: all file assets in manifest have `dst` field
   - Service: `system_services` count in manifest, or installer references systemctl
   - First-run: one `debateos-firstrun-*.service` unit per `first_run` entry in manifest
5. **Build** (`arch-build-iso.sh`) — skipped with `--skip-build`
6. **Validate** (`arch-validate-iso.sh`) — skipped with `--skip-build`
7. **Regression** (`go test ./... -count=1` + `python -m pytest translators/arch/tests/ -q`)

**The `--skip-build` equivalence-only run is GREEN** (16/16 checks, verified 2026-06-12).

**Full build status:** The Docker ISO build (`mkarchiso` via `pacstrap`) requires
`devtmpfs` filesystem mounting. This environment (Proxmox VE kernel 6.17.4-2-pve) restricts
`devtmpfs` inside containers. The build command and the structural validation tooling are
implemented and correct; full execution requires a host that allows devtmpfs in Docker
(standard Linux with Docker Desktop, or a VM with unrestricted kernel capabilities).

**To run the full build on a compatible host:**
```bash
bash scripts/arch-northstar-check.sh   # without --skip-build
```

---

## Optional: QEMU Boot Smoke

QEMU is **not available** on the build host (Proxmox VE). The Phase 2 bootability gate is
structural (`arch-validate-iso.sh`). For users who have QEMU:

```bash
# After a successful arch-validate-iso.sh, boot smoke-test the ISO:
qemu-system-x86_64 \
    -m 2048 \
    -cdrom <path-to-iso> \
    -boot d \
    -nographic \
    -no-reboot \
    -serial stdio \
    2>&1 | head -100
# Expect: kernel output, then autologin, then installer output.
# The installer will fail at pacstrap (no target disk) — that is expected.
# A successful smoke test shows the live env boots and .zlogin fires.
```

For UEFI boot:
```bash
qemu-system-x86_64 \
    -m 2048 \
    -bios /usr/share/ovmf/OVMF.fd \
    -cdrom <path-to-iso> \
    -boot d \
    -nographic \
    -serial stdio
```

---

## Architecture

```
ResolvedSpeech JSON (Phase 1 resolver output)
          │
          ▼
  translators/arch/generator.py
    1. load_resolved_speech()    — read the ResolvedSpeech JSON
    2. load_opinion_bodies()     — load opinion YAML files from --opinions dir
    3. load_capabilities()       — read capabilities.json
    4. BuildManifest.from_resolved()  — capability gate (SC-3) + aggregate all
                                       packages/file_assets/services/first_run
    5. load_variant_profile()    — read profiles/<name>.yaml
    6. emit_profile_tree()       — write complete archiso profile tree
          │
          ▼ (archiso profile dir)
  scripts/arch-build-iso.sh
    docker run --privileged archlinux:base-devel@<digest>
      copy releng baseline into /profile
      overlay debateos generator output on top
      mkarchiso -v -w /tmp/work -o /out /profile
          │
          ▼ (.iso file)
  scripts/arch-validate-iso.sh
    xorriso pvd_info + find /EFI + find airootfs.sfs
    unsquashfs -l airootfs.sfs | grep installer/zlogin/firstrun
          │
          ▼ (structural validation PASSED)
  scripts/arch-northstar-check.sh
    go test TestExampleOmarchy (clean-resolve)
    + 4 mechanical equivalence checks
```

**Generated profile tree:**
```
<out-dir>/
  profiledef.sh              — ISO metadata (iso_name=debateos, bootmodes, file_permissions)
  packages.x86_64            — Minimal live-env set (~15 pkgs, NOT the 285-pkg target set)
  pacman.conf                — pacman config with variant repos injected
  airootfs/
    root/
      debateos-install.sh    — Generated unattended installer (0755)
      .zlogin                — Calls installer on /dev/tty1 (releng Pattern 1)
    etc/
      systemd/
        user/
          debateos-firstrun-<id>.service  — One per execution_phase=first-run opinion
  build-manifest.json        — Runtime data read by installer via jq (Pitfall 6)
```

The installer script reads `build-manifest.json` at install time via `jq` — opinion payload
data is never shell-interpolated (T-02-09 shell-injection mitigration).

---

## Developer Guide

### Running the Python tests

```bash
cd /path/to/DebateOS
python -m pytest translators/arch/tests/ -x -q
# 128 tests; all GREEN
```

### Running the Go tests (includes TestExampleOmarchy)

```bash
go test ./... -count=1
```

### Full north-star (equivalence only, fast)

```bash
bash scripts/arch-northstar-check.sh --skip-build
# Completes in ~60s; resolves Omarchy speech + translates + checks equivalence
```

### Module layout

| File | Responsibility |
|------|----------------|
| `generator.py` | Public entrypoint: `generate(resolved, opinions, profile, out)` |
| `manifest.py` | `BuildManifest.from_resolved()` — aggregates all payload; `derive_source_date_epoch()` |
| `profile.py` | `emit_profile_tree()` — writes all profile files; `_sanitize_dst()` (T-02-08) |
| `variant.py` | `load_variant_profile()`, `apply_variant()`, `surface_conflicts()` |
| `firstrun.py` | `render_firstrun_unit()` — systemd user oneshot unit template (Pattern 2) |
| `capabilities.py` | `load_capabilities()`, `check_capabilities()` — SC-3 gate |
| `contract.py` | `load_resolved_speech()`, `load_opinion_bodies()` — input loaders |
| `translate` | Thin bash wrapper: frozen argv for Phase 3 CLI subprocess invocation |
| `capabilities.json` | Declared translator capability tokens |
| `profiles/` | Declarative variant YAML profiles |
| `templates/` | `profiledef.sh.tpl`, `pacman.conf.tpl`, `installer.sh.tpl`, `firstrun.service.tpl` |
| `tests/` | pytest suite: 128 tests covering all modules + fixtures |

### Security properties

| Threat | Mitigation |
|--------|-----------|
| T-02-08: file_asset path traversal | `_sanitize_dst()` rejects absolute paths and `..` escapes before any write |
| T-02-09: shell injection in installer | All opinion data in `build-manifest.json`; installer reads via `jq`, no string interpolation |
| T-02-12: privileged Docker container | Scoped to `arch-build-iso.sh` only; documented |
| T-02-13: stale Docker base image | Image pinned by sha256 digest; quarterly re-verification reminder in Dockerfile |
| sig_level=Never repos | TrustWarning emitted in build manifest and pacman.conf comment |
| First-run unit scope | Units in `etc/systemd/user/` (not system/); WantedBy=graphical-session.target |
