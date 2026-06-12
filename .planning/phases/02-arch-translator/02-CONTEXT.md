# Phase 2: Arch Translator - Context

**Gathered:** 2026-06-12
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous — recommended answers auto-accepted per owner directive + ADR process notes)

<domain>
## Phase Boundary

Phase 2 delivers `translators/arch/`: a shell/Python translator (D8) that consumes a resolved speech (Phase 1 canonical JSON) and emits a bootable, fully-unattended Arch installer ISO by wrapping mkarchiso inside Docker — plus the Omarchy-as-a-speech composition in `examples/omarchy/` (north star, invariant 6) and declarative variant profiles (vanilla Arch, CachyOS, Garuda) per D20.

Out of scope: Debian translator (Phase 4), CLI orchestration (Phase 3 — this phase exposes a translator entrypoint, not the `debateos` CLI), build-channel packaging/Actions (Phase 3), hardware-scanning installer (Phase 6/post-v1.0 — declared hardware only), registry/UI (Phase 5).

</domain>

<decisions>
## Implementation Decisions

### Translator Structure & Input Contract
- Python for the profile generator (resolved-speech JSON → archiso profile tree); shell only as a thin mkarchiso/Docker invocation layer. Go orchestrates later (Phase 3); translators emit.
- Input contract: the Phase 1 resolver's CanonicalJSON ResolvedSpeech, passed as a file path argument. This is the defined translator input contract per docs/11 — document it in translators/arch/README.md.
- `translators/arch/capabilities.json` declares supported opinion categories and translator_capabilities tokens. A composition-time capability check (invoked before any build) fails loudly when a required opinion needs an undeclared capability (SC-3); nice-to-have unsupported opinions are dropped with explanations.
- Output: an archiso profile directory + a bootable unattended ISO built inside Docker (archlinux base image with archiso installed). Determinism via SOURCE_DATE_EPOCH derived from the resolved-speech hash (D11 groundwork; full channel packaging is Phase 3).

### Unattended Install Mechanics
- Custom automated installer script embedded in the ISO (airootfs): preset partitioning defaults (single-disk, declared via speech defaults), pacstrap from the resolved package set, file-asset/config deployment, group memberships, service enablement, sysctl/kernel params, theme assets — all driven by a build-time manifest generated from the resolved speech. Zero install-time questions (SC-1).
- `execution_phase: first-run` opinions become systemd first-boot units (oneshot with condition flag file), mirroring Omarchy's install-time vs first-run split (SR-011).
- Declared hardware only: hardware-conditional resolution already happened in the resolver; the translator consumes the resolved (post-hardware) opinion set. NO install-time hardware scanning (D2).
- Host facts: QEMU is NOT available; Docker IS available (passwordless sudo present). Bootability gate = ISO structural validation (ISO9660 + boot entries + airootfs + installer present). QEMU boot smoke documented as optional manual step. North-star equivalence gate = mechanical rootfs checks: package-set diff vs resolved speech, file-asset presence, service enablement, first-run units — verified by inspecting the built profile/rootfs, not interactive UX.

### Variant Profiles (SC-4, D20)
- Declarative YAML profiles in `translators/arch/profiles/`: `vanilla-arch.yaml`, `cachyos.yaml`, `garuda.yaml` — each: repo list (+priority placement), keyring packages, kernel package(s), defaults (fs/bootloader/initramfs), pre-seeded opinions list (foundation-already-has-an-opinion markers from the Phase 0 delta study). No per-variant code forks; one generator consumes any profile.
- Stretch (SC-5, non-gating): generation works for CachyOS/Garuda profiles and emits declared, explainable differences (e.g. kernel-default conflict surfaced per EC corpus); full variant ISO boot validation is post-v1.0/non-gating.

### Omarchy Speech (north star)
- `examples/omarchy/`: all 134 OM-NNN opinions as schema-valid YAML grouped into the 32 evidence-derived points (research/omarchy-points.md) + one speech targeting vanilla-arch. Authoring is script-assisted from research/omarchy-opinion-inventory.md, then schema-validated (every file must pass ParseOpinion/ParsePoint/ParseSpeech) and resolver-resolved without hard conflicts.
- North-star automated gate: resolve examples/omarchy → translate → generated profile contains the expected package set, file assets, first-run units; ISO build completes in Docker and passes structural validation.

### TDD & Test Infrastructure (D19)
- pytest for the Python generator — TDD with RED commits before implementation; tests cover manifest generation, profile emission, capability gating, variant-profile application, first-run unit generation.
- Slow paths (Docker ISO build, full Omarchy ISO) are separate gated scripts (scripts/arch-build-iso.sh, scripts/arch-northstar-check.sh) run at wave/phase verification, not per-commit.
- Python: stdlib + PyYAML only (mirror minimal-deps discipline); pytest as dev dependency. Pin in translators/arch/requirements-dev.txt.

### Claude's Discretion
- archiso profile internals (which releng baseline to start from), exact installer script structure, manifest format details, Docker image tag pinning strategy.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- resolver/resolve CanonicalJSON ResolvedSpeech (the input contract); Explanation type with install-order output.
- research/omarchy-opinion-inventory.md (134 opinions w/ source citations), research/omarchy-points.md (32 groupings), research/arch-variants-delta.md (variant-profile sketch + repo/keyring/kernel data).
- examples/ harness pattern (Go test driving parse→resolve end-to-end) — extend for omarchy speech.
- schemas/ validation via resolver/parse — every authored YAML must pass it.

### Established Patterns
- TDD RED/GREEN commits (D19); coverage gates via scripts/; deterministic outputs; OS-agnostic opinions with translator-owned mechanics (invariant 1 — the translator is where Arch mechanics BELONG).
- Licensing: translators are code → AGPL (root LICENSE covers); examples/omarchy content → CC0 (examples/LICENSE exists).

### Integration Points
- Phase 3 CLI will invoke translators as subprocesses with the resolved-speech contract; keep the entrypoint argv-stable (translators/arch/translate <resolved.json> --profile <profile> --out <dir>).
- Phase 4 Debian translator mirrors this structure; keep generator logic cleanly separated from Arch specifics where reasonable.

</code_context>

<specifics>
## Specific Ideas

- North star (invariant 6): Omarchy reproducible as a speech on vanilla Arch with zero install-time questions — the phase gate.
- SC-3 visibility rule: unsupported required opinion → composition-time failure with a human-readable message naming the opinion and missing capability; never silent install-time failure.
- Use the pinned Omarchy commit (9cf1852) evidence only — no re-derivation from blog posts.
- Docker available, passwordless sudo, NO QEMU on host (bootability gate is structural; document QEMU smoke as optional).

</specifics>

<deferred>
## Deferred Ideas

- Full variant ISO boot validation (CachyOS/Garuda) — non-gating stretch, post-v1.0 completion.
- GitHub Actions reusable workflow + published Docker image — Phase 3.
- Direct-to-disk install target — post-v1.0.

</deferred>
