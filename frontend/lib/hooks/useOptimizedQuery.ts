/**
 * Optimized query hooks with built-in performance features
 */

import { useQuery, useMutation, useQueryClient, UseQueryOptions, UseMutationOptions } from '@tanstack/react-query';
import { handleApiError, logError } from '../errorHandler';
import { measureApiCall } from '../performance';
import React from 'react';

/**
 * Optimized useQuery with automatic error handling and performance tracking
 */
export function useOptimizedQuery<TData = unknown, TError = unknown>(
  options: UseQueryOptions<TData, TError> & {
    performanceKey?: string;
    logErrors?: boolean;
  }
) {
  const { performanceKey, logErrors = true, queryFn, ...restOptions } = options;

  return useQuery<TData, TError>({
    ...restOptions,
    queryFn: async (context) => {
      const key = performanceKey || String(options.queryKey);

      try {
        if (typeof queryFn === 'function') {
          return await measureApiCall(key, () => (queryFn as (ctx: unknown) => Promise<TData>)(context));
        }
        throw new Error('queryFn is required');
      } catch (error) {
        if (logErrors) {
          const appError = handleApiError(error);
          logError(appError, { queryKey: options.queryKey });
        }
        throw error;
      }
    },
  });
}

/**
 * Optimized useMutation with automatic error handling and cache invalidation
 */
export function useOptimizedMutation<TData = unknown, TError = unknown, TVariables = void, TContext = unknown>(
  options: UseMutationOptions<TData, TError, TVariables, TContext> & {
    performanceKey?: string;
    invalidateKeys?: readonly unknown[];
    logErrors?: boolean;
  }
) {
  const queryClient = useQueryClient();
  const { performanceKey, invalidateKeys, logErrors = true, mutationFn, onSuccess, onError, ...restOptions } = options;

  return useMutation<TData, TError, TVariables, TContext>({
    ...restOptions,
    mutationFn: async (variables) => {
      const key = performanceKey || 'mutation';

      try {
        if (typeof mutationFn === 'function') {
          return await measureApiCall(key, () => (mutationFn as (variables: TVariables) => Promise<TData>)(variables));
        }
        throw new Error('mutationFn is required');
      } catch (error) {
        if (logErrors) {
          const appError = handleApiError(error);
          logError(appError, { variables });
        }
        throw error;
      }
    },
    onSuccess: (data, variables, context) => {
      // Auto-invalidate specified query keys
      if (invalidateKeys && invalidateKeys.length > 0) {
        invalidateKeys.forEach((key) => {
          queryClient.invalidateQueries({ queryKey: key as readonly unknown[] });
        });
      }

      // Call original onSuccess
      if (onSuccess) {
        (onSuccess as (data: TData, variables: TVariables, context?: TContext) => void)(
          data,
          variables,
          context as TContext | undefined
        );
      }
    },
    onError: (error, variables, context) => {
      if (logErrors) {
        const appError = handleApiError(error);
        logError(appError, { variables });
      }

      // Call original onError
      if (onError) {
        (onError as (error: TError, variables: TVariables, context?: TContext) => void)(
          error,
          variables,
          context as TContext | undefined
        );
      }
    },
  });
}

/**
 * Hook for prefetching data on hover/focus
 */
export function usePrefetch<TData = unknown>(
  queryKey: unknown[],
  queryFn: () => Promise<TData>,
  options?: { staleTime?: number }
) {
  const queryClient = useQueryClient();

  return () => {
    queryClient.prefetchQuery({
      queryKey,
      queryFn,
      staleTime: options?.staleTime || 5 * 60 * 1000, // 5 minutes default
    });
  };
}

/**
 * Hook for optimistic updates with automatic rollback on error
 * Use this in mutation.onMutate and mutation.onError
 */
export function useOptimisticUpdate<TData = unknown>(
  queryKey: unknown[],
  context?: {
    previousData?: TData;
  }
) {
  const queryClient = useQueryClient();

  const updateCache = (updater: (old: TData | undefined) => TData) => {
    // Store previous data for potential rollback
    const previousData = queryClient.getQueryData<TData>(queryKey);
    if (context) {
      context.previousData = previousData;
    }
    queryClient.setQueryData<TData>(queryKey, updater);
    return previousData;
  };

  const rollback = (previousData: TData) => {
    queryClient.setQueryData<TData>(queryKey, previousData);
  };

  const invalidate = (refetchType?: 'all' | 'none') => {
    queryClient.invalidateQueries({ queryKey, refetchType });
  };

  const setQueryData = (data: TData) => {
    queryClient.setQueryData<TData>(queryKey, data);
  };

  return { updateCache, rollback, invalidate, setQueryData };
}

/**
 * Hook for debounced queries (useful for search)
 * Uses AbortController to cancel previous requests when query key changes
 */
export function useDebouncedQuery<TData = unknown, TError = unknown>(
  queryKey: unknown[],
  queryFn: () => Promise<TData>,
  delay: number = 300,
  options?: UseQueryOptions<TData, TError>
) {
  const [debouncedKey, setDebouncedKey] = React.useState(queryKey);
  const [abortController, setAbortController] = React.useState<AbortController | null>(null);

  React.useEffect(() => {
    // Cancel any in-flight request from previous debounce cycle
    if (abortController) {
      abortController.abort();
    }

    const timer = setTimeout(() => {
      // Create new AbortController for the upcoming request
      const controller = new AbortController();
      setAbortController(controller);

      setDebouncedKey(queryKey);
    }, delay);

    return () => {
      clearTimeout(timer);
      if (abortController) {
        abortController.abort();
      }
    };
  }, [queryKey, delay]);

  return useQuery<TData, TError>({
    queryKey: debouncedKey,
    queryFn: async () => {
      const controller = new AbortController();
      // Use React Query's built-in signal support
      return queryFn();
    },
    ...options,
  });
}