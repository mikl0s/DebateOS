<script lang="ts">
	/**
	 * Debate page — CSR only; ssr=false set in +page.ts.
	 *
	 * Loads debateos.wasm in onMount, then drives:
	 *   - DebateStage (pane stack + conflict overlays)
	 *   - ResolutionPanel (right sidebar with ExplanationCards)
	 *
	 * WASM call: debateosResolve called on every speech-store change
	 * with 150ms debounce (UI-SPEC §WASM call contract).
	 *
	 * Invariant 3: never reimplements resolution — only calls WASM + maps output.
	 * A3 gate: Forum-offline path works because WASM is client-side only.
	 */

	import { onMount } from 'svelte';
	import { base } from '$app/paths';
	import { debate, allOpinions, resolvedSpeechStore } from '$lib/stores/speech.js';
	import { mapExplanation, type ConflictView } from '$lib/conflict.js';
	import DebateStage from '$lib/components/DebateStage.svelte';
	import ResolutionPanel from '$lib/components/ResolutionPanel.svelte';
	import type { ResolvedSpeech, Opinion, ResolveInput } from '$lib/types.js';

	let wasmLoaded = $state(false);
	let wasmError = $state(false);
	let wasmErrorMsg = $state('');

	// Resolved speech state (from WASM)
	let resolvedSpeech = $state<ResolvedSpeech | null>(null);
	let conflictViews = $state<ConflictView[]>([]);
	let resolveError = $state<string | null>(null);
	let resolving = $state(false);

	// Debounce timer
	let resolveTimer: ReturnType<typeof setTimeout> | null = null;

	onMount(async () => {
		try {
			// Dynamic import deferred to onMount — Pitfall 3 guard (no SSR)
			const { loadDebateosWasm } = await import('$lib/wasm.js');
			await loadDebateosWasm(base);
			wasmLoaded = true;

			// Subscribe to speech-store changes and trigger debounced resolve
			const unsub = debate.subscribe(() => {
				scheduleResolve();
			});

			return unsub;
		} catch (e) {
			wasmError = true;
			wasmErrorMsg = e instanceof Error ? e.message : String(e);
			console.error('WASM load error:', e);
		}
	});

	function scheduleResolve() {
		if (resolveTimer) clearTimeout(resolveTimer);
		resolveTimer = setTimeout(() => runResolve(), 150);
	}

	async function runResolve() {
		const { debateosResolve, buildResolveInput } = await import('$lib/wasm.js');
		const state = debate.snapshot();

		// Nothing to resolve if no panes
		if (state.panes.length === 0) {
			resolvedSpeech = null;
			conflictViews = [];
			resolveError = null;
			resolvedSpeechStore.set(null); // clear shared store (IN-04)
			return;
		}

		resolving = true;
		resolveError = null;

		try {
			// Build the speech input for the resolver
			const speech = {
				schema: 1,
				foundation: state.foundation,
				points: state.panes.map((p) => ({ id: p.pointId })),
				opinions: undefined,
				hardware: undefined
			};

			// Collect all opinions flat
			const opinions: Opinion[] = state.panes.flatMap((p) => p.opinions);

			const input: ResolveInput = buildResolveInput(speech, opinions, state.hardware);

			// Call WASM resolver (invariant 3 — never reimplements resolution)
			const { resolved, error } = debateosResolve(input);
			resolvedSpeech = resolved;
			resolveError = error ?? null;

			// Publish to shared store so export page can download the real result (IN-04).
			resolvedSpeechStore.set(resolved);

			// Map resolver explanations to ConflictViews (A1 + A9)
			conflictViews = (resolved.explanations ?? []).map(mapExplanation);
		} catch (e) {
			resolveError = e instanceof Error ? e.message : String(e);
			conflictViews = [];
		} finally {
			resolving = false;
		}
	}

	// Is the speech build-ready? (Applied > 0, no hard conflicts)
	const isReady = $derived(
		resolvedSpeech !== null &&
		(resolvedSpeech.applied?.length ?? 0) > 0 &&
		!conflictViews.some((v) => v.state === 'hard')
	);

	const appliedCount = $derived(resolvedSpeech?.applied?.length ?? 0);

	// Handlers
	function handleRemovePane(paneId: string) {
		debate.removePane(paneId);
	}

	function handleDropOpinion(opinionId: string) {
		// Remove panes containing this opinion
		const state = debate.snapshot();
		const paneWithOp = state.panes.find((p) =>
			p.opinions.some((o) => o.id === opinionId)
		);
		if (paneWithOp) debate.removePane(paneWithOp.id);
	}

	function handleApplyPatch(_patchId: string) {
		// v1: patch application requires a registry lookup (deferred to Forum
		// integration). Show honest user feedback instead of silently re-resolving
		// with identical state (which looked like a broken button). (WR-02)
		resolveError = 'Patch application is not yet supported in v1. Add the patch opinion manually.';
	}

	// Demo: add conflicting points for testing (used by e2e via page.evaluate).
	// Gated on import.meta.env.DEV so these globals are absent in production builds
	// and cannot be abused to inject arbitrary opinion data. (WR-04)
	if (typeof window !== 'undefined' && import.meta.env.DEV) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		(window as any).debateAddTestPane = (pointId: string, pointName: string, opinions: Opinion[]) => {
			debate.addPane(pointId, pointName, opinions);
		};
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		(window as any).debateGetResolved = () => resolvedSpeech;
	}
</script>

<svelte:head>
	<title>Debate — DebateOS</title>
</svelte:head>

{#if !wasmLoaded && !wasmError}
	<!-- WasmLoadGate — UI-SPEC copy verbatim -->
	<div
		role="status"
		aria-live="polite"
		data-wasm-ready="false"
		style="
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			color: var(--color-text-secondary);
			font-size: var(--font-size-body);
		"
	>
		Loading the resolver…
	</div>
{:else if wasmError}
	<div
		role="alert"
		data-wasm-ready="error"
		style="
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			color: var(--color-conflict-hard);
			font-size: var(--font-size-body);
			text-align: center;
			max-width: 480px;
			margin: 0 auto;
		"
	>
		The resolver failed to load. Refresh to try again. Your debate is saved in this browser session.
	</div>
{:else}
	<!-- Debate compose layout: stage (left/center) + resolution panel (right) -->
	<div
		data-wasm-ready="true"
		style="
			flex: 1;
			display: flex;
			overflow: hidden;
		"
	>
		<!-- Left: Debate Stage (FoundationBar + PaneStack) -->
		<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; overflow: hidden;">
			<DebateStage
				foundation={$debate.foundation}
				panes={$debate.panes}
				{conflictViews}
				onRemovePane={handleRemovePane}
				onDropOpinion={handleDropOpinion}
				onApplyPatch={handleApplyPatch}
			/>

			<!-- Live resolve indicator -->
			{#if resolving}
				<div
					role="status"
					aria-live="polite"
					style="
						padding: var(--spacing-xs) var(--spacing-md);
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
						background-color: var(--color-surface-card);
						border-top: 1px solid var(--color-border-subtle);
					"
				>
					Resolving…
				</div>
			{:else if resolvedSpeech}
				<div
					role="status"
					aria-live="polite"
					style="
						padding: var(--spacing-xs) var(--spacing-md);
						font-size: var(--font-size-label);
						color: var(--color-conflict-compat);
						background-color: var(--color-surface-card);
						border-top: 1px solid var(--color-border-subtle);
					"
				>
					Resolved · {appliedCount} opinion{appliedCount !== 1 ? 's' : ''} applied
				</div>
			{/if}
		</div>

		<!-- Right: Resolution Panel -->
		<ResolutionPanel
			views={conflictViews}
			applied={resolvedSpeech?.applied ?? []}
			{isReady}
			onApplyPatch={handleApplyPatch}
		/>
	</div>
{/if}
