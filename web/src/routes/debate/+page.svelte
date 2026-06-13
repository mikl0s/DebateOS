<script lang="ts">
	// Debate page — CSR only; ssr=false set in +page.ts
	// WASM is loaded here. Wave-2 will implement full DebateStage.
	import { onMount } from 'svelte';
	import { base } from '$app/paths';
	// base is used to prefix WASM asset URLs (UI-02 dual-delivery)

	let wasmLoaded = $state(false);
	let wasmError = $state(false);
	let wasmErrorMsg = $state('');

	onMount(async () => {
		try {
			// Dynamic import of wasm.ts deferred to onMount to avoid SSR issues.
			const { loadDebateosWasm } = await import('$lib/wasm.js');
			await loadDebateosWasm(base);
			wasmLoaded = true;
		} catch (e) {
			wasmError = true;
			wasmErrorMsg = e instanceof Error ? e.message : String(e);
			console.error('WASM load error:', e);
		}
	});
</script>

<svelte:head>
	<title>Debate — DebateOS</title>
</svelte:head>

<div
	style="
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-3xl) var(--spacing-lg);
		gap: var(--spacing-lg);
	"
>
	{#if !wasmLoaded && !wasmError}
		<!-- WasmLoadGate — UI-SPEC copy verbatim -->
		<div
			role="status"
			aria-live="polite"
			data-wasm-ready="false"
			style="
				text-align: center;
				color: var(--color-text-secondary);
				font-size: var(--font-size-body);
			"
		>
			Loading the resolver…
		</div>
	{:else if wasmError}
		<div
			role="alert"
			style="
				text-align: center;
				color: var(--color-conflict-hard);
				font-size: var(--font-size-body);
				max-width: 480px;
			"
		>
			The resolver failed to load. Refresh to try again. Your debate is saved in this browser
			session.
		</div>
	{:else}
		<!-- WASM loaded — DebateStage will be implemented in Wave 2 -->
		<div
			data-wasm-ready="true"
			style="text-align: center; color: var(--color-text-secondary);"
		>
			<p style="font-size: var(--font-size-heading); font-weight: 600; color: var(--color-text-primary);">
				Your debate has no points yet.
			</p>
			<p style="font-size: var(--font-size-body); margin-top: var(--spacing-sm);">
				Browse the registry to find points from curators you trust, or add your own opinions
				directly. Your foundation is chosen — time to take a stand.
			</p>
			<a
				href="{base}/browse/"
				style="
					display: inline-flex;
					align-items: center;
					justify-content: center;
					min-height: var(--min-height-touch);
					padding: 0 var(--spacing-xl);
					margin-top: var(--spacing-lg);
					background-color: var(--color-accent-brand);
					color: #ffffff;
					font-size: var(--font-size-body);
					font-weight: 600;
					text-decoration: none;
					border-radius: 6px;
				"
			>
				Browse Points
			</a>
		</div>
	{/if}
</div>
