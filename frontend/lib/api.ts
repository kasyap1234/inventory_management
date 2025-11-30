import axios from 'axios';

import { csrfTokenManager, tokenStorage } from '@/lib/security';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api/v1';

if (!process.env.NEXT_PUBLIC_API_URL && typeof window !== 'undefined') {
  console.warn('NEXT_PUBLIC_API_URL is not defined, falling back to relative path /api/v1');
}

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000, // 30 seconds
  validateStatus: (status) => status >= 200 && status < 500, // Don't throw on 4xx errors
  withCredentials: true, // Send cookies with requests for HttpOnly cookie authentication
});

api.interceptors.request.use(
  async (config) => {
    const headers = config.headers ?? {};

    // No need to manually add Authorization header - tokens are sent as HttpOnly cookies
    // The backend reads the auth_token cookie automatically

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
  (response) => {
    // Handle 4xx errors that didn't throw
    if (response.status >= 400 && response.status < 500) {
      const error = new Error(response.data?.message || 'Request failed');
      (error as any).response = response;
      throw error;
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Handle CSRF token errors - retry once with a fresh token
    if (
      error.response?.status === 403 &&
      error.response?.data?.message?.toLowerCase().includes('csrf') &&
      !originalRequest._csrfRetry
    ) {
      originalRequest._csrfRetry = true;
      csrfTokenManager.clearToken();
      const newToken = await csrfTokenManager.getToken();
      if (newToken) {
        originalRequest.headers['X-CSRF-Token'] = newToken;
        return api.request(originalRequest);
      }
    }

    // Handle network errors
    if (!error.response) {
      console.error('Network error:', error.message);
      // Clear CSRF token on network errors in case it's stale
      csrfTokenManager.clearToken();
      return Promise.reject(new Error('Network error. Please check your connection.'));
    }

    // Handle timeout errors
    if (error.code === 'ECONNABORTED') {
      console.error('Request timeout');
      return Promise.reject(new Error('Request timeout. Please try again.'));
    }

    const status = error.response?.status;

    // With HttpOnly cookies, token refresh is handled automatically by the backend
    // When a 401 occurs, it means the refresh token is also invalid/expired
    if (status === 401) {
      tokenStorage.clear();
      csrfTokenManager.clearToken();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
      return Promise.reject(error);
    }

    return Promise.reject(error);
  }
);

// Role Management API
export const roleAPI = {
  list: () => api.get('/roles'),
  get: (id: string) => api.get(`/roles/${id}`),
  create: (data: { name: string; description?: string }) => api.post('/roles', data),
  update: (id: string, data: { name: string; description?: string }) => api.put(`/roles/${id}`, data),
  delete: (id: string) => api.delete(`/roles/${id}`),
  getPermissions: (id: string) => api.get(`/roles/${id}/permissions`),
  assignPermissions: (id: string, permissionIds: string[]) =>
    api.post(`/roles/${id}/permissions`, { permission_ids: permissionIds }),
  removePermission: (id: string, permissionId: string) =>
    api.delete(`/roles/${id}/permissions/${permissionId}`),
};

// Permission API
export const permissionAPI = {
  list: () => api.get('/permissions'),
};

// Notification Template API
export const notificationTemplateAPI = {
  list: (eventType?: string) =>
    api.get('/notification-templates', { params: eventType ? { event_type: eventType } : {} }),
  get: (id: string) => api.get(`/notification-templates/${id}`),
  create: (data: {
    name: string;
    type: 'email' | 'sms' | 'webhook' | 'in_app';
    event_type: string;
    subject?: string;
    body_template: string;
    variables?: Record<string, any>;
    is_active: boolean;
  }) => api.post('/notification-templates', data),
  update: (id: string, data: {
    name: string;
    type: 'email' | 'sms' | 'webhook' | 'in_app';
    event_type: string;
    subject?: string;
    body_template: string;
    variables?: Record<string, any>;
    is_active: boolean;
  }) => api.put(`/notification-templates/${id}`, data),
  delete: (id: string) => api.delete(`/notification-templates/${id}`),
  test: (id: string, testData: Record<string, any>) =>
    api.post(`/notification-templates/${id}/test`, { test_data: testData }),
};

// Alert Rule API
export const alertRuleAPI = {
  list: (eventType?: string) =>
    api.get('/alert-rules', { params: eventType ? { event_type: eventType } : {} }),
  get: (id: string) => api.get(`/alert-rules/${id}`),
  create: (data: {
    name: string;
    description?: string;
    event_type: string;
    conditions: Record<string, any>;
    actions: Array<{
      type: string;
      target: string;
      template_id?: string;
      custom_data?: Record<string, any>;
    }>;
    is_active: boolean;
  }) => api.post('/alert-rules', data),
  update: (id: string, data: {
    name?: string;
    description?: string;
    event_type?: string;
    conditions?: Record<string, any>;
    actions?: Array<{
      type: string;
      target: string;
      template_id?: string;
      custom_data?: Record<string, any>;
    }>;
    is_active?: boolean;
  }) => api.put(`/alert-rules/${id}`, data),
  delete: (id: string) => api.delete(`/alert-rules/${id}`),
  test: (id: string, testData: Record<string, any>) =>
    api.post(`/alert-rules/${id}/test`, { test_data: testData }),
};

type WebhookTestRequest = {
  target_url: string;
  method?: 'POST' | 'PUT' | 'PATCH';
  headers?: Record<string, string>;
  payload?: Record<string, any>;
  secret_id?: string;
  secret?: string;
  event_type?: string;
};

type WebhookTestResponse = {
  success: boolean;
  target_status?: number;
  response_headers?: Record<string, string>;
  response_body_snippet?: string;
  duration_ms: number;
  signature: { algorithm: string; header_name: string };
  error?: string;
};

// Attach webhooks helper to axios instance for convenient usage as api.webhooks.test(...)
(api as any).webhooks = {
  test: async (payload: WebhookTestRequest): Promise<WebhookTestResponse> => {
    const { data } = await api.post('/webhooks/test', payload);
    return data as WebhookTestResponse;
  },
};

export default api;
