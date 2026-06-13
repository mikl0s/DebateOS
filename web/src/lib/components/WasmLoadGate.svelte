<script lang="ts">
	/**
	 * WasmLoadGate — blocks the debate UI until debateos.wasm is loaded.
	 *
	 * role="status" aria-live="polite" per UI-SPEC §Screen reader.
	 * Copy verbatim from UI-SPEC §Copywriting contract.
	 */

	interface Props {
		loaded: boolean;
		error: boolean;
		errorMsg?: string;
	}

	let { loaded, error, errorMsg = '' }: Props = $props();
</script>

{#if !loaded && !error}
	<div
		role="status"
		aria-live="polite"
		data-wasm-ready="false"
		style="
			display: flex;
			align-items: center;
			justify-content: center;
			min-height: 200px;
			color: var(--color-text-secondary);
			font-size: var(--font-size-body);
			text-align: center;
		"
	>
		Loading the resolver…
	</div>
{:else if error}
	<div
		role="alert"
		data-wasm-ready="error"
		style="
			display: flex;
			align-items: center;
			justify-content: center;
			min-height: 200px;
			color: var(--color-conflict-hard);
			font-size: var(--font-size-body);
			text-align: center;
			max-width: 480px;
			margin: 0 auto;
		"
	>
		<p>
			The resolver failed to load. Refresh to try again. Your debate is saved in this browser
			session.
			{#if errorMsg}
				<br /><small style="font-size: var(--font-size-label); opacity: 0.7;">{errorMsg}</small>
			{/if}
		</p>
	</div>
{:else}
	<slot />
{/if}
