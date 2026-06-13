import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			// fallback: '404.html' enables the SPA fallback for CSR-only routes like /debate/
			// (prerender=false, ssr=false). The static server serves this file for any
			// unmatched path. Both GitHub Pages and go:embed net/http support this pattern.
			fallback: '404.html',
			precompress: false,
			strict: true
		}),
		paths: {
			// UI-02 dual-delivery seam:
			// Set BASE_PATH=/debateos for GitHub Pages deploy.
			// Leave empty (default) for localhost / cli/embed serve.
			base: process.env.BASE_PATH ?? ''
		},
		prerender: {
			// Ignore missing favicon (static asset; not a route; may have base-path prefix)
			handleHttpError: ({ path, message }) => {
				if (path.endsWith('/favicon.png')) return;
				throw new Error(message);
			}
		}
	}
};

export default config;
