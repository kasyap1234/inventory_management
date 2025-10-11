import { AxiosError } from 'axios';

/**
 * Base API error class for structured error handling
 */
export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public code?: string,
    public details?: Record<string, string[]>
  ) {
    super(message);
    this.name = 'ApiError';
    Object.setPrototypeOf(this, ApiError.prototype);
  }

  /**
   * Check if error is a specific HTTP status code
   */
  isStatus(status: number): boolean {
    return this.statusCode === status;
  }

  /**
   * Check if error is a validation error
   */
  isValidationError(): boolean {
    return this.statusCode === 422 || this.statusCode === 400;
  }

  /**
   * Check if error is an authentication error
   */
  isAuthError(): boolean {
    return this.statusCode === 401;
  }

  /**
   * Check if error is a permission error
   */
  isPermissionError(): boolean {
    return this.statusCode === 403;
  }

  /**
   * Check if error is a not found error
   */
  isNotFoundError(): boolean {
    return this.statusCode === 404;
  }
}

/**
 * Validation error with field-specific errors
 */
export class ValidationError extends Error {
  constructor(
    message: string,
    public errors: Record<string, string[]>
  ) {
    super(message);
    this.name = 'ValidationError';
    Object.setPrototypeOf(this, ValidationError.prototype);
  }

  /**
   * Get errors for a specific field
   */
  getFieldErrors(field: string): string[] {
    return this.errors[field] || [];
  }

  /**
   * Get the first error message for a field
   */
  getFirstFieldError(field: string): string | undefined {
    return this.getFieldErrors(field)[0];
  }

  /**
   * Check if a specific field has errors
   */
  hasFieldError(field: string): boolean {
    return !!this.errors[field]?.length;
  }
}

/**
 * Network error for connection issues
 */
export class NetworkError extends Error {
  constructor(message: string = 'Network connection failed') {
    super(message);
    this.name = 'NetworkError';
    Object.setPrototypeOf(this, NetworkError.prototype);
  }
}

/**
 * Timeout error for slow requests
 */
export class TimeoutError extends Error {
  constructor(message: string = 'Request timeout') {
    super(message);
    this.name = 'TimeoutError';
    Object.setPrototypeOf(this, TimeoutError.prototype);
  }
}

/**
 * Convert unknown error to structured ApiError
 */
export function handleApiError(error: unknown): ApiError {
  // Already an ApiError
  if (error instanceof ApiError) {
    return error;
  }

  // Axios error
  if (isAxiosError(error)) {
    const responseData = error.response?.data as Record<string, unknown> | undefined;
    const errorData = responseData?.error as Record<string, unknown> | undefined;
    
    const message = (typeof responseData?.message === 'string' ? responseData.message : undefined)
      || (typeof errorData?.message === 'string' ? errorData.message : undefined)
      || error.message;
    
    const statusCode = error.response?.status;
    const code = (typeof responseData?.code === 'string' ? responseData.code : undefined) || error.code;
    const details = (responseData?.details as Record<string, string[]> | undefined) 
      || (responseData?.errors as Record<string, string[]> | undefined);

    return new ApiError(message, statusCode, code, details);
  }

  // Network error
  if (error instanceof Error && error.message.includes('Network')) {
    return new ApiError(error.message, undefined, 'NETWORK_ERROR');
  }

  // Generic error
  if (error instanceof Error) {
    return new ApiError(error.message);
  }

  // Unknown error type
  return new ApiError('An unexpected error occurred');
}

/**
 * Type guard for AxiosError
 */
function isAxiosError(error: unknown): error is AxiosError {
  return (error as AxiosError).isAxiosError === true;
}

/**
 * Extract user-friendly error message from error object
 */
export function getErrorMessage(error: unknown): string {
  const apiError = handleApiError(error);
  return apiError.message;
}

/**
 * Extract validation errors from API response
 */
export function getValidationErrors(error: unknown): Record<string, string[]> | null {
  const apiError = handleApiError(error);
  return apiError.details || null;
}

/**
 * Check if error is retryable (network issues, 5xx errors)
 */
export function isRetryableError(error: unknown): boolean {
  const apiError = handleApiError(error);
  
  // Network errors are retryable
  if (apiError.code === 'NETWORK_ERROR') {
    return true;
  }

  // Server errors (5xx) are retryable
  if (apiError.statusCode && apiError.statusCode >= 500) {
    return true;
  }

  // Timeout errors are retryable
  if (error instanceof TimeoutError) {
    return true;
  }

  return false;
}
