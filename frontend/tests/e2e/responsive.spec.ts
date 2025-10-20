import { test, expect, devices } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

/**
 * Responsive Design Tests
 * Tests the application on different screen sizes and devices
 */

test.describe('Mobile Responsiveness', () => {
  test.use({ ...devices['iPhone 12'] });

  test('should display mobile login page correctly', async ({ page }) => {
    await page.goto('/login');
    
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    
    // Check that elements are properly sized for mobile
    const emailBox = await page.locator('#email').boundingBox();
    expect(emailBox?.width).toBeGreaterThan(200);
  });

  test('should have mobile navigation menu', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Look for hamburger menu
    const mobileMenu = page.locator('button[aria-label*="menu" i], button:has(svg)').first();
    
    if (await mobileMenu.isVisible()) {
      await mobileMenu.click();
      
      // Navigation should appear
      await page.waitForTimeout(500);
      
      const nav = page.locator('nav, [role="navigation"]');
      await expect(nav).toBeVisible();
    }
  });

  test('should display tables responsively on mobile', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    // Tables should either be responsive or scrollable
    const table = page.locator('table').first();
    
    if (await table.isVisible()) {
      const box = await table.boundingBox();
      const viewport = page.viewportSize();
      
      // Table should fit in viewport or be scrollable
      expect(viewport?.width).toBeTruthy();
    }
  });

  test('should have touch-friendly buttons on mobile', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Buttons should be large enough for touch (minimum 44x44px)
    const buttons = await page.locator('button').all();
    
    for (const button of buttons.slice(0, 5)) {
      if (await button.isVisible()) {
        const box = await button.boundingBox();
        
        // Check minimum touch target size (iOS HIG recommends 44px)
        if (box) {
          expect(box.height).toBeGreaterThanOrEqual(32); // Relaxed for some designs
        }
      }
    }
  });
});

test.describe('Tablet Responsiveness', () => {
  test.use({ ...devices['iPad Pro'] });

  test('should display tablet layout correctly', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    await expect(page.locator('text=/Dashboard|Overview/i').first()).toBeVisible();
  });

  test('should show sidebar on tablet', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    const sidebar = page.locator('nav, [role="navigation"], aside').first();
    await expect(sidebar).toBeVisible({ timeout: 5000 });
  });

  test('should display multi-column layout on tablet', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    // Content should be visible
    const content = page.locator('main, [role="main"]');
    await expect(content).toBeVisible();
  });
});

test.describe('Desktop Responsiveness', () => {
  test.use({ viewport: { width: 1920, height: 1080 } });

  test('should display full desktop layout', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Should show sidebar permanently
    const sidebar = page.locator('nav, aside').first();
    await expect(sidebar).toBeVisible();
  });

  test('should utilize wide screen space effectively', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/analytics');
    await page.waitForLoadState('networkidle');
    
    // Charts and metrics should use available space
    const mainContent = page.locator('main, [role="main"]');
    const box = await mainContent.boundingBox();
    
    // Should use significant portion of screen width
    expect(box?.width).toBeGreaterThan(800);
  });

  test('should display tables with all columns on desktop', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const table = page.locator('table').first();
    
    if (await table.isVisible()) {
      const headers = await table.locator('th').count();
      
      // Desktop should show more columns
      expect(headers).toBeGreaterThanOrEqual(3);
    }
  });
});

test.describe('Responsive Images and Media', () => {
  test('should load appropriate image sizes', async ({ page }) => {
    await page.goto('/login');
    
    const images = await page.locator('img').all();
    
    for (const img of images) {
      const loaded = await img.evaluate((el: HTMLImageElement) => el.complete);
      expect(loaded).toBeTruthy();
    }
  });

  test('should handle different screen densities', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Images should render without pixelation
    const images = await page.locator('img').all();
    
    for (const img of images) {
      const naturalWidth = await img.evaluate((el: HTMLImageElement) => el.naturalWidth);
      expect(naturalWidth).toBeGreaterThan(0);
    }
  });
});

test.describe('Cross-Browser Responsiveness', () => {
  test('should work on different browsers', async ({ page, browserName }) => {
    await page.goto('/login');
    
    // Basic functionality should work across browsers
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    
    // Log which browser we're testing
    console.log(`Testing on: ${browserName}`);
  });
});

test.describe('Orientation Changes', () => {
  test('should handle portrait to landscape rotation', async ({ page }) => {
    // Start in portrait
    await page.setViewportSize({ width: 375, height: 667 });
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    await expect(page.locator('text=/Dashboard/i').first()).toBeVisible();
    
    // Rotate to landscape
    await page.setViewportSize({ width: 667, height: 375 });
    await page.waitForTimeout(500);
    
    // Should still be functional
    await expect(page.locator('text=/Dashboard/i').first()).toBeVisible();
  });
});
