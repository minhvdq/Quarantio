declare global {
  interface Window {
    __CONFIG__: {
      BROKER_URL: string;
      TENANT_URL: string;
    };
  }
}

export const BROKER_URL =
  window.__CONFIG__?.BROKER_URL ??
  import.meta.env.VITE_BROKER_URL ??
  'http://localhost:8080';

export const TENANT_URL =
  window.__CONFIG__?.TENANT_URL ??
  import.meta.env.VITE_TENANT_URL ??
  'http://localhost:8082';
