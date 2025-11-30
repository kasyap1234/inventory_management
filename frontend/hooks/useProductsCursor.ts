import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { Product } from '@/types';

/**
 * Cursor-based pagination page info
 * More efficient than offset-based pagination for large datasets
 */
export interface CursorPageInfo {
    has_next_page: boolean;
    has_previous_page: boolean;
    start_cursor: string;
    end_cursor: string;
}

/**
 * Response shape for cursor-paginated products endpoint
 */
export interface CursorPaginatedProductsResponse {
    products: Product[];
    page_info: CursorPageInfo;
}

/**
 * Parameters for cursor-based pagination
 * Uses GraphQL-style cursor pagination (first/after for forward, last/before for backward)
 */
export interface CursorPaginationParams {
    /** Number of items to fetch (forward pagination) */
    first?: number;
    /** Cursor to start after (forward pagination) */
    after?: string;
    /** Number of items to fetch (backward pagination) */
    last?: number;
    /** Cursor to end before (backward pagination) */
    before?: string;
}

/**
 * Hook for products with cursor-based infinite scrolling pagination.
 * 
 * Benefits over offset-based pagination:
 * - Consistent performance regardless of page depth
 * - Stable results even with concurrent inserts/deletes
 * - Natural fit for infinite scroll UIs
 * 
 * @param pageSize - Number of products per page (default: 20)
 * 
 * @example
 * ```tsx
 * const { data, fetchNextPage, hasNextPage, isFetching } = useProductsCursor(20);
 * 
 * // Load more products
 * if (hasNextPage) {
 *   await fetchNextPage();
 * }
 * 
 * // All products across all fetched pages
 * const allProducts = data?.pages.flatMap(page => page.products) ?? [];
 * ```
 */
export function useProductsCursor(pageSize: number = 20) {
    const queryClient = useQueryClient();

    const query = useInfiniteQuery<CursorPaginatedProductsResponse, Error>({
        queryKey: ['products-cursor', pageSize],
        queryFn: async ({ pageParam }) => {
            const params = new URLSearchParams();
            params.set('first', String(pageSize));

            if (pageParam) {
                params.set('after', pageParam as string);
            }

            const response = await api.get(`/products?${params.toString()}`);
            return response.data;
        },
        initialPageParam: null as string | null,
        getNextPageParam: (lastPage) => {
            if (lastPage.page_info?.has_next_page && lastPage.page_info?.end_cursor) {
                return lastPage.page_info.end_cursor;
            }
            return undefined;
        },
        getPreviousPageParam: (firstPage) => {
            if (firstPage.page_info?.has_previous_page && firstPage.page_info?.start_cursor) {
                return firstPage.page_info.start_cursor;
            }
            return undefined;
        },
        staleTime: 5 * 60 * 1000, // 5 minutes
        gcTime: 10 * 60 * 1000, // 10 minutes
        refetchOnWindowFocus: false,
    });

    // Flatten all pages into a single array of products
    const allProducts = query.data?.pages.flatMap(page => page.products) ?? [];

    const createProduct = useMutation({
        mutationFn: async (data: Partial<Product>) => {
            if (!data.name || !data.unit_price || data.unit_price <= 0) {
                throw new Error('Invalid product data: name and positive unit_price are required');
            }
            const response = await api.post('/products', data);
            return response.data;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['products-cursor'] });
            queryClient.invalidateQueries({ queryKey: ['products'] });
        },
    });

    const updateProduct = useMutation({
        mutationFn: async ({ id, data }: { id: string; data: Partial<Product> }) => {
            if (!id) {
                throw new Error('Product ID is required for update');
            }
            const response = await api.put(`/products/${id}`, data);
            return response.data;
        },
        onSuccess: (_, variables) => {
            queryClient.invalidateQueries({ queryKey: ['products-cursor'] });
            queryClient.invalidateQueries({ queryKey: ['products'] });
            queryClient.invalidateQueries({ queryKey: ['product', variables.id] });
        },
    });

    const deleteProduct = useMutation({
        mutationFn: async (id: string) => {
            if (!id) {
                throw new Error('Product ID is required for deletion');
            }
            await api.delete(`/products/${id}`);
        },
        onSuccess: (_, id) => {
            queryClient.invalidateQueries({ queryKey: ['products-cursor'] });
            queryClient.invalidateQueries({ queryKey: ['products'] });
            queryClient.removeQueries({ queryKey: ['product', id] });
        },
    });

    return {
        // Data
        products: allProducts,
        pages: query.data?.pages ?? [],

        // Infinite scroll controls
        fetchNextPage: query.fetchNextPage,
        fetchPreviousPage: query.fetchPreviousPage,
        hasNextPage: query.hasNextPage,
        hasPreviousPage: query.hasPreviousPage,
        isFetchingNextPage: query.isFetchingNextPage,
        isFetchingPreviousPage: query.isFetchingPreviousPage,

        // Query state
        isLoading: query.isLoading,
        isFetching: query.isFetching,
        error: query.error,
        refetch: query.refetch,

        // Mutations
        createProduct,
        updateProduct,
        deleteProduct,
    };
}

/**
 * Utility function to build cursor pagination query params
 * @param params - Cursor pagination parameters
 * @returns URLSearchParams with cursor pagination params
 */
export function buildCursorParams(params: CursorPaginationParams): URLSearchParams {
    const searchParams = new URLSearchParams();

    if (params.first !== undefined) {
        searchParams.set('first', String(params.first));
    }
    if (params.after) {
        searchParams.set('after', params.after);
    }
    if (params.last !== undefined) {
        searchParams.set('last', String(params.last));
    }
    if (params.before) {
        searchParams.set('before', params.before);
    }

    return searchParams;
}
