# 11 — Repository Layout (Monorepo)

Create this structure as the project is built. Directories appear as their phase delivers them; the tree below is the v1.0 target. Keep Go code under a single module rooted at the repo (e.g. `module github.com/<owner>/debateos`).

```
debateos/
├── README.md                      # project overview (points at docs/)
├── LICENSE                        # AGPL-3.0 (code)
├── go.mod / go.sum                # single Go module for resolver + cli + forum
│
├── docs/                          # THIS folder — founding context (source of truth)
│
├── research/                      # Phase 0 outputs
│   ├── omarchy-opinion-inventory.md
│   ├── omarchy-points.md
│   ├── schema-requirements.md
│   └── open-questions.md
│
├── schemas/                       # Phase 1 — Opinion/Point/Speech YAML schema
│   ├── LICENSE                    # CC0-1.0 (schemas are content)
│   ├── opinion.schema.yaml
│   ├── point.schema.yaml
│   └── speech.schema.yaml
│
├── resolver/                      # Phase 1 — Go resolver library (native + WASM targets)
│   ├── parse/  graph/  resolve/  patch/  hardware/
│   └── wasm/                       # WASM entrypoint (js/wasm build tag)
│
├── cli/                           # Phase 3 — Go CLI (debateos compose|validate|build|pane)
│   └── embed/                      # embedded static Debate UI assets (from web/ build)
│
├── translators/                   # Phase 2 & 4 — shell/Python per foundation
│   ├── arch/                       # wraps mkarchiso  (+ structure for arch variants)
│   └── debian/                     # wraps live-build / preseed
│
├── build/                         # Phase 3 — build channels
│   ├── docker/                     # Dockerfile: resolver + translators + ISO tooling
│   └── actions/                    # reusable GitHub Actions workflow + template repo docs
│
├── web/                           # Phase 5 — Visual Debate UI (SvelteKit + adapter-static + Tailwind)
│   └── src/                        # calls the Go-WASM resolver; static build → Pages AND cli/embed
│
├── registry/                      # Phase 5 — static index generator (→ GitHub Pages)
│
├── forum/                         # Phase 5 — The Forum: Go (chi) + embedded SQLite discovery service
│   ├── api/  index/                # read-mostly GitHub indexer, OAuth (GitHub), ratings/threads
│   ├── store/                      # repository interface + sqlc queries (SQLite default, Postgres optional)
│   ├── migrations/                 # SQL schema + FTS5 (sqlc-compatible; portable to Postgres)
│   └── deploy/                     # Oracle A1 / VM notes: single static arm64 binary + one .db file
│
├── examples/                      # Phase 1+ — example opinions/points/speeches
│   ├── LICENSE                     # CC0-1.0
│   └── omarchy/                    # the Omarchy-as-a-speech composition (Phase 2 north star)
│
└── .github/workflows/             # CI: lint/test Go, build WASM, validate schemas/examples
```

## Module & build notes

- **One Go module** covers `resolver/`, `cli/`, `forum/`. The resolver builds two ways: native (default) and `GOOS=js GOARCH=wasm` (the `resolver/wasm` entrypoint) for the web UI.
- **`web/` build output** is consumed twice: published to GitHub Pages (registry + composer) and embedded into the CLI via `go:embed` under `cli/embed/` so `debateos compose` can serve it offline.
- **Translators** are intentionally outside the Go module (shell/Python); the Docker build image and CLI invoke them as subprocesses with a defined input contract (the resolved-speech JSON/YAML).
- **Licensing split:** `LICENSE` (AGPL-3.0) at root governs code; `schemas/LICENSE` and `examples/LICENSE` (CC0-1.0) govern schema + content so opinions/points/speeches stay maximally remixable.

## Community repos (out of this monorepo)

Curators' point/speech repositories live as **separate GitHub repos** (plain YAML), discovered via the static registry index and The Forum. The monorepo ships the tooling, schema, translators, reference examples, and the Forum service — not the community content.
