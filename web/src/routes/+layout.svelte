<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<!-- AppShell: TopNav + page content -->
<div class="app-shell" style="min-height: 100vh; display: flex; flex-direction: column;">
	<nav
		aria-label="DebateOS navigation"
		style="
			background-color: var(--color-surface-card);
			border-bottom: 1px solid var(--color-border-subtle);
			height: var(--height-foundation-bar);
			display: flex;
			align-items: center;
			padding: 0 var(--spacing-lg);
			gap: var(--spacing-lg);
		"
	>
		<!-- DebateOS wordmark -->
		<a
			href="/"
			style="
				font-size: var(--font-size-heading);
				font-weight: 600;
				line-height: var(--line-height-heading);
				color: var(--color-text-primary);
				text-decoration: none;
				letter-spacing: -0.02em;
			"
		>
			DebateOS
		</a>

		<!-- Nav links -->
		<div style="display: flex; gap: var(--spacing-md); margin-left: auto; align-items: center;">
			<a
				href="/debate/"
				style="
					font-size: var(--font-size-label);
					color: var(--color-text-secondary);
					text-decoration: none;
					min-height: var(--min-height-touch);
					display: flex;
					align-items: center;
					padding: 0 var(--spacing-sm);
				"
			>
				Debate
			</a>
			<a
				href="/browse/"
				style="
					font-size: var(--font-size-label);
					color: var(--color-text-secondary);
					text-decoration: none;
					min-height: var(--min-height-touch);
					display: flex;
					align-items: center;
					padding: 0 var(--spacing-sm);
				"
			>
				Browse
			</a>
			<a
				href="https://github.com/mikl0s/debateos"
				target="_blank"
				rel="noopener noreferrer"
				aria-label="DebateOS on GitHub"
				style="
					font-size: var(--font-size-label);
					color: var(--color-text-secondary);
					text-decoration: none;
					min-height: var(--min-height-touch);
					display: flex;
					align-items: center;
					padding: 0 var(--spacing-sm);
				"
			>
				GitHub
			</a>
		</div>
	</nav>

	<!-- Page content -->
	<main style="flex: 1; display: flex; flex-direction: column;">
		{@render children()}
	</main>

	<!-- Footer -->
	<footer
		style="
			background-color: var(--color-surface-card);
			border-top: 1px solid var(--color-border-subtle);
			padding: var(--spacing-md) var(--spacing-lg);
			text-align: center;
			font-size: var(--font-size-label);
			color: var(--color-text-secondary);
		"
	>
		No conclusions — only debates.
	</footer>
</div>
