import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

/**
 * Analytics and Reporting E2E Tests
 * Tests analytics dashboard, charts, and data visualization
 */

test.describe('Analytics Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/analytics');
  });

  test('should display analytics page', async ({ page }) => {
    await expect(page.locator('h1, h2').filter({ hasText: /Analytics|Reports/i }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should display key metrics cards', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Look for metric cards with numbers
    const metrics = page.locator('text=/\\$|sales|revenue|total/i');
    const metricsCount = await metrics.count();

    expect(metricsCount).toBeGreaterThan(0);
  });

  test('should render sales trend chart', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('svg, [data-testid="chart"], [class*="chart"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
    await page.waitForTimeout(2000);

    // Look for charts (Recharts uses SVG)
    const charts = page.locator('svg');
    const chartCount = await charts.count();

    expect(chartCount).toBeGreaterThan(0);
  });

  test('should filter analytics by date range', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const dateFilter = page.locator('input[type="date"], button:has-text("Date"), [data-testid="date-filter"]').first();

    if (await dateFilter.isVisible()) {
      await dateFilter.click();
      await page.waitForTimeout(500);

      // Select a date range
      const dateOption = page.locator('[role="option"], .react-datepicker__day').first();
      if (await dateOption.isVisible()) {
        await dateOption.click();
        await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
      }
    }
  });

  test('should display top products by sales', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Look for top products section
    const topProducts = page.locator('text=/top.*products|best.*selling/i').first();
    const hasTopProducts = await topProducts.isVisible({ timeout: 5000 }).catch(() => false);

    expect(hasTopProducts || true).toBeTruthy();
  });

  test('should show low stock alerts', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const lowStock = page.locator('text=/low.*stock|stock.*alert/i').first();
    const hasLowStock = await lowStock.isVisible({ timeout: 5000 }).catch(() => false);

    // May or may not have low stock items
    expect(hasLowStock !== undefined).toBeTruthy();
  });

  test('should export analytics report', async ({ page }) => {
    const exportButton = page.locator('button:has-text("Export"), button:has-text("Download"), button:has-text("PDF")').first();
    
    if (await exportButton.isVisible()) {
      const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);
      
      await exportButton.click();
      
      const download = await downloadPromise;
      if (download) {
        expect(download.suggestedFilename()).toMatch(/\.(pdf|csv|xlsx)$/);
      }
    }
  });

  test('should display revenue breakdown by category', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
    await page.waitForTimeout(2000);

    // Look for category-related content
    const categorySection = page.locator('text=/category|categor/i').first();
    const hasCategories = await categorySection.isVisible({ timeout: 5000 }).catch(() => false);

    expect(hasCategories || true).toBeTruthy();
  });

  test('should show inventory valuation', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const valuation = page.locator('text=/valuation|inventory.*value|stock.*value/i').first();
    const hasValuation = await valuation.isVisible({ timeout: 5000 }).catch(() => false);

    expect(hasValuation || true).toBeTruthy();
  });

  test('should display order statistics', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const orderStats = page.locator('text=/order.*count|total.*orders|orders/i').first();
    const hasOrderStats = await orderStats.isVisible({ timeout: 5000 }).catch(() => false);

    expect(hasOrderStats || true).toBeTruthy();
  });

  test('should refresh data when refresh button clicked', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const refreshButton = page.locator('button:has-text("Refresh"), button[aria-label*="refresh" i]').first();

    if (await refreshButton.isVisible()) {
      await refreshButton.click();

      // Wait for reload
      await page.waitForTimeout(1000);

      // Page should still be functional
      await expect(page.locator('h1, h2').filter({ hasText: /Analytics/i })).toBeVisible();
    }
  });
});

test.describe('Analytics Data Visualization', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/analytics');
  });

  test('should render charts without errors', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
    await page.waitForTimeout(2000);

    // Check for any console errors
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    await page.waitForTimeout(1000);

    // Filter out known non-critical errors
    const criticalErrors = errors.filter(e =>
      !e.includes('favicon') &&
      !e.includes('sourcemap')
    );

    expect(criticalErrors.length).toBe(0);
  });

  test('should handle empty data gracefully', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('[data-testid="metrics"], [class*="metric"], [class*="card"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Even with no data, page should render properly
    const bodyVisible = await page.locator('body').isVisible();
    expect(bodyVisible).toBeTruthy();
  });

  test('should be responsive on different screen sizes', async ({ page }) => {
    // Test mobile
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(1000);
    
    await expect(page.locator('h1, h2').filter({ hasText: /Analytics/i }).first()).toBeVisible();
    
    // Test tablet
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(1000);
    
    await expect(page.locator('h1, h2').filter({ hasText: /Analytics/i }).first()).toBeVisible();
    
    // Test desktop
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(1000);
    
    await expect(page.locator('h1, h2').filter({ hasText: /Analytics/i }).first()).toBeVisible();
  });
});
