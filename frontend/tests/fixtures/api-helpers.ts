import { APIRequestContext, expect } from '@playwright/test';

/**
 * API helper functions for integration tests
 */

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function createAuthHeaders(token: string) {
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
}

export async function apiLogin(request: APIRequestContext, email: string, password: string) {
  const response = await request.post(`${API_BASE_URL}/api/v1/auth/login`, {
    data: { email, password },
  });
  
  expect(response.ok()).toBeTruthy();
  const data = await response.json();
  return data.data?.access_token || data.access_token || data.token;
}

export async function apiCreateProduct(
  request: APIRequestContext,
  token: string,
  productData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/products`, {
    headers: await createAuthHeaders(token),
    data: productData,
  });
  
  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiDeleteProduct(
  request: APIRequestContext,
  token: string,
  productId: string
) {
  const response = await request.delete(`${API_BASE_URL}/api/v1/products/${productId}`, {
    headers: await createAuthHeaders(token),
  });
  
  expect(response.ok()).toBeTruthy();
}

export async function apiCreateOrder(
  request: APIRequestContext,
  token: string,
  orderData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/orders`, {
    headers: await createAuthHeaders(token),
    data: orderData,
  });
  
  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiDeleteOrder(
  request: APIRequestContext,
  token: string,
  orderId: string
) {
  const response = await request.delete(`${API_BASE_URL}/api/v1/orders/${orderId}`, {
    headers: await createAuthHeaders(token),
  });

  expect(response.ok()).toBeTruthy();
}

export async function apiGetProducts(
  request: APIRequestContext,
  token: string,
  params?: { page?: number; limit?: number; search?: string }
) {
  const queryParams = new URLSearchParams();
  if (params?.page) queryParams.set('page', params.page.toString());
  if (params?.limit) queryParams.set('limit', params.limit.toString());
  if (params?.search) queryParams.set('search', params.search);
  
  const url = `${API_BASE_URL}/api/v1/products${queryParams.toString() ? '?' + queryParams.toString() : ''}`;
  
  const response = await request.get(url, {
    headers: await createAuthHeaders(token),
  });
  
  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiCreateWarehouse(
  request: APIRequestContext,
  token: string,
  warehouseData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/warehouses`, {
    headers: await createAuthHeaders(token),
    data: warehouseData,
  });

  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiDeleteWarehouse(
  request: APIRequestContext,
  token: string,
  warehouseId: string
) {
  const response = await request.delete(`${API_BASE_URL}/api/v1/warehouses/${warehouseId}`, {
    headers: await createAuthHeaders(token),
  });

  expect(response.ok()).toBeTruthy();
}

export async function apiCreateDistributor(
  request: APIRequestContext,
  token: string,
  distributorData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/distributors`, {
    headers: await createAuthHeaders(token),
    data: distributorData,
  });

  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiDeleteDistributor(
  request: APIRequestContext,
  token: string,
  distributorId: string
) {
  const response = await request.delete(`${API_BASE_URL}/api/v1/distributors/${distributorId}`, {
    headers: await createAuthHeaders(token),
  });

  expect(response.ok()).toBeTruthy();
}

export async function apiCreateInventory(
  request: APIRequestContext,
  token: string,
  inventoryData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/inventory`, {
    headers: await createAuthHeaders(token),
    data: inventoryData,
  });

  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiAdjustInventory(
  request: APIRequestContext,
  token: string,
  adjustmentData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/inventory/adjust`, {
    headers: await createAuthHeaders(token),
    data: adjustmentData,
  });

  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiListInventories(
  request: APIRequestContext,
  token: string,
  params?: { limit?: number; offset?: number }
) {
  const queryParams = new URLSearchParams();
  if (params?.limit) queryParams.set('limit', params.limit.toString());
  if (params?.offset) queryParams.set('offset', params.offset.toString());

  const response = await request.get(
    `${API_BASE_URL}/api/v1/inventory${queryParams.toString() ? '?' + queryParams.toString() : ''}`,
    {
      headers: await createAuthHeaders(token),
    }
  );

  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiUpdateInventory(
  request: APIRequestContext,
  token: string,
  inventoryData: any
) {
  const response = await request.post(`${API_BASE_URL}/api/v1/inventory/adjust`, {
    headers: await createAuthHeaders(token),
    data: inventoryData,
  });
  
  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function apiGetAnalytics(
  request: APIRequestContext,
  token: string
) {
  const response = await request.get(`${API_BASE_URL}/api/v1/analytics/dashboard`, {
    headers: await createAuthHeaders(token),
  });
  
  expect(response.ok()).toBeTruthy();
  return await response.json();
}

export async function cleanupTestData(
  request: APIRequestContext,
  token: string,
  productIds: string[] = []
) {
  // Delete test products
  for (const id of productIds) {
    try {
      await apiDeleteProduct(request, token, id);
    } catch (error) {
      console.warn(`Failed to cleanup product ${id}:`, error);
    }
  }
}
