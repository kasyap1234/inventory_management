import axios from 'axios';

import { csrfTokenManager, tokenStorage } from '@/lib/security';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/v1';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use(
  async (config) => {
    const headers = config.headers ?? {};

    const accessToken = tokenStorage.getAccessToken();
    if (accessToken) {
      headers.Authorization = `Bearer ${accessToken}`;
    }

    const method = config.method?.toLowerCase();
    if (method && !['get', 'head', 'options'].includes(method)) {
      const csrfToken = await csrfTokenManager.getToken();
      if (csrfToken) {
        headers['X-CSRF-Token'] = csrfToken;
      }
    }

    config.headers = headers;
    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const status = error.response?.status;
    const requestConfig = error.config as typeof error.config & { __isRetryRequest?: boolean };

    if (status === 401 && requestConfig && !requestConfig.__isRetryRequest) {
      const refreshToken = tokenStorage.getRefreshToken();
      if (!refreshToken) {
        tokenStorage.clear();
        csrfTokenManager.clearToken();
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }

      try {
        requestConfig.__isRetryRequest = true;
        const response = await api.post('/auth/refresh', { refresh_token: refreshToken });
        const { access_token, refresh_token: newRefreshToken } = response.data;
        tokenStorage.setTokens(access_token, newRefreshToken);
        csrfTokenManager.clearToken();

        requestConfig.headers = {
          ...(requestConfig.headers ?? {}),
          Authorization: `Bearer ${access_token}`,
        };

        return api(requestConfig);
      } catch (refreshError) {
        tokenStorage.clear();
        csrfTokenManager.clearToken();
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default api;
