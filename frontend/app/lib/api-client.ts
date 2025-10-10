// Optimized API client with request batching and caching
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

import { csrfTokenManager, tokenStorage } from '@/lib/security';

// Request queue for batching
interface QueuedRequest {
  config: AxiosRequestConfig;
  resolve: (value: any) => void;
  reject: (reason: any) => void;
}

class APIClient {
  private client: AxiosInstance;
  private requestQueue: Map<string, QueuedRequest[]> = new Map();
  private batchTimeout: NodeJS.Timeout | null = null;
  private readonly BATCH_DELAY = 50; // 50ms delay for batching

  constructor(baseURL: string) {
    this.client = axios.create({
      baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor
    this.client.interceptors.request.use(
      async (config) => {
        const headers = config.headers ?? {};
        const token = tokenStorage.getAccessToken();
        if (token) {
          headers.Authorization = `Bearer ${token}`;
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

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          // Handle token refresh or redirect to login
          if (typeof window !== 'undefined') {
            tokenStorage.clear();
            csrfTokenManager.clearToken();
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );
  }

  // Standard GET request with caching
  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  // Standard POST request
  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  // Standard PUT request
  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  // Standard DELETE request
  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  // Batch multiple GET requests
  async batchGet<T = any>(urls: string[]): Promise<T[]> {
    const promises = urls.map(url => this.get<T>(url));
    return Promise.all(promises);
  }

  // Upload file with progress tracking
  async uploadFile(url: string, file: File, onProgress?: (progress: number) => void): Promise<any> {
    const formData = new FormData();
    formData.append('file', file);

    return this.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  }
}

// Create singleton instance
const apiBaseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/v1';
export const apiClient = new APIClient(apiBaseURL);

// Export for convenience
export default apiClient;
