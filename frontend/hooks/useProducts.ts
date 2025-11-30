import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import api from '@/lib/api';
import { Product } from '@/types';

export interface PaginationParams {
  page?: number;
  pageSize?: number;
}

export interface PaginatedProductsResponse {
  products: Product[];
  total?: number;
  page?: number;
  pageSize?: number;
  totalPages?: number;
}

export function useProducts(params?: PaginationParams) {
  const queryClient = useQueryClient();

  const page = params?.page ?? 1;
  const pageSize = params?.pageSize ?? 20;

  const { data: products, isLoading, error, refetch } = useQuery<PaginatedProductsResponse>({
    queryKey: ['products', { page, pageSize }],
    queryFn: async () => {
      const response = await api.get(`/products?page=${page}&limit=${pageSize}`);
      return response.data;
    },
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
  });

  // Prefetch next page for smoother pagination experience
  useEffect(() => {
    const totalPages = products?.totalPages ?? 0;
    const hasNextPage = page < totalPages;

    if (hasNextPage) {
      const nextPage = page + 1;
      queryClient.prefetchQuery({
        queryKey: ['products', { page: nextPage, pageSize }],
        queryFn: async () => {
          const response = await api.get(`/products?page=${nextPage}&limit=${pageSize}`);
          return response.data;
        },
        staleTime: 5 * 60 * 1000,
      });
    }
  }, [page, pageSize, products?.totalPages, queryClient]);

  const createProduct = useMutation({
    mutationFn: async (data: Partial<Product>) => {
      if (!data.name || !data.unit_price || data.unit_price <= 0) {
        throw new Error('Invalid product data: name and positive unit_price are required');
      }
      const response = await api.post('/products', data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
    onError: (error: any) => {
      console.error('Failed to create product:', error);
    },
    retry: 2,
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
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['product', variables.id] });
    },
    onError: (error: any) => {
      console.error('Failed to update product:', error);
    },
    retry: 2,
  });

  const deleteProduct = useMutation({
    mutationFn: async (id: string) => {
      if (!id) {
        throw new Error('Product ID is required for deletion');
      }
      await api.delete(`/products/${id}`);
    },
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.removeQueries({ queryKey: ['product', id] });
    },
    onError: (error: any) => {
      console.error('Failed to delete product:', error);
    },
    retry: 1,
  });

  const uploadProductImage = useMutation({
    mutationFn: async ({ productId, file }: { productId: string; file: File }) => {
      if (!productId || !file) {
        throw new Error('Product ID and file are required');
      }

      // Validate file type
      const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
      if (!validTypes.includes(file.type)) {
        throw new Error('Invalid file type. Only JPEG, PNG, GIF, and WebP images are allowed');
      }

      // Validate file size (max 5MB)
      const maxSize = 5 * 1024 * 1024;
      if (file.size > maxSize) {
        throw new Error('File size exceeds 5MB limit');
      }

      const formData = new FormData();
      formData.append('image', file);

      const response = await api.post(`/products/${productId}/images`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      queryClient.invalidateQueries({ queryKey: ['product', variables.productId] });
    },
    onError: (error: any) => {
      console.error('Failed to upload product image:', error);
    },
    retry: 1,
  });

  return {
    products: products?.products || [],
    isLoading,
    error,
    refetch,
    createProduct,
    updateProduct,
    deleteProduct,
    uploadProductImage,
  };
}
