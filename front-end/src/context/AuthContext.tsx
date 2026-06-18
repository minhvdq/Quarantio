import React, { createContext, useContext, useState, useCallback } from 'react';
import { TENANT_URL } from '../config';

const K_ACCESS = 'gm_access';
const K_REFRESH = 'gm_refresh';
const K_TENANT = 'gm_tenant';
const K_ROLE = 'gm_role';
const K_EMAIL = 'gm_email';

interface AuthData {
  access_token: string;
  refresh_token: string;
  tenant_id?: string;
  role?: string;
  user?: { email: string; email_verified?: boolean };
}

interface AuthContextValue {
  accessToken: string;
  refreshToken: string;
  tenantId: string;
  role: string;
  email: string;
  login: (data: AuthData) => void;
  logout: () => void;
  tryRefresh: () => Promise<boolean>;
  storeAuth: (data: AuthData) => void;
  setAccessToken: (token: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [accessToken, setAccessTokenState] = useState(() => localStorage.getItem(K_ACCESS) || '');
  const [refreshToken, setRefreshToken] = useState(() => localStorage.getItem(K_REFRESH) || '');
  const [tenantId, setTenantId] = useState(() => localStorage.getItem(K_TENANT) || '');
  const [role, setRole] = useState(() => localStorage.getItem(K_ROLE) || '');
  const [email, setEmail] = useState(() => localStorage.getItem(K_EMAIL) || '');

  const setAccessToken = useCallback((token: string) => {
    setAccessTokenState(token);
    localStorage.setItem(K_ACCESS, token);
  }, []);

  const storeAuth = useCallback((data: AuthData) => {
    const at = data.access_token;
    const rt = data.refresh_token;
    const tid = data.tenant_id || '';
    const r = data.role || '';
    const em = data.user?.email || '';
    setAccessTokenState(at);
    setRefreshToken(rt);
    setTenantId(tid);
    setRole(r);
    setEmail(em);
    localStorage.setItem(K_ACCESS, at);
    localStorage.setItem(K_REFRESH, rt);
    localStorage.setItem(K_TENANT, tid);
    localStorage.setItem(K_ROLE, r);
    localStorage.setItem(K_EMAIL, em);
    localStorage.setItem('gm_verified', data.user?.email_verified ? '1' : '0');
  }, []);

  const login = storeAuth;

  const tryRefresh = useCallback(async (): Promise<boolean> => {
    const rt = localStorage.getItem(K_REFRESH) || refreshToken;
    if (!rt) return false;
    try {
      const res = await fetch(`${TENANT_URL}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: rt }),
      });
      if (!res.ok) return false;
      const d = await res.json();
      setAccessToken(d.access_token);
      return true;
    } catch {
      return false;
    }
  }, [refreshToken, setAccessToken]);

  const logout = useCallback(() => {
    const rt = localStorage.getItem(K_REFRESH) || refreshToken;
    if (rt) {
      fetch(`${TENANT_URL}/auth/logout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: rt }),
      }).catch(() => {});
    }
    [K_ACCESS, K_REFRESH, K_TENANT, K_ROLE, K_EMAIL].forEach((k) =>
      localStorage.removeItem(k)
    );
    setAccessTokenState('');
    setRefreshToken('');
    setTenantId('');
    setRole('');
    setEmail('');
  }, [refreshToken]);

  return (
    <AuthContext.Provider
      value={{ accessToken, refreshToken, tenantId, role, email, login, logout, tryRefresh, storeAuth, setAccessToken }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
