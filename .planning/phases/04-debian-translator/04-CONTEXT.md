# Phase 4: Debian Translator - Context

**Gathered:** 2026-06-13
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous — recommended answers auto-accepted per owner directive + ADR process notes)

<domain>
## Phase Boundary

Phase 4 delivers `translators/debian/`: a translator (shell/Python per D8) that consumes the SAME resolved-speech contract as the Arch translator and emits a bootable, fully-unattended Debian installer by wrapping live-build/preseed, built inside Docker — plus the dual-foundation proof (one representative speech → both Arch and Debian installers), the leaked-Arch-assumption audit + fixes (DEB-03), and the translator ownership model documentation (COMM-01).

Out of scope: registry/Forum/UI (Phase 5), hardware-scanning installer (post-v1.0), Fedora/other foundations (post-v1.0/community), making Omarchy itself build on Debian (Omarchy is Arch-specific by nature — the dual-foundation proof uses a deliberately foundation-neutral representative speech, NOT Omarchy).

</domain>

<decisions>
## Implementation Decisions

### Translator Structure (DEB-01/DEB-02, mirrors Arch)
- Same Python-generator + thin-shell pattern as translators/arch. Reuse the foundation-neutral generator logic where it already is (manifest building, capability gating, first-run unit emission, dst sanitization) — extract any Arch-specific bits the Phase 2/code-review audit reveals into translator-local code, keep the shared core shared. Prefer a small shared Python module (e.g. translators/common/) over copy-paste if the refactor is clean; otherwise document the duplication and its rationale.
- Same argv-stable entrypoint: translators/debian/translate <resolved.json> --opinions <dir> --profile <name> --out <dir>.
- Wraps live-build (lb config/lb build) producing a Debian live/installer ISO with a preseed for fully-unattended install. Preseed (debian-installer) drives partitioning/base install; late_command (or a first-boot systemd unit) applies opinions (packages via apt, file assets, services, sysctl, kernel params, groups). execution_phase first-run → systemd first-boot units (same flag-file pattern as Arch).
- translators/debian/capabilities.json declares supported opinion categories/capability tokens; required+unsupported → loud CapabilityError at composition time (SC-2); nice-to-have+unsupported → visible drop. Debian capability set will legitimately differ from Arch (e.g. no pacman repos; apt sources instead) — that is the POINT of the dual-foundation proof.

### Variant Profiles (Debian family)
- translators/debian/profiles/: at minimum debian.yaml (stable). Structure mirrors the Arch variant-profile schema (repos→apt sources, keyring, kernel pkg, defaults fs/bootloader). One generator, no per-variant fork. Ubuntu/other Debian-family variants are post-v1.0 (the profile structure must accommodate them, like Arch accommodates CachyOS/Garuda).

### Dual-Foundation Proof (DEB-01, the gate)
- A representative foundation-neutral speech in examples/dual-foundation/: a handful of points using opinions whose intent is genuinely OS-agnostic (e.g. a base CLI toolset, a desktop bundle, a service) — chosen so BOTH the Arch and Debian translators can effectuate every REQUIRED opinion. Authored from scratch (not Omarchy), schema-valid, resolves clean.
- Gate: resolve the representative speech ONCE → run translators/arch/translate AND translators/debian/translate on the SAME resolved.json → both emit valid profile trees; mechanical equivalence checks per foundation (Arch: pacman package set + assets; Debian: apt package set + preseed + assets). ISO build is the slow gate (Docker); on THIS host the Debian live-build path may hit the same privileged/loop limitations as mkarchiso — if so, the gate is profile-emission + structural validation where buildable, with full ISO build documented as deferred-to-capable-host (same policy as Phases 2/3). Verify which works in research.

### Leaked-Arch-Assumption Audit (DEB-03)
- Audit schema fields, resolver logic, and the shared generator for Arch-specific assumptions surfaced by implementing a second foundation. Likely suspects: package naming assumptions, repo model (pacman SigLevel vs apt signed-by), bootloader/initramfs naming, install_phase enum values that are Arch-pipeline-shaped, mkinitcpio vs initramfs-tools. Fix genuinely-leaked ones (schema/resolver changes are allowed and EXPECTED here — this is the phase that earns invariant 1); document each finding + fix. Anything that's correctly translator-owned (not leaked) is recorded as "correctly isolated."
- Any schema change re-runs the full Go + existing-example test suites (no regressions); the Arch translator + examples/omarchy must still pass.

### Ownership Model Docs (COMM-01)
- docs/ownership-model.md (or a section in translator READMEs): distributions own their translators (Ubuntu controls Ubuntu's), curators own points/speeches, community PRs welcome; how to add a new translator (the entrypoint contract + capabilities.json + profile schema). Concise, AI-and-human readable.

### TDD (D19)
- pytest for the Debian generator/preseed-emission logic; RED before GREEN. Reuse the Arch test patterns. Dual-foundation equivalence + ISO builds are gated scripts (scripts/dual-foundation-check.sh, scripts/debian-build-iso.sh, scripts/debian-validate-iso.sh) run at wave/phase verification.
- Full Go suite + Arch pytest + examples must stay green after any shared-code/schema change (regression gate).

### Claude's Discretion
- Whether to extract translators/common/ now or document duplication; live-build profile internals; preseed template structure; exact representative-speech contents.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- translators/arch/* (manifest.py, capabilities.py, contract.py, variant.py, firstrun.py, profile.py, translate) — the pattern + the foundation-neutral pieces to share.
- resolver/* (resolved-speech contract), cli/build (build subcommand — must become foundation-aware: foundation field → translator dir selection, data-driven not hardcoded; check current state and wire Debian in).
- scripts/* gate pattern; examples/ harness (Go) pattern for the representative speech.
- research/arch-variants-delta.md (model for a future debian-variants delta if needed).

### Established Patterns
- TDD RED/GREEN; coverage gates; deterministic builds (SOURCE_DATE_EPOCH); minimal deps; environment-blocked ISO documented; invariant 1 (translator owns mechanics).
- Licensing: translator code AGPL (root LICENSE); examples CC0 (examples/LICENSE).

### Integration Points
- cli/build must select the translator by the speech's foundation field (Phase 3 left this — verify; if hardcoded to arch, make it data-driven now: foundation "arch*"→translators/arch, "debian"/"ubuntu"→translators/debian).
- Phase 5 (registry/UI/Forum) consumes nothing new structural here, but foundation-compatibility metadata (which foundations a point/speech supports) may surface — keep capability data machine-readable.

</code_context>

<specifics>
## Specific Ideas

- The whole point of Phase 4: prove the abstraction is real by building a SECOND foundation and fixing whatever Arch-shaped leaks that surfaces — invariant 1 is earned here, not assumed.
- Dual-foundation proof speech is foundation-NEUTRAL and authored fresh (Omarchy is Arch-specific — do not try to build it on Debian).
- Debian capability set differs from Arch by design; the capability gate makes unsupported required opinions fail loudly at composition time on each foundation.
- Host limits: Docker yes, no QEMU, Proxmox devtmpfs blocks privileged loop mounts — Debian live-build likely needs the same capable-host deferral; confirm in research and document.
</specifics>

<deferred>
## Deferred Ideas

- Ubuntu + other Debian-family variant profiles → post-v1.0 (structure must accommodate).
- Full Debian ISO build on this host if devtmpfs-blocked → deferred-to-capable-host (documented), like Phases 2/3.
- translators/common/ extraction if not cleanly doable in-phase → document duplication, defer refactor.

</deferred>
