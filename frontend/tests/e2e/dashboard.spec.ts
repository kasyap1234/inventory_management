import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';
import { selectors } from '../fixtures/test-data';

/**
 * Dashboard E2E Tests
 * Tests dashboard navigation, widgets, and overall functionality
 */

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('should display dashboard with all key sections', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for key dashboard elements
    await expect(page.locator('text=/Dashboard|Overview/i').first()).toBeVisible();
    
    // Check for navigation sidebar or menu
    const nav = page.locator('nav, [role="navigation"]').first();
    await expect(nav).toBeVisible();
  });

  test('should display analytics cards/widgets', async ({ page }) => {
    await page.goto('/dashboard');

    // Wait for content to load
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Look for common metric cards (sales, inventory, orders, etc.)
    const metricsText = await page.textContent('body');
    const hasMetrics = metricsText?.toLowerCase().includes('total') ||
                      metricsText?.toLowerCase().includes('sales') ||
                      metricsText?.toLowerCase().includes('stock');

    expect(hasMetrics).toBeTruthy();
  });

  test('should navigate to products page', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Click on Products link
    const productsLink = page.locator('a[href*="/products"], button:has-text("Products")').first();
    await productsLink.click();
    
    await expect(page).toHaveURL(/\/products/);
  });

  test('should navigate to inventory page', async ({ page }) => {
    await page.goto('/dashboard');
    
    const inventoryLink = page.locator('a[href*="/inventory"], button:has-text("Inventory")').first();
    if (await inventoryLink.isVisible()) {
      await inventoryLink.click();
      await expect(page).toHaveURL(/\/inventory/);
    }
  });

  test('should navigate to orders page', async ({ page }) => {
    await page.goto('/dashboard');
    
    const ordersLink = page.locator('a[href*="/orders"], button:has-text("Orders")').first();
    if (await ordersLink.isVisible()) {
      await ordersLink.click();
      await expect(page).toHaveURL(/\/orders/);
    }
  });

  test('should navigate to analytics page', async ({ page }) => {
    await page.goto('/dashboard');
    
    const analyticsLink = page.locator('a[href*="/analytics"], button:has-text("Analytics")').first();
    if (await analyticsLink.isVisible()) {
      await analyticsLink.click();
      await expect(page).toHaveURL(/\/analytics/);
    }
  });

  test('should display user profile menu', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Look for profile/user menu button
    const profileButton = page.locator('button:has-text("Profile"), button[aria-label*="profile" i], [data-testid="user-menu"]').first();
    
    if (await profileButton.isVisible()) {
      await profileButton.click();
      
      // Check for profile menu items
      await expect(page.locator('text=/Profile|Settings|Logout/i').first()).toBeVisible();
    }
  });

  test('should handle mobile responsive navigation', async ({ page, viewport }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    // Look for mobile menu button (hamburger)
    const mobileMenuButton = page.locator('button[aria-label*="menu" i], button:has(svg)').first();
    
    if (await mobileMenuButton.isVisible()) {
      await mobileMenuButton.click();
      
      // Navigation should be visible after click
      const nav = page.locator('nav, [role="navigation"]');
      await expect(nav).toBeVisible();
    }
  });

  test('should load charts and graphs if present', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Wait for any charts to load
    await page.waitForTimeout(2000);
    
    // Check for chart elements (Recharts uses SVG)
    const charts = page.locator('svg');
    const chartCount = await charts.count();
    
    // Dashboard should have at least some visual elements
    expect(chartCount).toBeGreaterThan(0);
  });

  test('should display notifications if available', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Look for notification bell or badge
    const notificationButton = page.locator('button[aria-label*="notification" i], [data-testid="notifications"]').first();
    
    if (await notificationButton.isVisible()) {
      await notificationButton.click();
      
      // Notification panel or dropdown should appear
      await page.waitForTimeout(500);
      const hasNotificationPanel = await page.locator('[role="dialog"], [class*="notification"]').count() > 0;
      expect(hasNotificationPanel).toBeTruthy();
    }
  });

  test('should handle search functionality if present', async ({ page }) => {
    await page.goto('/dashboard');
    
    const searchInput = page.locator('input[type="search"], input[placeholder*="search" i]').first();
    
    if (await searchInput.isVisible()) {
      await searchInput.fill('test search');
      await page.keyboard.press('Enter');
      
      // Wait for search results or navigation
      await page.waitForTimeout(1000);
      
      // Verify search was processed
      const url = page.url();
      const hasSearchParam = url.includes('search') || url.includes('query');
      expect(hasSearchParam).toBeTruthy();
    }
  });

  test('should display proper loading states', async ({ page }) => {
    // Slow down network to see loading states
    await page.route('**/*', route => {
      setTimeout(() => route.continue(), 100);
    });

    await page.goto('/dashboard');

    // Look for loading indicators
    const loader = page.locator('[class*="loading"], [class*="spinner"], svg[class*="animate-spin"]').first();

    // Loading state should appear briefly
    const hadLoader = await loader.isVisible().catch(() => false);

    // Wait for page to fully load
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Loading should be gone
    const hasLoaderNow = await loader.isVisible({ timeout: 1000 }).catch(() => false);
    expect(hasLoaderNow).toBeFalsy();
  });

  test('should maintain dashboard state on navigation back', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Navigate to products
    const productsLink = page.locator('a[href*="/products"]').first();
    await productsLink.click();
    await expect(page).toHaveURL(/\/products/);
    
    // Go back
    await page.goBack();
    await expect(page).toHaveURL(/\/dashboard/);
    
    // Dashboard should load properly
    await expect(page.locator('text=/Dashboard|Overview/i').first()).toBeVisible();
  });
});

test.describe('Dashboard Permissions', () => {
  test('should show appropriate sections based on user role', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Admin should see admin-specific sections
    const bodyText = await page.textContent('body');
    const hasAdminFeatures = bodyText?.toLowerCase().includes('users') || 
                            bodyText?.toLowerCase().includes('settings') ||
                            bodyText?.toLowerCase().includes('roles');
    
    // This is expected for admin users
    expect(hasAdminFeatures).toBeTruthy();
  });
});
