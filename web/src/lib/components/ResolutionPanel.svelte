<script lang="ts">
	/**
	 * ResolutionPanel — right sidebar showing conflict explanations + build readiness.
	 *
	 * UI-SPEC §Core Debate Components:
	 *   ResolutionPanel
	 *   ├── ExplanationCard[]  — one per Explanation from resolver
	 *   └── BuildReadyBanner   — shown when Applied[] > 0 and no hard conflicts
	 *
	 * Copy verbatim from UI-SPEC §Copywriting contract.
	 */

	import type { ConflictView } from '$lib/conflict.js';
	import ExplanationCard from './ExplanationCard.svelte';
	import { base } from '$app/paths';

	interface Props {
		views: ConflictView[];
		applied: string[];
		isReady: boolean; // true = no hard conflicts
		onApplyPatch?: (patchId: string) => void;
	}

	let { views, applied, isReady, onApplyPatch }: Props = $props();

	// Filter non-info views for display in the panel
	const significantViews = $derived(views.filter((v) => v.state !== 'info'));
	const hasConflicts = $derived(views.some((v) => v.state === 'hard'));
</script>

<aside
	style="
		display: flex;
		flex-direction: column;
		gap: var(--spacing-md);
		padding: var(--spacing-md);
		background-color: var(--color-surface-card);
		border-left: 1px solid var(--color-border-subtle);
		overflow-y: auto;
		min-width: 280px;
		max-width: 380px;
	"
>
	<h2
		style="
			font-size: var(--font-size-heading);
			font-weight: 600;
			color: var(--color-text-primary);
			margin: 0;
		"
	>
		Resolution Panel
	</h2>

	{#if significantViews.length === 0 && applied.length === 0}
		<p style="color: var(--color-text-secondary); font-size: var(--font-size-body); margin: 0;">
			No conflicts detected. Add more points to your debate.
		</p>
	{/if}

	<!-- ExplanationCards -->
	{#each significantViews as view (view.state + view.text.slice(0, 20))}
		<ExplanationCard {view} {onApplyPatch} />
	{/each}

	<!-- BuildReadyBanner — shown when applied > 0 and no hard conflicts (UI-SPEC) -->
	{#if isReady && applied.length > 0 && !hasConflicts}
		<div
			style="
				border: 1px solid var(--color-conflict-compat);
				border-radius: 8px;
				padding: var(--spacing-md);
				background-color: rgba(34, 197, 94, 0.08);
				display: flex;
				flex-direction: column;
				gap: var(--spacing-sm);
			"
		>
			<h3
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-conflict-compat);
					margin: 0;
				"
			>
				Your speech is ready.
			</h3>
			<p style="font-size: var(--font-size-body); color: var(--color-text-secondary); margin: 0;">
				No conflicts remain. Run the command below to build your installer, or download the
				resolved speech YAML to build later.
			</p>
			<a
				href="{base}/export/"
				style="
					display: inline-flex;
					align-items: center;
					justify-content: center;
					min-height: var(--min-height-touch);
					padding: 0 var(--spacing-xl);
					background-color: var(--color-accent-brand);
					color: #ffffff;
					font-size: var(--font-size-body);
					font-weight: 600;
					text-decoration: none;
					border-radius: 6px;
					align-self: flex-start;
				"
			>
				Proceed to Build
			</a>
		</div>
	{/if}
</aside>
