/**
 * wasm.test.ts — unit tests for the WASM loader pure helpers.
 *
 * These tests cover the pure, testable functions (no WASM loaded):
 *   - parseResolveOutput: parses raw JSON from window.debateosResolve
 *   - buildResolveInput: constructs typed ResolveInput with predicates guard
 *
 * Test plan (from plan Task 2 <behavior> block):
 *   TestParseResolveOutput: result present + error present → returns {resolved, error}
 *   TestParseResolveOutputErrorOnly: no result → throws Error containing error message
 *   TestBuildInput: predicates is always an array (Pitfall 6 guard)
 */

import { describe, it, expect } from 'vitest';
import { parseResolveOutput, buildResolveInput } from './wasm.js';
import type { ResolvedSpeech, Speech, Opinion, HardwareProfile } from './types.js';

// ─── TestParseResolveOutput ───────────────────────────────────────────────────

describe('parseResolveOutput', () => {
	it('returns {resolved, error} when result is present alongside an error (hard conflict path)', () => {
		// Canonical ResolvedSpeech — mirrors resolve.ResolvedSpeech JSON shape
		const resolvedSpeech: ResolvedSpeech = {
			schema: 1,
			foundation: 'arch',
			applied: ['OM-001'],
			dropped: ['OM-002'],
			explanations: [
				{
					text: 'OM-002 conflicts with required OM-001.',
					rule: 'rule2',
					opinions_involved: ['OM-001', 'OM-002'],
					dropped: ['OM-002'],
					kept: ['OM-001']
				}
			]
		};

		// Raw string as debateosResolve returns it:
		// outer JSON: { result: "<inner JSON string>", error: "hard conflict" }
		const raw = JSON.stringify({
			result: JSON.stringify(resolvedSpeech),
			error: 'hard conflict'
		});

		const out = parseResolveOutput(raw);

		expect(out.resolved).toBeDefined();
		expect(out.resolved.schema).toBe(1);
		expect(out.resolved.foundation).toBe('arch');
		expect(out.resolved.explanations).toHaveLength(1);
		expect(out.resolved.explanations[0].rule).toBe('rule2');
		expect(out.error).toBe('hard conflict');
	});

	it('returns {resolved} with no error on clean resolve', () => {
		const resolvedSpeech: ResolvedSpeech = {
			schema: 1,
			foundation: 'debian',
			applied: ['OM-001', 'OM-006'],
			explanations: []
		};

		const raw = JSON.stringify({ result: JSON.stringify(resolvedSpeech) });
		const out = parseResolveOutput(raw);

		expect(out.resolved.applied).toHaveLength(2);
		expect(out.error).toBeUndefined();
	});

	// TestParseResolveOutputErrorOnly
	it('throws Error containing the error message when result is absent (unrecoverable)', () => {
		const raw = JSON.stringify({ error: 'empty input' });

		expect(() => parseResolveOutput(raw)).toThrowError(/empty input/);
	});

	it('throws Error when no result and no error message', () => {
		const raw = JSON.stringify({});

		expect(() => parseResolveOutput(raw)).toThrowError(/resolver returned no result/);
	});

	it('propagates explanations array even on hard conflict', () => {
		const resolvedSpeech: ResolvedSpeech = {
			schema: 1,
			foundation: 'arch',
			applied: [],
			dropped: ['OM-010'],
			explanations: [
				{
					text: 'Hard conflict between OM-010 and OM-020.',
					rule: 'rule2',
					opinions_involved: ['OM-010', 'OM-020']
				}
			]
		};

		const raw = JSON.stringify({
			result: JSON.stringify(resolvedSpeech),
			error: 'hard conflict'
		});

		const out = parseResolveOutput(raw);
		expect(out.resolved.explanations).toHaveLength(1);
		expect(out.error).toBe('hard conflict');
	});
});

// ─── TestBuildInput ───────────────────────────────────────────────────────────

describe('buildResolveInput', () => {
	const mockSpeech: Speech = {
		schema: 1,
		foundation: 'arch',
		points: []
	};

	const mockOpinions: Opinion[] = [];

	it('predicates is an array when hardware.predicates is provided', () => {
		// Pitfall 6 guard: predicates must be string[] not object
		const hw: HardwareProfile = { predicates: ['hw-nvidia-gpu', 'hw-amd-cpu'] };
		const input = buildResolveInput(mockSpeech, mockOpinions, hw);

		expect(Array.isArray(input.hardware.predicates)).toBe(true);
		expect(input.hardware.predicates).toEqual(['hw-nvidia-gpu', 'hw-amd-cpu']);
	});

	it('predicates defaults to [] when not provided (Pitfall 6 guard)', () => {
		// Partial hardware with no predicates
		const hw: Partial<HardwareProfile> = { facts: { cpu: 'intel' } };
		const input = buildResolveInput(mockSpeech, mockOpinions, hw as HardwareProfile);

		expect(Array.isArray(input.hardware.predicates)).toBe(true);
		expect(input.hardware.predicates).toEqual([]);
	});

	it('predicates defaults to [] when explicitly undefined', () => {
		const hw = { predicates: undefined as unknown as string[] };
		const input = buildResolveInput(mockSpeech, mockOpinions, hw);

		expect(Array.isArray(input.hardware.predicates)).toBe(true);
		expect(input.hardware.predicates).toHaveLength(0);
	});

	it('passes speech and opinions through unchanged', () => {
		const input = buildResolveInput(mockSpeech, mockOpinions, { predicates: [] });

		expect(input.speech).toBe(mockSpeech);
		expect(input.opinions).toBe(mockOpinions);
	});

	it('forwards facts and pci_ids when provided', () => {
		const hw: HardwareProfile = {
			predicates: [],
			facts: { cpu: 'amd', ram: '16G' },
			pci_ids: ['10de:1234']
		};
		const input = buildResolveInput(mockSpeech, mockOpinions, hw);

		expect(input.hardware.facts).toEqual({ cpu: 'amd', ram: '16G' });
		expect(input.hardware.pci_ids).toEqual(['10de:1234']);
	});
});
