// Prerender the entire SvelteKit app as static output (adapter-static requirement).
// trailingSlash ensures consistent URLs for GitHub Pages and go:embed serving.
// Source: https://svelte.dev/docs/kit/adapter-static
export const prerender = true;
export const trailingSlash = 'always';
