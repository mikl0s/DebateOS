<script lang="ts">
	// Landing page — empty debate state
	// BRND-01: uses canonical DebateOS vocabulary only.
	// Forbidden: config, preset, package set, distro (A6).
	// Verbatim copy from UI-SPEC §Copywriting contract.
</script>

<svelte:head>
	<title>DebateOS — That's just your opinion, man.</title>
	<meta
		name="description"
		content="Compose your system through opinions, points, and speeches. No conclusions — only debates."
	/>
</svelte:head>

<section
	aria-label="DebateOS landing"
	style="
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-3xl) var(--spacing-lg);
		text-align: center;
		gap: var(--spacing-lg);
	"
>
	<!-- Display tagline — UI-SPEC §Copywriting contract, used in landing/empty state only -->
	<h1
		style="
			font-size: var(--font-size-display);
			font-weight: 600;
			line-height: var(--line-height-display);
			color: var(--color-text-primary);
			margin: 0;
		"
	>
		That's just your opinion, man.
	</h1>

	<!-- Sub-tagline — UI-SPEC §Copywriting contract -->
	<p
		style="
			font-size: var(--font-size-heading);
			font-weight: 600;
			line-height: var(--line-height-heading);
			color: var(--color-text-secondary);
			margin: 0;
		"
	>
		No conclusions — only debates.
	</p>

	<!-- Description -->
	<p
		style="
			font-size: var(--font-size-body);
			line-height: var(--line-height-body);
			color: var(--color-text-secondary);
			max-width: 560px;
			margin: 0;
		"
	>
		Compose your system from curators' points. Resolve conflicts through structured debate.
		Build a bootable unattended installer — zero cost, no central service required.
	</p>

	<!-- Primary CTA — UI-SPEC "Start Debating" -->
	<div style="display: flex; gap: var(--spacing-md); flex-wrap: wrap; justify-content: center;">
		<a
			href="/debate/"
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
				border: none;
				cursor: pointer;
			"
		>
			Start Debating
		</a>

		<a
			href="/browse/"
			style="
				display: inline-flex;
				align-items: center;
				justify-content: center;
				min-height: var(--min-height-touch);
				padding: 0 var(--spacing-xl);
				background-color: transparent;
				color: var(--color-text-primary);
				font-size: var(--font-size-body);
				font-weight: 600;
				text-decoration: none;
				border-radius: 6px;
				border: 1px solid var(--color-border-subtle);
				cursor: pointer;
			"
		>
			Browse Points
		</a>
	</div>

	<!-- How it works — using canonical vocabulary only -->
	<div
		style="
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: var(--spacing-lg);
			max-width: 800px;
			width: 100%;
			margin-top: var(--spacing-2xl);
		"
	>
		<div
			style="
				background-color: var(--color-surface-card);
				border: 1px solid var(--color-border-subtle);
				border-radius: 8px;
				padding: var(--spacing-lg);
				text-align: left;
			"
		>
			<h2
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					line-height: var(--line-height-heading);
					color: var(--color-text-primary);
					margin: 0 0 var(--spacing-sm) 0;
				"
			>
				Opinions
			</h2>
			<p
				style="
					font-size: var(--font-size-label);
					line-height: var(--line-height-label);
					color: var(--color-text-secondary);
					margin: 0;
				"
			>
				Atomic, foundation-agnostic configuration decisions. Each Opinion knows what it wants —
				not how to get it.
			</p>
		</div>

		<div
			style="
				background-color: var(--color-surface-card);
				border: 1px solid var(--color-border-subtle);
				border-radius: 8px;
				padding: var(--spacing-lg);
				text-align: left;
			"
		>
			<h2
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					line-height: var(--line-height-heading);
					color: var(--color-text-primary);
					margin: 0 0 var(--spacing-sm) 0;
				"
			>
				Points
			</h2>
			<p
				style="
					font-size: var(--font-size-label);
					line-height: var(--line-height-label);
					color: var(--color-text-secondary);
					margin: 0;
				"
			>
				Curated bundles of Opinions from trusted curators. Browse Points from the registry and add
				them to your Debate.
			</p>
		</div>

		<div
			style="
				background-color: var(--color-surface-card);
				border: 1px solid var(--color-border-subtle);
				border-radius: 8px;
				padding: var(--spacing-lg);
				text-align: left;
			"
		>
			<h2
				style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					line-height: var(--line-height-heading);
					color: var(--color-text-primary);
					margin: 0 0 var(--spacing-sm) 0;
				"
			>
				Speech
			</h2>
			<p
				style="
					font-size: var(--font-size-label);
					line-height: var(--line-height-label);
					color: var(--color-text-secondary);
					margin: 0;
				"
			>
				Your resolved composition. The WASM resolver settles all conflicts and produces a
				Foundation-specific build Speech.
			</p>
		</div>
	</div>
</section>
