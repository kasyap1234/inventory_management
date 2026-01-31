import axios, { AxiosRequestConfig } from 'axios';
import { csrfTokenManager, tokenStorage } from '@/lib/security';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/v1';

if (!process.env.NEXT_PUBLIC_API_URL && typeof window !== 'undefined') {
  console.warn('NEXT_PUBLIC_API_URL is not defined, falling back to relative path /v1');
}

// Request deduplication: Track in-flight requests
const pendingRequests = new Map<string, AbortController>();

function generateRequestKey(config: AxiosRequestConfig): string {
  return `${config.method}:${config.url}:${JSON.stringify(config.params)}:${JSON.stringify(config.data)}`;
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

    // Request deduplication for GET requests
    if (config.method?.toLowerCase() === 'get' && config.url) {
      const requestKey = generateRequestKey(config);
      const existingController = pendingRequests.get(requestKey);

      if (existingController) {
        // Abort the previous identical request
        existingController.abort();
      }

      // Create new AbortController for this request
      const controller = new AbortController();
      pendingRequests.set(requestKey, controller);
      config.signal = controller.signal;

      // Clean up after request completes (handled in response interceptor)
    }

    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => {
    // Clean up pending request after success
    if (response.config.method?.toLowerCase() === 'get' && response.config.url) {
      const requestKey = generateRequestKey(response.config);
      pendingRequests.delete(requestKey);
    }

    // Handle 4xx errors that didn't throw
    if (response.status >= 400 && response.status < 500) {
      const message = response.data?.message || response.data?.error?.message || 'Request failed';
      const error = new Error(message) as Error & { response: typeof response; code: string; status: number };
      error.response = response;
      error.code = `HTTP_${response.status}`;
      error.name = 'HttpError';
      throw error;
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Clean up pending request on error (unless it was aborted)
    if (originalRequest && originalRequest.method?.toLowerCase() === 'get' && originalRequest.url) {
      const requestKey = generateRequestKey(originalRequest);
      pendingRequests.delete(requestKey);
    }

    // If request was aborted, propagate the abort error
    if (error.name === 'CanceledError' || error.code === 'ERR_CANCELED') {
      return Promise.reject(new Error('Request cancelled'));
    }

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
      const networkError = new Error('Network error. Please check your connection.') as Error & { code: string; isRetryable: boolean };
      networkError.code = 'NETWORK_ERROR';
      networkError.isRetryable = true;
      return Promise.reject(networkError);
    }

    // Handle timeout errors
    if (error.code === 'ECONNABORTED') {
      console.error('Request timeout');
      const timeoutError = new Error('Request timeout. Please try again.') as Error & { code: string; isRetryable: boolean };
      timeoutError.code = 'TIMEOUT';
      timeoutError.isRetryable = true;
      return Promise.reject(timeoutError);
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

// Export function to cancel in-flight requests
export function cancelRequest(config: AxiosRequestConfig) {
  const requestKey = generateRequestKey(config);
  const controller = pendingRequests.get(requestKey);
  if (controller) {
    controller.abort();
    pendingRequests.delete(requestKey);
  }
}

// Export function to cancel all in-flight requests
export function cancelAllRequests() {
  pendingRequests.forEach((controller) => {
    controller.abort();
  });
  pendingRequests.clear();
}

export default api;
