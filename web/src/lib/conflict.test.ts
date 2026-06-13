/**
 * Tests for conflict.ts — pure mapping from Explanation to ConflictView.
 *
 * TDD RED: these tests are written BEFORE conflict.ts exists.
 * Assertions:
 *   A9: text field carries Explanation.text verbatim (TestExplanationVerbatim)
 *   A1: triple encoding — state/icon/label mapped per UI-SPEC (TestMapExplanation)
 */

import { describe, it, expect } from 'vitest';
import { mapExplanation } from './conflict.js';
import type { Explanation } from './types.js';

function makeExp(overrides: Partial<Explanation>): Explanation {
	return {
		text: 'Test explanation text',
		rule: 'no-conflict',
		...overrides
	};
}

describe('TestMapExplanation', () => {
	it('maps rule2 → hard conflict', () => {
		const view = mapExplanation(makeExp({ rule: 'rule2' }));
		expect(view.state).toBe('hard');
		expect(view.icon).toBe('AlertTriangle');
		expect(view.label).toBe('Hard conflict');
	});

	it('maps rule1 → warn (will be dropped)', () => {
		const view = mapExplanation(makeExp({ rule: 'rule1' }));
		expect(view.state).toBe('warn');
		expect(view.icon).toBe('Info');
		expect(view.label).toBe('Will be dropped');
	});

	it('maps hardware-skip → hardware mismatch', () => {
		const view = mapExplanation(makeExp({ rule: 'hardware-skip' }));
		expect(view.state).toBe('hardware');
		expect(view.icon).toBe('Cpu');
		expect(view.label).toBe('Hardware mismatch');
	});

	it('maps no-conflict → compatible', () => {
		const view = mapExplanation(makeExp({ rule: 'no-conflict' }));
		expect(view.state).toBe('compat');
		expect(view.icon).toBe('CheckCircle2');
		expect(view.label).toBe('Compatible');
	});

	it('maps rule3 → warn', () => {
		const view = mapExplanation(makeExp({ rule: 'rule3' }));
		expect(view.state).toBe('warn');
		expect(view.icon).toBe('Info');
		expect(view.label).toBe('Will be dropped');
	});

	it('maps rule4 → info (patch applied)', () => {
		const view = mapExplanation(makeExp({ rule: 'rule4' }));
		expect(view.state).toBe('patch');
		expect(view.icon).toBe('Puzzle');
		expect(view.label).toBe('Patch applied');
	});

	it('maps ordering → info', () => {
		const view = mapExplanation(makeExp({ rule: 'ordering' }));
		expect(view.state).toBe('info');
		expect(view.icon).toBe('Info');
		expect(view.label).toBe('Ordering applied');
	});

	it('maps cycle → hard (cycle error)', () => {
		const view = mapExplanation(makeExp({ rule: 'cycle' }));
		expect(view.state).toBe('hard');
		expect(view.icon).toBe('AlertTriangle');
		expect(view.label).toBe('Circular dependency');
	});

	it('maps sysctl-collision → warn', () => {
		const view = mapExplanation(makeExp({ rule: 'sysctl-collision' }));
		expect(view.state).toBe('warn');
		expect(view.icon).toBe('Info');
		expect(view.label).toBe('Sysctl collision');
	});

	it('maps hardware-apply → info', () => {
		const view = mapExplanation(makeExp({ rule: 'hardware-apply' }));
		expect(view.state).toBe('info');
		expect(view.icon).toBe('Cpu');
		expect(view.label).toBe('Hardware applied');
	});

	it('maps unknown rule → info fallback', () => {
		const view = mapExplanation(makeExp({ rule: 'unknown-rule' }));
		expect(view.state).toBe('info');
		expect(view.icon).toBe('Info');
		expect(view.label).toBe('unknown-rule');
	});
});

describe('TestPatchBadge', () => {
	it('adds patch badge when patch_offered is non-empty', () => {
		const view = mapExplanation(
			makeExp({ rule: 'rule2', patch_offered: 'OM-015-patch' })
		);
		// Base state is still hard (rule2)
		expect(view.state).toBe('hard');
		// Patch flag is set
		expect(view.patchOffered).toBe('OM-015-patch');
		expect(view.hasPatch).toBe(true);
	});

	it('no patch badge when patch_offered is empty', () => {
		const view = mapExplanation(makeExp({ rule: 'rule2', patch_offered: '' }));
		expect(view.hasPatch).toBe(false);
		expect(view.patchOffered).toBe('');
	});

	it('no patch badge when patch_offered is undefined', () => {
		const view = mapExplanation(makeExp({ rule: 'rule2' }));
		expect(view.hasPatch).toBe(false);
	});
});

describe('TestExplanationVerbatim', () => {
	it('carries text verbatim from Explanation (A9 guard)', () => {
		const verbatim = 'Required opinion OM-015 conflicts with OM-015-greetd: both declare conflicting display managers.';
		const view = mapExplanation(makeExp({ rule: 'rule2', text: verbatim }));
		// text must be IDENTICAL — no substitution, no reformatting
		expect(view.text).toBe(verbatim);
	});

	it('carries text verbatim even for empty string', () => {
		const view = mapExplanation(makeExp({ rule: 'no-conflict', text: '' }));
		expect(view.text).toBe('');
	});

	it('carries text verbatim for unicode content', () => {
		const unicode = 'Opinion ＯＭ－０１５ conflicts. Special chars: <>&"\'';
		const view = mapExplanation(makeExp({ text: unicode }));
		expect(view.text).toBe(unicode);
	});

	it('carries opinions_involved verbatim', () => {
		const view = mapExplanation(
			makeExp({ rule: 'rule2', opinions_involved: ['OM-015', 'OM-015-greetd'] })
		);
		expect(view.opinionsInvolved).toEqual(['OM-015', 'OM-015-greetd']);
	});

	it('carries kept and dropped verbatim', () => {
		const view = mapExplanation(
			makeExp({ rule: 'rule2', kept: ['OM-015'], dropped: ['OM-015-greetd'] })
		);
		expect(view.kept).toEqual(['OM-015']);
		expect(view.dropped).toEqual(['OM-015-greetd']);
	});

	it('carries trust_warning verbatim', () => {
		const warning = 'This point includes a repository with signature verification disabled.';
		const view = mapExplanation(makeExp({ trust_warning: warning }));
		expect(view.trustWarning).toBe(warning);
	});

	it('carries alternative_suggestion verbatim', () => {
		const suggestion = 'Consider greetd (OM-015-greetd) instead.';
		const view = mapExplanation(makeExp({ alternative_suggestion: suggestion }));
		expect(view.alternativeSuggestion).toBe(suggestion);
	});
});
