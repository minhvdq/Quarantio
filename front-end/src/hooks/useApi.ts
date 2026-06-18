import { useCallback } from 'react';
import { useAuth } from '../context/AuthContext';

export function useApi() {
  const { accessToken, tryRefresh, logout, setAccessToken } = useAuth();

  const apiFetch = useCallback(
    async (url: string, opts: RequestInit = {}): Promise<Response> => {
      const token = localStorage.getItem('gm_access') || accessToken;
      opts = {
        ...opts,
        headers: {
          ...(opts.headers || {}),
          Authorization: 'Bearer ' + token,
        },
      };
      let res = await fetch(url, opts);
      if (res.status === 401) {
        const refreshed = await tryRefresh();
        if (refreshed) {
          const newToken = localStorage.getItem('gm_access') || '';
          setAccessToken(newToken);
          opts = {
            ...opts,
            headers: {
              ...(opts.headers || {}),
              Authorization: 'Bearer ' + newToken,
            },
          };
          res = await fetch(url, opts);
        } else {
          logout();
          return res;
        }
      }
      return res;
    },
    [accessToken, tryRefresh, logout, setAccessToken]
  );

  return { apiFetch };
}
