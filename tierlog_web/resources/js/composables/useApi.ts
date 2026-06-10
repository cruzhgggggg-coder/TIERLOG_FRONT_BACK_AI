import { useAuthStore } from '@/stores/auth';

export function useApi() {
  const auth = useAuthStore();

  async function get<T = unknown>(path: string, options?: RequestInit & { auth?: boolean }) {
    return auth.api<T>(path, { ...options, method: 'GET' });
  }

  async function post<T = unknown>(path: string, body?: unknown, options?: RequestInit & { auth?: boolean }) {
    return auth.api<T>(path, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async function put<T = unknown>(path: string, body?: unknown, options?: RequestInit & { auth?: boolean }) {
    return auth.api<T>(path, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async function patch<T = unknown>(path: string, body?: unknown, options?: RequestInit & { auth?: boolean }) {
    return auth.api<T>(path, {
      ...options,
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async function del<T = unknown>(path: string, options?: RequestInit & { auth?: boolean }) {
    return auth.api<T>(path, { ...options, method: 'DELETE' });
  }

  return { get, post, put, patch, del, raw: auth.api };
}
