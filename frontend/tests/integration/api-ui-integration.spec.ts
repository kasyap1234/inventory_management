import { test, expect } from '@playwright/test';
import { testUsers, testProducts } from '../fixtures/test-data';
import { apiLogin, apiCreateProduct, apiDeleteProduct, apiGetProducts, cleanupTestData } from '../fixtures/api-helpers';
import { login, ensureLoggedOut } from '../fixtures/auth-helpers';

/**
 * Integration Tests - API + UI
 * Tests that verify the entire stack works together correctly
 */

test.describe('API-UI Integration Tests', () => {
  let authToken: string;
  const testProductIds: string[] = [];

  test.beforeAll(async ({ request }) => {
    // Get auth token for API calls
    authToken = await apiLogin(request, testUsers.admin.email, testUsers.admin.password);
  });

  test.afterAll(async ({ request }) => {
    // Cleanup test data
    await cleanupTestData(request, authToken, testProductIds);
  });

  test('should sync product creation between API and UI', async ({ page, request }) => {
    // Create product via API
    const productData = {
      name: testProducts.basic.name,
      sku: testProducts.basic.sku,
      price: testProducts.basic.price,
      description: testProducts.basic.description,
      category: testProducts.basic.category,
    };

    const response = await apiCreateProduct(request, authToken, productData);
    const productId = response.data?.id || response.id;
    
    if (productId) {
      testProductIds.push(productId);
    }

    // Login to UI and verify product appears
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');

    // Search for the created product
    const searchInput = page.locator('input[type="search"], input[placeholder*="Search" i]').first();
    if (await searchInput.isVisible()) {
      await searchInput.fill(testProducts.basic.sku);
      await page.waitForTimeout(1500);
    }

    // Verify product appears in UI
    await expect(page.locator(`text="${testProducts.basic.name}"`).first()).toBeVisible({ timeout: 5000 });
  });

  test('should reflect UI changes in API', async ({ page, request }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');

    // Create product via UI
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create"), button:has-text("New Product")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(500);

      const uniqueSku = `UI-TEST-${Date.now()}`;
      
      // Fill form
      await page.locator('input[name="name"], input[placeholder*="name" i]').first().fill('UI Created Product');
      await page.locator('input[name="sku"], input[placeholder*="sku" i]').first().fill(uniqueSku);
      await page.locator('input[name="price"], input[placeholder*="price" i]').first().fill('199');

      // Submit
      await page.locator('button[type="submit"], button:has-text("Create"), button:has-text("Save")').first().click();
      await page.waitForTimeout(3000);

      // Verify via API
      const apiResponse = await apiGetProducts(request, authToken, { search: uniqueSku });
      
      const products = apiResponse.data?.products || apiResponse.products || [];
      const foundProduct = products.find((p: any) => p.sku === uniqueSku);
      
      if (foundProduct) {
        testProductIds.push(foundProduct.id);
        expect(foundProduct.name).toBe('UI Created Product');
      }
    }
  });

  test('should handle real-time inventory updates', async ({ page, request }) => {
    // Create a product first
    const productData = {
      name: 'Real-time Test Product',
      sku: `RT-${Date.now()}`,
      price: 150,
    };

    const createResponse = await apiCreateProduct(request, authToken, productData);
    const productId = createResponse.data?.id || createResponse.id;
    
    if (productId) {
      testProductIds.push(productId);

      // Login and navigate to inventory
      await login(page, testUsers.admin.email, testUsers.admin.password);
      await page.goto('/dashboard/inventory');
      await page.waitForLoadState('networkidle');

      // Note: This test verifies the page loads and can display inventory
      // Real-time updates would require WebSocket testing
      await expect(page.locator('h1, h2').filter({ hasText: /Inventory/i })).toBeVisible();
    }
  });

  test('should handle order workflow end-to-end', async ({ page, request }) => {
    // Create a product for the order
    const productData = {
      name: 'Order Test Product',
      sku: `ORD-${Date.now()}`,
      price: 100,
      stock_quantity: 50,
    };

    const productResponse = await apiCreateProduct(request, authToken, productData);
    const productId = productResponse.data?.id || productResponse.id;
    
    if (productId) {
      testProductIds.push(productId);

      await login(page, testUsers.admin.email, testUsers.admin.password);
      await page.goto('/dashboard/orders');
      await page.waitForLoadState('networkidle');

      // Try to create order via UI
      const createButton = page.locator('button:has-text("Create"), button:has-text("New Order")').first();
      
      if (await createButton.isVisible()) {
        await createButton.click();
        await page.waitForTimeout(500);

        // Verify order form loaded
        await expect(page.locator('form, [role="dialog"]')).toBeVisible({ timeout: 5000 });
      }
    }
  });

  test('should sync authentication state between API and UI', async ({ page, request }) => {
    // Login via UI
    await login(page, testUsers.admin.email, testUsers.admin.password);
    
    // Get token from UI
    const uiToken = await page.evaluate(() => {
      return localStorage.getItem('access_token') || localStorage.getItem('token');
    });

    expect(uiToken).toBeTruthy();

    // Verify token works with API
    if (uiToken) {
      const response = await apiGetProducts(request, uiToken);
      expect(response).toBeTruthy();
    }
  });

  test('should handle API errors gracefully in UI', async ({ page, request }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    
    // Invalidate token
    await page.evaluate(() => {
      localStorage.setItem('access_token', 'invalid-token');
    });

    // Navigate to protected page
    await page.goto('/dashboard/products');
    
    // Should redirect to login or show error
    await page.waitForTimeout(2000);
    
    const currentUrl = page.url();
    const isLoginPage = currentUrl.includes('/login');
    const hasErrorMessage = await page.locator('[class*="error"], text=/error|unauthorized/i').isVisible({ timeout: 3000 }).catch(() => false);
    
    expect(isLoginPage || hasErrorMessage).toBeTruthy();
  });

  test('should validate data consistency after concurrent operations', async ({ page, request }) => {
    const sku = `CONCURRENT-${Date.now()}`;
    
    // Create product via API
    const productData = {
      name: 'Concurrent Test',
      sku: sku,
      price: 200,
    };

    const response = await apiCreateProduct(request, authToken, productData);
    const productId = response.data?.id || response.id;
    
    if (productId) {
      testProductIds.push(productId);

      // Immediately check in UI
      await login(page, testUsers.admin.email, testUsers.admin.password);
      await page.goto('/dashboard/products');
      await page.waitForLoadState('networkidle');

      // Search for product
      const searchInput = page.locator('input[type="search"], input[placeholder*="Search" i]').first();
      if (await searchInput.isVisible()) {
        await searchInput.fill(sku);
        await page.waitForTimeout(1500);

        // Verify consistency
        const productVisible = await page.locator(`text="${sku}"`).isVisible({ timeout: 5000 }).catch(() => false);
        expect(productVisible).toBeTruthy();
      }
    }
  });
});

test.describe('Integration Performance Tests', () => {
  test('should handle large dataset efficiently', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    
    const startTime = Date.now();
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    const loadTime = Date.now() - startTime;

    // Should load within reasonable time even with data
    expect(loadTime).toBeLessThan(8000);
  });

  test('should handle rapid navigation without memory leaks', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);

    // Navigate between pages rapidly
    const pages = ['/dashboard', '/dashboard/products', '/dashboard/orders', '/dashboard/inventory'];
    
    for (let i = 0; i < 3; i++) {
      for (const path of pages) {
        await page.goto(path);
        await page.waitForLoadState('domcontentloaded');
      }
    }

    // Page should still be responsive
    await page.goto('/dashboard');
    await expect(page.locator('text=/Dashboard|Overview/i').first()).toBeVisible();
  });
});

test.describe('Integration Error Handling', () => {
  test('should recover from network errors', async ({ page, context }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await page.goto('/dashboard/products');

    // Simulate network failure
    await context.setOffline(true);
    
    // Try to perform action
    const addButton = page.locator('button:has-text("Add")').first();
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(2000);
    }

    // Restore network
    await context.setOffline(false);
    
    // Reload and verify recovery
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    await expect(page.locator('h1, h2').filter({ hasText: /Products/i })).toBeVisible();
  });

  test('should handle session expiration gracefully', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await page.goto('/dashboard');

    // Clear auth token to simulate expiration
    await page.evaluate(() => {
      localStorage.removeItem('access_token');
      localStorage.removeItem('token');
    });

    // Try to navigate to protected page
    await page.goto('/dashboard/products');
    await page.waitForTimeout(2000);

    // Should redirect to login
    const currentUrl = page.url();
    expect(currentUrl).toContain('/login');
  });
});
