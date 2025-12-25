import { QueryClient } from '@tanstack/react-query';

// Helper function to get appropriate staleTime based on query key
function getStaleTime(queryKey: unknown[]): number {
  const key = queryKey[0] as string;

  // Dynamic data - shorter cache
  if (key === 'inventory' || key === 'orders') {
    return 30 * 1000; // 30 seconds
  }

  // Moderately dynamic data
  if (key === 'dashboard' || key === 'products') {
    return 2 * 60 * 1000; // 2 minutes
  }

  // Reference data - longer cache
  if (key === 'warehouses' || key === 'categories' || key === 'suppliers' || key === 'distributors' || key === 'roles' || key === 'permissions') {
    return 10 * 60 * 1000; // 10 minutes
  }

  // User/session data
  if (key === 'user' || key === 'tenants') {
    return 5 * 60 * 1000; // 5 minutes
  }

  // Analytics data
  if (key === 'analytics') {
    return 60 * 1000; // 1 minute
  }

  // Default
  return 5 * 60 * 1000; // 5 minutes
}

// Create a query client with optimized defaults
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Retry configuration
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),

      // Caching configuration - use function for dynamic staleTime
      staleTime: getStaleTime([]),
      gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)

      // Refetch configuration
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
      refetchOnMount: false, // Changed from true to prevent unnecessary refetches

      // Error handling
      throwOnError: false,

      // Network mode
      networkMode: 'online',
    },
    mutations: {
      // Retry configuration for mutations
      retry: 2,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),

      // Network mode
      networkMode: 'online',

      // Error handling
      throwOnError: false,
    },
  },
});

// Query key factory for consistent key generation
export const queryKeys = {
  // Products
  products: {
    all: ['products'] as const,
    lists: () => [...queryKeys.products.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.products.lists(), filters] as const,
    details: () => [...queryKeys.products.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.products.details(), id] as const,
  },

  // Dashboard
  dashboard: {
    all: ['dashboard'] as const,
    analytics: () => [...queryKeys.dashboard.all, 'analytics'] as const,
    lowStock: () => [...queryKeys.dashboard.all, 'low-stock'] as const,
    invoices: () => [...queryKeys.dashboard.all, 'unpaid-invoices'] as const,
  },

  // Inventory
  inventory: {
    all: ['inventory'] as const,
    lists: () => [...queryKeys.inventory.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.inventory.lists(), filters] as const,
    details: () => [...queryKeys.inventory.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.inventory.details(), id] as const,
  },

  // Orders
  orders: {
    all: ['orders'] as const,
    lists: () => [...queryKeys.orders.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.orders.lists(), filters] as const,
    details: () => [...queryKeys.orders.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.orders.details(), id] as const,
  },

  // Invoices
  invoices: {
    all: ['invoices'] as const,
    lists: () => [...queryKeys.invoices.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.invoices.lists(), filters] as const,
    details: () => [...queryKeys.invoices.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.invoices.details(), id] as const,
  },

  // Analytics
  analytics: {
    all: ['analytics'] as const,
    salesTrends: (params?: Record<string, unknown>) => [...queryKeys.analytics.all, 'sales-trends', params] as const,
    topProducts: (params?: Record<string, unknown>) => [...queryKeys.analytics.all, 'top-products', params] as const,
    gstTotals: (params?: Record<string, unknown>) => [...queryKeys.analytics.all, 'gst-totals', params] as const,
  },
};

// Helper function to selectively invalidate queries based on operation type
export const invalidateRelatedQueries = (
  queryClient: QueryClient,
  entityType: string,
  operation: 'create' | 'update' | 'delete' | 'list',
  itemId?: string
) => {
  switch (entityType) {
    case 'product':
      if (operation === 'delete' && itemId) {
        // For delete, remove specific item and update lists
        queryClient.removeQueries({ queryKey: queryKeys.products.detail(itemId) });
        queryClient.invalidateQueries({ queryKey: queryKeys.products.lists(), refetchType: 'none' });
      } else if (operation === 'create') {
        // For create, only invalidate list queries (not needed for specific detail)
        queryClient.invalidateQueries({ queryKey: queryKeys.products.lists() });
      } else {
        // For update, invalidate specific item and lists
        if (itemId) {
          queryClient.invalidateQueries({ queryKey: queryKeys.products.detail(itemId), refetchType: 'none' });
        }
        queryClient.invalidateQueries({ queryKey: queryKeys.products.lists(), refetchType: 'none' });
      }
      // Only invalidate inventory/dashboard if list changed
      if (operation !== 'update' && operation !== 'delete') {
        queryClient.invalidateQueries({ queryKey: queryKeys.inventory.lists(), refetchType: 'none' });
      }
      break;

    case 'inventory':
      if (operation === 'update' || operation === 'create') {
        queryClient.invalidateQueries({ queryKey: queryKeys.inventory.lists(), refetchType: 'none' });
      } else if (operation === 'delete' && itemId) {
        queryClient.removeQueries({ queryKey: queryKeys.inventory.detail(itemId) });
        queryClient.invalidateQueries({ queryKey: queryKeys.inventory.lists(), refetchType: 'none' });
      }
      // Dashboard low stock needs refresh on any inventory change
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.lowStock(), refetchType: 'none' });
      break;

    case 'order':
      if (operation === 'delete' && itemId) {
        queryClient.removeQueries({ queryKey: queryKeys.orders.detail(itemId) });
        queryClient.invalidateQueries({ queryKey: queryKeys.orders.lists(), refetchType: 'none' });
      } else if (operation === 'create' || operation === 'update') {
        queryClient.invalidateQueries({ queryKey: queryKeys.orders.lists(), refetchType: 'none' });
        if (itemId) {
          queryClient.invalidateQueries({ queryKey: queryKeys.orders.detail(itemId), refetchType: 'none' });
        }
      }
      // Analytics and dashboard only need refresh on new orders
      if (operation === 'create') {
        queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.all, refetchType: 'none' });
        queryClient.invalidateQueries({ queryKey: queryKeys.analytics.all, refetchType: 'none' });
      }
      break;

    case 'invoice':
      if (operation === 'delete' && itemId) {
        queryClient.removeQueries({ queryKey: queryKeys.invoices.detail(itemId) });
        queryClient.invalidateQueries({ queryKey: queryKeys.invoices.lists(), refetchType: 'none' });
      } else if (operation === 'create' || operation === 'update') {
        queryClient.invalidateQueries({ queryKey: queryKeys.invoices.lists(), refetchType: 'none' });
      }
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.invoices(), refetchType: 'none' });
      break;

    case 'user':
    case 'role':
      queryClient.invalidateQueries({ queryKey: [entityType], refetchType: 'none' });
      break;

    default:
      break;
  }
};
