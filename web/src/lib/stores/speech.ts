/**
 * speech.ts — Svelte store for debate composition state.
 *
 * Holds: foundation, panes (list of added points), opinions, hardware profile.
 * Exposes: addPane / removePane / updateHardware / resetDebate / setFoundation.
 *
 * No resolution logic — resolution is called by the page component via debateosResolve.
 * Invariant 3: this store never makes resolve decisions.
 */

import { writable, derived, get } from 'svelte/store';
import type { Opinion, HardwareProfile } from '$lib/types.js';

// ────────────────────────────────────────────────────────────────────────────
// Pane type (a point added to the debate)
// ────────────────────────────────────────────────────────────────────────────

export interface Pane {
	/** Unique pane ID (generated on add). */
	id: string;
	/** Point ID (matches speech point reference). */
	pointId: string;
	/** Display name for the point. */
	pointName: string;
	/** Curator handle (optional). */
	curator?: string;
	/** Opinions belonging to this pane. */
	opinions: Opinion[];
}

// ────────────────────────────────────────────────────────────────────────────
// Debate state
// ────────────────────────────────────────────────────────────────────────────

export interface DebateState {
	foundation: string;
	panes: Pane[];
	hardware: HardwareProfile;
}

function createDebateStore() {
	const initial: DebateState = {
		foundation: 'arch',
		panes: [],
		hardware: { predicates: [] }
	};

	const { subscribe, set, update } = writable<DebateState>(initial);

	let paneCounter = 0;

	return {
		subscribe,

		/** Set the debate foundation. */
		setFoundation(foundation: string) {
			update((s) => ({ ...s, foundation }));
		},

		/** Add a point pane to the debate. */
		addPane(pointId: string, pointName: string, opinions: Opinion[], curator?: string): string {
			const id = `pane-${++paneCounter}`;
			update((s) => ({
				...s,
				panes: [
					...s.panes,
					{ id, pointId, pointName, curator, opinions }
				]
			}));
			return id;
		},

		/** Remove a pane by ID. */
		removePane(paneId: string) {
			update((s) => ({
				...s,
				panes: s.panes.filter((p) => p.id !== paneId)
			}));
		},

		/** Reorder panes (for drag-and-drop). */
		reorderPanes(panes: Pane[]) {
			update((s) => ({ ...s, panes }));
		},

		/** Update the hardware profile. */
		updateHardware(hardware: HardwareProfile) {
			update((s) => ({ ...s, hardware }));
		},

		/** Reset the debate to empty state. */
		resetDebate() {
			paneCounter = 0;
			set(initial);
		},

		/** Get current value snapshot. */
		snapshot(): DebateState {
			return get({ subscribe });
		}
	};
}

export const debate = createDebateStore();

/** Derived: flat list of all opinions across all panes. */
export const allOpinions = derived(debate, ($debate) =>
	$debate.panes.flatMap((p) => p.opinions)
);

/** Derived: count of panes. */
export const paneCount = derived(debate, ($debate) => $debate.panes.length);
