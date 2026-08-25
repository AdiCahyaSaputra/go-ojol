import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { refreshSession } from '@/lib/auth/refresh-session';
import { clearSession, getSession } from '@/lib/auth/token-storage';

const AUTH_SKIP_PATHS = ['/api/auth/login', '/api/auth/register', '/api/auth/refresh'];

export const axiosClient = axios.create({
  baseURL: process.env.EXPO_PUBLIC_API_URL,
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  },
});

function shouldSkipAuth(url?: string): boolean {
  if (!url) {
    return false;
  }
  return AUTH_SKIP_PATHS.some((path) => url.includes(path));
}

axiosClient.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  if (shouldSkipAuth(config.url)) {
    return config;
  }

  const session = await getSession();
  if (session?.accessToken) {
    config.headers.Authorization = `Bearer ${session.accessToken}`;
  }

  return config;
});

let refreshPromise: Promise<void> | null = null;

axiosClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    if (
      error.response?.status !== 401 ||
      !originalRequest ||
      originalRequest._retry ||
      shouldSkipAuth(originalRequest.url)
    ) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      if (!refreshPromise) {
        refreshPromise = refreshSession()
          .then(() => undefined)
          .finally(() => {
            refreshPromise = null;
          });
      }

      await refreshPromise;

      const session = await getSession();
      if (session?.accessToken) {
        originalRequest.headers.Authorization = `Bearer ${session.accessToken}`;
      }

      return axiosClient(originalRequest);
    } catch (refreshError) {
      await clearSession();
      return Promise.reject(refreshError);
    }
  },
);
