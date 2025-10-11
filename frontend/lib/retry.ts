/**
 * Retry a promise-based function with exponential backoff
 * Useful for handling transient network failures
 */
export async function withRetry<T>(
  fn: () => Promise<T>,
  options: {
    maxRetries?: number;
    initialDelayMs?: number;
    maxDelayMs?: number;
    backoffMultiplier?: number;
    shouldRetry?: (error: unknown, attempt: number) => boolean;
    onRetry?: (error: unknown, attempt: number, delayMs: number) => void;
  } = {}
): Promise<T> {
  const {
    maxRetries = 3,
    initialDelayMs = 1000,
    maxDelayMs = 10000,
    backoffMultiplier = 2,
    shouldRetry = () => true,
    onRetry,
  } = options;

  let lastError: Error;
  let attempt = 0;

  while (attempt <= maxRetries) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;
      attempt++;

      // Check if we've exhausted retries
      if (attempt > maxRetries) {
        throw lastError;
      }

      // Check if we should retry this error
      if (!shouldRetry(error, attempt)) {
        throw lastError;
      }

      // Calculate delay with exponential backoff
      const exponentialDelay = initialDelayMs * Math.pow(backoffMultiplier, attempt - 1);
      const delayMs = Math.min(exponentialDelay, maxDelayMs);

      // Add jitter to prevent thundering herd
      const jitter = Math.random() * 0.3 * delayMs; // ±30% jitter
      const finalDelay = delayMs + jitter;

      // Call retry callback if provided
      onRetry?.(error, attempt, finalDelay);

      // Wait before retrying
      await sleep(finalDelay);
    }
  }

  // This should never be reached, but TypeScript needs it
  throw lastError!;
}

/**
 * Sleep for a specified number of milliseconds
 */
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Retry configuration presets for common scenarios
 */
export const retryPresets = {
  /**
   * Quick retry for fast-failing operations
   */
  quick: {
    maxRetries: 2,
    initialDelayMs: 500,
    maxDelayMs: 2000,
  },
  
  /**
   * Standard retry for most API calls
   */
  standard: {
    maxRetries: 3,
    initialDelayMs: 1000,
    maxDelayMs: 10000,
  },
  
  /**
   * Aggressive retry for critical operations
   */
  aggressive: {
    maxRetries: 5,
    initialDelayMs: 2000,
    maxDelayMs: 30000,
  },
} as const;

/**
 * Check if an error is retryable (network errors, 5xx status codes)
 */
export function isRetryableError(error: unknown): boolean {
  if (!error || typeof error !== 'object') {
    return false;
  }

  const err = error as { response?: { status?: number }; code?: string; message?: string };

  // Network errors (no response received)
  if (!err.response) {
    return true;
  }

  // Server errors (5xx)
  const status = err.response.status;
  if (status && status >= 500 && status < 600) {
    return true;
  }

  // Rate limit errors (429) can be retried
  if (status === 429) {
    return true;
  }

  // Timeout errors
  if (err.code === 'ECONNABORTED' || err.message?.includes('timeout')) {
    return true;
  }

  return false;
}

/**
 * Retry wrapper specifically for API calls
 * Uses standard retry configuration and checks for retryable errors
 */
export async function retryApiCall<T>(
  fn: () => Promise<T>,
  customOptions?: Partial<Parameters<typeof withRetry>[1]>
): Promise<T> {
  return withRetry(fn, {
    ...retryPresets.standard,
    shouldRetry: isRetryableError,
    ...customOptions,
    onRetry: (error, attempt, delayMs) => {
      console.warn(`API call failed (attempt ${attempt}), retrying in ${Math.round(delayMs)}ms...`, error);
      customOptions?.onRetry?.(error, attempt, delayMs);
    },
  });
}
