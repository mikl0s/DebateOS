// Debate/Compose page — CSR only (WASM cannot run in SSR/Node context).
// Pitfall 3 guard: export ssr = false prevents SvelteKit from pre-rendering
// this page server-side, where window.debateosResolve would be undefined.
export const ssr = false;
export const prerender = false;
