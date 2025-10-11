/**
 * Standard API response wrapper
 */
export interface ApiResponse<T = unknown> {
  data: T;
  message?: string;
}

/**
 * API error response structure
 */
export interface ApiErrorResponse {
  message: string;
  code?: string;
  details?: Record<string, string[]>;
  error?: {
    message: string;
    code?: string;
  };
}

/**
 * Paginated response wrapper
 */
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
}

/**
 * List response (non-paginated)
 */
export interface ListResponse<T> {
  [key: string]: T[];
}
