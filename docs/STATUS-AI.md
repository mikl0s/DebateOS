# DebateOS — Machine-Readable Status Snapshot

> AUDIENCE: AI assistants. Condensed facts, no narrative. Updated: 2026-06-12. Source of truth for decisions: docs/09-decisions.md (D1–D20 + 7 invariants). Repo: github.com/mikl0s/DebateOS. Module: github.com/mikl0s/debateos (go 1.24).

## PROJECT
- Product: DebateOS — compose OS configurations ("speeches") from curators' atomic OS-agnostic "opinions" grouped in "points"; resolver detects/explains conflicts; per-foundation "translators" emit unattended installer ISOs. Zero-cost infra; Git/GitHub = registry; optional Forum.
- v1.0 = Phases 0–5 (sequential, each gates next). Phase 6 (hw-scanning installer) + Fedora/GitLab/post-install-reconciliation = post-v1.0, forbidden in v1.0.
- Methodology: TDD mandatory (D19): RED commit before GREEN, resolver coverage ≥90% enforced; WASM/native parity proven by automated tests.
- North star (invariant 6): Omarchy reproducible as a speech on vanilla Arch, zero install-time questions.

## PHASE STATUS
| Phase | Name | Status |
|---|---|---|
| 0 | Omarchy Research & Arch-Variant Study | COMPLETE (verified) |
| 1 | Schema & Resolver Core | COMPLETE (verified) |
| 2 | Arch Translator | PLANNED (5 plans, 3 waves) — executing next |
| 3 | CLI & Build Channels | not started |
| 4 | Debian Translator | not started |
| 5 | Registry, Forum & Debate UI | not started |

## PHASE 0 OUTPUTS (research/)
- omarchy-opinion-inventory.md: 134 atomic opinions (OM-001..OM-134) from basecamp/omarchy @ 9cf1852 (4.0.0.alpha, no tags). 6 install phases: preflight→packaging→config→login→post-install→first-run (ordering load-bearing). 38 hardware-conditional. 13 first-run scripts.
- omarchy-points.md: 32 point groupings, full OM coverage, no orphans.
- schema-requirements.md: SR-001..SR-022 schema floor (status, deps/conflicts/patches, compound hw predicates AND/OR/NOT+set-membership, install_phase+ordering, translator capabilities, file assets, custom repos w/ 4 trust levels (sig_level incl. Never), runtime tool installs (npm), execution_phase install-time|first-run, script payloads, DM, bootloader, services(deferred), sysctl, kernel params, groups, MIME, themes, point/speech shapes).
- arch-variants-delta.md: CachyOS = layered perf repos (cachyos,v3,v4 ABOVE arch repos), linux-cachyos kernel family, sysctl pre-seeds, x86-64-v3/v4 binaries. Garuda = [garuda]+[chaotic-aur], MANDATORY dracut (conflicts mkinitcpio) + GRUB (vs Omarchy limine) + btrfs/snapper + Dr460nized theming pre-seeded. 4–5 hard conflicts Garuda↔Omarchy. Declarative variant-profile YAML sketch (repos+keyring+kernel+defaults+pre-seeded opinions).
- resolver-edge-cases.md: 27 EC scenarios (Given/When/Then + expected explanation + provenance), covers 4 docs/04 rules + ordering/cycles/hw behaviors; seeded Phase 1 tests 1:1.
- open-questions.md: 10 OQs (migrations primitive deferred post-v1.0; execution-phase + runtime-tool-install became schema fields).

## PHASE 1 OUTPUTS (code, all tests green)
- schemas/: opinion|point|speech.schema.json (JSON Schema 2020-12, additionalProperties:false, schema:1 required), README traceability SR-001..022, CC0.
- resolver/types.go: Opinion/Point/Speech/HardwareExpr + sub-types; no float fields (canonical JSON parity).
- resolver/parse: strict YAML (KnownFields) + embedded schema validation. resolver/graph: Kahn toposort, deterministic (phase order then lex ID), cycles name opinions. resolver/hardware: recursive EvalCondition (depth cap 32), profile = predicates/facts/pci_ids. resolver/patch: FindPatch over known_patches. resolver/resolve: docs/04 hierarchy exact (required>nice; req-vs-req hard unless patch→offered; nice-vs-nice deterministic default; patches override), accumulated multi-conflict errors, sysctl per-key collision detection, SigLevel=Never trust warnings, hardware-skip swap suggestions (in-composition, AlternativeSuggestion field), Explanation on every decision, CanonicalJSON deterministic.
- resolver/wasm: js/wasm entrypoint debateosResolve. Parity: scripts/wasm-parity-test.sh (golden files, byte-identical native vs WASM). Coverage gate: scripts/check-coverage.sh (93.5% ≥ 90%).
- examples/: 4 evidence-derived compositions (omarchy-mini, two-point-clean, conflicting, hardware-conditional) + Go e2e harness.
- Deps: go.yaml.in/yaml/v3 v3.0.4 (maintained yaml.v3 fork) + santhosh-tekuri/jsonschema/v6 only. NO SAT solver (D6).
- Code review done, 8 findings fixed (phase tie-break, pci_ids plumbing, multi-conflict accumulation, trust-warning scope, hw depth limit, sysctl-collision continuation).

## PHASE 2 PLAN (Arch Translator — current work)
Research verified live in archlinux Docker (archiso 88-1):
- Baseline: releng profile copy; autologin + .zlogin hook launches /root/debateos-install.sh; profiledef.sh/packages.x86_64/airootfs structure.
- Build: mkarchiso in --privileged archlinux:base-devel (digest-pinned) container; SOURCE_DATE_EPOCH natively supported (derive from resolved-speech sha256) → deterministic ISO.
- Install model: live ISO stays minimal (~releng+installer deps, ~800MB–1.2GB); resolved speech packages (~285) pacstrapped at install time w/ network; first-run opinions → flag-file oneshot systemd units.
- Capability gate (ARCH-03): translators/arch/capabilities.json declares 34 tokens derived from OM corpus; required+unsupported → loud CapabilityError (names opinion+capability) at composition time; nice-to-have+unsupported → visible drop.
- Input contract (frozen for Phase 3 CLI): `translate <resolved.json> --opinions <dir> --profile <name> --out <dir>` (ResolvedSpeech carries IDs+explanations only; opinion bodies come from --opinions dir).
- Variant profiles (ARCH-04/D20): translators/arch/profiles/{vanilla-arch,cachyos,garuda}.yaml — declarative (repos+keyring+kernel+defaults+pre-seeded opinions), ONE generator no forks; garuda.yaml marks 4 hard conflicts vs Omarchy; [UNVERIFIED] repo URLs tagged. Variant ISO boot = non-gating stretch.
- Plans: 02-01 core TDD (loader+capability gate+manifest) | 02-03 variant YAML | 02-04 examples/omarchy authoring (134 opinions→32 points→speech, schema-validated, resolves clean) [wave 1, parallel] → 02-02 profile emitter TDD (+translate entrypoint) [wave 2] → 02-05 Docker build + ISO structural validation (xorriso/unsquashfs; no QEMU on host) + north-star rootfs equivalence checks [wave 3].
- North-star gate (ARCH-02): resolve examples/omarchy → translate → profile contains expected packages/file-assets/first-run units → ISO builds in Docker → structural validation passes. UX equivalence = manual-only (documented).

## ENVIRONMENT
- Host: Go 1.24.1, Node 24, Docker running, passwordless sudo, NO QEMU. wasm_exec at $(go env GOROOT)/lib/wasm/.
- Quality gates per phase: code review (gsd) + fixes, goal-backward verification, TDD review (RED/GREEN commit audit).

## INVARIANTS (always enforced)
1 opinions OS-agnostic / 2 docs/04 hierarchy + patches first-class / 3 human-readable YAML+explanations / 4 no central service in build path / 5 zero cost / 6 Omarchy north star / 7 private pane never leaves user.
