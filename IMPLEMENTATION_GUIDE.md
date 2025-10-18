# Implementation Guide for Optimizations

## Quick Start

### 1. Update Your Query Client Configuration

Replace your existing QueryClient setup with the optimized version:

```typescript
// app/providers.tsx or app/layout.tsx
import { queryClient } from '@/lib/queryClient';
import { QueryClientProvider } from '@tanstack/react-query';

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

### 2. Use Centralized Error Handling

Update your API calls to use the error handler:

```typescript
import { handleApiError, formatErrorMessage } from '@/lib/errorHandler';

// In your components
const { mutate, error } = useMutation({
  mutationFn: createProduct,
  onError: (err) => {
    const appError = handleApiError(err);
    toast.error(formatErrorMessage(appError));
  },
});
```

### 3. Use Query Key Factory

Replace hardcoded query keys with the factory:

```typescript
import { queryKeys } from '@/lib/queryClient';

// Before
useQuery({ queryKey: ['products'] });

// After
useQuery({ queryKey: queryKeys.products.all });
```

### 4. Add Performance Tracking

Track critical operations:

```typescript
import { measureApiCall, usePerformanceTracking } from '@/lib/performance';

// In components
function ProductList() {
  usePerformanceTracking('ProductList');
  
  const { data } = useQuery({
    queryKey: queryKeys.products.all,
    queryFn: () => measureApiCall('fetchProducts', () => api.get('/products')),
  });
}
```

## Backend Integration

### Update Go Services

The backend services have been updated with comprehensive validation. No additional configuration needed, but ensure you handle the new error responses:

```typescript
// Frontend error handling for new backend errors
{
  "error": "validation_error",
  "message": "Quantity must be positive",
  "details": {
    "quantity": "Quantity must be positive"
  }
}
```

## Testing the Optimizations

### 1. Test Network Resilience

```bash
# Simulate slow network
curl --limit-rate 1k http://localhost:8080/v1/products

# Simulate timeout
curl --max-time 1 http://localhost:8080/v1/products
```

### 2. Test Cache Behavior

```typescript
// In browser console
import { queryClient } from '@/lib/queryClient';

// Check cache
console.log(queryClient.getQueryData(['products']));

// Invalidate cache
queryClient.invalidateQueries({ queryKey: ['products'] });
```

### 3. Test Error Handling

```typescript
// Trigger validation error
createProduct.mutate({ name: '', unit_price: -1 });

// Should show: "Invalid product data: name and positive unit_price are required"
```

### 4. Monitor Performance

```typescript
import { performanceMonitor } from '@/lib/performance';

// View metrics
console.log(performanceMonitor.getSummary());

// Export for analysis
console.log(performanceMonitor.exportMetrics());
```

## Migration Checklist

- [ ] Update QueryClient configuration
- [ ] Replace hardcoded query keys with factory
- [ ] Add error handling to all mutations
- [ ] Add performance tracking to critical paths
- [ ] Test retry logic with network failures
- [ ] Test cache invalidation
- [ ] Verify validation errors display correctly
- [ ] Monitor performance metrics
- [ ] Update API documentation
- [ ] Train team on new patterns

## Common Patterns

### Pattern 1: Optimistic Updates

```typescript
const updateProduct = useMutation({
  mutationFn: (data) => api.put(`/products/${data.id}`, data),
  onMutate: async (newProduct) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: queryKeys.products.all });
    
    // Snapshot previous value
    const previous = queryClient.getQueryData(queryKeys.products.all);
    
    // Optimistically update
    queryClient.setQueryData(queryKeys.products.all, (old: any) => ({
      ...old,
      products: old.products.map((p: any) => 
        p.id === newProduct.id ? newProduct : p
      ),
    }));
    
    return { previous };
  },
  onError: (err, newProduct, context) => {
    // Rollback on error
    queryClient.setQueryData(queryKeys.products.all, context?.previous);
  },
  onSettled: () => {
    // Refetch after mutation
    queryClient.invalidateQueries({ queryKey: queryKeys.products.all });
  },
});
```

### Pattern 2: Dependent Queries

```typescript
// Product details depend on product ID
const { data: product } = useQuery({
  queryKey: queryKeys.products.detail(productId),
  queryFn: () => api.get(`/products/${productId}`),
  enabled: !!productId, // Only run if productId exists
});

// Images depend on product
const { data: images } = useQuery({
  queryKey: ['product-images', productId],
  queryFn: () => api.get(`/products/${productId}/images`),
  enabled: !!product, // Only run if product loaded
});
```

### Pattern 3: Infinite Queries

```typescript
const {
  data,
  fetchNextPage,
  hasNextPage,
  isFetchingNextPage,
} = useInfiniteQuery({
  queryKey: queryKeys.products.lists(),
  queryFn: ({ pageParam = 0 }) => 
    api.get(`/products?limit=20&offset=${pageParam}`),
  getNextPageParam: (lastPage, pages) => {
    const nextOffset = pages.length * 20;
    return lastPage.products.length === 20 ? nextOffset : undefined;
  },
  staleTime: 5 * 60 * 1000,
});
```

### Pattern 4: Parallel Queries

```typescript
function Dashboard() {
  const queries = useQueries({
    queries: [
      {
        queryKey: queryKeys.dashboard.analytics(),
        queryFn: analyticsService.getDashboardAnalytics,
      },
      {
        queryKey: queryKeys.dashboard.lowStock(),
        queryFn: () => analyticsService.getLowStockReport({ threshold: 10 }),
      },
      {
        queryKey: queryKeys.dashboard.invoices(),
        queryFn: () => invoiceService.getUnpaid(),
      },
    ],
  });

  const isLoading = queries.some(q => q.isLoading);
  const hasError = queries.some(q => q.error);
}
```

## Performance Tips

### 1. Reduce Bundle Size

```typescript
// Use dynamic imports for large components
const HeavyComponent = dynamic(() => import('./HeavyComponent'), {
  loading: () => <Spinner />,
  ssr: false,
});
```

### 2. Memoize Expensive Calculations

```typescript
const expensiveValue = useMemo(() => {
  return products.reduce((sum, p) => sum + p.price * p.quantity, 0);
}, [products]);
```

### 3. Debounce Search Inputs

```typescript
import { useDebouncedValue } from '@/hooks/useDebounce';

const [search, setSearch] = useState('');
const debouncedSearch = useDebouncedValue(search, 500);

const { data } = useQuery({
  queryKey: ['products', 'search', debouncedSearch],
  queryFn: () => api.get(`/products/search?q=${debouncedSearch}`),
  enabled: debouncedSearch.length > 2,
});
```

### 4. Virtual Scrolling for Large Lists

```typescript
import { useVirtualizer } from '@tanstack/react-virtual';

const rowVirtualizer = useVirtualizer({
  count: products.length,
  getScrollElement: () => parentRef.current,
  estimateSize: () => 50,
});
```

## Troubleshooting

### Issue: Queries Not Refetching

**Solution**: Check staleTime and gcTime settings

```typescript
// Force refetch
queryClient.invalidateQueries({ queryKey: queryKeys.products.all });

// Or refetch manually
refetch();
```

### Issue: Too Many API Calls

**Solution**: Increase staleTime or disable refetchOnWindowFocus

```typescript
useQuery({
  queryKey: ['data'],
  queryFn: fetchData,
  staleTime: 10 * 60 * 1000, // 10 minutes
  refetchOnWindowFocus: false,
});
```

### Issue: Cache Not Clearing

**Solution**: Use proper cache invalidation

```typescript
// Remove specific query
queryClient.removeQueries({ queryKey: ['product', id] });

// Clear all cache
queryClient.clear();
```

### Issue: Slow Initial Load

**Solution**: Implement prefetching

```typescript
// Prefetch on hover
const prefetchProduct = (id: string) => {
  queryClient.prefetchQuery({
    queryKey: queryKeys.products.detail(id),
    queryFn: () => api.get(`/products/${id}`),
  });
};

<Link 
  href={`/products/${id}`}
  onMouseEnter={() => prefetchProduct(id)}
>
  View Product
</Link>
```

## Monitoring & Debugging

### Enable React Query DevTools

```typescript
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

<QueryClientProvider client={queryClient}>
  <App />
  <ReactQueryDevtools initialIsOpen={false} />
</QueryClientProvider>
```

### Log Performance Metrics

```typescript
// In production
if (process.env.NODE_ENV === 'production') {
  window.addEventListener('load', () => {
    setTimeout(() => {
      const metrics = performanceMonitor.getSummary();
      // Send to analytics
      analytics.track('performance_metrics', metrics);
    }, 5000);
  });
}
```

## Next Steps

1. **Monitor Production**: Set up error tracking (Sentry, LogRocket)
2. **A/B Testing**: Test performance improvements with real users
3. **Optimize Images**: Implement next/image for automatic optimization
4. **Service Worker**: Add offline support with Workbox
5. **CDN**: Serve static assets from CDN
6. **Database Indexes**: Ensure proper indexes on frequently queried fields

## Support

For questions or issues:
1. Check this guide
2. Review the OPTIMIZATION_SUMMARY.md
3. Check React Query documentation
4. Review error logs in browser console
