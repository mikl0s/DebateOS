/**
 * WASM loader and typed resolver wrapper for DebateOS.
 *
 * Invariant 3: the UI NEVER reimplements resolution logic.
 * All resolution goes through window.debateosResolve() (the Go WASM binary).
 * This module only loads the WASM, forwards JSON, and parses the response.
 *
 * T-05-04 mitigation: pure helpers (buildResolveInput, parseResolveOutput) only
 * parse and forward — they make no resolution decisions.
 *
 * Pitfall 3 guard: loadDebateosWasm() must be called from onMount() only.
 * Never import at module level in SSR contexts.
 */

import type {
	HardwareProfile,
	Opinion,
	ParsedResolveOutput,
	ResolveInput,
	ResolveOutput,
	ResolvedSpeech,
	Speech
} from './types.js';

export type {
	HardwareProfile,
	Opinion,
	ParsedResolveOutput,
	ResolveInput,
	ResolveOutput,
	ResolvedSpeech,
	Speech
};

let wasmReady = false;

/**
 * loadDebateosWasm — loads debateos.wasm and registers window.debateosResolve.
 *
 * Must be called from onMount() (CSR only — Pitfall 3).
 * Idempotent: subsequent calls return immediately if WASM is already loaded.
 *
 * @param base - The SvelteKit base path (from $app/paths). Empty string for localhost.
 */
export async function loadDebateosWasm(base = ''): Promise<void> {
	if (wasmReady) return;

	// wasm_exec.js adds globalThis.Go — copy at build time from GOROOT (T-05-03).
	// Using dynamic import with the base-prefixed URL.
	await import(/* @vite-ignore */ `${base}/wasm_exec.js`);

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const go = new (globalThis as any).Go();

	const result = await WebAssembly.instantiateStreaming(
		fetch(`${base}/debateos.wasm`), // Pitfall 4 guard: base prefix for dual delivery
		go.importObject
	);

	// go.run() is non-blocking (WASM main() does select{}).
	// init() in main.go registers window.debateosResolve synchronously before main().
	go.run(result.instance);

	wasmReady = true;
}

/**
 * buildResolveInput — constructs a valid ResolveInput from components.
 *
 * Pitfall 6 guard: enforces predicates as string[] (never an object).
 * This is a PURE function — testable without WASM loaded.
 */
export function buildResolveInput(
	speech: Speech,
	opinions: Opinion[],
	hardware: Partial<HardwareProfile>
): ResolveInput {
	return {
		speech,
		opinions,
		hardware: {
			predicates: hardware.predicates ?? [], // MUST be array (Pitfall 6)
			facts: hardware.facts,
			pci_ids: hardware.pci_ids
		}
	};
}

/**
 * parseResolveOutput — parses the raw JSON string returned by window.debateosResolve.
 *
 * This is a PURE function — testable without WASM loaded.
 *
 * Result behavior (mirrors resolver/wasm/main.go resolveOutput contract):
 * - result present, no error → clean resolve
 * - result present + error → hard conflict (partial RS with explanations)
 * - no result → unrecoverable error → throws Error
 *
 * @throws Error if raw JSON has no result field (unrecoverable resolver error).
 */
export function parseResolveOutput(raw: string): ParsedResolveOutput {
	const out: ResolveOutput = JSON.parse(raw);

	if (!out.result) {
		// No result means an unrecoverable error (empty input, parse failure, etc.)
		throw new Error(out.error ?? 'resolver returned no result');
	}

	const resolved: ResolvedSpeech = JSON.parse(out.result);

	return {
		resolved,
		error: out.error
	};
}

/**
 * debateosResolve — calls window.debateosResolve and parses the output.
 *
 * Invariant 3: this function calls the WASM resolver; it does not decide anything.
 * Requires loadDebateosWasm() to have been called first.
 *
 * @throws Error if resolver returns no result (pass-through from parseResolveOutput).
 */
export function debateosResolve(input: ResolveInput): ParsedResolveOutput {
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const raw: string = (globalThis as any).debateosResolve(JSON.stringify(input));
	return parseResolveOutput(raw);
}
