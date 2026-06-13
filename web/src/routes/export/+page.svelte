<script lang="ts">
	/**
	 * Export / Speech Export screen.
	 *
	 * UI-SPEC §Screens: "Resolved speech YAML preview + build command + download".
	 * UI-SPEC §Build-stage naming (BRND-01 rhetoric names).
	 * Copy verbatim from UI-SPEC §Copywriting contract.
	 *
	 * IN-04 fix: wired to the actual resolved speech from the debate store.
	 * The resolvedSpeechStore is populated by the debate page after every WASM
	 * resolve cycle. The export page reads it and renders/downloads the real
	 * resolved speech YAML instead of a hardcoded example.
	 *
	 * A6: no forbidden terms (config/preset/distro/package set) in visible text.
	 */

	import { base } from '$app/paths';
	import { resolvedSpeechStore } from '$lib/stores/speech.js';
	import type { ResolvedSpeech } from '$lib/types.js';

	// Build-stage display names per UI-SPEC §Build-stage naming
	const buildStages = [
		{ key: 'resolve',   label: 'Settling the Debate' },
		{ key: 'translate', label: "Finding Your Foundation's Voice" },
		{ key: 'build',     label: 'Writing the Final Argument' },
		{ key: 'package',   label: 'Sealing the Speech' }
	];

	// Reactive: use the real resolved speech if available, else null.
	$: resolved = $resolvedSpeechStore;

	// Generate YAML from the actual resolved speech (IN-04).
	// This is a lightweight serialisation — the full YAML is produced by `debateos build`
	// from the downloaded file; we reproduce the key fields here for preview + download.
	function resolvedSpeechToYaml(rs: ResolvedSpeech): string {
		const lines: string[] = [`schema: ${rs.schema}`, `foundation: ${rs.foundation}`];

		if (rs.applied && rs.applied.length > 0) {
			lines.push('applied:');
			for (const id of rs.applied) {
				lines.push(`  - ${id}`);
			}
		} else {
			lines.push('applied: []');
		}

		if (rs.dropped && rs.dropped.length > 0) {
			lines.push('dropped:');
			for (const id of rs.dropped) {
				lines.push(`  - ${id}`);
			}
		} else {
			lines.push('dropped: []');
		}

		if (rs.explanations && rs.explanations.length > 0) {
			lines.push('explanations:');
			for (const exp of rs.explanations) {
				lines.push(`  - rule: ${exp.rule}`);
				const escaped = exp.text.replace(/"/g, '\\"');
				lines.push(`    text: "${escaped}"`);
			}
		}

		return lines.join('\n') + '\n';
	}

	$: yamlContent = resolved ? resolvedSpeechToYaml(resolved) : null;
	$: buildCommand = resolved
		? `debateos build --speech resolved-speech.yaml`
		: null;
	$: hasResolved = resolved !== null && (resolved.applied?.length ?? 0) > 0;

	function downloadYaml() {
		if (!yamlContent) return;
		const blob = new Blob([yamlContent], { type: 'text/yaml' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = 'resolved-speech.yaml';
		a.click();
		URL.revokeObjectURL(url);
	}
</script>

<svelte:head>
	<title>Export — DebateOS</title>
</svelte:head>

<div
	style="
		flex: 1;
		padding: var(--spacing-3xl) var(--spacing-lg);
		max-width: 800px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: var(--spacing-2xl);
	"
>
	{#if !hasResolved}
		<!-- No resolved speech yet — honest placeholder state (IN-04) -->
		<div style="display: flex; flex-direction: column; gap: var(--spacing-sm);">
			<h1
				style="
					font-size: var(--font-size-display);
					font-weight: 600;
					color: var(--color-text-primary);
					margin: 0;
				"
			>
				No resolved speech yet.
			</h1>
			<p style="font-size: var(--font-size-body); color: var(--color-text-secondary); margin: 0;">
				Return to the Debate page, add your points and opinions, and resolve them before exporting.
			</p>
			<a
				href="{base}/debate/"
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
					border: none;
					border-radius: 6px;
					text-decoration: none;
					width: fit-content;
				"
			>
				Go to Debate
			</a>
		</div>
	{:else}
		<!-- Speech is ready — show real content (IN-04) -->

		<!-- Heading — UI-SPEC copy verbatim -->
		<div style="display: flex; flex-direction: column; gap: var(--spacing-sm);">
			<h1
				style="
					font-size: var(--font-size-display);
					font-weight: 600;
					color: var(--color-text-primary);
					margin: 0;
				"
			>
				Your speech is ready.
			</h1>
			<p style="font-size: var(--font-size-body); color: var(--color-text-secondary); margin: 0;">
				No conflicts remain. Run the command below to build your installer, or download the resolved
				speech YAML to build later.
			</p>
		</div>

		<!-- Build command -->
		<div style="display: flex; flex-direction: column; gap: var(--spacing-sm);">
			<h2
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-text-primary);
					margin: 0;
				"
			>
				Build command
			</h2>
			<pre
				style="
					font-family: var(--font-mono);
					font-size: var(--font-size-code);
					line-height: var(--line-height-code);
					background-color: var(--color-surface-card);
					border: 1px solid var(--color-border-subtle);
					border-radius: 6px;
					padding: var(--spacing-md);
					overflow-x: auto;
					margin: 0;
					color: var(--color-text-primary);
				"
			>{buildCommand}</pre>
		</div>

		<!-- Build stages (BRND-01 rhetoric names) -->
		<div style="display: flex; flex-direction: column; gap: var(--spacing-sm);">
			<h2
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-text-primary);
					margin: 0;
				"
			>
				Build stages
			</h2>
			<div style="display: flex; flex-direction: column; gap: var(--spacing-xs);">
				{#each buildStages as stage, i}
					<div
						style="
							display: flex;
							align-items: center;
							gap: var(--spacing-md);
							padding: var(--spacing-sm) var(--spacing-md);
							background-color: var(--color-surface-card);
							border: 1px solid var(--color-border-subtle);
							border-radius: 6px;
						"
					>
						<span
							style="
								width: 24px;
								height: 24px;
								border-radius: 999px;
								background-color: var(--color-accent-brand);
								color: #ffffff;
								font-size: var(--font-size-label);
								font-weight: 600;
								display: flex;
								align-items: center;
								justify-content: center;
								flex-shrink: 0;
							"
						>
							{i + 1}
						</span>
						<span style="font-size: var(--font-size-body); color: var(--color-text-primary);">
							{stage.label}
						</span>
					</div>
				{/each}
			</div>
		</div>

		<!-- Resolved speech YAML preview — real data (IN-04) -->
		<div style="display: flex; flex-direction: column; gap: var(--spacing-sm);">
			<h2
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-text-primary);
					margin: 0;
				"
			>
				Resolved speech
			</h2>
			<pre
				data-testid="resolved-yaml"
				style="
					font-family: var(--font-mono);
					font-size: var(--font-size-code);
					line-height: var(--line-height-code);
					background-color: var(--color-surface-card);
					border: 1px solid var(--color-border-subtle);
					border-radius: 6px;
					padding: var(--spacing-md);
					overflow-x: auto;
					margin: 0;
					color: var(--color-text-primary);
					max-height: 320px;
					overflow-y: auto;
				"
			>{yamlContent}</pre>
		</div>

		<!-- Actions -->
		<div style="display: flex; gap: var(--spacing-md); flex-wrap: wrap;">
			<!-- Primary CTA: download (accent-brand — A5) -->
			<button
				onclick={downloadYaml}
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
					border: none;
					border-radius: 6px;
					cursor: pointer;
				"
			>
				Download Resolved Speech
			</button>

			<!-- Secondary: back to debate -->
			<a
				href="{base}/debate/"
				style="
					display: inline-flex;
					align-items: center;
					justify-content: center;
					min-height: var(--min-height-touch);
					padding: 0 var(--spacing-xl);
					border: 1px solid var(--color-border-subtle);
					border-radius: 6px;
					color: var(--color-text-secondary);
					font-size: var(--font-size-body);
					text-decoration: none;
				"
			>
				Back to Debate
			</a>
		</div>
	{/if}
</div>
