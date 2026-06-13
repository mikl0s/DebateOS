# Dual-Foundation Proof (DEB-02)

This directory contains the **dual-foundation representative speech** — the
concrete proof that DebateOS's abstraction is real (Invariant 1): a single,
foundation-neutral speech that both the Arch translator and the Debian
translator can effectuate without any distribution-specific assumption leaking
through the schema.

## What this is

`speech.yaml` (targeting `foundation: debian`) is a **fresh, Omarchy-independent**
composition. Omarchy is Arch-specific by nature; this speech is deliberately
authored from scratch using only OS-agnostic opinion categories.

The same `resolved.json` produced from this speech is fed to:
1. `translators/arch/translate` — produces an Arch profile tree
2. `translators/debian/translate` — produces a Debian profile tree

Both translators effectuate every `required` opinion. This is verified
end-to-end by `scripts/dual-foundation-check.sh` (Plan 04-05) and at the
resolve level by `TestExampleDualFoundation` in `examples/dual_foundation_test.go`.

## Opinions (all 5 are required)

| ID     | Category     | Capability Token           | Why OS-agnostic |
|--------|-------------|----------------------------|-----------------|
| DF-001 | package-install | install-packages          | `git`, `curl`, `vim` exist under the same upstream names in both Arch and Debian repos |
| DF-002 | config-file  | deploy-config-file-tree    | File copy to `etc/motd`; no distribution mechanic beyond writing a file |
| DF-003 | service      | enable-systemd-service     | `systemd-timesyncd.service` is the same unit name on both foundations; enabling a systemd unit is foundation-neutral |
| DF-004 | sysctl       | write-sysctl-drop-in       | `/etc/sysctl.d/` exists on both Arch and Debian; the kernel parameter key/value is identical |
| DF-005 | user-group   | add-user-to-group          | The `video` group exists on both foundations; the `usermod`/`gpasswd` mechanic is translator-owned |

## What is NOT here

- No Arch-specific tokens: no `configure-mkinitcpio-hooks-and-modules`, no `manage-limine-bootloader-installation`, no `add-custom-package-repo` with pacman sig_level mechanics.
- No Debian-specific tokens: no apt sources, no preseed mechanics, no `dpkg` configurations.
- The Omarchy speech (`examples/omarchy/`) is intentionally untouched; it remains Arch-only.

## Resolution properties

- All 5 opinions are `required` with no `depends_on` or `conflicts` entries.
- No `hardware_condition` — applies on any baseline hardware profile.
- Resolves **clean**: 0 hard conflicts, 0 dropped, 0 skipped, 5 applied.
- `TestExampleDualFoundation` in `examples/dual_foundation_test.go` is the automated clean-resolution gate.

## Licensing

CC0-1.0 — see `examples/LICENSE`.
