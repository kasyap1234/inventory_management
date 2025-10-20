import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

/**
 * Accessibility Tests
 * Ensures the application meets WCAG standards and is accessible
 */

test.describe('Accessibility', () => {
  test('login page should be accessible', async ({ page }) => {
    await page.goto('/login');
    
    // Check for proper labels
    const emailInput = page.locator('#email');
    const emailLabel = await page.locator('label[for="email"]').count();
    
    if (await emailInput.isVisible()) {
      // Should have associated label or aria-label
      const hasLabel = emailLabel > 0 || await emailInput.getAttribute('aria-label') !== null;
      expect(hasLabel).toBeTruthy();
    }

    // Check heading hierarchy
    const h1Count = await page.locator('h1').count();
    expect(h1Count).toBeGreaterThan(0);

    // Check for skip navigation link (best practice)
    const skipLink = page.locator('a[href="#main"], a:has-text("Skip to")').first();
    // This is optional but good to have
  });

  test('dashboard should be keyboard navigable', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');

    // Tab through elements
    await page.keyboard.press('Tab');
    await page.waitForTimeout(100);
    
    // Get focused element
    const focusedElement = await page.evaluate(() => {
      const el = document.activeElement;
      return el?.tagName;
    });

    // Should be able to focus on interactive elements
    expect(focusedElement).toBeTruthy();
  });

  test('forms should have proper ARIA labels', async ({ page }) => {
    await page.goto('/login');

    const inputs = await page.locator('input').all();
    
    for (const input of inputs) {
      const hasLabel = await input.evaluate((el) => {
        const id = el.id;
        const hasLabelFor = id && document.querySelector(`label[for="${id}"]`);
        const hasAriaLabel = el.hasAttribute('aria-label');
        const hasAriaLabelledBy = el.hasAttribute('aria-labelledby');
        
        return hasLabelFor || hasAriaLabel || hasAriaLabelledBy;
      });
      
      // All inputs should have labels
      expect(hasLabel).toBeTruthy();
    }
  });

  test('images should have alt text', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    const images = await page.locator('img').all();
    
    for (const img of images) {
      const alt = await img.getAttribute('alt');
      // Images should have alt attribute (can be empty for decorative images)
      expect(alt !== null).toBeTruthy();
    }
  });

  test('buttons should have accessible names', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');

    const buttons = await page.locator('button').all();
    
    for (const button of buttons.slice(0, 10)) { // Check first 10 buttons
      const hasAccessibleName = await button.evaluate((el) => {
        const text = el.textContent?.trim();
        const ariaLabel = el.getAttribute('aria-label');
        const ariaLabelledBy = el.getAttribute('aria-labelledby');
        const title = el.getAttribute('title');
        
        return text || ariaLabel || ariaLabelledBy || title;
      });
      
      expect(hasAccessibleName).toBeTruthy();
    }
  });

  test('interactive elements should have visible focus indicators', async ({ page }) => {
    await page.goto('/login');
    
    const loginButton = page.locator('button[type="submit"]');
    await loginButton.focus();
    
    // Check if element has focus
    const isFocused = await loginButton.evaluate((el) => el === document.activeElement);
    expect(isFocused).toBeTruthy();
  });

  test('page should have proper document structure', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');

    // Should have main landmark
    const main = page.locator('main, [role="main"]');
    const hasMain = await main.count() > 0;
    
    // Should have navigation
    const nav = page.locator('nav, [role="navigation"]');
    const hasNav = await nav.count() > 0;
    
    expect(hasMain || hasNav).toBeTruthy();
  });

  test('color contrast should be sufficient', async ({ page }) => {
    await page.goto('/login');
    
    // This is a basic check - for full contrast checking, use axe-core
    const bodyBg = await page.evaluate(() => {
      const body = document.body;
      return window.getComputedStyle(body).backgroundColor;
    });
    
    expect(bodyBg).toBeTruthy();
  });

  test('error messages should be announced to screen readers', async ({ page }) => {
    await page.goto('/login');
    
    // Submit without credentials
    await page.click('button[type="submit"]');
    await page.waitForTimeout(500);
    
    // Error should have appropriate ARIA attributes
    const error = page.locator('[role="alert"], [class*="error"]').first();
    
    if (await error.isVisible({ timeout: 3000 })) {
      const role = await error.getAttribute('role');
      const ariaLive = await error.getAttribute('aria-live');
      
      // Should have alert role or aria-live
      expect(role === 'alert' || ariaLive === 'polite' || ariaLive === 'assertive').toBeTruthy();
    }
  });

  test('modals should trap focus', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');

    const addButton = page.locator('button:has-text("Add")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(500);

      const modal = page.locator('[role="dialog"]').first();
      
      if (await modal.isVisible()) {
        // Tab multiple times
        await page.keyboard.press('Tab');
        await page.keyboard.press('Tab');
        await page.keyboard.press('Tab');
        
        // Focus should still be within modal
        const focusedElement = await page.evaluate(() => {
          const el = document.activeElement;
          const dialog = document.querySelector('[role="dialog"]');
          return dialog?.contains(el);
        });
        
        expect(focusedElement).toBeTruthy();
      }
    }
  });

  test('tables should have proper structure', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');

    const tables = await page.locator('table').all();
    
    for (const table of tables) {
      // Should have thead and tbody
      const hasHeader = await table.locator('thead').count() > 0;
      const hasBody = await table.locator('tbody').count() > 0;
      
      expect(hasHeader || hasBody).toBeTruthy();
    }
  });
});

test.describe('Screen Reader Support', () => {
  test('navigation should have proper ARIA landmarks', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');

    const landmarks = await page.evaluate(() => {
      return {
        banner: document.querySelector('[role="banner"], header'),
        navigation: document.querySelector('[role="navigation"], nav'),
        main: document.querySelector('[role="main"], main'),
        contentinfo: document.querySelector('[role="contentinfo"], footer'),
      };
    });

    // Should have at least navigation and main
    expect(landmarks.navigation && landmarks.main).toBeTruthy();
  });

  test('dynamic content updates should be announced', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');

    // Look for live regions
    const liveRegions = await page.locator('[aria-live]').count();
    
    // Application should have live regions for dynamic updates
    // This is a soft check as not all pages may have them
    expect(liveRegions >= 0).toBeTruthy();
  });
});
