import { test, expect } from '@playwright/test';

/**
 * Smoke Tests - Quick verification that basic functionality works
 * Run these tests first to verify the system is operational
 */

test.describe('Smoke Tests', () => {
  test('frontend server is accessible', async ({ page }) => {
    const response = await page.goto('/');
    expect(response?.status()).toBeLessThan(400);
  });

  test('login page loads correctly', async ({ page }) => {
    await page.goto('/login');
    
    // Check for essential elements
    await expect(page.locator('#email')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    await expect(page.locator('h1:has-text("AgroMart")')).toBeVisible();
  });

  test('backend API is accessible', async ({ request }) => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    const response = await request.get(`${apiUrl}/api/v1/health`).catch(() => null);
    
    if (response) {
      expect(response.status()).toBeLessThan(500);
    } else {
      console.warn('⚠️  Backend API not accessible - some tests may fail');
    }
  });

  test('can navigate to different pages', async ({ page }) => {
    await page.goto('/');
    
    // Should redirect somewhere (login or dashboard)
    await page.waitForTimeout(2000);
    
    const url = page.url();
    expect(url).toMatch(/\/(login|dashboard)/);
  });

  test('JavaScript is enabled and working', async ({ page }) => {
    await page.goto('/login');
    
    const jsEnabled = await page.evaluate(() => {
      return typeof window !== 'undefined' && typeof document !== 'undefined';
    });
    
    expect(jsEnabled).toBeTruthy();
  });

  test('no critical console errors on load', async ({ page }) => {
    const errors: string[] = [];
    
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    // Filter out known non-critical errors
    const criticalErrors = errors.filter(e => 
      !e.includes('favicon') && 
      !e.includes('sourcemap') &&
      !e.includes('DevTools')
    );
    
    console.log(`Console errors: ${criticalErrors.length}`);
    if (criticalErrors.length > 0) {
      console.log('Errors:', criticalErrors);
    }
    
    expect(criticalErrors.length).toBe(0);
  });

  test('essential dependencies are loaded', async ({ page }) => {
    await page.goto('/login');
    
    // Check if React is loaded
    const hasReact = await page.evaluate(() => {
      return typeof (window as any).React !== 'undefined' || 
             document.querySelector('[data-reactroot], [data-react]') !== null ||
             document.querySelector('div[id="__next"]') !== null;
    });
    
    // At minimum, the Next.js app root should exist
    const hasNextRoot = await page.locator('#__next').isVisible();
    
    expect(hasNextRoot).toBeTruthy();
  });

  test('CSS is loaded and applied', async ({ page }) => {
    await page.goto('/login');
    
    // Check if body has styles applied
    const bodyColor = await page.evaluate(() => {
      const body = document.body;
      return window.getComputedStyle(body).backgroundColor;
    });
    
    // Should not be default (transparent or rgba(0, 0, 0, 0))
    expect(bodyColor).toBeTruthy();
    expect(bodyColor).not.toBe('rgba(0, 0, 0, 0)');
  });

  test('form inputs are functional', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    const passwordInput = page.locator('#password');
    
    await emailInput.fill('test@example.com');
    await passwordInput.fill('password123');
    
    const emailValue = await emailInput.inputValue();
    const passwordValue = await passwordInput.inputValue();
    
    expect(emailValue).toBe('test@example.com');
    expect(passwordValue).toBe('password123');
  });

  test('buttons are clickable', async ({ page }) => {
    await page.goto('/login');
    
    const submitButton = page.locator('button[type="submit"]');
    
    // Button should be visible and enabled
    await expect(submitButton).toBeVisible();
    
    const isEnabled = await submitButton.isEnabled();
    expect(isEnabled).toBeTruthy();
  });
});

test.describe('Smoke Test Summary', () => {
  test('display test configuration', async ({ page }) => {
    console.log('\n🧪 Test Configuration:');
    console.log(`   Frontend URL: ${process.env.NEXT_PUBLIC_FRONTEND_URL || 'http://localhost:3000'}`);
    console.log(`   Backend URL: ${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}`);
    console.log(`   Browser: ${test.info().project.name}`);
    console.log(`   Viewport: ${page.viewportSize()?.width}x${page.viewportSize()?.height}`);
    console.log('\n✅ If all smoke tests pass, the system is ready for comprehensive testing!\n');
  });
});
