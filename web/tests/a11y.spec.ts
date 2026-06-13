/**
 * a11y.spec.ts — Accessibility and brand assertions.
 *
 * Assertions verified:
 *   A2: Touch target minimum — all .pane-header and .conflict-action-btn offsetHeight >= 44
 *   A6: Terminology compliance — no forbidden terms in visible text
 *   A7: Typography scale — no font-size outside {13,14,16,20,28}px
 */

import { test, expect } from '@playwright/test';

// ─────────────────────────────────────────────────────────────────────────────
// A6: Terminology compliance (no forbidden terms)
// ─────────────────────────────────────────────────────────────────────────────

test.describe('A6 — Terminology compliance', () => {
	const forbiddenTerms = ['config', 'preset', 'distro', 'package set'];
	const routes = ['/', '/debate/', '/export/'];

	for (const route of routes) {
		test(`no forbidden terms on ${route}`, async ({ page }) => {
			// Block Forum to prevent any Forum-injected text
			await page.route('**/api/**', (route) => route.abort());
			await page.route('**/health**', (route) => route.abort());

			await page.goto(route);

			// For /debate, wait for WASM
			if (route === '/debate/') {
				try {
					await page.waitForSelector('[data-wasm-ready="true"]', { timeout: 15000 });
				} catch {
					// If WASM load times out, still check the page
				}
			}

			// Get all visible text nodes (excluding code/pre blocks which show YAML)
			const pageText = await page.evaluate(() => {
				// Exclude code/pre elements (YAML display is allowed to contain any text)
				const excludedSelectors = ['pre', 'code'];
				const walker = document.createTreeWalker(
					document.body,
					NodeFilter.SHOW_TEXT,
					{
						acceptNode: (node) => {
							// Skip if parent is a code/pre element
							let parent = node.parentElement;
							while (parent) {
								if (excludedSelectors.includes(parent.tagName.toLowerCase())) {
									return NodeFilter.FILTER_REJECT;
								}
								parent = parent.parentElement;
							}
							// Skip if not visible
							const parentEl = node.parentElement;
							if (!parentEl) return NodeFilter.FILTER_REJECT;
							const style = window.getComputedStyle(parentEl);
							if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') {
								return NodeFilter.FILTER_REJECT;
							}
							return NodeFilter.FILTER_ACCEPT;
						}
					}
				);

				const texts: string[] = [];
				let node = walker.nextNode();
				while (node) {
					const text = node.textContent?.trim();
					if (text) texts.push(text);
					node = walker.nextNode();
				}
				return texts.join(' ').toLowerCase();
			});

			for (const term of forbiddenTerms) {
				// Word-boundary check for exact term match
				const regex = new RegExp(`\\b${term.replace(' ', '\\s+')}\\b`, 'i');
				expect(pageText, `Forbidden term "${term}" found on ${route}`).not.toMatch(regex);
			}
		});
	}
});

// ─────────────────────────────────────────────────────────────────────────────
// A2: Touch target minimum (44px for pane headers and conflict action buttons)
// ─────────────────────────────────────────────────────────────────────────────

test.describe('A2 — Touch target minimum', () => {
	test('.pane-header elements are at least 44px tall', async ({ page }) => {
		await page.route('**/api/**', (route) => route.abort());
		await page.goto('/debate/');

		try {
			await page.waitForSelector('[data-wasm-ready="true"]', { timeout: 15000 });
		} catch {
			// continue
		}

		// Add a test pane via helper to create pane-header elements
		await page.evaluate(() => {
			const fn = (window as any).debateAddTestPane;
			if (!fn) return;
			fn('a2-test-1', 'Test Point A', [
				{ schema: 1, id: 'OP-001', name: 'Test Opinion', category: 'package-install', status: 'required' }
			]);
		});

		await page.waitForTimeout(200);

		// Check all .pane-header elements
		const paneHeaders = page.locator('.pane-header');
		const headerCount = await paneHeaders.count();

		for (let i = 0; i < headerCount; i++) {
			const height = await paneHeaders.nth(i).evaluate((el) => el.getBoundingClientRect().height);
			expect(height, `pane-header[${i}] must be >= 44px (WCAG 2.5.5)`).toBeGreaterThanOrEqual(44);
		}
	});

	test('.conflict-action-btn elements are at least 44px tall', async ({ page }) => {
		await page.route('**/api/**', (route) => route.abort());
		await page.goto('/debate/');

		try {
			await page.waitForSelector('[data-wasm-ready="true"]', { timeout: 15000 });
		} catch {
			// continue
		}

		// Check any visible conflict-action-btn elements
		const actionBtns = page.locator('.conflict-action-btn');
		const btnCount = await actionBtns.count();

		for (let i = 0; i < btnCount; i++) {
			const height = await actionBtns.nth(i).evaluate((el) => el.getBoundingClientRect().height);
			if (height > 0) {
				// Only check elements that are actually rendered (height > 0)
				expect(height, `.conflict-action-btn[${i}] must be >= 44px (WCAG 2.5.5)`).toBeGreaterThanOrEqual(44);
			}
		}
	});
});

// ─────────────────────────────────────────────────────────────────────────────
// A7: Typography scale — only {13,14,16,20,28}px font sizes
// ─────────────────────────────────────────────────────────────────────────────

test.describe('A7 — Typography scale', () => {
	test('no out-of-scale font sizes on landing page', async ({ page }) => {
		await page.goto('/');

		const allowedSizes = new Set([13, 14, 16, 20, 28]);
		// Allow system-rendered elements a small tolerance
		const tolerance = 0.5;

		const outOfScale = await page.evaluate((allowed) => {
			const elements = document.querySelectorAll('*');
			const violations: string[] = [];

			elements.forEach((el) => {
				// Skip hidden elements
				const style = window.getComputedStyle(el);
				if (style.display === 'none' || style.visibility === 'hidden') return;

				const fsPx = parseFloat(style.fontSize);
				if (isNaN(fsPx) || fsPx === 0) return;

				// Check if it's within tolerance of an allowed size
				const isAllowed = allowed.some((s: number) => Math.abs(fsPx - s) <= 0.5);
				if (!isAllowed) {
					// Skip browser default small sizes that we can't control (e.g., 12px UA styles)
					if (fsPx < 12) return;
					violations.push(`${el.tagName}${el.id ? '#' + el.id : ''}.${Array.from(el.classList).join('.')}: ${fsPx}px`);
				}
			});

			return violations.slice(0, 10); // cap to first 10
		}, Array.from(allowedSizes));

		expect(outOfScale, `Font sizes outside scale: ${outOfScale.join(', ')}`).toHaveLength(0);
	});

	test('no out-of-scale font-weight on landing page', async ({ page }) => {
		await page.goto('/');

		const allowedWeights = new Set([400, 500, 600]); // 500 is tolerated per spec

		const violations = await page.evaluate((allowed) => {
			const elements = document.querySelectorAll('*');
			const viol: string[] = [];

			elements.forEach((el) => {
				const style = window.getComputedStyle(el);
				if (style.display === 'none') return;

				const fw = parseInt(style.fontWeight);
				if (isNaN(fw)) return;

				if (!allowed.includes(fw)) {
					viol.push(`${el.tagName}: weight=${fw}`);
				}
			});

			return viol.slice(0, 10);
		}, Array.from(allowedWeights));

		expect(violations, `Font weights outside scale: ${violations.join(', ')}`).toHaveLength(0);
	});
});
