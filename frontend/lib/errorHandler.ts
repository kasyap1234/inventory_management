import { AxiosError } from 'axios';

export interface AppError {
  message: string;
  code?: string;
  status?: number;
  details?: Record<string, unknown>;
}

/**
 * Centralized error handler for API errors
 */
export const handleApiError = (error: unknown): AppError => {
  // Handle Axios errors
  if (error instanceof Error && 'isAxiosError' in error) {
    const axiosError = error as AxiosError<{ message?: string; errors?: unknown; details?: unknown }>;

    // Network errors
    if (!axiosError.response) {
      if (axiosError.code === 'ECONNABORTED') {
        return {
          message: 'Request timeout. Please try again.',
          code: 'TIMEOUT',
          status: 408,
        };
      }
      return {
        message: 'Network error. Please check your connection.',
        code: 'NETWORK_ERROR',
        status: 0,
      };
    }

    // Server errors
    const { status, data } = axiosError.response;

    // Handle specific status codes
    switch (status) {
      case 400:
        return {
          message: data?.message || 'Invalid request. Please check your input.',
          code: 'BAD_REQUEST',
          status: 400,
          details: (typeof data?.errors === 'object' && data?.errors !== null ? data.errors as Record<string, unknown> : typeof data?.details === 'object' && data?.details !== null ? data.details as Record<string, unknown> : undefined),
        };

      case 401:
        return {
          message: 'Authentication required. Please log in.',
          code: 'UNAUTHORIZED',
          status: 401,
        };

      case 403:
        return {
          message: 'You do not have permission to perform this action.',
          code: 'FORBIDDEN',
          status: 403,
        };

      case 404:
        return {
          message: data?.message || 'Resource not found.',
          code: 'NOT_FOUND',
          status: 404,
        };

      case 409:
        return {
          message: data?.message || 'Conflict. Resource already exists.',
          code: 'CONFLICT',
          status: 409,
        };

      case 422:
        return {
          message: data?.message || 'Validation failed.',
          code: 'VALIDATION_ERROR',
          status: 422,
          details: (typeof data?.errors === 'object' && data?.errors !== null ? data.errors as Record<string, unknown> : typeof data?.details === 'object' && data?.details !== null ? data.details as Record<string, unknown> : undefined),
        };

      case 429:
        return {
          message: 'Too many requests. Please try again later.',
          code: 'RATE_LIMIT',
          status: 429,
        };

      case 500:
        return {
          message: 'Server error. Please try again later.',
          code: 'INTERNAL_SERVER_ERROR',
          status: 500,
        };

      case 503:
        return {
          message: 'Service temporarily unavailable. Please try again later.',
          code: 'SERVICE_UNAVAILABLE',
          status: 503,
        };

      default:
        return {
          message: data?.message || 'An unexpected error occurred.',
          code: 'UNKNOWN_ERROR',
          status,
        };
    }
  }

  // Handle standard JavaScript errors
  if (error instanceof Error) {
    return {
      message: error.message,
      code: 'ERROR',
    };
  }

  // Handle unknown errors
  return {
    message: 'An unexpected error occurred.',
    code: 'UNKNOWN_ERROR',
  };
};

/**
 * Format error for display to user
 */
export const formatErrorMessage = (error: AppError): string => {
  if (error.details) {
    const detailMessages = Object.entries(error.details)
      .map(([field, message]) => `${field}: ${message}`)
      .join(', ');
    return `${error.message} (${detailMessages})`;
  }
  return error.message;
};

/**
 * Check if error is retryable
 */
export const isRetryableError = (error: AppError): boolean => {
  const retryableCodes = [
    'TIMEOUT',
    'NETWORK_ERROR',
    'RATE_LIMIT',
    'SERVICE_UNAVAILABLE',
    'INTERNAL_SERVER_ERROR',
  ];

  return retryableCodes.includes(error.code || '') || (error.status || 0) >= 500;
};

/**
 * Log error for monitoring
 */
export const logError = (error: AppError, context?: Record<string, unknown>) => {
  if (process.env.NODE_ENV === 'development') {
    console.error('Application Error:', {
      ...error,
      context,
      timestamp: new Date().toISOString(),
    });
  }

  // In production, send to error tracking service (e.g., Sentry)
  // if (process.env.NODE_ENV === 'production') {
  //   Sentry.captureException(error, { extra: context });
  // }
};
