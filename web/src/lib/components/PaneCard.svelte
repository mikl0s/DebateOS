<script lang="ts">
	/**
	 * PaneCard — per-point pane in the debate stage.
	 *
	 * UI-SPEC §Core Debate Components:
	 *   PaneCard: role="region" aria-label="[Point name] pane" (A2 — screen reader).
	 *   PaneHeader: 44px min touch target (A2 — WCAG 2.5.5).
	 *   Glass overlay: surface-pane rgba background.
	 *
	 * Copy verbatim from UI-SPEC §Copywriting contract.
	 */

	import type { Pane } from '$lib/stores/speech.js';
	import type { ConflictView } from '$lib/conflict.js';
	import ConflictOverlay from './ConflictOverlay.svelte';
	import { XCircle } from '@lucide/svelte';

	interface Props {
		pane: Pane;
		conflictViews: ConflictView[]; // views relevant to this pane's opinions
		onRemove?: (paneId: string) => void;
		onDrop?: (opinionId: string) => void;
		onApplyPatch?: (patchId: string) => void;
		isActive?: boolean;
	}

	let { pane, conflictViews = [], onRemove, onDrop, onApplyPatch, isActive = false }: Props = $props();

	let showRemoveConfirm = $state(false);

	function handleRemove() {
		if (showRemoveConfirm) {
			onRemove?.(pane.id);
			showRemoveConfirm = false;
		} else {
			showRemoveConfirm = true;
		}
	}

	function cancelRemove() {
		showRemoveConfirm = false;
	}

	// Filter conflict views that relate to this pane's opinions
	const paneOpinionIds = $derived(new Set(pane.opinions.map((o) => o.id)));
	const paneConflicts = $derived(
		conflictViews.filter(
			(v) =>
				v.state !== 'info' &&
				v.opinionsInvolved.some((id) => paneOpinionIds.has(id))
		)
	);
</script>

<div
	role="region"
	aria-label="{pane.pointName} pane"
	data-pane-id={pane.id}
	data-pane-active={isActive ? 'true' : 'false'}
	style="
		background-color: var(--color-surface-pane);
		border: 1px solid {isActive ? 'var(--color-accent-brand)' : 'var(--color-border-subtle)'};
		border-radius: 8px;
		overflow: hidden;
		{isActive ? 'box-shadow: 0 0 0 2px var(--color-accent-brand);' : ''}
	"
>
	<!-- Pane header — 44px min touch target (A2) -->
	<div
		class="pane-header"
		style="
			min-height: var(--min-height-touch);
			padding: 0 var(--spacing-md);
			display: flex;
			align-items: center;
			gap: var(--spacing-sm);
			background-color: var(--color-surface-card);
			border-bottom: 1px solid var(--color-border-subtle);
		"
	>
		<!-- Point name -->
		<div style="flex: 1; display: flex; flex-direction: column; gap: 2px;">
			<span
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-text-primary);
					line-height: var(--line-height-heading);
				"
			>
				{pane.pointName}
			</span>
			{#if pane.curator}
				<span
					style="
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
					"
				>
					by {pane.curator}
				</span>
			{/if}
		</div>

		<!-- Remove button -->
		{#if !showRemoveConfirm}
			<button
				class="conflict-action-btn"
				onclick={handleRemove}
				aria-label="Remove {pane.pointName} pane"
				style="
					min-height: var(--min-height-touch);
					min-width: var(--min-height-touch);
					display: flex;
					align-items: center;
					justify-content: center;
					border: none;
					background: transparent;
					color: var(--color-text-secondary);
					cursor: pointer;
					border-radius: 4px;
					padding: 0 var(--spacing-sm);
				"
			>
				<XCircle size={18} aria-hidden="true" />
			</button>
		{:else}
			<!-- Destructive confirm pattern (inline, no modal — UI-SPEC) -->
			<div style="display: flex; align-items: center; gap: var(--spacing-xs);">
				<span style="font-size: var(--font-size-label); color: var(--color-destructive);">
					Remove pane?
				</span>
				<button
					onclick={cancelRemove}
					style="
						min-height: var(--min-height-touch);
						padding: 0 var(--spacing-sm);
						border: 1px solid var(--color-border-subtle);
						border-radius: 4px;
						background: transparent;
						font-size: var(--font-size-label);
						cursor: pointer;
						color: var(--color-text-secondary);
					"
				>
					Cancel
				</button>
				<button
					onclick={handleRemove}
					class="conflict-action-btn"
					style="
						min-height: var(--min-height-touch);
						padding: 0 var(--spacing-sm);
						border: 1px solid var(--color-destructive);
						border-radius: 4px;
						background: var(--color-destructive);
						color: #ffffff;
						font-size: var(--font-size-label);
						cursor: pointer;
					"
				>
					Remove
				</button>
			</div>
		{/if}
	</div>

	<!-- Opinion list -->
	<div style="padding: var(--spacing-sm) var(--spacing-md); display: flex; flex-direction: column; gap: var(--spacing-xs);">
		{#each pane.opinions as opinion}
			<div
				style="
					padding: var(--spacing-xs) var(--spacing-sm);
					border-radius: 4px;
					background-color: var(--color-surface-base);
					border: 1px solid var(--color-border-subtle);
					display: flex;
					align-items: center;
					gap: var(--spacing-sm);
				"
			>
				<span style="font-size: var(--font-size-label); color: var(--color-text-primary); flex: 1;">
					{opinion.name}
				</span>
				<span
					style="
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
						text-transform: capitalize;
					"
				>
					{opinion.status}
				</span>
			</div>
		{/each}
	</div>

	<!-- Conflict overlays for this pane's opinions -->
	{#if paneConflicts.length > 0}
		<div style="padding: 0 var(--spacing-md) var(--spacing-md);">
			{#each paneConflicts as cv}
				<ConflictOverlay
					view={cv}
					opinionsLabel={cv.opinionsInvolved.join(' vs ')}
					{onDrop}
					{onApplyPatch}
				/>
			{/each}
		</div>
	{/if}
</div>
