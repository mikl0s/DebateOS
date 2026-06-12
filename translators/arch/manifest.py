"""
manifest.py — BuildManifest dataclass for the Arch translator.

Aggregates the full payload from a ResolvedSpeech + opinion bodies into a
single typed in-memory structure that the profile emitter (Plan 02) and
the installer script (Plans 03-05) consume.

Key design decisions (from 02-CONTEXT.md and 02-RESEARCH.md):
- install_order from ResolvedSpeech is authoritative; never re-sort.
- packages.x86_64 (live env) is OUT of this manifest — that is fixed
  releng-baseline territory (Pitfall 2). This manifest is the TARGET install
  set only.
- All opinion payload strings are serialized as JSON data (Pitfall 6 /
  T-02-01): the installer reads build-manifest.json, never shell-interpolates
  raw strings.
- check_capabilities (ARCH-03 / SC-3) is called BEFORE any assembly begins.
- custom_repos with sig_level="Never" are flagged in trust_warnings (T-02-02).
- execution_phase="first-run" opinions become systemd first-boot units;
  they are collected into the first_run list, distinct from install-time ops.
- SOURCE_DATE_EPOCH is derived deterministically from resolved-speech bytes
  (BLD-03 groundwork) via derive_source_date_epoch.
"""

import hashlib
import struct
from dataclasses import dataclass, field
from typing import List, Dict, Any

from capabilities import check_capabilities


# ---------------------------------------------------------------------------
# derive_source_date_epoch (Pattern 3 from 02-RESEARCH.md)
# ---------------------------------------------------------------------------

_MIN_EPOCH = 1577836800  # 2020-01-01T00:00:00Z
_MAX_EPOCH = 2208988800  # 2040-01-01T00:00:00Z


def derive_source_date_epoch(content_bytes: bytes) -> int:
    """Derive a deterministic SOURCE_DATE_EPOCH from resolved-speech bytes.

    Algorithm (02-RESEARCH.md §Pattern 3):
    1. SHA-256 hash of ``content_bytes``.
    2. Take the first 4 bytes as a big-endian uint32.
    3. Clamp to [2020-01-01, 2040-01-01) via modulo + MIN.

    Guarantees: same bytes → same integer; value always in valid epoch range.
    Uses only stdlib hashlib + struct (no custom crypto).

    Args:
        content_bytes: The raw bytes of the ResolvedSpeech JSON (or any stable
            byte representation of the resolved speech content).

    Returns:
        A stable int in [1577836800, 2208988800).
    """
    digest = hashlib.sha256(content_bytes).digest()
    raw = struct.unpack(">I", digest[:4])[0]
    return _MIN_EPOCH + (raw % (_MAX_EPOCH - _MIN_EPOCH))


# ---------------------------------------------------------------------------
# BuildManifest dataclass
# ---------------------------------------------------------------------------


@dataclass
class BuildManifest:
    """Fully-aggregated, typed build manifest derived from a resolved speech.

    Every field is a plain Python type (list, dict, int, str) so that
    ``to_dict()`` produces a JSON-serializable payload without custom encoding
    (T-02-01 data-as-JSON pattern, Pitfall 6).

    Fields
    ------
    foundation : str
        Target foundation from the resolved speech (e.g. "arch").
    install_order : List[str]
        Authoritative opinion ID order from ResolvedSpeech.install_order.
    target_packages : List[str]
        Union of Opinion.packages across all applied opinions, in install_order,
        deduplicated with first-occurrence semantics.
    remove_packages : List[str]
        Union of Opinion.remove_packages across applied opinions.
    file_assets : List[dict]
        Aggregated FileAsset records ({src, dst}) across applied opinions.
    system_services : List[dict]
        ServiceDecl records with deferred=False (enable at install time).
    deferred_services : List[dict]
        ServiceDecl records with deferred=True (enable at first-boot time).
    first_run : List[dict]
        Opinions with execution_phase=="first-run"; each entry is
        {"id": ..., "script_payload": ...}.
    sysctl_params : List[dict]
        SysctlParam records ({key, value, drop_in_file}) across applied opinions.
    kernel_params : List[dict]
        KernelParam records ({key, value}) across applied opinions.
    group_memberships : List[dict]
        GroupMembership records ({group}) across applied opinions.
    mime_associations : List[dict]
        MimeAssoc records ({mime_pattern, app_id}) across applied opinions.
    themes : List[dict]
        ThemeDecl records ({bundle_dir, symlinks, is_default}) across applied.
    custom_repos : List[dict]
        RepoDecl records ({name, url, sig_level, priority, keyring}) across applied.
    trust_warnings : List[str]
        Human-readable warnings for every custom_repo with sig_level="Never"
        (T-02-02).  Surfaced as pacman.conf comments in Plan 02.
    source_date_epoch : int
        Deterministic SOURCE_DATE_EPOCH derived from resolved-speech bytes.
    """

    foundation: str
    install_order: List[str]
    target_packages: List[str]
    remove_packages: List[str]
    file_assets: List[Dict[str, Any]]
    system_services: List[Dict[str, Any]]
    deferred_services: List[Dict[str, Any]]
    first_run: List[Dict[str, Any]]
    sysctl_params: List[Dict[str, Any]]
    kernel_params: List[Dict[str, Any]]
    group_memberships: List[Dict[str, Any]]
    mime_associations: List[Dict[str, Any]]
    themes: List[Dict[str, Any]]
    custom_repos: List[Dict[str, Any]]
    trust_warnings: List[str]
    source_date_epoch: int

    # -------------------------------------------------------------------------
    # Factory classmethod
    # -------------------------------------------------------------------------

    @classmethod
    def from_resolved(
        cls,
        resolved: dict,
        opinions_index: dict,
        capabilities: set,
        resolved_bytes: bytes,
    ) -> "BuildManifest":
        """Construct a BuildManifest from a resolved speech + opinion bodies.

        Steps:
        1. Run the capability gate (check_capabilities) — raises CapabilityError
           before any assembly if a required opinion needs an undeclared capability.
        2. Iterate over install_order aggregating all payload fields.
        3. Derive source_date_epoch from resolved_bytes.

        Args:
            resolved: Loaded ResolvedSpeech dict (keys: applied, skipped,
                dropped, install_order, explanations).
            opinions_index: Dict mapping opinion ID str → opinion dict.
            capabilities: Set of declared capability token strings.
            resolved_bytes: Raw bytes of the ResolvedSpeech JSON, used to
                derive the deterministic SOURCE_DATE_EPOCH.

        Returns:
            A fully-populated BuildManifest.

        Raises:
            CapabilityError: If a required opinion needs an undeclared capability
                (ARCH-03 / SC-3 gate fires before any assembly).
        """
        # SC-3 / ARCH-03: gate runs BEFORE any manifest assembly.
        check_capabilities(resolved, opinions_index, capabilities)

        install_order = resolved.get("install_order", [])
        foundation = resolved.get("foundation", "")

        # Aggregation accumulators
        target_packages: List[str] = []
        seen_packages: set = set()
        remove_packages: List[str] = []
        file_assets: List[Dict[str, Any]] = []
        system_services: List[Dict[str, Any]] = []
        deferred_services: List[Dict[str, Any]] = []
        first_run: List[Dict[str, Any]] = []
        sysctl_params: List[Dict[str, Any]] = []
        kernel_params: List[Dict[str, Any]] = []
        group_memberships: List[Dict[str, Any]] = []
        mime_associations: List[Dict[str, Any]] = []
        themes: List[Dict[str, Any]] = []
        custom_repos: List[Dict[str, Any]] = []
        trust_warnings: List[str] = []

        for opinion_id in install_order:
            opinion = opinions_index.get(opinion_id)
            if opinion is None:
                continue

            exec_phase = opinion.get("execution_phase", "")

            if exec_phase == "first-run":
                # First-run opinions: collect id + script_payload for
                # systemd oneshot unit generation in Plan 02.
                entry: Dict[str, Any] = {"id": opinion_id}
                if "script_payload" in opinion:
                    entry["script_payload"] = opinion["script_payload"]
                first_run.append(entry)
                # First-run opinions do NOT contribute to the install-time package
                # or service sets — they run in the live user session.
                continue

            # Packages — dedup preserving first occurrence.
            for pkg in opinion.get("packages", []):
                if pkg not in seen_packages:
                    target_packages.append(pkg)
                    seen_packages.add(pkg)

            # Remove packages
            remove_packages.extend(opinion.get("remove_packages", []))

            # File assets
            for fa in opinion.get("file_assets", []):
                file_assets.append(dict(fa))

            # Services — split by deferred flag.
            for svc in opinion.get("services", []):
                svc_copy = dict(svc)
                if svc.get("deferred", False):
                    deferred_services.append(svc_copy)
                else:
                    system_services.append(svc_copy)

            # Sysctl params
            for param in opinion.get("sysctl_params", []):
                sysctl_params.append(dict(param))

            # Kernel params
            for param in opinion.get("kernel_params", []):
                kernel_params.append(dict(param))

            # Group memberships
            for gm in opinion.get("group_memberships", []):
                group_memberships.append(dict(gm))

            # MIME associations
            for ma in opinion.get("mime_associations", []):
                mime_associations.append(dict(ma))

            # Theme
            if "theme" in opinion and opinion["theme"] is not None:
                themes.append(dict(opinion["theme"]))

            # Custom repos + trust warnings for sig_level=Never (T-02-02)
            for repo in opinion.get("custom_repos", []):
                repo_copy = dict(repo)
                custom_repos.append(repo_copy)
                if repo.get("sig_level") == "Never":
                    trust_warnings.append(
                        f"WARNING: repo '{repo['name']}' has sig_level=Never "
                        f"(opinion {opinion_id}) — unsigned packages accepted; "
                        f"verify repo source before use."
                    )

        source_date_epoch = derive_source_date_epoch(resolved_bytes)

        return cls(
            foundation=foundation,
            install_order=list(install_order),
            target_packages=target_packages,
            remove_packages=remove_packages,
            file_assets=file_assets,
            system_services=system_services,
            deferred_services=deferred_services,
            first_run=first_run,
            sysctl_params=sysctl_params,
            kernel_params=kernel_params,
            group_memberships=group_memberships,
            mime_associations=mime_associations,
            themes=themes,
            custom_repos=custom_repos,
            trust_warnings=trust_warnings,
            source_date_epoch=source_date_epoch,
        )

    # -------------------------------------------------------------------------
    # Serialization
    # -------------------------------------------------------------------------

    def to_dict(self) -> Dict[str, Any]:
        """Serialize the manifest to a plain JSON-serializable dict.

        The installer script reads this as ``build-manifest.json`` at runtime —
        it must not contain dataclass instances or any non-JSON-native types
        (T-02-01: data-as-JSON, Pitfall 6).

        Returns:
            A ``dict`` with all manifest fields as plain Python types.
        """
        return {
            "foundation": self.foundation,
            "install_order": self.install_order,
            "target_packages": self.target_packages,
            "remove_packages": self.remove_packages,
            "file_assets": self.file_assets,
            "system_services": self.system_services,
            "deferred_services": self.deferred_services,
            "first_run": self.first_run,
            "sysctl_params": self.sysctl_params,
            "kernel_params": self.kernel_params,
            "group_memberships": self.group_memberships,
            "mime_associations": self.mime_associations,
            "themes": self.themes,
            "custom_repos": self.custom_repos,
            "trust_warnings": self.trust_warnings,
            "source_date_epoch": self.source_date_epoch,
        }
