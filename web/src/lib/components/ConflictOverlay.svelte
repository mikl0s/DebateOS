<script lang="ts">
	/**
	 * ConflictOverlay — absolute-positioned conflict region (A1 triple encoding).
	 *
	 * UI-SPEC §Conflict Visualization Spec:
	 *   - Background: conflict color at reduced opacity (fill cue)
	 *   - Border: 2px solid conflict color (border cue)
	 *   - ConflictBadge: icon + text label (icon + text cues — A1)
	 *   - data-conflict-state attr for test assertions
	 *   - role="status" aria-label="[State]: A vs B" (UI-SPEC §Screen reader)
	 *
	 * Color-only is FORBIDDEN — triple encoding enforced (T-05-12 mitigation).
	 */

	import type { ConflictView } from '$lib/conflict.js';
	import ConflictBadge from './ConflictBadge.svelte';

	interface Props {
		view: ConflictView;
		opinionsLabel?: string; // e.g. "OM-015 vs OM-015-greetd"
		onDrop?: (opinionId: string) => void;
		onApplyPatch?: (patchId: string) => void;
	}

	let { view, opinionsLabel = '', onDrop, onApplyPatch }: Props = $props();

	// Map state to token color and opacity
	const styleMap: Record<string, { bg: string; bgOpacity: number; border: string; borderStyle: string }> = {
		hard:     { bg: 'var(--color-conflict-hard)',  bgOpacity: 0.18, border: 'var(--color-conflict-hard)',  borderStyle: 'solid' },
		warn:     { bg: 'var(--color-conflict-warn)',  bgOpacity: 0.14, border: 'var(--color-conflict-warn)',  borderStyle: 'solid' },
		hardware: { bg: 'var(--color-conflict-warn)',  bgOpacity: 0.14, border: 'var(--color-conflict-warn)',  borderStyle: 'dashed' },
		compat:   { bg: 'var(--color-conflict-compat)', bgOpacity: 0.12, border: 'var(--color-conflict-compat)', borderStyle: 'dashed' },
		patch:    { bg: 'var(--color-accent-brand)',   bgOpacity: 0.10, border: 'var(--color-accent-brand)',   borderStyle: 'solid' },
		info:     { bg: 'transparent',                bgOpacity: 0,    border: 'var(--color-border-subtle)',   borderStyle: 'solid' }
	};

	const style = $derived(styleMap[view.state] ?? styleMap.info);

	const ariaLabel = $derived(
		`${view.label}${opinionsLabel ? ': ' + opinionsLabel : ''}`
	);

	// Compute the CSS background-color value with opacity embedded in rgba
	// for Playwright assertions to detect the red fill on hard conflicts.
	const bgColor = $derived(
		view.state === 'hard'
			? `rgba(239, 68, 68, ${style.bgOpacity})`
			: view.state === 'warn' || view.state === 'hardware'
			? `rgba(245, 158, 11, ${style.bgOpacity})`
			: view.state === 'compat'
			? `rgba(34, 197, 94, ${style.bgOpacity})`
			: view.state === 'patch'
			? `rgba(99, 102, 241, ${style.bgOpacity})`
			: 'transparent'
	);

	// For Playwright: data-conflict-bg-rgb exposes the pure base RGB for color assertions.
	const bgRgb = $derived(
		view.state === 'hard' ? '239,68,68' :
		view.state === 'warn' || view.state === 'hardware' ? '245,158,11' :
		view.state === 'compat' ? '34,197,94' :
		view.state === 'patch' ? '99,102,241' : ''
	);
</script>

{#if view.state !== 'info'}
	<div
		role="status"
		aria-label={ariaLabel}
		data-conflict-state={view.state}
		data-conflict-bg-rgb={bgRgb}
		style="
			position: relative;
			border-radius: 6px;
			padding: var(--spacing-sm) var(--spacing-md);
			background-color: {bgColor};
			border: 2px {style.borderStyle} {style.border};
			display: flex;
			flex-direction: column;
			gap: var(--spacing-sm);
		"
	>
		<!-- Badge: icon + label = 2nd + 3rd encoding (A1) -->
		<!-- Note: background-color above is 1st encoding (color cue) -->
		<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;">
			<ConflictBadge state={view.state} icon={view.icon} label={view.label} />

			{#if view.hasPatch}
				<ConflictBadge state="patch" icon="Puzzle" label="Patch available" />
			{/if}
		</div>

		<!-- Explanation text (body copy if non-empty) -->
		{#if view.text}
			<p
				style="
					font-size: var(--font-size-body);
					color: var(--color-text-secondary);
					margin: 0;
					line-height: var(--line-height-body);
				"
			>
				{view.text}
			</p>
		{/if}

		<!-- Actions (conflict-action-btn class for A2 touch target assertion) -->
		{#if view.state === 'hard' || view.state === 'warn'}
			<div style="display: flex; gap: var(--spacing-sm); flex-wrap: wrap;">
				{#each view.dropped as droppedId}
					{#if onDrop}
						<button
							class="conflict-action-btn"
							onclick={() => onDrop?.(droppedId)}
							style="
								min-height: var(--min-height-touch);
								padding: 0 var(--spacing-md);
								border: 1px solid var(--color-destructive);
								border-radius: 6px;
								background: transparent;
								color: var(--color-destructive);
								font-size: var(--font-size-label);
								cursor: pointer;
								display: flex;
								align-items: center;
							"
						>
							Drop '{droppedId}'
						</button>
					{/if}
				{/each}

				{#if view.hasPatch && onApplyPatch}
					<button
						class="conflict-action-btn"
						onclick={() => onApplyPatch?.(view.patchOffered)}
						style="
							min-height: var(--min-height-touch);
							padding: 0 var(--spacing-md);
							border: 1px solid var(--color-accent-brand);
							border-radius: 6px;
							background: var(--color-accent-brand);
							color: #ffffff;
							font-size: var(--font-size-label);
							cursor: pointer;
							display: flex;
							align-items: center;
						"
					>
						Apply Patch
					</button>
				{/if}
			</div>
		{/if}
	</div>
{/if}
