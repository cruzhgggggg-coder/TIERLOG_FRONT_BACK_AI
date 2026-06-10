import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import type { User } from '@/types';

const STORAGE_KEY = 'tierlog.auth';
const API_URL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';

interface StoredAuth {
  user: User;
  accessToken: string;
  refreshToken: string;
}

function parseJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const base64 = token.split('.')[1];
    const json = atob(base64.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(json);
  } catch {
    return null;
  }
}

function isTokenExpired(token: string, bufferSeconds = 30): boolean {
  const payload = parseJwtPayload(token);
  if (!payload || typeof payload.exp !== 'number') return true;
  return Date.now() / 1000 >= payload.exp - bufferSeconds;
}

export function translateError(err: unknown): string {
  if (typeof err === 'string') return err;
  if (err && typeof err === 'object') {
    const obj = err as Record<string, unknown>;
    if (typeof obj.message === 'string') return obj.message;
    if (typeof obj.error === 'string') return obj.error;
    if (typeof obj.detail === 'string') return obj.detail;
    if (Array.isArray(obj.errors)) {
      return obj.errors
        .map((e: unknown) => {
          if (typeof e === 'string') return e;
          if (e && typeof e === 'object' && typeof (e as Record<string, unknown>).message === 'string') {
            return (e as Record<string, unknown>).message as string;
          }
          return String(e);
        })
        .join(', ');
    }
  }
  return 'An unexpected error occurred';
}

export const useAuthStore = defineStore('auth', () => {
  const booting = ref(true);
  const user = ref<User | null>(null);
  const accessToken = ref<string | null>(null);
  const refreshToken = ref<string | null>(null);

  const isAuthenticated = computed(() => !!user.value && !!accessToken.value);

  function persist() {
    if (user.value && accessToken.value && refreshToken.value) {
      const data: StoredAuth = {
        user: user.value,
        accessToken: accessToken.value,
        refreshToken: refreshToken.value,
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  }

  function loadFromStorage(): StoredAuth | null {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      const data = JSON.parse(raw) as StoredAuth;
      if (data.user && data.accessToken && data.refreshToken) return data;
      return null;
    } catch {
      return null;
    }
  }

  async function doRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false;
    try {
      const res = await fetch(`${API_URL}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken.value }),
      });
      if (!res.ok) return false;
      const data = await res.json();
      accessToken.value = data.access_token;
      if (data.refresh_token) refreshToken.value = data.refresh_token;
      persist();
      return true;
    } catch {
      return false;
    }
  }

  async function initialize() {
    const stored = loadFromStorage();
    if (stored) {
      user.value = stored.user;
      accessToken.value = stored.accessToken;
      refreshToken.value = stored.refreshToken;
      if (isTokenExpired(stored.accessToken)) {
        const refreshed = await doRefresh();
        if (!refreshed) {
          user.value = null;
          accessToken.value = null;
          refreshToken.value = null;
          localStorage.removeItem(STORAGE_KEY);
        }
      }
    }
    booting.value = false;
  }

  async function login(email: string, password: string): Promise<{ ok: boolean; error?: string }> {
    try {
      const res = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });
      const data = await res.json();
      if (!res.ok) return { ok: false, error: translateError(data) };
      user.value = data.user;
      accessToken.value = data.access_token;
      refreshToken.value = data.refresh_token;
      persist();
      return { ok: true };
    } catch (err) {
      return { ok: false, error: translateError(err) };
    }
  }

  async function register(payload: {
    name: string;
    email: string;
    password: string;
    role: string;
  }): Promise<{ ok: boolean; error?: string }> {
    try {
      const res = await fetch(`${API_URL}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const data = await res.json();
      if (!res.ok) return { ok: false, error: translateError(data) };
      user.value = data.user;
      accessToken.value = data.access_token;
      refreshToken.value = data.refresh_token;
      persist();
      return { ok: true };
    } catch (err) {
      return { ok: false, error: translateError(err) };
    }
  }

  async function logout() {
    try {
      if (accessToken.value) {
        await fetch(`${API_URL}/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken.value}`,
          },
        });
      }
    } catch {
      // ignore logout errors
    }
    user.value = null;
    accessToken.value = null;
    refreshToken.value = null;
    localStorage.removeItem(STORAGE_KEY);
  }

  async function api<T = unknown>(
    path: string,
    options: RequestInit & { auth?: boolean } = {},
  ): Promise<{ ok: boolean; data?: T; error?: string; status: number }> {
    const { auth = true, headers: customHeaders, ...rest } = options;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(customHeaders as Record<string, string>),
    };

    if (auth && accessToken.value) {
      if (isTokenExpired(accessToken.value)) {
        const refreshed = await doRefresh();
        if (!refreshed) {
          await logout();
          return { ok: false, error: 'Session expired', status: 401 };
        }
      }
      headers['Authorization'] = `Bearer ${accessToken.value}`;
    }

    try {
      let res = await fetch(`${API_URL}${path}`, { headers, ...rest });

      if (res.status === 401 && auth && refreshToken.value) {
        const refreshed = await doRefresh();
        if (refreshed && accessToken.value) {
          headers['Authorization'] = `Bearer ${accessToken.value}`;
          res = await fetch(`${API_URL}${path}`, { headers, ...rest });
        } else {
          await logout();
          return { ok: false, error: 'Session expired', status: 401 };
        }
      }

      const contentType = res.headers.get('content-type') || '';
      let data: unknown;

      if (contentType.includes('application/json')) {
        data = await res.json();
      } else {
        data = await res.text();
      }

      if (!res.ok) {
        return { ok: false, error: translateError(data), status: res.status, data: data as T };
      }

      return { ok: true, data: data as T, status: res.status };
    } catch (err) {
      return { ok: false, error: translateError(err), status: 0 };
    }
  }

  watch(user, () => {
    persist();
  }, { deep: true });

  return {
    booting,
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    initialize,
    login,
    register,
    logout,
    refresh: doRefresh,
    api,
  };
});
