import { QueryClient } from '@tanstack/react-query';

// Create a query client with optimized defaults
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Retry configuration
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      
      // Caching configuration
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
      
      // Refetch configuration
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
      refetchOnMount: true,
      
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
    list: (filters: Record<string, any>) => [...queryKeys.products.lists(), filters] as const,
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
    list: (filters: Record<string, any>) => [...queryKeys.inventory.lists(), filters] as const,
    details: () => [...queryKeys.inventory.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.inventory.details(), id] as const,
  },
  
  // Orders
  orders: {
    all: ['orders'] as const,
    lists: () => [...queryKeys.orders.all, 'list'] as const,
    list: (filters: Record<string, any>) => [...queryKeys.orders.lists(), filters] as const,
    details: () => [...queryKeys.orders.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.orders.details(), id] as const,
  },
  
  // Invoices
  invoices: {
    all: ['invoices'] as const,
    lists: () => [...queryKeys.invoices.all, 'list'] as const,
    list: (filters: Record<string, any>) => [...queryKeys.invoices.lists(), filters] as const,
    details: () => [...queryKeys.invoices.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.invoices.details(), id] as const,
  },
  
  // Analytics
  analytics: {
    all: ['analytics'] as const,
    salesTrends: (params?: Record<string, any>) => [...queryKeys.analytics.all, 'sales-trends', params] as const,
    topProducts: (params?: Record<string, any>) => [...queryKeys.analytics.all, 'top-products', params] as const,
    gstTotals: (params?: Record<string, any>) => [...queryKeys.analytics.all, 'gst-totals', params] as const,
  },
};

// Helper function to invalidate related queries
export const invalidateRelatedQueries = (queryClient: QueryClient, entityType: string, id?: string) => {
  switch (entityType) {
    case 'product':
      queryClient.invalidateQueries({ queryKey: queryKeys.products.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.inventory.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.all });
      break;
    case 'inventory':
      queryClient.invalidateQueries({ queryKey: queryKeys.inventory.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.lowStock() });
      break;
    case 'order':
      queryClient.invalidateQueries({ queryKey: queryKeys.orders.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.analytics.all });
      break;
    case 'invoice':
      queryClient.invalidateQueries({ queryKey: queryKeys.invoices.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.dashboard.invoices() });
      break;
    default:
      break;
  }
};
