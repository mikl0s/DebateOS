<script lang="ts">
	/**
	 * ExplanationCard — renders one resolver Explanation with verbatim text (A9).
	 *
	 * UI-SPEC §Explanation rendering:
	 *   - Rule badge (pill): "Rule 1" / "Rule 2" / etc.
	 *   - Explanation.Text verbatim (A9 contract — never substituted)
	 *   - Opinions involved (opinion-name chips)
	 *   - Kept row (green check + names)
	 *   - Dropped row (red X + names)
	 *   - Patch row (Puzzle icon + PatchOffered ID + Apply button)
	 *   - Trust warning banner (amber, verbatim)
	 */

	import type { ConflictView } from '$lib/conflict.js';
	import { CheckCircle2, XCircle, Puzzle, AlertTriangle } from '@lucide/svelte';

	interface Props {
		view: ConflictView;
		onApplyPatch?: (patchId: string) => void;
	}

	let { view, onApplyPatch }: Props = $props();

	// Map rule code → display badge text
	const ruleBadgeMap: Record<string, string> = {
		hard:     'Rule 2',
		warn:     'Rule 1',
		hardware: 'Hardware',
		compat:   'Compatible',
		patch:    'Rule 4',
		info:     'Info'
	};

	const ruleBadge = $derived(ruleBadgeMap[view.state] ?? view.state);

	// Badge color per state
	const badgeColorMap: Record<string, string> = {
		hard:     'var(--color-conflict-hard)',
		warn:     'var(--color-conflict-warn)',
		hardware: 'var(--color-conflict-warn)',
		compat:   'var(--color-conflict-compat)',
		patch:    'var(--color-accent-brand)',
		info:     'var(--color-text-secondary)'
	};

	const badgeColor = $derived(badgeColorMap[view.state] ?? badgeColorMap.info);
</script>

<section
	style="
		background-color: var(--color-surface-card);
		border: 1px solid var(--color-border-subtle);
		border-radius: 8px;
		padding: var(--spacing-md);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
	"
>
	<!-- Rule badge -->
	<div style="display: flex; align-items: center; gap: var(--spacing-sm);">
		<span
			style="
				display: inline-block;
				padding: 2px var(--spacing-sm);
				border-radius: 999px;
				background-color: {badgeColor};
				color: #ffffff;
				font-size: var(--font-size-label);
				font-weight: 600;
			"
		>
			{ruleBadge}
		</span>
	</div>

	<!-- Explanation text — A9: verbatim, never substituted -->
	{#if view.text}
		<p
			class="explanation-text"
			style="
				font-size: var(--font-size-body);
				line-height: var(--line-height-body);
				color: var(--color-text-primary);
				margin: 0;
			"
		>
			{view.text}
		</p>
	{/if}

	<!-- Opinions involved chips -->
	{#if view.opinionsInvolved.length > 0}
		<div style="display: flex; flex-wrap: wrap; gap: var(--spacing-xs);">
			{#each view.opinionsInvolved as opId}
				<span
					style="
						padding: 2px var(--spacing-sm);
						background-color: var(--color-surface-base);
						border: 1px solid var(--color-border-subtle);
						border-radius: 4px;
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
					"
				>
					{opId}
				</span>
			{/each}
		</div>
	{/if}

	<!-- Kept row -->
	{#if view.kept.length > 0}
		<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;">
			<CheckCircle2 size={16} color="var(--color-conflict-compat)" aria-hidden="true" />
			<span style="font-size: var(--font-size-label); color: var(--color-conflict-compat); font-weight: 600;">
				Kept:
			</span>
			{#each view.kept as keptId}
				<span style="font-size: var(--font-size-label); color: var(--color-text-secondary);">
					{keptId}
				</span>
			{/each}
		</div>
	{/if}

	<!-- Dropped row -->
	{#if view.dropped.length > 0}
		<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;">
			<XCircle size={16} color="var(--color-conflict-hard)" aria-hidden="true" />
			<span style="font-size: var(--font-size-label); color: var(--color-conflict-hard); font-weight: 600;">
				Dropped:
			</span>
			{#each view.dropped as droppedId}
				<span style="font-size: var(--font-size-label); color: var(--color-text-secondary);">
					{droppedId}
				</span>
			{/each}
		</div>
	{/if}

	<!-- Patch row -->
	{#if view.hasPatch}
		<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;">
			<Puzzle size={16} color="var(--color-accent-brand)" aria-hidden="true" />
			<span style="font-size: var(--font-size-label); color: var(--color-accent-brand);">
				Patch available: {view.patchOffered}
			</span>
			{#if onApplyPatch}
				<button
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
					Apply Resolution
				</button>
			{/if}
		</div>
	{/if}

	<!-- Trust warning banner -->
	{#if view.trustWarning}
		<div
			style="
				display: flex;
				align-items: flex-start;
				gap: var(--spacing-sm);
				padding: var(--spacing-sm) var(--spacing-md);
				background-color: rgba(245, 158, 11, 0.12);
				border: 1px solid var(--color-conflict-warn);
				border-radius: 6px;
			"
		>
			<AlertTriangle
				size={16}
				color="var(--color-conflict-warn)"
				style="flex-shrink: 0; margin-top: 2px;"
				aria-hidden="true"
			/>
			<p style="margin: 0; font-size: var(--font-size-label); color: var(--color-text-primary);">
				{view.trustWarning}
			</p>
		</div>
	{/if}
</section>
