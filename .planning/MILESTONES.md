# Milestones

## v1.0 DebateOS MVP (Shipped: 2026-06-13)

**Phases completed:** 6 phases, 29 plans, 42 tasks

**Key accomplishments:**

- 134 atomic OM-NNN opinions covering all 6 Omarchy install phases at commit 9cf1852, with category, OS-agnostic intent, source citations, hardware conditions, and 7 identified schema surprises
- 134 OM-NNN opinions grouped into 32 evidence-driven candidate points; 22 OS-agnostic SR-NNN schema requirements derived with full evidence backing including all 7 confirmed schema surprises
- CachyOS + Garuda delta catalog from 6 freshly cloned repos with declarative
- 27 EC-NNN Given/When/Then resolver scenarios (full docs/04 coverage, all 6 collision classes) plus 10 open questions seeding the Phase 1 TDD harness and schema design.
- Kahn toposort with container/heap lexicographic tie-breaking, BuildGraph from Opinion ordering edges, and cycle detection naming offending opinions — all built test-first against EC-035/EC-036/cross-phase/determinism gates.
- 1. `hardware.HardwareProfile` as a distinct package-local struct
- 1. [Rule 3 - Design] EC-038 PCIIDs not decodable from resolver.HardwareProfile
- `main.go`
- Python package skeleton with SC-3 capability gate (CapabilityError + 29-token capabilities.json), contract loaders (ResolvedSpeech JSON + opinion bodies), and BuildManifest dataclass aggregating the full payload in install_order with deterministic SOURCE_DATE_EPOCH — 43 pytest GREEN, RED-before-GREEN commits for both TDD tasks (D19)
- Archiso profile tree emitter (ARCH-01) + variant application (ARCH-04) on top of Plan 01 BuildManifest: load_variant_profile, apply_variant (above_core ordering, keyring-first, trust-warning capture), surface_conflicts, emit_profile_tree (profiledef.sh/packages.x86_64/pacman.conf/installer 0755/.zlogin/first-run units/build-manifest.json), file_asset dst sanitization (T-02-08), and argv-stable translate entrypoint — 128 pytest GREEN, RED-before-GREEN for all three TDD tasks (D19)
- Three declarative variant profile YAML files (vanilla-arch, cachyos, garuda) authored from the verified delta study — ARCH-04 satisfied: one generator, no per-variant code forks; garuda's four hard Omarchy conflicts captured as structured data; all [UNVERIFIED] items tagged in-line; README documents schema and invariant
- 134 OS-agnostic OM-NNN opinion YAMLs + 32 curated points + one vanilla-Arch speech that resolves clean (Applied=99, Skipped=35, Hard conflicts=0), proving ARCH-02 with a CC0-licensed, script-generated, schema-valid example corpus.
- Docker mkarchiso build script + ISO structural validator + `cmd/resolve-json` Phase 3 CLI seed + north-star equivalence gate (16/16 PASS on --skip-build) + translator README documenting the frozen input contract, capabilities, variant profiles, and optional QEMU smoke step, with ARCH-01..04 marked Complete.
- stdlib flag.FlagSet subcommand dispatch with testable Run() int exit codes, DEBATEOS_DIR-first config resolution, Runner/FakeRunner interface, and compose/validate subcommands backed by the shared resolve pipeline — filippo.io/age v1.3.1 promoted to direct dep for pane plan.
- Age X25519 identity local-only secrets management with 0600-gated pane.yaml, FakeRunner-isolated git backup/restore, and lossless age encrypt/decrypt round-trip.
- debateos build resolve→epoch→translate→docker via FakeRunner-testable Runner with --dry-run/--skip-iso gates; private-injection.tar emitted locally with sanitized target-relative paths and debateos-private.json manifest.
- Multi-stage Docker image (CGO_ENABLED=0 + digest-pinned archlinux) + GHA reusable workflow + determinism/secret-free/coverage gates all passing; Phase 3 finalized in REQUIREMENTS.md and ROADMAP.md.
- Foundation-neutral shared translator core (contract loaders, BuildManifest, first-run renderer) extracted to translators/common/ with arch re-exporting via shims; all regression gates green.
- Foundation-neutral dual-foundation speech (5 opinions, 2 points, foundation: debian) with TestExampleDualFoundation clean-resolve gate proving DEB-02.
- Debian translator with capability gate (45 tokens, no mkinitcpio/limine), apt sig_level → signed-by/trusted=yes mapping, and live-build config/ tree emitter (preseed.cfg + chroot hook + package-lists) via RED/GREEN TDD.
- `foundationRegistry` in `cli/build/build.go` dispatches `translators/arch/translate` or `translators/debian/translate` by `rs.Foundation`, plus complete DEB-03 6-finding audit confirming `build.go` is the only genuine leak
- DEB-02 dual-foundation proof gate (20/20 PASS), Debian lb-build Docker wrapper + digest-pinned Dockerfile + translator README (DEB-01), and COMM-01 ownership model with how-to-add-a-translator; DEB-01/02/03/COMM-01 all Complete; Phase 4 finalized.
- Go static registry generator (REG-01) — YAML → deterministic JSON index + static browse HTML with foundation-compat from capabilities.json; go.mod owns all Phase-5 deps (chi/sqlite v1.46.1/oauth2).
- SvelteKit 5 + adapter-static + Tailwind v4 @theme token contract, typed Go-WASM loader with invariant-3 guard, dual-delivery BASE_PATH seam, BRND-01 landing page, and build-wasm.sh.
- Forum storage core: sqlc-generated queries over modernc.org/sqlite FTS5 with subscription/rating store and chi read API gated on injected IdentityFn — all tested against in-memory SQLite.
- WASM-driven conflict visualization (triple-encoded ConflictOverlay), pure mapExplanation() with verbatim A9 text, ResolutionPanel + export screen, 13 Playwright e2e tests all GREEN (A1/A2/A3/A6/A7/A9).
- GitHub OAuth web flow behind OAuthProvider interface (fake in tests, CSRF-validated, token discarded), conflict thread CRUD with patch PR URLs, idempotent Reindex from registry index for DB-loss recovery, and single arm64-buildable forumctl binary.
- go:embed dual-delivery SvelteKit (BASE_PATH= embed, BASE_PATH=/debateos Pages), compose --serve offline UI, four-gate coverage all passing (resolver 93.5%, cli 85.8%, registry 85.4%, forum 85.3%), invariant-4 offline guard, GitHub Actions registry CI, Oracle A1 deployment notes, end-to-end docs, REG-01 Complete — v1.0 milestone closed.

---
