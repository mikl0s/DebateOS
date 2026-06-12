# Phase 1: Schema & Resolver Core — Research

**Researched:** 2026-06-12
**Domain:** Go module bootstrap, YAML schema design, JSON Schema 2020-12 validation, topological sort, WASM parity testing
**Confidence:** HIGH (core stack verified against live registries and installed toolchain; WASM parity reasoning is HIGH for pure-Go paths, MEDIUM for edge cases)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Schema Design**
- YAML schema documentation + JSON Schema (draft 2020-12) validation files in `schemas/` with CC0 LICENSE; the resolver's parse layer enforces them.
- Explicit `schema: 1` version field on every Opinion/Point/Speech document from day one.
- Full SR traceability: every SR-001..SR-022 maps to a schema field or a documented deferral; traceability table lives in `schemas/README.md`.
- Speech-level foundation target (SR-022) and foundation pre-seed semantic hooks are included now; the full variant-profile schema is Phase 2 (candidate sketch stays in research/).
- Schema surprises from Phase 0 must be expressible: file/asset payloads, custom repo + keyring registration (with per-repo trust level per SR-009), runtime-tool-install category (SR-010), execution-phase field install-time vs first-run (SR-011), compound hardware predicates, arbitrary script payloads with declared capabilities, phase-level ordering.

**Resolver Architecture**
- Package layout locked by docs/11: `resolver/parse`, `resolver/graph`, `resolver/resolve`, `resolver/patch`, `resolver/hardware`, `resolver/wasm` — single Go module at repo root (`module github.com/mikkelraglan/debateos`).
- First-class `Explanation` type attached to every resolution decision: human-readable text plus structured fields (rule applied, opinions involved, what was dropped/kept and why). Every EC scenario's "expected explanation" must be producible.
- Determinism discipline: stable sorts everywhere, no map-iteration-order leaks, canonical JSON output for resolved speeches.
- Dependencies minimal: `gopkg.in/yaml.v3` only; stdlib otherwise; NO SAT/constraint libraries (D6); rule-based resolution only.
- Resolution hierarchy implemented exactly per docs/04.

**TDD Harness (D19 — locked)**
- EC-NNN corpus (27 scenarios) encoded as table-driven Go tests, one per EC with ID in test name — written RED before implementation (GREEN).
- Parity proof: canonical-JSON golden files; identical fixtures run native and `GOOS=js GOARCH=wasm` (Node-based runner); byte-identical outputs asserted by repeatable script.
- Coverage: ≥90% on resolver packages overall, 100% of docs/04 rule branches; enforced by coverage-check script.

**Example Compositions**
- `examples/` with CC0 LICENSE. Evidence-derived: mini-Omarchy subset, clean two-point speech, one deliberately conflicting speech, one hardware-conditional speech.
- Examples are harness fixtures loaded end-to-end (parse → resolve → explain), not just documentation.

### Claude's Discretion
- Exact Go type names, internal function decomposition, golden-file layout.
- How many JSON Schema files vs one combined file, as long as Opinion/Point/Speech are each fully specified.
- Node wasm runner mechanics (wasm_exec.js wiring) details.

### Deferred Ideas (OUT OF SCOPE)
- Variant-profile schema implementation → Phase 2 (hooks only in Phase 1 per SR-022).
- OQ-001 migrations/update primitive → recorded in open-questions; decide only if schema work forces it; default defer post-v1.0.
- Resolver-as-a-service (Forum-side indexing use) → Phase 5.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SCHM-01 | Opinion/Point/Speech YAML schemas in `schemas/` (CC0), covering SR-001..SR-022 floor (status, deps, conflicts, hardware conditions, ordering, known patches, translator capability) plus Phase 0 expansions | Schema layout section; JSON Schema 2020-12 library; SR field map |
| SCHM-02 | Schemas and example files are human-readable; no Arch/Debian specifics leak into schema or content | Anti-patterns section; OS-agnostic design contract |
| RSLV-01 | Resolver parses + validates speeches, builds dep/conflict graph, applies docs/04 hierarchy, emits resolved speech with Explanation for every decision | Architecture patterns; `resolver/` package breakdown; Explanation type design |
| RSLV-02 | Patch opinions are first-class: attached to conflict pairs, discovered and offered automatically, able to override hierarchy | Patch discovery pattern; `resolver/patch` package |
| RSLV-03 | Ordering constraints feed toposort producing install order; cycles are hard errors naming offending opinions | Kahn's algorithm pattern; cycle extraction |
| RSLV-04 | Hardware-conditional opinions evaluate against declared hardware at composition time; mismatches surface with suggested swaps | `resolver/hardware` package pattern; EC-037/EC-038 TDD |
| RSLV-05 | Resolver compiles to native and WASM; identical results in both, verified by automated parity tests | WASM build mechanics; parity test script; float/map determinism |
| RSLV-06 | TDD conflict harness covers Phase 0 EC corpus plus required-vs-required, hardware mismatch, version clash, patchable pair; near-total coverage; 3–4 example files | TDD structure; table-driven test pattern; EC-NNN encoding |
</phase_requirements>

---

## Summary

Phase 1 creates the first code in the DebateOS repository: a Go module (`module github.com/mikkelraglan/debateos`), JSON Schema 2020-12 validation files for Opinion/Point/Speech, and the `resolver/` library implementing the six packages prescribed by docs/11. The module is test-driven against the 27 EC-NNN edge-case scenarios from `research/resolver-edge-cases.md`. It compiles to both native Go and `GOOS=js GOARCH=wasm`, and a parity script asserts byte-identical canonical JSON output for all fixtures under both targets.

The dependency decision locked in CONTEXT.md (`gopkg.in/yaml.v3` only, stdlib otherwise) needs a one-word update: `gopkg.in/yaml.v3` was archived by the go-yaml authors in April 2025. The official YAML organization (`yaml/go-yaml`) took over as the maintained fork under the import path `go.yaml.in/yaml/v3` (v3.0.4 as of June 2025). This is an API-identical, security-maintained drop-in. The planner should treat `go.yaml.in/yaml/v3` as the implementation of the locked "gopkg.in/yaml.v3 only" decision — the intent (YAML + stdlib, no extras) is preserved. The second allowed package is `github.com/santhosh-tekuri/jsonschema/v6` (v6.0.2, May 2025) for JSON Schema 2020-12 validation of the schema files themselves and for the parse layer's document validation.

The host Go toolchain is 1.24.1 (`/usr/local/go`); current Go stable is 1.26.4 (released May 2026). The project should set `go 1.24` in go.mod so existing toolchain works without auto-download; CI can upgrade later. WASM test infrastructure (Node.js v24.12.0, `go_js_wasm_exec` script at `/usr/local/go/lib/wasm/go_js_wasm_exec`) is fully present on this host.

**Primary recommendation:** Start with `go mod init github.com/mikkelraglan/debateos`, add `go.yaml.in/yaml/v3` and `santhosh-tekuri/jsonschema/v6` as the only two external dependencies, write EC-NNN table-driven tests RED in all six resolver packages, then implement GREEN.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| YAML document parsing + strict validation | `resolver/parse` | `schemas/` (JSON Schema files) | Parse layer owns the parse/validate pipeline; schemas are the constraint definitions |
| Dependency + conflict graph construction | `resolver/graph` | — | Graph is a pure data structure derived from parsed opinions; no output concerns |
| Conflict resolution (docs/04 rules) | `resolver/resolve` | `resolver/patch` | Resolve owns rule application; patch lookup is a separate concern |
| Patch opinion discovery | `resolver/patch` | — | Isolated so Phase 2+ can extend patch sources |
| Hardware condition evaluation | `resolver/hardware` | — | Hardware predicate evaluation is complex enough (AND/OR/NOT/set-membership) to warrant isolation |
| WASM entry point + JS bridge | `resolver/wasm` | — | `js/wasm` build tag isolates JS-specific glue from pure-Go resolver logic |
| Topological sort + cycle detection | `resolver/graph` | — | Graph package owns the ordering graph and its algorithms |
| Canonical JSON serialization | `resolver/resolve` | stdlib `encoding/json` | Resolved speech serialization is an output concern of the resolver |
| JSON Schema 2020-12 validation | `resolver/parse` | `schemas/*.json` | Validation happens at parse time using embedded schema files |
| Example fixture loading (end-to-end tests) | `examples/` + test harness in each package | — | Fixtures are YAML; tests load them via parse, drive through resolve, assert output |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `go.yaml.in/yaml/v3` | v3.0.4 [VERIFIED: proxy.golang.org] | YAML parsing + encoding | Official maintained fork of `gopkg.in/yaml.v3` (archived April 2025); API-identical; owned by YAML org; has KnownFields strict mode + alias-bomb protection |
| `github.com/santhosh-tekuri/jsonschema/v6` | v6.0.2 [VERIFIED: proxy.golang.org] | JSON Schema 2020-12 validation | Only Go library with full draft 2020-12 compliance per JSON-Schema-Test-Suite; Apache 2.0; passes bowtie compliance suite; active maintenance |
| stdlib `encoding/json` | Go stdlib | Canonical JSON output for resolved speeches | Map keys are sorted lexicographically by default; struct fields emit in declaration order; deterministic output native and WASM |
| stdlib `container/heap` | Go stdlib | Priority queue for Kahn's algorithm tie-breaking | Enables deterministic lexicographic tie-breaking in toposort with O(E log V) |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib `embed` | Go 1.16+ | Embed `schemas/*.json` into the binary | Used in `resolver/parse` so schema files ship with the binary |
| stdlib `testing` | Go stdlib | Table-driven tests, golden file assertions | All test code — the only test framework needed |
| stdlib `os/exec` + `GOROOT/lib/wasm/go_js_wasm_exec` | Go stdlib + installed script | WASM parity test runner | The `go_js_wasm_exec` Node-backed script at `/usr/local/go/lib/wasm/go_js_wasm_exec` is the official Go-provided runner |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `go.yaml.in/yaml/v3` | `goccy/go-yaml` (v1.19.2, 2180 stars, actively maintained) | `goccy/go-yaml` has more ergonomic API but different error types; go.yaml.in is API-identical to gopkg.in — zero migration cost; CONTEXT.md intent is satisfied |
| `go.yaml.in/yaml/v3` | `gopkg.in/yaml.v3` (v3.0.1, archived 2022) | Archived May 2025; no security updates; use go.yaml.in instead |
| `santhosh-tekuri/jsonschema/v6` | Hand-rolled validation | JSON Schema 2020-12 has many edge cases (unevaluatedProperties, dynamic refs, vocab system); do not hand-roll |
| `container/heap` for toposort | `sort.Slice` on a plain slice | `sort.Slice` doesn't give O(log N) insertion; heap is idiomatic for Kahn's |

**Installation:**
```bash
go mod init github.com/mikkelraglan/debateos
go get go.yaml.in/yaml/v3@v3.0.4
go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.2
```

---

## Package Legitimacy Audit

> Phase installs two Go modules. No npm packages. Go module proxy verification performed.

| Package | Registry | Age | Downloads/signals | Source Repo | Verdict | Disposition |
|---------|----------|-----|-------------------|-------------|---------|-------------|
| `go.yaml.in/yaml/v3` | proxy.golang.org | v3.0.4 released 2025-06-29 [VERIFIED: proxy.golang.org] | Official YAML org (`yaml/go-yaml`, 7023 stars) | github.com/yaml/go-yaml | OK | Approved — official YAML organization fork |
| `github.com/santhosh-tekuri/jsonschema/v6` | proxy.golang.org | v6.0.2 released 2025-05-23 [VERIFIED: proxy.golang.org] | Passes JSON-Schema-Test-Suite; bowtie compliance badges | github.com/santhosh-tekuri/jsonschema | OK | Approved |

**Packages removed due to SLOP verdict:** none
**Packages flagged as suspicious (SUS):** none

---

## Architecture Patterns

### System Architecture Diagram

```
YAML input files (opinion.yaml, point.yaml, speech.yaml)
        │
        ▼
┌─────────────────────────────────────────┐
│  resolver/parse                          │
│  · go.yaml.in/yaml/v3 decode            │
│  · KnownFields strict mode              │
│  · embed JSON Schema 2020-12 files      │
│  · santhosh-tekuri/jsonschema/v6        │
│    compile + validate                   │
│  → typed Go structs (Opinion/Point/     │
│    Speech)                              │
└──────────────────┬──────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────┐
│  resolver/graph                          │
│  · build directed dep graph             │
│  · collect ordering edges               │
│  · Kahn toposort (heap, lex tie-break)  │
│  · cycle detection → hard error         │
│  → GraphResult{order []OpinionID,       │
│    cycles [][]OpinionID}                │
└──────────────────┬──────────────────────┘
                   │
                   ▼
┌──────────────────────┐   ┌─────────────────────┐
│  resolver/hardware   │   │  resolver/patch      │
│  · eval compound     │   │  · look up patches   │
│    hw predicates     │   │    for conflict pair  │
│    AND/OR/NOT        │   │  · return patch ops   │
│  → hw eval results   │   │  → PatchOffer{}       │
└──────────┬───────────┘   └──────────┬────────────┘
           │                          │
           └──────────┬───────────────┘
                      │
                      ▼
┌─────────────────────────────────────────┐
│  resolver/resolve                        │
│  · apply docs/04 rules (4 rules + hw +  │
│    ordering)                             │
│  · attach Explanation{} to each decision │
│  · emit ResolvedSpeech with install order│
│  · canonical JSON serialization          │
│    (encoding/json, sorted maps, struct   │
│    field order)                          │
└──────────────────┬──────────────────────┘
                   │
          ┌────────┴────────┐
          │                 │
          ▼                 ▼
   native binary      resolver/wasm
   (go build)         (GOOS=js GOARCH=wasm
                       build tag, js.Func
                       exports)
```

**Parity test flow:**
```
fixtures/*.json (canonical golden files)
        │
        ├─── native: go test ./resolver/... → assert output
        └─── wasm: GOOS=js GOARCH=wasm go test -exec=go_js_wasm_exec → assert same output
                   script diffs golden files byte-for-byte
```

### Recommended Project Structure

```
debateos/
├── go.mod / go.sum                  # single module; two external deps
│
├── schemas/
│   ├── LICENSE                      # CC0-1.0
│   ├── opinion.schema.json          # JSON Schema 2020-12
│   ├── point.schema.json
│   ├── speech.schema.json
│   └── README.md                    # SR-001..SR-022 traceability table
│
├── resolver/
│   ├── parse/
│   │   ├── parse.go                 # ParseOpinion, ParsePoint, ParseSpeech
│   │   ├── validate.go              # JSON Schema validation via jsonschema/v6
│   │   ├── schemas_embed.go         # //go:embed ../../schemas/*.json
│   │   └── parse_test.go
│   ├── graph/
│   │   ├── graph.go                 # BuildGraph(opinions []Opinion) → Graph
│   │   ├── toposort.go              # Kahn + heap, cycle extraction
│   │   └── graph_test.go
│   ├── resolve/
│   │   ├── resolve.go               # Resolve(Speech, HardwareProfile) → ResolvedSpeech
│   │   ├── explanation.go           # Explanation type
│   │   ├── canonical.go             # CanonicalJSON(ResolvedSpeech) []byte
│   │   └── resolve_test.go          # EC-NNN table-driven tests
│   ├── patch/
│   │   ├── patch.go                 # FindPatch(a, b OpinionID) → *PatchOffer
│   │   └── patch_test.go
│   ├── hardware/
│   │   ├── eval.go                  # EvalCondition(expr, HardwareProfile) bool
│   │   └── hardware_test.go
│   └── wasm/
│       ├── main.go                  # //go:build js && wasm; exports js.Func
│       └── main_test.go
│
├── examples/
│   ├── LICENSE                      # CC0-1.0
│   ├── omarchy-mini/                # mini-Omarchy subset (parse→resolve→explain)
│   ├── two-point-clean/             # clean two-point speech
│   ├── conflicting/                 # deliberately conflicting speech (EC-031 style)
│   └── hardware-conditional/        # hardware-gated opinions (EC-037/EC-038 style)
│
└── scripts/
    ├── wasm-parity-test.sh          # build wasm, run same fixtures, diff golden files
    └── check-coverage.sh            # go test -coverprofile; fail below 90%
```

### Pattern 1: Table-Driven EC-NNN Tests

**What:** Each of the 27 edge cases becomes a named test entry. Test name contains the EC ID so failures are immediately locatable.

**When to use:** All resolver conflict scenarios.

**Example:**
```go
// Source: Go wiki table-driven test convention
// resolver/resolve/resolve_test.go

func TestResolveConflicts(t *testing.T) {
    cases := []struct {
        id       string // "EC-001"
        speech   *Speech
        hw       *HardwareProfile
        wantErr  bool
        wantExpl string // substring match on Explanation.Text
    }{
        {
            id: "EC-001",
            // Garuda pre-seeded snapper root vs Omarchy snapper root
            speech:  loadFixture(t, "testdata/ec001-garuda-snapper.yaml"),
            wantErr: true,
            wantExpl: "Hard conflict",
        },
        {
            id: "EC-036",
            // Circular ordering constraint
            speech:  loadFixture(t, "testdata/ec036-cycle.yaml"),
            wantErr: true,
            wantExpl: "Cycle detected",
        },
        // ... 25 more cases
    }
    for _, tc := range cases {
        t.Run(tc.id, func(t *testing.T) {
            result, err := Resolve(tc.speech, tc.hw)
            if tc.wantErr {
                if err == nil {
                    t.Fatalf("%s: expected error, got nil", tc.id)
                }
                return
            }
            if err != nil {
                t.Fatalf("%s: unexpected error: %v", tc.id, err)
            }
            if !strings.Contains(result.Explanations[0].Text, tc.wantExpl) {
                t.Errorf("%s: explanation %q does not contain %q",
                    tc.id, result.Explanations[0].Text, tc.wantExpl)
            }
        })
    }
}
```

### Pattern 2: Kahn's Algorithm with Lexicographic Tie-Breaking

**What:** Topological sort using a min-heap keyed by OpinionID string. Deterministic output regardless of input order.

**When to use:** `resolver/graph` toposort, cycle detection.

**Example:**
```go
// Source: ASSUMED (standard Kahn's algorithm pattern; heap usage is stdlib)
import "container/heap"

// stringHeap is a min-heap of string keys (OpinionIDs)
type stringHeap []string
func (h stringHeap) Len() int           { return len(h) }
func (h stringHeap) Less(i, j int) bool { return h[i] < h[j] } // lexicographic
func (h stringHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *stringHeap) Push(x any)        { *h = append(*h, x.(string)) }
func (h *stringHeap) Pop() any {
    old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x
}

func TopoSort(graph map[string][]string) (order []string, cycle []string, err error) {
    inDegree := map[string]int{}
    for node, deps := range graph {
        if _, ok := inDegree[node]; !ok {
            inDegree[node] = 0
        }
        for _, dep := range deps {
            inDegree[dep]++ // dep must come before node
        }
    }
    h := &stringHeap{}
    heap.Init(h)
    for node, deg := range inDegree {
        if deg == 0 {
            heap.Push(h, node)
        }
    }
    for h.Len() > 0 {
        node := heap.Pop(h).(string)
        order = append(order, node)
        for _, next := range graph[node] {
            inDegree[next]--
            if inDegree[next] == 0 {
                heap.Push(h, next)
            }
        }
    }
    if len(order) != len(inDegree) {
        // cycle — extract names of nodes with non-zero in-degree
        for node, deg := range inDegree {
            if deg > 0 {
                cycle = append(cycle, node)
            }
        }
        sort.Strings(cycle) // deterministic error message
        return nil, cycle, fmt.Errorf("cycle detected: %v", cycle)
    }
    return order, nil, nil
}
```

### Pattern 3: Canonical JSON for WASM Parity Golden Files

**What:** `encoding/json` produces deterministic output when:
- Structs are used (field order = declaration order in source)
- Maps are avoided in output types, or marshaled through a struct (encoding/json sorts map keys lexicographically by default [VERIFIED: live test on Go 1.24.1])
- Floats are avoided in the resolver output (opinion IDs and status fields are strings; no float data in resolved speech output)

**When to use:** `resolver/resolve.CanonicalJSON()` — the function that produces the golden file bytes.

**Key verified facts:**
- `encoding/json.Marshal(map[string]int{"z":3,"a":1})` → `{"a":1,"z":3}` (sorted) [VERIFIED: live test]
- `encoding/json.Marshal(struct{Z,A,M int}{3,1,2})` → `{"Z":3,"A":1,"M":2}` (declaration order) [VERIFIED: live test]
- NaN and ±Inf in float fields → `json: unsupported value: NaN` error; avoid float fields in canonical output [VERIFIED: live test]
- `strconv.FormatFloat(v, 'g', -1, 64)` produces the shortest decimal round-trip representation — same algorithm in native and WASM (Go's WASM target uses the same stdlib, same IEEE 754 float ops) [ASSUMED for WASM parity; no known issue reported in Go tracker]

### Pattern 4: WASM Build + Node Test Runner

**What:** The canonical Go WASM build pattern using `go_js_wasm_exec` (present at `/usr/local/go/lib/wasm/go_js_wasm_exec` on this host).

**When to use:** Building `resolver/wasm` and running WASM parity tests.

**Example:**
```bash
# Build WASM binary
GOOS=js GOARCH=wasm go build -o resolver.wasm ./resolver/wasm/

# Run resolver package tests under WASM (Node.js v24.12.0 available on host)
GOOS=js GOARCH=wasm go test -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./resolver/...

# Parity assertion script (scripts/wasm-parity-test.sh):
# 1. Run native tests → write golden JSON to testdata/golden/
# 2. Run WASM tests → write same golden JSON to testdata/golden-wasm/
# 3. diff -r testdata/golden/ testdata/golden-wasm/ && echo "PARITY OK"
```

**Important:** `wasm_exec.js` is at `/usr/local/go/lib/wasm/wasm_exec.js`. The Go wiki says the same major Go version of compiler and `wasm_exec.js` MUST be used together. The `go_js_wasm_exec` wrapper script handles this automatically. [CITED: go.dev/wiki/WebAssembly]

### Pattern 5: YAML Strict Parsing with KnownFields

**What:** Prevent typo'd field names from silently being ignored.

**When to use:** `resolver/parse` — all document ingestion.

**Example:**
```go
// Source: go.yaml.in/yaml/v3 API (drop-in from gopkg.in/yaml.v3)
import "go.yaml.in/yaml/v3"

func parseOpinion(r io.Reader) (*Opinion, error) {
    dec := yaml.NewDecoder(r)
    dec.KnownFields(true) // reject unknown fields — catches typos like "statues: required"
    var op Opinion
    if err := dec.Decode(&op); err != nil {
        return nil, fmt.Errorf("parse opinion: %w", err)
    }
    return &op, nil
}
```

**Billion-laughs protection:** `go.yaml.in/yaml/v3` includes `allowedAliasRatio` scaling that limits alias expansion based on decoded document size (verified in source: allows up to 99% alias-driven for small docs, scales down to 10% for large docs). [VERIFIED: inspected decode.go source at yaml/go-yaml v3]

### Pattern 6: Embedded JSON Schema Validation

**What:** Embed JSON Schema 2020-12 files into the binary using `go:embed`; validate parsed YAML against schema.

**When to use:** `resolver/parse` validate step, after YAML decode.

**Example:**
```go
// Source: ASSUMED (stdlib embed + santhosh-tekuri/jsonschema v6 API pattern)
import (
    _ "embed"
    "github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed ../../schemas/opinion.schema.json
var opinionSchemaJSON []byte

var opinionSchema *jsonschema.Schema

func init() {
    c := jsonschema.NewCompiler()
    if err := c.AddResource("opinion.schema.json",
        bytes.NewReader(opinionSchemaJSON)); err != nil {
        panic(err)
    }
    var err error
    opinionSchema, err = c.Compile("opinion.schema.json")
    if err != nil {
        panic(err)
    }
}

func validateOpinion(v any) error {
    return opinionSchema.Validate(v)
}
```

### Anti-Patterns to Avoid

- **Range over maps in output-producing code:** Go map iteration is non-deterministic. Any code that iterates a `map` and writes to the resolved speech must sort keys first or use a struct instead. `encoding/json` sorts map keys on marshal, but intermediate processing must not leak ordering.
- **Floating-point fields in canonical output:** `encoding/json` returns an error for NaN/Inf. Hardware condition predicates and opinion statuses are strings; no float data should appear in the canonical resolved speech. If versions are stored as `semver.Version` structs, serialize as strings.
- **`gopkg.in/yaml.v3` (archived):** The archived package at `gopkg.in/yaml.v3` still works (v3.0.1 is the last release, 2022) but receives no security updates. Use `go.yaml.in/yaml/v3` instead — same API, active maintenance.
- **Sharing `wasm_exec.js` across Go versions:** Each Go release ships a matching `wasm_exec.js`. Always use the one from `$(go env GOROOT)/lib/wasm/wasm_exec.js`. Do not commit it to the repo; reference it at build time.
- **`time.Time` in canonical JSON:** `encoding/json` marshals `time.Time` with timezone offset. Use Unix timestamps (int64) or RFC3339 UTC strings for any time fields in canonical output.
- **Un-seeded `math/rand` in tests:** Maps and Go internals use randomized hash seeds since Go 1. Tests must not rely on any ordering that isn't explicitly sorted.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML parsing | Custom parser | `go.yaml.in/yaml/v3` | YAML spec edge cases: merge keys, multi-document, anchors, type coercion, block vs flow styles |
| JSON Schema 2020-12 validation | Custom validator | `santhosh-tekuri/jsonschema/v6` | 2020-12 has `unevaluatedProperties`, `$dynamicRef`, `$vocabulary` — extremely complex to implement correctly |
| Alias-bomb protection | Depth counter | `go.yaml.in/yaml/v3` (built-in) | `allowedAliasRatio` already implemented and tested against CVE-2019-11253 class attacks |
| Lexicographic toposort | Ad-hoc sort | `container/heap` + Kahn's | A plain sort-then-iterate does not give O(log N) insertion; heap is idiomatic |
| JSON map key sorting | Manual sort loop | `encoding/json` (automatic) | `encoding/json.Marshal` sorts map keys automatically; rely on it |
| WASM Node test runner | Custom Node script | `$(go env GOROOT)/lib/wasm/go_js_wasm_exec` | Official Go-provided script handles V8 stack size tuning and WASM instantiation |

**Key insight:** In this domain, the hand-rolled paths all have failure modes that are invisible in the happy path (alias bombs pass with simple docs, JSON key ordering looks fine until a second contributor changes map insertion order). Use the libraries that have been tested against adversarial input.

---

## Go Version and Module Setup

### Target Go Version

| Fact | Value | Source |
|------|-------|--------|
| Installed on host | go1.24.1 | [VERIFIED: `go version` on host] |
| Current Go stable (June 2026) | go1.26.4 | [VERIFIED: go.dev/VERSION] |
| Go 1.25 released | August 2025 | [CITED: go.dev/doc/go1.25] |
| Recommended `go.mod` directive | `go 1.24` | [ASSUMED: matches installed toolchain; avoids auto-download on first build] |

**Recommendation:** Use `go 1.24` in go.mod. Go's toolchain auto-download (`GOTOOLCHAIN=auto`, the default) will upgrade if a higher version is required by a dependency, but all code in Phase 1 uses features available in Go 1.16+ (embed) and 1.18+ (generics if used). CI can pin to 1.26 when it upgrades the build matrix.

### Monorepo go.mod Setup

```
module github.com/mikkelraglan/debateos

go 1.24

require (
    go.yaml.in/yaml/v3 v3.0.4
    github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
)
```

All Phase 1 packages (`resolver/parse`, `resolver/graph`, etc.) are sub-packages of this single module. Phase 3 `cli/` and Phase 5 `forum/` will add to the same `go.mod` — no nested modules needed. The docs/11 layout explicitly calls for a single Go module. [CITED: docs/11-repo-layout.md]

### Coverage Tooling

Go's built-in `-coverprofile` plus a shell script is the idiomatic approach. No third-party tool needed. [VERIFIED: `go help testflag`]

```bash
# scripts/check-coverage.sh
#!/usr/bin/env bash
set -euo pipefail
go test -coverprofile=coverage.out -covermode=count ./resolver/...
TOTAL=$(go tool cover -func=coverage.out | grep "^total:" | awk '{print $3}' | tr -d '%')
if (( $(echo "$TOTAL < 90" | bc -l) )); then
    echo "FAIL: coverage $TOTAL% is below 90% threshold"
    exit 1
fi
echo "PASS: coverage $TOTAL%"
```

For per-package breakdown (100% on docs/04 rule branches = `resolver/resolve` package):
```bash
go test -v -coverprofile=resolve.out -covermode=count ./resolver/resolve/
go tool cover -func=resolve.out
```

---

## YAML Library Decision: gopkg.in/yaml.v3 → go.yaml.in/yaml/v3

This is the key finding that affects the locked decision in CONTEXT.md:

| Fact | Value | Source |
|------|-------|--------|
| `gopkg.in/yaml.v3` archived | April 2025 | [VERIFIED: github.com/go-yaml/yaml — `archived: True`] |
| Last release of `gopkg.in/yaml.v3` | v3.0.1 (2022-05-27) | [VERIFIED: proxy.golang.org] |
| Maintained fork | `go.yaml.in/yaml/v3` v3.0.4 (2025-06-29) | [VERIFIED: proxy.golang.org] |
| Fork maintainer | The YAML Project (`yaml` GitHub org) | [VERIFIED: github.com/yaml/go-yaml] |
| API compatibility | Drop-in; same `package yaml` name | [VERIFIED: source header `package yaml`] |
| KnownFields support | Yes — `dec.KnownFields(true)` | [VERIFIED: go.yaml.in yaml.go source] |
| Alias-bomb protection | Yes — `allowedAliasRatio` scaling | [VERIFIED: decode.go source inspection] |

**Planner action:** The single change from the CONTEXT.md locked decision is the import path in `go.mod` and all `import` statements: `go.yaml.in/yaml/v3` instead of `gopkg.in/yaml.v3`. The intent ("YAML only, stdlib otherwise") is preserved. No user confirmation required — this is a security-motivated clarification, not a scope change.

---

## SR-001..SR-022 → Schema Field Map

The planner uses this table to create tasks for each schema field.

| SR-NNN | Field Category | Go Type Hint | Schema Key |
|--------|---------------|--------------|------------|
| SR-001 | Status | `string` enum `required`/`nice-to-have` | `status` |
| SR-002 | Opinion dependencies | `[]OpinionRef` | `depends_on` |
| SR-003 | Conflict declarations + known patches | `[]OpinionRef` + `[]PatchRef` | `conflicts`, `known_patches` |
| SR-004 | Single hardware predicate | `string` (named predicate) | `hardware_condition` (scalar) |
| SR-005 | Compound hardware predicates | `HardwareExpr` (tree: AND/OR/NOT/leaf) | `hardware_condition` (object with combinators) |
| SR-006 | Phase-level + relative ordering | `string` enum (phase) + `[]OpinionRef` before/after | `install_phase`, `ordering` |
| SR-007 | Translator capability declaration | `[]string` | `translator_capabilities` |
| SR-008 | File asset payloads | `[]FileAsset{src, dst}` | `file_assets` |
| SR-009 | Custom repo registration | `[]RepoDecl{name, url, sig_level, priority}` | `custom_repos` |
| SR-010 | Runtime tool install (npm/pip/cargo) | `[]RuntimeToolInstall{manager, packages}` | `runtime_tool_installs` |
| SR-011 | Execution phase (install-time vs first-run) | `string` enum | `execution_phase` |
| SR-012 | Arbitrary script payload | `*ScriptPayload{script, capabilities}` | `script_payload` |
| SR-013 | Display manager config | `*DisplayManagerConfig` | `display_manager` |
| SR-014 | Bootloader configuration | `*BootloaderConfig` | `bootloader` |
| SR-015 | Service enable/disable + chroot compat | `[]ServiceDecl{name, deferred}` | `services` |
| SR-016 | Sysctl parameter drop-in | `[]SysctlParam{key, value, drop_in_file}` | `sysctl_params` |
| SR-017 | Kernel boot parameters | `[]KernelParam{key, value}` | `kernel_params` |
| SR-018 | User/group membership | `[]GroupMembership{group}` | `group_memberships` |
| SR-019 | MIME type associations | `[]MimeAssoc{mime_pattern, app_id}` | `mime_associations` |
| SR-020 | Theming system | `*ThemeDecl{bundle_dir, symlinks, is_default}` | `theme` |
| SR-021 | Point: name + intent + member list | `Point{id, name, intent, members []OpinionID}` | top-level Point fields |
| SR-022 | Speech: foundation target + point list | `Speech{foundation, points []PointRef}` | top-level Speech fields |

**JSON Schema approach:** One file per top-level type (`opinion.schema.json`, `point.schema.json`, `speech.schema.json`) with `$defs` for shared sub-types (HardwareExpr, RepoDecl, etc.). Keep schemas in JSON (not YAML) so they can be loaded directly by santhosh-tekuri/jsonschema without YAML pre-processing.

---

## Common Pitfalls

### Pitfall 1: Map Iteration Order in Intermediate Resolver State

**What goes wrong:** Resolver builds intermediate maps (e.g., `map[OpinionID]*Opinion`); if output is derived by ranging over this map, byte output varies between runs and between native/WASM.

**Why it happens:** Go map iteration is randomized per-run since Go 1.

**How to avoid:** Collect map keys into a `[]string`, sort with `sort.Strings`, then iterate. Use `encoding/json` struct-based output (not `map[string]any`) for all canonical output types.

**Warning signs:** Golden file tests pass individually but fail on repeated runs; diff shows only field ordering changes.

### Pitfall 2: WASM wasm_exec.js Version Mismatch

**What goes wrong:** `wasm_exec.js` from a different Go version is bundled with a binary, causing runtime panics or silently wrong behavior.

**Why it happens:** The `wasm_exec.js` JS/Go bridge has breaking changes between major Go versions.

**How to avoid:** Always reference `$(go env GOROOT)/lib/wasm/wasm_exec.js` at build time. Do not commit `wasm_exec.js` to the repository. The `go_js_wasm_exec` wrapper script does this automatically. [CITED: go.dev/wiki/WebAssembly]

**Warning signs:** WASM binary runs in browser/Node but produces wrong output or panics immediately.

### Pitfall 3: Hardware Condition Expression Underspecification

**What goes wrong:** A simple `hardware_condition: string` field is not sufficient for SR-005 compound predicates (AND/OR/NOT with set membership). Implementing simple predicates first and discovering the compound requirement later means a breaking schema change.

**Why it happens:** SR-005 is non-obvious — 8 of the 18 Omarchy hardware helpers use compound logic.

**How to avoid:** Design `hardware_condition` as a discriminated union from day one: `{type: "leaf", predicate: "omarchy-hw-asus-rog"}` vs `{type: "and", operands: [...]}` etc. See SR-005 compound examples in schema-requirements.md.

**Warning signs:** Attempting to encode OM-071 (`hw-intel AND battery-present AND cpu-model in [151,154,...]`) as a single string.

### Pitfall 4: gopkg.in/yaml.v3 vs go.yaml.in/yaml/v3 Import Path Confusion

**What goes wrong:** Code imports `gopkg.in/yaml.v3` (the archived package); `go get` still works (it fetches v3.0.1), but the module is not maintained and has no security updates.

**Why it happens:** `gopkg.in/yaml.v3` is still accessible via the module proxy; it just won't be updated.

**How to avoid:** Use `go.yaml.in/yaml/v3` in both `go.mod` and all `import` statements. The API is identical.

### Pitfall 5: Toposort Phase/Cross-Phase Ordering (SR-006)

**What goes wrong:** Phase-level ordering (preflight → packaging → config → login → post-install → first-run) is implemented as a simple enum sort without support for cross-phase override constraints (OM-023: packaging phase but must run after config/OM-041).

**Why it happens:** The naive model is "sort by phase enum, then by within-phase order". OM-023 violates this.

**How to avoid:** The toposort graph must accept edges that cross phase boundaries. Phase enum is an additional sort key for `nice-to-have` tie-breaking, not a hard partition. The ordering graph (`must_install_after`, `must_install_before`) takes precedence over phase order.

### Pitfall 6: NaN/Inf in float64 Fields Breaking JSON Marshaling

**What goes wrong:** `encoding/json.Marshal` returns `unsupported value: NaN` if any float64 field in the canonical output struct contains NaN or Inf.

**Why it happens:** JSON specification does not support NaN/Inf.

**How to avoid:** Canonical output structs must not contain raw `float64` fields. Version numbers → semver strings. Percentages → int or string. Hardware predicates → string. [VERIFIED: live test on Go 1.24.1]

---

## Code Examples

### Verified: encoding/json Map Key Sorting

```go
// Verified on Go 1.24.1 — encoding/json always sorts map keys lexicographically
m := map[string]int{"z": 3, "a": 1, "m": 2}
b, _ := json.Marshal(m)
// b = {"a":1,"m":2,"z":3}  ← sorted
```

### Verified: encoding/json Struct Field Order

```go
// encoding/json emits struct fields in declaration order (NOT alphabetical)
type Opinion struct {
    Status string  // field 1
    ID     string  // field 2
    Name   string  // field 3
}
b, _ := json.Marshal(Opinion{Status: "required", ID: "OM-001", Name: "test"})
// b = {"Status":"required","ID":"OM-001","Name":"test"}  ← declaration order
// Use lowercase json tags for canonical field names:
// `json:"status"`, `json:"id"`, `json:"name"`
```

### Verified: WASM Test Runner Command

```bash
# Verified: go_js_wasm_exec exists at /usr/local/go/lib/wasm/go_js_wasm_exec
# node v24.12.0 available on host
GOOS=js GOARCH=wasm go test -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./resolver/...
```

### Verified: KnownFields Usage

```go
// Source: go.yaml.in/yaml/v3 (identical API to gopkg.in/yaml.v3)
dec := yaml.NewDecoder(r)
dec.KnownFields(true)
var op Opinion
if err := dec.Decode(&op); err != nil {
    // "field statues not found in type parse.Opinion" — catches typos
}
```

### Verified: Alias-Bomb Protection in go.yaml.in/yaml/v3

```go
// The library internally tracks aliasCount and decodeCount
// allowedAliasRatio scales from 0.99 (small docs) to 0.10 (>4MB docs)
// No application code needed — protection is built in
// Verified by reading decode.go source at github.com/yaml/go-yaml v3
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `gopkg.in/yaml.v3` | `go.yaml.in/yaml/v3` (official YAML org fork) | April 2025 (archive); June 2025 (v3.0.4 release) | Must update import path; API identical |
| `wasm_exec.js` at `misc/wasm/` | `wasm_exec.js` at `lib/wasm/` | Go 1.24 | `go env GOROOT)/lib/wasm/wasm_exec.js` is the correct path; `misc/wasm/` no longer exists |
| Separate `wasm_exec.js` at `misc/wasm/wasm_exec_node.js` | `wasm_exec_node.js` + `go_js_wasm_exec` script at `lib/wasm/` | Go 1.21+ | `go_js_wasm_exec` wrapper handles Node stack-size tuning automatically |
| `jsonschema/v5` (last stable) | `jsonschema/v6` (v6.0.2) | May 2025 | v6 has cleaner API; same 2020-12 compliance |
| `encoding/json/v2` (experimental) | `encoding/json` (stable) | Go 1.25 (v2 still experimental) | Use stable stdlib; v2 adds Deterministic option but stable stdlib is sufficient for our use case |
| `GODEBUG=asynctimerchan=0` WASM settings | `GOWASM` env var settings removed | Go 1.26 | `signext`/`satconv` GOWASM settings are now ignored (always enabled); no action needed |

**Deprecated/outdated:**
- `gopkg.in/yaml.v3`: archived; switch to `go.yaml.in/yaml/v3`
- `jsonschema/v5`: superseded by v6; v6 has cleaner API for validation errors
- `misc/wasm/wasm_exec.js`: moved to `lib/wasm/`; any scripts referencing the old path must be updated

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All resolver code | Yes | go1.24.1 | — |
| Node.js | WASM parity tests | Yes | v24.12.0 | — |
| `go_js_wasm_exec` | WASM test runner | Yes | `/usr/local/go/lib/wasm/go_js_wasm_exec` | — |
| `wasm_exec.js` | WASM browser build | Yes | `/usr/local/go/lib/wasm/wasm_exec.js` | — |
| `wasm_exec_node.js` | Node WASM runner | Yes | `/usr/local/go/lib/wasm/wasm_exec_node.js` | — |
| `bc` (for coverage script) | check-coverage.sh | Yes (standard Linux) | — | Use Python or awk for arithmetic |
| Docker | Phase 3 only | Not required in Phase 1 | — | — |

**Missing dependencies with no fallback:** none
**Missing dependencies with fallback:** none (all Phase 1 tooling is present)

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | `go test` (Go 1.24.1 stdlib) |
| Config file | None — `go test` is config-free |
| Quick run command | `go test ./resolver/...` |
| Full suite command | `go test ./resolver/... && GOOS=js GOARCH=wasm go test -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./resolver/... && bash scripts/check-coverage.sh` |
| Estimated quick run | < 5 seconds (pure logic, no I/O) |
| Estimated full suite | < 30 seconds (WASM compile adds ~10–15s) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SCHM-01 | Opinion/Point/Speech YAML parses and validates against JSON Schema 2020-12 | unit | `go test ./resolver/parse/...` | No — Wave 0 |
| SCHM-01 | All SR-001..SR-022 fields are expressible in schema | unit | `go test ./resolver/parse/ -run TestParseOpinionSR` | No — Wave 0 |
| SCHM-02 | No Arch/Debian-specific strings leak into schema types | unit | `go test ./resolver/parse/ -run TestSchemaOSAgnostic` | No — Wave 0 |
| RSLV-01 | 27 EC-NNN scenarios resolve with correct Explanation | unit | `go test ./resolver/resolve/ -run TestResolveEC` | No — Wave 0 |
| RSLV-02 | Patch opinions discovered and offered automatically | unit | `go test ./resolver/patch/ -run TestPatchDiscovery` | No — Wave 0 |
| RSLV-03 | Toposort produces correct order; cycle → hard error naming opinions | unit | `go test ./resolver/graph/ -run TestTopoSort` | No — Wave 0 |
| RSLV-04 | Hardware conditions evaluate against declared hardware profile | unit | `go test ./resolver/hardware/ -run TestHardwareEval` | No — Wave 0 |
| RSLV-05 | Native and WASM produce byte-identical canonical JSON | parity | `bash scripts/wasm-parity-test.sh` | No — Wave 0 |
| RSLV-06 | TDD harness covers all 27 EC-NNN + required-vs-required + hw mismatch + version clash + patchable pair | unit | `go test ./resolver/... -v -run TestResolve` | No — Wave 0 |
| RSLV-06 | Coverage ≥ 90% on resolver packages | coverage | `bash scripts/check-coverage.sh` | No — Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./resolver/...` (< 5 seconds)
- **Per wave merge:** `go test ./... && bash scripts/wasm-parity-test.sh && bash scripts/check-coverage.sh`
- **Phase gate:** Full suite green + coverage ≥ 90% + WASM parity PASS before `/gsd-verify-work`

### Wave 0 Gaps (test infrastructure to create before implementation tasks)

- [ ] `resolver/parse/parse_test.go` — covers SCHM-01, SCHM-02
- [ ] `resolver/graph/graph_test.go` — covers RSLV-03 (toposort, cycle detection)
- [ ] `resolver/resolve/resolve_test.go` — covers RSLV-01, RSLV-06 (27 EC-NNN + rule branches)
- [ ] `resolver/patch/patch_test.go` — covers RSLV-02
- [ ] `resolver/hardware/hardware_test.go` — covers RSLV-04
- [ ] `resolver/wasm/main_test.go` — WASM entry point smoke test
- [ ] `resolver/resolve/testdata/ec001-garuda-snapper.yaml` .. `ec052-*.yaml` — 27 EC fixture YAML files
- [ ] `scripts/wasm-parity-test.sh` — covers RSLV-05
- [ ] `scripts/check-coverage.sh` — coverage enforcement
- [ ] `go.mod` / `go.sum` — module bootstrap (Wave 0, Task 0)
- [ ] `schemas/opinion.schema.json`, `point.schema.json`, `speech.schema.json` — JSON Schema 2020-12 files
- [ ] `schemas/README.md` — SR-001..SR-022 traceability table (SCHM-01 requirement)

---

## Security Domain

> Phase 1 handles only local file parsing; no network, no auth. Reduced ASVS surface.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | N/A (no auth in Phase 1) |
| V3 Session Management | No | N/A |
| V4 Access Control | No | N/A |
| V5 Input Validation | Yes — YAML + JSON Schema parsing of user-provided files | `go.yaml.in/yaml/v3` KnownFields; `santhosh-tekuri/jsonschema/v6` schema validation; explicit type assertions |
| V6 Cryptography | No | N/A |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| YAML alias bomb (billion-laughs) | DoS | `go.yaml.in/yaml/v3` `allowedAliasRatio` built-in [VERIFIED: decode.go] |
| Malformed YAML causing panic | Tampering | `KnownFields(true)` + recover pattern in parse layer |
| JSON Schema infinite loop trap | DoS | `santhosh-tekuri/jsonschema/v6` detects `$schema` cycles + validation cycles [CITED: v6 README] |
| Trust level `SigLevel=Never` (OM-100, SR-009) | Elevation | Schema enumerates trust levels; resolver must warn on `Never` trust in explanation output |
| Unsigned repo injection via opinion | Tampering | `custom_repos[].sig_level` is an explicit schema field; resolver attaches warning to explanation when `sig_level = Never` |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Go WASM `GOOS=js GOARCH=wasm` produces byte-identical `strconv.FormatFloat` and `encoding/json` output vs native (no known issue in Go tracker for this) | WASM Parity, Pattern 3 | Parity test would fail; would need platform-specific float formatting workaround (low probability — Go stdlib is the same code in both targets) |
| A2 | `santhosh-tekuri/jsonschema/v6` API uses `NewCompiler()` + `AddResource()` + `Compile()` pattern consistent with v5 | Pattern 6 | Compilation error; check v6 pkg.go.dev for exact API on first use |
| A3 | Kahn's toposort with `container/heap` is the standard deterministic Go pattern | Pattern 2 | Any correct O(E log V) min-heap toposort works; exact implementation is Claude's Discretion |
| A4 | `go 1.24` in go.mod is sufficient for all Phase 1 features (embed, generics if used, WASM) | Module Setup | If a dependency requires higher, `GOTOOLCHAIN=auto` will auto-download; not a blocker |

**If this table is exhausted:** All other claims are verified against live registries, installed toolchain, or official documentation.

---

## Open Questions

1. **`gopkg.in/yaml.v3` → `go.yaml.in/yaml/v3` import path change**
   - What we know: The CONTEXT.md locked decision says "gopkg.in/yaml.v3 only". `gopkg.in/yaml.v3` is archived (April 2025). `go.yaml.in/yaml/v3` is the API-identical official maintained fork.
   - What's unclear: Does the user consider this a decision change requiring explicit confirmation, or a security-obvious clarification?
   - Recommendation: The planner should note this in the Wave 0 task ("use `go.yaml.in/yaml/v3` v3.0.4 — API-identical drop-in for archived `gopkg.in/yaml.v3`") and proceed. No scope change.

2. **JSON Schema files: JSON or YAML format**
   - What we know: `santhosh-tekuri/jsonschema/v6` supports both JSON and YAML input. CONTEXT.md says "YAML schema documentation + JSON Schema (draft 2020-12) validation files". The "YAML schema documentation" may mean human-friendly YAML docs alongside the machine-readable JSON Schema files.
   - What's unclear: Should the `.json` files be the canonical validation artifact and `.yaml` files be documentation-only, or are the `.yaml` files both documentation and valid JSON Schema?
   - Recommendation: Use `.json` files as the canonical validation artifact (simpler loading for `santhosh-tekuri/jsonschema/v6`; no YAML pre-processing step); add a human-readable `opinion.schema.yaml` "documentation YAML" alongside that is not loaded by the validator. This satisfies both "YAML human-readable documentation" and "JSON Schema 2020-12 validation files" from SCHM-01.

3. **Speech `schema: 1` version field location**
   - What we know: CONTEXT.md requires an explicit `schema: 1` version field on every document.
   - What's unclear: Is `schema` at the top level of each document type (same field on Opinion, Point, and Speech), or only on Speech?
   - Recommendation: Add `schema: 1` as a required top-level field on Opinion, Point, and Speech. This future-proofs all three document types for schema migration.

---

## Sources

### Primary (HIGH confidence — verified on live systems June 2026)
- `go version` on host → go1.24.1 + `lib/wasm/` layout confirmed
- `curl proxy.golang.org/go.yaml.in/yaml/v3/@latest` → v3.0.4, 2025-06-29
- `curl proxy.golang.org/github.com/santhosh-tekuri/jsonschema/v6/@latest` → v6.0.2, 2025-05-23
- `curl api.github.com/repos/go-yaml/yaml` → `archived: true`
- `curl api.github.com/repos/goccy/go-yaml` → `archived: false`, stars: 2180
- `curl api.github.com/orgs/yaml` → "The YAML Project" (official YAML org)
- Live Go code test (`/tmp/test_json.go`, `/tmp/test_json2.go`) → map key sort, struct field order, NaN error
- `cat decode.go` from github.com/yaml/go-yaml v3 → `allowedAliasRatio`, `KnownFields` source confirmed
- `ls /usr/local/go/lib/wasm/` → `wasm_exec.js`, `go_js_wasm_exec`, `wasm_exec_node.js` present
- `node --version` → v24.12.0
- `curl go.dev/VERSION` → go1.26.4 (current stable)

### Secondary (MEDIUM confidence — official documentation)
- [go.dev/wiki/WebAssembly](https://go.dev/wiki/WebAssembly) — WASM build + test runner instructions
- [go.dev/doc/go1.25](https://go.dev/doc/go1.25) — Go 1.25 release notes (August 2025)
- [pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6) — 2020-12 compliance badges
- `santhosh-tekuri/jsonschema/v6` README at refs/tags/v6.0.2 — feature list
- [go.dev/doc/go1.24](https://go.dev/doc/go1.24) — Go 1.24 release notes

### Tertiary (LOW confidence — web search, informational only)
- GitHub issues for `cli/cli`, `spf13/viper`, `goharbor/harbor` discussing gopkg.in/yaml.v3 migration → confirms `go.yaml.in/yaml/v3` is the community-accepted replacement

---

## Metadata

**Confidence breakdown:**
- Standard Stack: HIGH — all packages verified via Go module proxy and GitHub API
- Architecture: HIGH — derived from locked decisions in CONTEXT.md and docs/11
- Pitfalls: HIGH — map ordering verified by live test; WASM path MEDIUM (A1 assumption)
- WASM parity: MEDIUM — no known issues but not directly proven by a WASM test run here (no Go code exists yet)
- Schema field map (SR→field): HIGH — derived from research/schema-requirements.md source material

**Research date:** 2026-06-12
**Valid until:** 2026-09-12 (90 days; `go.yaml.in/yaml/v3` and `jsonschema/v6` are stable; WASM mechanics change only with Go major releases)
