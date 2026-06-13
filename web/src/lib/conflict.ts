/**
 * conflict.ts — pure mapping from resolver Explanation to ConflictView.
 *
 * Invariant 3: this module NEVER reimplements resolution logic.
 * It only maps resolver output (Explanation) to view state (ConflictView).
 *
 * T-05-10 mitigation: all resolution via WASM; this module only maps output.
 * A9 guard: text field carries Explanation.text VERBATIM — never substituted.
 */

import type { Explanation } from './types.js';

/**
 * ConflictState — the UI-SPEC state for a conflict overlay.
 *
 * Maps to triple-encoding (color + icon + label) per UI-SPEC §Conflict Visualization Spec.
 */
export type ConflictState =
	| 'hard'     // rule2/cycle — red (conflict-hard), AlertTriangle
	| 'warn'     // rule1/rule3/sysctl-collision — amber (conflict-warn), Info
	| 'hardware' // hardware-skip — amber, Cpu
	| 'compat'   // no-conflict — green (conflict-compat), CheckCircle2
	| 'patch'    // rule4 — accent-brand, Puzzle
	| 'info';    // ordering/hardware-apply/fallback — no overlay, Info

/**
 * ConflictView — the complete view state for one Explanation.
 *
 * Passed to ConflictOverlay / ExplanationCard components.
 * A9 contract: text is Explanation.text VERBATIM.
 */
export interface ConflictView {
	/** Conflict state for color/icon triple encoding (A1). */
	state: ConflictState;

	/** Lucide icon name for shape cue (A1 — second encoding). */
	icon: string;

	/** Visible text label for text cue (A1 — third encoding). */
	label: string;

	/** Verbatim Explanation.text — A9 contract: never substituted. */
	text: string;

	/** Opinion IDs involved in this decision (verbatim from resolver). */
	opinionsInvolved: string[];

	/** Opinion IDs retained (verbatim from resolver). */
	kept: string[];

	/** Opinion IDs dropped (verbatim from resolver). */
	dropped: string[];

	/** Patch opinion ID offered (empty string when none). */
	patchOffered: string;

	/** True when patch_offered is non-empty — drives patch badge rendering. */
	hasPatch: boolean;

	/** Trust warning text (verbatim from resolver; empty when none). */
	trustWarning: string;

	/** Alternative suggestion text (verbatim from resolver; empty when none). */
	alternativeSuggestion: string;
}

/**
 * mapExplanation — maps one resolver Explanation to a ConflictView.
 *
 * This is the ONLY mapping function in the UI.
 * A9: text is copied verbatim — no substitution, no reformatting.
 *
 * @param e - Explanation from ResolvedSpeech.Explanations[]
 * @returns ConflictView with triple-encoded state for rendering
 */
export function mapExplanation(e: Explanation): ConflictView {
	const { state, icon, label } = mapRule(e.rule);

	return {
		state,
		icon,
		label,
		text: e.text, // A9: verbatim — NEVER substitute with UI prose
		opinionsInvolved: e.opinions_involved ?? [],
		kept: e.kept ?? [],
		dropped: e.dropped ?? [],
		patchOffered: e.patch_offered ?? '',
		hasPatch: Boolean(e.patch_offered && e.patch_offered.length > 0),
		trustWarning: e.trust_warning ?? '',
		alternativeSuggestion: e.alternative_suggestion ?? ''
	};
}

/**
 * mapRule — maps Explanation.rule to {state, icon, label} triple.
 *
 * Per UI-SPEC §Conflict Visualization Spec table.
 */
function mapRule(rule: string): { state: ConflictState; icon: string; label: string } {
	switch (rule) {
		case 'rule2':
			return { state: 'hard', icon: 'AlertTriangle', label: 'Hard conflict' };

		case 'cycle':
			return { state: 'hard', icon: 'AlertTriangle', label: 'Circular dependency' };

		case 'rule1':
		case 'rule3':
			return { state: 'warn', icon: 'Info', label: 'Will be dropped' };

		case 'sysctl-collision':
			return { state: 'warn', icon: 'Info', label: 'Sysctl collision' };

		case 'hardware-skip':
			return { state: 'hardware', icon: 'Cpu', label: 'Hardware mismatch' };

		case 'hardware-apply':
			return { state: 'info', icon: 'Cpu', label: 'Hardware applied' };

		case 'no-conflict':
			return { state: 'compat', icon: 'CheckCircle2', label: 'Compatible' };

		case 'rule4':
			return { state: 'patch', icon: 'Puzzle', label: 'Patch applied' };

		case 'ordering':
			return { state: 'info', icon: 'Info', label: 'Ordering applied' };

		default:
			// Fallback: pass the rule name as label so it's visible in debug
			return { state: 'info', icon: 'Info', label: rule };
	}
}
