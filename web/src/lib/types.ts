/**
 * TypeScript types mirroring the DebateOS resolver JSON shapes.
 * All field names match the Go JSON tags exactly (resolver/types.go, resolver/resolve/explanation.go).
 *
 * Key invariant: never add resolution logic here — types only.
 * The resolver is the Go WASM binary; JS/TS only parses and forwards.
 */

// ────────────────────────────────────────────────────────────────────────────
// Input types (mirroring resolver/types.go)
// ────────────────────────────────────────────────────────────────────────────

/** Reference to another opinion (OpinionRef). */
export interface OpinionRef {
	id: string;
	version?: string;
}

/** Reference to a patch opinion that resolves a known conflict. */
export interface PatchRef {
	id: string;
	resolves?: string;
	rationale?: string;
}

/** Ordering constraints for install order. */
export interface Ordering {
	before?: OpinionRef[];
	after?: OpinionRef[];
}

/** Hardware condition expression (discriminated union). */
export interface HardwareExpr {
	type: 'leaf' | 'and' | 'or' | 'not';
	predicate?: string;
	values?: string[];
	match?: string;
	operands?: HardwareExpr[];
}

/** File asset payload. */
export interface FileAsset {
	src: string;
	dst: string;
}

/** Custom repository declaration. */
export interface RepoDecl {
	name: string;
	url: string;
	sig_level: 'Required' | 'RequiredDatabaseOptional' | 'OptionalTrustAll' | 'Never';
	priority?: number;
	keyring?: string;
}

/** Runtime tool install (npm/pip/etc). */
export interface RuntimeToolInstall {
	manager: string;
	packages: string[];
}

/** Script payload. */
export interface ScriptPayload {
	script: string;
	capabilities?: string[];
}

/** Display manager configuration. */
export interface DisplayManagerConfig {
	name: string;
	greeter?: string;
	auto_login?: boolean;
}

/** Bootloader configuration. */
export interface BootloaderConfig {
	name: string;
	timeout?: number;
	snapshot?: boolean;
}

/** Service enable/disable declaration. */
export interface ServiceDecl {
	name: string;
	enable: boolean;
	deferred?: boolean;
}

/** Sysctl parameter. */
export interface SysctlParam {
	key: string;
	value: string;
	drop_in_file?: string;
}

/** Kernel boot parameter. */
export interface KernelParam {
	key: string;
	value?: string;
}

/** Group membership. */
export interface GroupMembership {
	group: string;
}

/** MIME association. */
export interface MimeAssoc {
	mime_pattern: string;
	app_id: string;
}

/** Theme bundle declaration. */
export interface ThemeDecl {
	bundle_dir: string;
	symlinks?: Record<string, string>;
	is_default?: boolean;
}

/**
 * Opinion — the atomic unit of DebateOS.
 * Mirrors resolver.Opinion (resolver/types.go).
 */
export interface Opinion {
	schema: number;
	id: string;
	name: string;
	category: string;
	intent?: string;
	status: 'required' | 'nice-to-have';

	depends_on?: OpinionRef[];
	conflicts?: OpinionRef[];
	known_patches?: PatchRef[];

	hardware_condition?: HardwareExpr;

	install_phase?: string;
	ordering?: Ordering;

	translator_capabilities?: string[];

	packages?: string[];
	remove_packages?: string[];
	file_assets?: FileAsset[];
	custom_repos?: RepoDecl[];
	runtime_tool_installs?: RuntimeToolInstall[];
	execution_phase?: string;
	script_payload?: ScriptPayload;
	display_manager?: DisplayManagerConfig;
	bootloader?: BootloaderConfig;
	services?: ServiceDecl[];
	sysctl_params?: SysctlParam[];
	kernel_params?: KernelParam[];
	group_memberships?: GroupMembership[];
	mime_associations?: MimeAssoc[];
	theme?: ThemeDecl;
}

/** Member of a Point bundle. */
export interface PointMember {
	id: string;
	status: 'required' | 'nice-to-have';
}

/** Point — a curated bundle of opinions (resolver.Point). */
export interface Point {
	schema: number;
	id: string;
	name: string;
	intent?: string;
	curator?: string;
	members: PointMember[];
}

/** Reference to a point in a speech. */
export interface PointRef {
	id: string;
	version?: string;
}

/**
 * HardwareProfile — declares the target machine's hardware facts.
 * CRITICAL: predicates is string[] NOT an object (Pitfall 6 guard).
 * resolver/types.go: HardwareProfile.Predicates is []string.
 */
export interface HardwareProfile {
	predicates: string[]; // MUST be array — resolver uses []string
	facts?: Record<string, string>;
	pci_ids?: string[];
}

/**
 * Speech — the user's complete composition.
 * Mirrors resolver.Speech (resolver/types.go).
 */
export interface Speech {
	schema: number;
	id?: string;
	name?: string;
	foundation: string;
	points: PointRef[];
	opinions?: OpinionRef[];
	hardware?: HardwareProfile;
}

// ────────────────────────────────────────────────────────────────────────────
// Output types (mirroring resolver/resolve/explanation.go)
// ────────────────────────────────────────────────────────────────────────────

/**
 * Explanation — one resolver decision record.
 * Mirrors resolve.Explanation (resolver/resolve/explanation.go).
 */
export interface Explanation {
	text: string;
	rule: string; // "rule1" | "rule2" | "rule3" | "rule4" | "hardware-skip" | "hardware-apply" | "ordering" | "cycle" | "sysctl-collision" | "no-conflict"
	opinions_involved?: string[];
	dropped?: string[];
	kept?: string[];
	patch_offered?: string;
	trust_warning?: string;
	alternative_suggestion?: string;
}

/**
 * ResolvedSpeech — the output of resolve.Resolve.
 * Mirrors resolve.ResolvedSpeech (resolver/resolve/explanation.go).
 * NOTE: result is ALWAYS present even on hard conflict (partial RS with explanations).
 */
export interface ResolvedSpeech {
	schema: number;
	foundation: string;
	install_order?: string[];
	applied?: string[];
	skipped?: string[];
	dropped?: string[];
	explanations: Explanation[];
}

// ────────────────────────────────────────────────────────────────────────────
// WASM call contract types
// ────────────────────────────────────────────────────────────────────────────

/**
 * ResolveInput — the JSON payload accepted by window.debateosResolve.
 * Mirrors resolveInput (resolver/wasm/main.go).
 */
export interface ResolveInput {
	speech: Speech;
	opinions: Opinion[];
	hardware: HardwareProfile;
}

/**
 * ResolveOutput — the raw JSON response from window.debateosResolve.
 * result is a JSON-encoded ResolvedSpeech string (present even on hard conflict).
 * error is non-empty when at least one hard conflict exists.
 */
export interface ResolveOutput {
	result?: string; // JSON-encoded ResolvedSpeech
	error?: string;
}

/**
 * ParsedResolveOutput — the parsed result of parseResolveOutput().
 * resolved is always present when parsing succeeds.
 * error is present on hard conflicts (alongside resolved).
 */
export interface ParsedResolveOutput {
	resolved: ResolvedSpeech;
	error?: string;
}
