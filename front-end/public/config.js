// Overridden at runtime by docker-entrypoint.sh in production.
// In dev, src/config.ts falls back to import.meta.env.VITE_* values.
window.__CONFIG__ = window.__CONFIG__ || {};
