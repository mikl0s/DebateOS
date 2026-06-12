# 07 — Roadmap (v1.0 = Phases 0–5)

v1.0 spans **Phases 0 through 5**. Phase 6 is documented for continuity but is **post-v1.0** and must not be built in this run. Each phase lists its goal, deliverables, success criteria, and gating dependency. Phases are sequential unless noted; Phase 0 gates everything.

---

## Phase 0 — Omarchy Research (gates everything)

**Goal:** Derive the baseline opinion schema from real-world data, not theory.

**Deliverables** (full brief in `08-omarchy-research.md`):
- `research/omarchy-opinion-inventory.md` — every post-base-Arch-install decision as a candidate atomic opinion, with category + metadata observations.
- `research/omarchy-points.md` — proposed natural point groupings.
- `research/schema-requirements.md` — minimum expressive surface an opinion/point/speech schema must support, evidenced by the inventory.
- `research/open-questions.md` — surprises and anything that can't be made OS-agnostic (→ translator capability requirements).

**Success criteria:** A complete, evidence-backed inventory exists; the minimum metadata surface is justified by real Omarchy decisions; schema design can begin with a known floor.

**Gates:** Phase 1 (schema). Do not design the schema before this is done.

---

## Phase 1 — Schema & Resolver Core

**Goal:** A versioned YAML schema and a working Go resolver.

**Deliverables:**
- `schemas/` — Opinion / Point / Speech YAML schema, grounded in Phase 0 findings + the conflict model (`04`).
- `resolver/` — Go library: parse, validate, dependency graph, conflict detection, the resolution hierarchy, patch application, hardware-aware checks. Compiles native; WASM build target wired up.
- **Conflict test harness** — sample speeches with deliberate conflicts: required-vs-required, hardware mismatch, version clash, and at least one patchable pair.
- 3–4 example opinion/point/speech files, including one deliberately conflicting composition.

**Success criteria:** The resolver correctly resolves every test-harness scenario per the `04` rules and explains each resolution; the WASM build produces identical results to native.

**Gates:** Phase 2 (translators consume resolved speeches).

---

## Phase 2 — Arch Translator (proof of concept)

**Goal:** Turn resolved speeches into real Arch installers and **reproduce Omarchy**.

**Deliverables:**
- `translators/arch/` — translate resolved speeches into concrete Arch build instructions; wrap `mkarchiso` for unattended installer output. Structured so 1–2 Arch variants can follow closely.
- A DebateOS speech that expresses Omarchy as points + opinions on vanilla Arch.

**Success criteria (NORTH STAR):** **Omarchy is reproducible as a speech on vanilla Arch** — building the speech produces an installed system equivalent to Omarchy. If this works, the model is validated.

**Gates:** Confidence to generalize (Phases 3–5).

---

## Phase 3 — CLI & Build Channels

**Goal:** A user-facing CLI and both build channels.

**Deliverables:**
- `cli/` — Go CLI: `compose`, `validate`, `build`; private-pane management in `$HOME`; ability to serve the embedded Debate UI on localhost.
- `build/docker/` — Docker build image (resolver + translators + ISO tooling); the local/private build path.
- `build/actions/` — reusable GitHub Actions workflow; the distributed-CI build path. Same Docker image internally.
- Secrets model implemented per `05` (first-boot injection; no secrets in shared artifacts).

**Success criteria:** A user can build the Omarchy speech end-to-end via both Docker and GitHub Actions, deterministically.

---

## Phase 4 — Debian Translator

**Goal:** A second foundation, to prove the abstraction is real.

**Deliverables:**
- `translators/debian/` — translate resolved speeches into Debian build instructions; wrap `live-build`/preseed for unattended installer output.
- Schema/translator-capability adjustments surfaced by Debian (documented).

**Success criteria:** A representative speech builds installers for **both** Arch and Debian from the same resolved input; anything that leaked Arch-specific assumptions into the schema is identified and fixed. (One translator is a build script; two is an architecture.)

---

## Phase 5 — Registry, The Forum & Visual Debate UI

**Goal:** Discovery, reputation, collaborative conflict resolution, and the visual composition experience. Adoption-critical.

**Deliverables:**
- `registry/` — generator for the static registry index (points/speeches discovery) on GitHub Pages.
- `web/` — Visual Debate UI: SvelteKit + `adapter-static` + Tailwind; foundation + glass panes, red/green conflict overlays, suggested resolutions, hardware-aware checks; runs the Go-WASM resolver client-side. Built static for GitHub Pages **and** embeddable in the CLI.
- `forum/` — The Forum: Go (chi) + embedded SQLite discovery service per `05`/`06` (thin swappable `store` interface + sqlc; SQLite default, Postgres optional later; read-mostly GitHub index, GitHub OAuth only, no code execution, rebuildable). Search (FTS5), subscriptions, ratings, conflict-resolution threads.
- Deployment notes for hosting The Forum on a small VM (single static binary + one SQLite file; no separate DB server).

**Success criteria:** A user can discover points via The Forum, compose a speech in the Visual Debate UI with live conflict visualization, and proceed to build — with the core path still functional if The Forum is offline.

---

## Phase 6 — Hardware-Scanning Installer (POST-v1.0 — DO NOT BUILD)

Installer-side hardware detection: kernel/driver selection at install time, fully zero-question flow ("click → install Linux on this machine → it just works"). Documented for continuity only; out of scope for this run.

## Later / community (post-v1.0)

- Additional translators (Fedora, others) — community-owned; distributions invited to own theirs.
- Direct-to-disk install target.
- Full GitLab parity.
- Post-install reconciliation (applying speech changes to a running system).
