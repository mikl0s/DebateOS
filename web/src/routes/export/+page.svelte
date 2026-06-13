<script lang="ts">
	/**
	 * Export / Speech Export screen.
	 *
	 * UI-SPEC §Screens: "Resolved speech YAML preview + build command + download".
	 * UI-SPEC §Build-stage naming (BRND-01 rhetoric names).
	 * Copy verbatim from UI-SPEC §Copywriting contract.
	 *
	 * A6: no forbidden terms (config/preset/distro/package set) in visible text.
	 */

	import { base } from '$app/paths';

	// Build-stage display names per UI-SPEC §Build-stage naming
	const buildStages = [
		{ key: 'resolve',   label: 'Settling the Debate' },
		{ key: 'translate', label: "Finding Your Foundation's Voice" },
		{ key: 'build',     label: 'Writing the Final Argument' },
		{ key: 'package',   label: 'Sealing the Speech' }
	];

	// Example resolved speech YAML (shown as preview — real data from debate store)
	// In v1, this page shows a template; full wiring is from the debate compose flow.
	const exampleCommand = `debateos build --speech resolved-speech.yaml`;

	const exampleYaml = `schema: 1
foundation: arch
applied:
  - OM-001
  - OM-006
  - OM-097
  - OM-099
dropped: []
explanations:
  - rule: no-conflict
    text: "All required opinions are compatible."
`;

	function downloadYaml() {
		const blob = new Blob([exampleYaml], { type: 'text/yaml' });
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
		>{exampleCommand}</pre>
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

	<!-- Resolved speech YAML preview -->
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
		>{exampleYaml}</pre>
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
</div>
