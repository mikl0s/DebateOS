<script lang="ts">
	/**
	 * DebateStage — the main debate compose view.
	 *
	 * UI-SPEC §Core Debate Components:
	 *   DebateStage
	 *   ├── FoundationBar     (56px strip)
	 *   ├── PaneStack         (scrollable vertical stack)
	 *   │   └── PaneCard[]    (per-point panes with conflict overlays)
	 *   └── ConflictOverlay / ResolutionPanel (right sidebar via slot/parent)
	 *
	 * On every pane add/remove, the parent (+page.svelte) calls debateosResolve
	 * with 150ms debounce. This component just receives the resolved conflict views.
	 *
	 * Copy verbatim from UI-SPEC §Copywriting contract (empty state).
	 */

	import type { Pane } from '$lib/stores/speech.js';
	import type { ConflictView } from '$lib/conflict.js';
	import FoundationBar from './FoundationBar.svelte';
	import PaneCard from './PaneCard.svelte';
	import { base } from '$app/paths';

	interface Props {
		foundation: string;
		panes: Pane[];
		conflictViews: ConflictView[];
		onRemovePane?: (paneId: string) => void;
		onSwapFoundation?: () => void;
		onDropOpinion?: (opinionId: string) => void;
		onApplyPatch?: (patchId: string) => void;
		activePaneId?: string;
	}

	let {
		foundation,
		panes,
		conflictViews = [],
		onRemovePane,
		onSwapFoundation,
		onDropOpinion,
		onApplyPatch,
		activePaneId
	}: Props = $props();
</script>

<div
	style="
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow: hidden;
	"
>
	<!-- Foundation strip: fixed 56px -->
	<FoundationBar {foundation} onSwap={onSwapFoundation} />

	<!-- Pane stack: scrollable -->
	<div
		style="
			flex: 1;
			overflow-y: auto;
			padding: var(--spacing-md);
			display: flex;
			flex-direction: column;
			gap: var(--spacing-md);
		"
	>
		{#if panes.length === 0}
			<!-- Empty state — UI-SPEC §Copywriting contract verbatim -->
			<div
				style="
					display: flex;
					flex-direction: column;
					align-items: center;
					justify-content: center;
					text-align: center;
					padding: var(--spacing-3xl) var(--spacing-lg);
					gap: var(--spacing-md);
				"
			>
				<h2
					style="
						font-size: var(--font-size-display);
						font-weight: 600;
						color: var(--color-text-primary);
						margin: 0;
					"
				>
					Your debate has no points yet.
				</h2>
				<p
					style="
						font-size: var(--font-size-body);
						color: var(--color-text-secondary);
						max-width: 480px;
						margin: 0;
						line-height: var(--line-height-body);
					"
				>
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
		{:else}
			{#each panes as pane (pane.id)}
				<PaneCard
					{pane}
					{conflictViews}
					onRemove={onRemovePane}
					onDrop={onDropOpinion}
					{onApplyPatch}
					isActive={pane.id === activePaneId}
				/>
			{/each}
		{/if}
	</div>
</div>
