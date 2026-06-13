# DEB-03 Arch-Leak Audit

**Audit date:** 2026-06-13
**Auditor:** Phase 4 planning (04-RESEARCH.md §Arch-Leak Audit)
**Status:** COMPLETE — all 6 findings documented; 1 genuine leak fixed

---

## Why This Audit Exists

DebateOS's core design guarantee is stated in **Invariant 1**: translators own their mechanics.
The schema, resolver, and CLI layer must be foundation-agnostic — they describe *what* the user
wants (opinions, points, speeches) without embedding assumptions about *how* any particular OS
distribution implements those wants.

Implementing a second foundation (Debian/Ubuntu) surfaces any Arch-shaped assumptions that
leaked into the shared infrastructure. This audit identifies every such assumption, classifies
it, and documents the action taken. The result proves that Invariant 1 is real, not assumed.

---

## Findings Table

| # | Finding | Location | Type | Action |
|---|---------|----------|------|--------|
| 1 | `sig_level` enum | `schemas/opinion.schema.json` — `repoDecl.sig_level` | Intentional abstraction | Document apt mapping in `translators/debian/variant.py`; no schema change |
| 2 | `install_phase` enum | `schemas/opinion.schema.json` — `install_phase` | Foundation-neutral | No action; both translators map phases identically |
| 3 | `mkinitcpio` capability tokens | `translators/arch/capabilities.json` | Correctly isolated | Debian `capabilities.json` omits them; opinions requiring them CapabilityError on Debian |
| 4 | `limine` bootloader tokens | `translators/arch/capabilities.json` | Correctly isolated | Debian uses GRUB2; `bootloader` schema field is abstract; limine tokens absent from Debian caps |
| 5 | `build.go` hardcoded to Arch | `cli/build/build.go` | Genuine leak — **FIXED** | Refactored via `foundationRegistry` map (this plan, Task 1) |
| 6 | `keyring` field interpretation | `schemas/opinion.schema.json` — `repoDecl.keyring` | Minor asymmetry | Documented as translator-interpreted; no schema change for v1.0 |

---

## Finding Detail

### Finding 1 — `sig_level` enum (Intentional Abstraction)

**Schema field:** `repoDecl.sig_level` — enum values: `Required`, `RequiredDatabaseOptional`,
`OptionalTrustAll`, `Never`

**Appearance of leak:** The enum names were modelled after pacman's `SigLevel` directive.

**Actual assessment:** The values express a *trust intent*, not a pacman-specific mechanism:
- `Required` → "require cryptographic signatures on packages and database"
- `RequiredDatabaseOptional` → "require signatures on the database; packages may be unsigned"
- `OptionalTrustAll` → "accept unsigned packages; emit trust warning"
- `Never` → "skip signature checks entirely; emit loud warning"

The Debian translator maps these intents to apt `sources.list.d` options:

| `sig_level` | apt option | Notes |
|-------------|------------|-------|
| `Required` | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | Keyring from `repoDecl.keyring` |
| `RequiredDatabaseOptional` | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | No direct apt equiv; treated as Required |
| `OptionalTrustAll` | `[trusted=yes]` | Emits trust warning in output |
| `Never` | `[trusted=yes]` | Emits LOUD WARNING comment |

**Disposition:** Intentional abstraction. The enum is correctly shared. No schema change needed.
The mapping is implemented in `translators/debian/variant.py` (`apply_variant()`).

---

### Finding 2 — `install_phase` enum (Foundation-Neutral)

**Schema field:** `install_phase` — enum values: `preflight`, `packaging`, `config`, `login`,
`post-install`, `first-run`

**Appearance of leak:** Phase names could be Arch-pipeline-shaped.

**Actual assessment:** The values describe *when in the lifecycle* an opinion effectuates,
independent of the distribution:
- `packaging` → package install time (pacman on Arch; apt on Debian)
- `first-run` → first user login, via systemd oneshot unit (same flag-file pattern on both)
- Other phases map symmetrically on both foundations

Both translators use identical phase logic: `packaging` → install packages; `first-run` →
systemd user oneshot unit with flag-file guard.

**Disposition:** Foundation-neutral by design. No action needed; no schema change.

---

### Finding 3 — `mkinitcpio` Capability Tokens (Correctly Isolated)

**Location:** `translators/arch/capabilities.json`

**Tokens:** `configure-mkinitcpio-hooks-and-modules`, `write-mkinitcpio-config-drop-in`,
`write-mkinitcpio-module-configuration`, `write-mkinitcpio-module-list`,
`write-mkinitcpio-module-list-configuration`, `install-initramfs-hooks`

**Assessment:** These tokens appear in Arch opinions that configure mkinitcpio (Arch's initramfs
generator). Debian uses `initramfs-tools` instead. The Debian translator simply does **not**
declare these tokens in its `capabilities.json`. Any required opinion that declares one of
these capability tokens will fire a `CapabilityError` at composition time on the Debian
foundation — which is the *correct* behaviour (SC-2 gate). The dual-foundation proof speech
uses only capability tokens declared by both translators.

**Note on `translators/arch/manifest.py`:** The `from_resolved()` method reads the `foundation`
field from `resolved.json` but only as a pass-through value stored in the manifest. It does
**not** branch on foundation to gate any logic. This is correct: the capability gate
(`check_capabilities()`) performs foundation-specific gating before `BuildManifest` is
constructed.

**Disposition:** Correctly isolated. No schema change. Debian `capabilities.json` omits
mkinitcpio tokens; `configure-initramfs-tools-hooks` (or equivalent) can be added when
Debian opinions that configure `initramfs-tools` are authored.

---

### Finding 4 — `limine` Bootloader Tokens (Correctly Isolated)

**Location:** `translators/arch/capabilities.json`

**Tokens:** `manage-limine-bootloader-installation`, `write-bootloader-entry-tool-drop-in`,
`manage-efi-boot-entries`

**Assessment:** Limine is the bootloader used by Omarchy on Arch. Debian uses GRUB2. The
`bootloader` schema field (`{name, timeout, snapshot}`) is correctly abstract — it describes
the desired bootloader in neutral terms. Each translator maps the abstract bootloader spec to
its distribution-specific mechanism (limine config on Arch; grub2 config on Debian).

The Debian translator declares `configure-grub2-bootloader` (or equivalent) rather than
the limine tokens. Opinions requiring `manage-limine-bootloader-installation` will
CapabilityError on Debian — correct behaviour.

**Disposition:** Correctly isolated. No schema change. The `bootloader` schema field is
foundation-agnostic.

---

### Finding 5 — `build.go` Hardcoded to Arch (Genuine Leak — FIXED)

**Location:** `cli/build/build.go`

**Leak (before this fix):**
```go
profileFlag := fs.String("profile", "vanilla-arch", ...)  // ← Arch default
profileDir := filepath.Join(outDir, "arch-profile")        // ← Arch hardcode
translateBin := "translators/arch/translate"               // ← Arch hardcode
```

**Impact:** `debateos build` with a Debian speech would still invoke the Arch translator,
silently producing an incorrect Arch profile tree instead of a Debian one.

**Fix applied in this plan (Task 1):**
```go
type foundationConfig struct {
    TranslateBin   string
    ProfileDir     string
    DefaultProfile string
}

var foundationRegistry = map[string]foundationConfig{
    "arch":   {"translators/arch/translate",   "arch-profile",   "vanilla-arch"},
    "debian": {"translators/debian/translate", "debian-profile", "debian"},
    // Future: "ubuntu": {"translators/debian/translate", "debian-profile", "ubuntu"},
}

fc, ok := foundationRegistry[rs.Foundation]
if !ok {
    fmt.Fprintf(stderr, "build: unknown foundation %q — no translator registered\n", rs.Foundation)
    return 1
}
effectiveProfile := *profileFlag
if effectiveProfile == "" {
    effectiveProfile = fc.DefaultProfile
}
```

The `--profile` flag default is now `""`, resolved to the foundation's `DefaultProfile` when
not explicitly set. Existing users who pass `--profile vanilla-arch` explicitly are unaffected
(backward compatible — Pitfall 6 from 04-RESEARCH.md resolved).

**Security note (T-04-10):** Unknown foundations return an error rather than falling through
to an arbitrary translator binary. The registry is a closed set of compile-time constants —
`rs.Foundation` (curator-declared string) selects from a closed set, not an arbitrary path.

**Test coverage:** `TestBuildFoundationDispatch`, `TestBuildArchUnchanged`,
`TestBuildExplicitProfileOverride`, `TestBuildUnknownFoundation` in `cli/build/build_test.go`.

**Disposition:** Genuine leak — FIXED. This is the only code change required by DEB-03.

---

### Finding 6 — `keyring` Field Interpretation (Minor Asymmetry — Documented)

**Schema field:** `repoDecl.keyring` — string

**Asymmetry:**
- Arch translator: `keyring` is a **pacman package name** (e.g. `cachyos-keyring`), installed
  via `pacman -S <keyring>` before the custom repo is first accessed.
- Debian translator: `keyring` is a **URL or file path** pointing to a `.asc` armoured GPG
  key, fetched and placed in `config/archives/debateos-REPONAME.key.chroot` (live-build installs
  it before package fetch via the `config/archives/` mechanism).

**Assessment:** Both interpretations fit the existing `string` schema type. The field semantics
are "translator-interpreted" — Arch translators treat it as a pacman package name; Debian
translators treat it as a URL or path. For v1.0 this is acceptable because:
1. Speechs typically target a single foundation; a `custom_repo` with a `keyring` is
   foundation-specific by nature.
2. If a future foundation-neutral speech needs to declare custom repos, a `keyring_url` field
   (or a union type) can be added in a future schema version.

**Disposition:** Minor asymmetry — documented. No schema change for v1.0.

---

## Schema / Capability Adjustments Summary

**The only code change required by DEB-03 is in `cli/build/build.go` (Finding 5).**

| Component | Changed? | Notes |
|-----------|----------|-------|
| `schemas/opinion.schema.json` | No | `sig_level`, `install_phase`, `keyring` all correctly abstracted |
| `resolver/` | No | Foundation field is already a pass-through string; no Arch assumptions |
| `translators/arch/*.py` | No | All Arch-specific logic correctly owned by the Arch translator |
| `translators/arch/capabilities.json` | No | mkinitcpio/limine tokens are translator-owned |
| `translators/debian/capabilities.json` | No (new file) | Correctly omits Arch-specific tokens |
| `cli/build/build.go` | YES — FIXED | `foundationRegistry` replaces hardcoded Arch references |

**DEB-03 conclusion:** The DebateOS abstraction is foundation-agnostic end-to-end. The single
genuine infrastructure leak was in the CLI dispatch layer (`build.go`), which has been fixed.
No schema field required modification. The capability gate (SC-2) correctly partitions
foundation-specific mechanics at translator boundaries.

---

## Cross-References

- `translators/debian/variant.py` — `apply_variant()` implements the `sig_level` → apt mapping (Finding 1)
- `translators/debian/capabilities.json` — omits mkinitcpio and limine tokens (Findings 3, 4)
- `cli/build/build.go` — `foundationRegistry` implementation (Finding 5)
- `cli/build/build_test.go` — `TestBuildFoundationDispatch`, `TestBuildArchUnchanged`,
  `TestBuildExplicitProfileOverride`, `TestBuildUnknownFoundation`
- `04-RESEARCH.md §Arch-Leak Audit` — source of all 6 findings with evidence citations
- `docs/09-decisions.md` — Invariant 1: translators own their mechanics
