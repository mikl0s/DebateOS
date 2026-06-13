<script lang="ts">
	/**
	 * FoundationBar — fixed 56px strip showing foundation name + version + swap button.
	 *
	 * UI-SPEC: height=56px (--height-foundation-bar).
	 * Uses canonical vocabulary: "Foundation", not "distro" / "preset" / "config" (A6).
	 */

	interface Props {
		foundation: string;
		onSwap?: () => void;
	}

	let { foundation, onSwap }: Props = $props();
</script>

<div
	style="
		height: var(--height-foundation-bar);
		background-color: var(--color-surface-card);
		border-bottom: 1px solid var(--color-border-subtle);
		display: flex;
		align-items: center;
		padding: 0 var(--spacing-lg);
		gap: var(--spacing-md);
		flex-shrink: 0;
	"
>
	<span
		style="
			font-size: var(--font-size-label);
			color: var(--color-text-secondary);
			text-transform: uppercase;
			letter-spacing: 0.06em;
		"
	>
		Foundation
	</span>
	<span
		style="
			font-size: var(--font-size-heading);
			font-weight: 600;
			color: var(--color-text-primary);
			text-transform: capitalize;
		"
	>
		{foundation}
	</span>

	<div style="flex: 1;"></div>

	{#if onSwap}
		<button
			onclick={onSwap}
			style="
				min-height: var(--min-height-touch);
				padding: 0 var(--spacing-md);
				border: 1px solid var(--color-border-subtle);
				border-radius: 6px;
				background: transparent;
				color: var(--color-text-secondary);
				font-size: var(--font-size-label);
				cursor: pointer;
				display: flex;
				align-items: center;
				gap: var(--spacing-xs);
			"
		>
			Swap Foundation
		</button>
	{/if}
</div>
