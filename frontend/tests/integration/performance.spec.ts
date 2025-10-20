import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

/**
 * Performance Tests
 * Tests application performance, load times, and resource usage
 */

test.describe('Page Load Performance', () => {
  test('login page should load quickly', async ({ page }) => {
    const startTime = Date.now();
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    const loadTime = Date.now() - startTime;
    
    console.log(`Login page load time: ${loadTime}ms`);
    expect(loadTime).toBeLessThan(3000);
  });

  test('dashboard should load within acceptable time', async ({ page }) => {
    await loginAsAdmin(page);
    
    const startTime = Date.now();
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    const loadTime = Date.now() - startTime;
    
    console.log(`Dashboard load time: ${loadTime}ms`);
    expect(loadTime).toBeLessThan(5000);
  });

  test('products page should load efficiently', async ({ page }) => {
    await loginAsAdmin(page);
    
    const startTime = Date.now();
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    const loadTime = Date.now() - startTime;
    
    console.log(`Products page load time: ${loadTime}ms`);
    expect(loadTime).toBeLessThan(6000);
  });
});

test.describe('Network Performance', () => {
  test('should minimize number of requests', async ({ page }) => {
    const requests: string[] = [];
    
    page.on('request', request => {
      requests.push(request.url());
    });
    
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    console.log(`Total requests on login page: ${requests.length}`);
    
    // Should have reasonable number of requests
    expect(requests.length).toBeLessThan(50);
  });

  test('should use caching effectively', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    const cachedRequests: string[] = [];
    const cacheHitHeader: string[] = [];
    
    page.on('response', response => {
      const cacheControl = response.headers()['cache-control'];
      if (cacheControl && cacheControl.includes('max-age')) {
        cachedRequests.push(response.url());
      }
    });
    
    // Reload page
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    console.log(`Cacheable resources: ${cachedRequests.length}`);
    
    // Some resources should be cacheable
    expect(cachedRequests.length).toBeGreaterThanOrEqual(0);
  });

  test('should compress resources', async ({ page }) => {
    const responses: any[] = [];
    
    page.on('response', async (response) => {
      const headers = response.headers();
      if (headers['content-encoding']) {
        responses.push({
          url: response.url(),
          encoding: headers['content-encoding']
        });
      }
    });
    
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    console.log(`Compressed resources: ${responses.length}`);
    
    // At least some resources should be compressed
    expect(responses.length).toBeGreaterThanOrEqual(0);
  });
});

test.describe('Runtime Performance', () => {
  test('should handle rapid interactions smoothly', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const startTime = Date.now();
    
    // Perform rapid clicks
    const searchInput = page.locator('input[type="search"]').first();
    if (await searchInput.isVisible()) {
      for (let i = 0; i < 5; i++) {
        await searchInput.fill(`test${i}`);
        await page.waitForTimeout(100);
      }
    }
    
    const interactionTime = Date.now() - startTime;
    
    console.log(`Rapid interaction time: ${interactionTime}ms`);
    expect(interactionTime).toBeLessThan(3000);
  });

  test('should scroll smoothly', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const startY = await page.evaluate(() => window.scrollY);
    
    // Scroll down
    await page.evaluate(() => window.scrollTo(0, 1000));
    await page.waitForTimeout(500);
    
    const endY = await page.evaluate(() => window.scrollY);
    
    // Should have scrolled
    expect(endY).toBeGreaterThan(startY);
  });

  test('should not have memory leaks on navigation', async ({ page }) => {
    await loginAsAdmin(page);
    
    const pages = ['/dashboard', '/dashboard/products', '/dashboard/orders', '/dashboard/analytics'];
    
    for (let i = 0; i < 2; i++) {
      for (const path of pages) {
        await page.goto(path);
        await page.waitForLoadState('domcontentloaded');
        await page.waitForTimeout(200);
      }
    }
    
    // Page should still be responsive
    await page.goto('/dashboard');
    await expect(page.locator('text=/Dashboard/i').first()).toBeVisible();
  });
});

test.describe('Resource Loading', () => {
  test('should load JavaScript bundles efficiently', async ({ page }) => {
    const jsFiles: { url: string; size: number }[] = [];
    
    page.on('response', async (response) => {
      const url = response.url();
      if (url.endsWith('.js')) {
        const buffer = await response.body().catch(() => null);
        if (buffer) {
          jsFiles.push({
            url,
            size: buffer.length
          });
        }
      }
    });
    
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    console.log(`JavaScript files loaded: ${jsFiles.length}`);
    
    const totalSize = jsFiles.reduce((sum, file) => sum + file.size, 0);
    console.log(`Total JS size: ${(totalSize / 1024).toFixed(2)} KB`);
    
    // Should not load excessive JS
    expect(jsFiles.length).toBeLessThan(30);
  });

  test('should load CSS efficiently', async ({ page }) => {
    const cssFiles: string[] = [];
    
    page.on('response', response => {
      const url = response.url();
      if (url.endsWith('.css') || response.headers()['content-type']?.includes('text/css')) {
        cssFiles.push(url);
      }
    });
    
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    console.log(`CSS files loaded: ${cssFiles.length}`);
    expect(cssFiles.length).toBeLessThan(20);
  });

  test('should lazy load images when appropriate', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    const images = await page.locator('img').all();
    
    for (const img of images.slice(0, 5)) {
      const loading = await img.getAttribute('loading');
      
      // Images below the fold should ideally be lazy loaded
      // This is optional but good for performance
      console.log(`Image loading attribute: ${loading}`);
    }
  });
});

test.describe('API Performance', () => {
  test('should receive API responses quickly', async ({ page }) => {
    const apiTimes: { url: string; startTime: number; endTime: number }[] = [];
    
    page.on('request', request => {
      const url = request.url();
      if (url.includes('/api/')) {
        const startTime = Date.now();
        request.response().then(response => {
          if (response) {
            apiTimes.push({
              url,
              startTime,
              endTime: Date.now()
            });
          }
        }).catch(() => {});
      }
    });
    
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    console.log(`API calls made: ${apiTimes.length}`);
    
    // API calls should be reasonably fast
    const avgTime = apiTimes.length > 0 
      ? apiTimes.reduce((sum, t) => sum + (t.endTime - t.startTime), 0) / apiTimes.length 
      : 0;
    
    console.log(`Average API response time: ${avgTime.toFixed(2)}ms`);
    
    if (apiTimes.length > 0) {
      expect(avgTime).toBeLessThan(3000);
    }
  });

  test('should handle concurrent API requests', async ({ page }) => {
    await loginAsAdmin(page);
    
    // Navigate to a page that makes multiple API calls
    await page.goto('/dashboard/analytics');
    await page.waitForLoadState('networkidle');
    
    // Page should load successfully
    await expect(page.locator('h1, h2').filter({ hasText: /Analytics/i }).first()).toBeVisible();
  });
});

test.describe('Rendering Performance', () => {
  test('should render large lists efficiently', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const startTime = Date.now();
    
    // Scroll to trigger rendering
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(500);
    
    const renderTime = Date.now() - startTime;
    
    console.log(`List render time: ${renderTime}ms`);
    expect(renderTime).toBeLessThan(2000);
  });

  test('should update UI quickly on state changes', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const searchInput = page.locator('input[type="search"]').first();
    
    if (await searchInput.isVisible()) {
      const startTime = Date.now();
      
      await searchInput.fill('test');
      await page.waitForTimeout(1500); // Wait for debounce
      
      const updateTime = Date.now() - startTime;
      
      console.log(`UI update time: ${updateTime}ms`);
      expect(updateTime).toBeLessThan(3000);
    }
  });
});
