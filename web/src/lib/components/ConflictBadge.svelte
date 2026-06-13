<script lang="ts">
	/**
	 * ConflictBadge — triple-encoded conflict indicator (A1).
	 *
	 * Renders: color fill + lucide icon + visible text label.
	 * Per UI-SPEC §Conflict Visualization Spec state table.
	 *
	 * Uses @lucide/svelte for tree-shakeable icon imports.
	 */

	import type { ConflictState } from '$lib/conflict.js';
	import {
		AlertTriangle,
		Info,
		CheckCircle2,
		Cpu,
		Puzzle
	} from '@lucide/svelte';

	interface Props {
		state: ConflictState;
		icon: string;
		label: string;
	}

	let { state, icon, label }: Props = $props();

	// Map state to token color
	const colorMap: Record<ConflictState | string, string> = {
		hard: 'var(--color-conflict-hard)',
		warn: 'var(--color-conflict-warn)',
		hardware: 'var(--color-conflict-warn)',
		compat: 'var(--color-conflict-compat)',
		patch: 'var(--color-accent-brand)',
		info: 'var(--color-text-secondary)'
	};

	const color = $derived(colorMap[state] ?? colorMap.info);
</script>

<div
	style="
		display: inline-flex;
		align-items: center;
		gap: var(--spacing-xs);
		color: {color};
		font-size: var(--font-size-label);
		font-weight: 600;
	"
>
	<!-- Icon: shape cue (second encoding of triple A1) -->
	{#if icon === 'AlertTriangle'}
		<AlertTriangle size={16} aria-hidden="true" data-icon="AlertTriangle" />
	{:else if icon === 'CheckCircle2'}
		<CheckCircle2 size={16} aria-hidden="true" data-icon="CheckCircle2" />
	{:else if icon === 'Cpu'}
		<Cpu size={16} aria-hidden="true" data-icon="Cpu" />
	{:else if icon === 'Puzzle'}
		<Puzzle size={16} aria-hidden="true" data-icon="Puzzle" />
	{:else}
		<Info size={16} aria-hidden="true" data-icon="Info" />
	{/if}

	<!-- Text label: text cue (third encoding of triple A1) -->
	<span>{label}</span>
</div>
