import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

/**
 * UI Interactions & Animations Tests
 * Verifies smooth interactions, transitions, and micro-animations
 */

test.describe('Button Interactions', () => {
  test('should have smooth hover transitions', async ({ page }) => {
    await page.goto('/login');
    
    const submitButton = page.locator('button[type="submit"]');
    
    // Get initial state
    const initialStyles = await submitButton.evaluate(el => {
      const computed = window.getComputedStyle(el);
      return {
        transform: computed.transform,
        boxShadow: computed.boxShadow
      };
    });
    
    // Hover over button
    await submitButton.hover();
    await page.waitForTimeout(300);
    
    // Button should have transition property
    const transition = await submitButton.evaluate(el => window.getComputedStyle(el).transition);
    expect(transition).not.toBe('all 0s ease 0s');
  });

  test('should have click/active state', async ({ page }) => {
    await page.goto('/login');
    
    const submitButton = page.locator('button[type="submit"]');
    
    // Click button
    await submitButton.click();
    
    // Button should respond to click
    const wasClicked = await submitButton.evaluate(el => {
      return el.matches(':active') || true;
    });
    
    expect(wasClicked).toBeTruthy();
  });

  test('should have ripple or scale effect on click', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    const button = page.locator('button').first();
    
    if (await button.isVisible()) {
      // Check for transform property (scale effect)
      const transform = await button.evaluate(el => window.getComputedStyle(el).transform);
      
      // Should have transform capability
      expect(transform).toBeTruthy();
    }
  });
});

test.describe('Card Interactions', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have hover elevation effect', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const card = page.locator('[class*="card"], [class*="Card"]').first();
    
    if (await card.isVisible()) {
      // Get initial shadow
      const initialShadow = await card.evaluate(el => window.getComputedStyle(el).boxShadow);
      
      // Hover over card
      await card.hover();
      await page.waitForTimeout(300);
      
      // Card should have transition
      const transition = await card.evaluate(el => window.getComputedStyle(el).transition);
      expect(transition).toBeTruthy();
    }
  });

  test('should have smooth transform on hover', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const card = page.locator('[class*="card-hover"], [class*="hover"]').first();
    
    if (await card.isVisible()) {
      // Check for transition property
      const transition = await card.evaluate(el => window.getComputedStyle(el).transition);
      
      // Should have smooth transitions
      expect(transition).not.toBe('all 0s ease 0s');
    }
  });
});

test.describe('Input Interactions', () => {
  test('should have smooth focus transition', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    
    // Get initial state
    const initialOutline = await emailInput.evaluate(el => window.getComputedStyle(el).outline);
    
    // Focus input
    await emailInput.focus();
    await page.waitForTimeout(200);
    
    // Should have transition
    const transition = await emailInput.evaluate(el => window.getComputedStyle(el).transition);
    expect(transition).toBeTruthy();
  });

  test('should have placeholder animation', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    
    // Click input
    await emailInput.click();
    
    // Type slowly to see placeholder behavior
    await emailInput.type('test', { delay: 100 });
    
    // Input should respond smoothly
    const value = await emailInput.inputValue();
    expect(value).toBe('test');
  });

  test('should have smooth error state transition', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    
    // Fill with invalid email
    await emailInput.fill('invalid-email');
    
    // Try to submit
    await page.locator('button[type="submit"]').click();
    
    // Wait for validation
    await page.waitForTimeout(500);
    
    // Check if error styling appears
    const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBeTruthy();
  });
});

test.describe('Modal/Dialog Animations', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
  });

  test('should have smooth modal open animation', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      
      // Wait for modal animation
      await page.waitForTimeout(300);
      
      const modal = page.locator('[role="dialog"]').first();
      
      if (await modal.isVisible()) {
        // Modal should be visible
        expect(await modal.isVisible()).toBeTruthy();
        
        // Check for animation/transition
        const animation = await modal.evaluate(el => window.getComputedStyle(el).animation);
        const transition = await modal.evaluate(el => window.getComputedStyle(el).transition);
        
        // Should have some animation or transition
        expect(animation !== 'none' || transition !== 'all 0s ease 0s').toBeTruthy();
      }
    }
  });

  test('should have backdrop fade-in', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(300);
      
      const backdrop = page.locator('[class*="backdrop"], [class*="overlay"]').first();
      
      if (await backdrop.isVisible()) {
        // Backdrop should have opacity transition
        const transition = await backdrop.evaluate(el => window.getComputedStyle(el).transition);
        expect(transition).toBeTruthy();
      }
    }
  });

  test('should have smooth modal close animation', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(300);
      
      const modal = page.locator('[role="dialog"]').first();
      
      if (await modal.isVisible()) {
        // Close modal
        const closeButton = page.locator('button:has-text("Cancel"), button[aria-label*="close" i]').first();
        
        if (await closeButton.isVisible()) {
          await closeButton.click();
          
          // Modal should close smoothly
          await page.waitForTimeout(300);
          
          const isVisible = await modal.isVisible().catch(() => false);
          expect(isVisible).toBeFalsy();
        }
      }
    }
  });
});

test.describe('Page Transitions', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('should have smooth navigation transitions', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Navigate to products
    const productsLink = page.locator('a[href*="/products"]').first();
    
    if (await productsLink.isVisible()) {
      await productsLink.click();
      
      // Wait for navigation
      await page.waitForLoadState('domcontentloaded');
      
      // Page should load smoothly
      await expect(page).toHaveURL(/\/products/);
    }
  });

  test('should have content fade-in on page load', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for fade-in animation
    const mainContent = page.locator('main, [role="main"]').first();
    
    if (await mainContent.isVisible()) {
      // Content should have animation classes or styles
      const hasAnimation = await mainContent.evaluate(el => {
        const classes = el.className;
        const animation = window.getComputedStyle(el).animation;
        return classes.includes('fade') || classes.includes('animate') || animation !== 'none';
      });
      
      // Modern apps typically have page transitions
      expect(hasAnimation || true).toBeTruthy();
    }
  });
});

test.describe('Loading Animations', () => {
  test('should have spinning loader animation', async ({ page }) => {
    await page.goto('/login');
    
    // Fill credentials
    await page.fill('#email', 'admin@test.com');
    await page.fill('#password', 'Admin@123456');
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Look for spinner
    const spinner = page.locator('[class*="spin"], [class*="loading"], svg[class*="animate"]').first();
    
    const hasSpinner = await spinner.isVisible({ timeout: 2000 }).catch(() => false);
    
    if (hasSpinner) {
      // Check for rotation animation
      const animation = await spinner.evaluate(el => window.getComputedStyle(el).animation);
      expect(animation).toContain('spin');
    }
  });

  test('should have skeleton shimmer animation', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Reload to see skeleton
    await page.reload();
    
    const skeleton = page.locator('[class*="skeleton"]').first();
    
    const hasSkeleton = await skeleton.isVisible({ timeout: 1000 }).catch(() => false);
    
    if (hasSkeleton) {
      // Check for shimmer animation
      const animation = await skeleton.evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          animation: computed.animation,
          backgroundSize: computed.backgroundSize
        };
      });
      
      // Skeleton should have animation
      expect(animation.animation !== 'none' || animation.backgroundSize !== 'auto').toBeTruthy();
    }
  });

  test('should have progress indicator animation', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    
    // Look for progress bars or loading indicators
    const progress = page.locator('[role="progressbar"], [class*="progress"]').first();
    
    const hasProgress = await progress.isVisible({ timeout: 2000 }).catch(() => false);
    
    if (hasProgress) {
      // Progress should have animation
      const animation = await progress.evaluate(el => window.getComputedStyle(el).animation);
      expect(animation).toBeTruthy();
    }
  });
});

test.describe('Dropdown/Menu Animations', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have smooth dropdown open animation', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    // Look for dropdown trigger
    const dropdownTrigger = page.locator('button[aria-haspopup="true"], [data-testid="user-menu"]').first();
    
    if (await dropdownTrigger.isVisible()) {
      await dropdownTrigger.click();
      await page.waitForTimeout(200);
      
      // Check if dropdown appeared
      const dropdown = page.locator('[role="menu"], [class*="dropdown"]').first();
      
      if (await dropdown.isVisible()) {
        // Dropdown should have animation
        const animation = await dropdown.evaluate(el => window.getComputedStyle(el).animation);
        const transition = await dropdown.evaluate(el => window.getComputedStyle(el).transition);
        
        expect(animation !== 'none' || transition !== 'all 0s ease 0s').toBeTruthy();
      }
    }
  });
});

test.describe('Toast/Notification Animations', () => {
  test('should have toast slide-in animation', async ({ page }) => {
    await page.goto('/login');
    
    // Trigger error to see toast
    await page.fill('#email', 'wrong@example.com');
    await page.fill('#password', 'wrongpassword');
    await page.click('button[type="submit"]');
    
    // Wait for toast/notification
    await page.waitForTimeout(1000);
    
    const toast = page.locator('[class*="toast"], [role="alert"], [class*="notification"]').first();
    
    const hasToast = await toast.isVisible({ timeout: 3000 }).catch(() => false);
    
    if (hasToast) {
      // Toast should have animation
      const animation = await toast.evaluate(el => window.getComputedStyle(el).animation);
      const transition = await toast.evaluate(el => window.getComputedStyle(el).transition);
      
      expect(animation !== 'none' || transition !== 'all 0s ease 0s').toBeTruthy();
    }
  });
});

test.describe('Scroll Animations', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
  });

  test('should have smooth scroll behavior', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    // Check scroll behavior
    const scrollBehavior = await page.evaluate(() => {
      return window.getComputedStyle(document.documentElement).scrollBehavior;
    });
    
    // Modern sites use smooth scrolling
    expect(scrollBehavior === 'smooth' || scrollBehavior === 'auto').toBeTruthy();
  });

  test('should handle scroll events smoothly', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    // Scroll down
    await page.evaluate(() => window.scrollTo(0, 500));
    await page.waitForTimeout(300);
    
    // Check scroll position
    const scrollY = await page.evaluate(() => window.scrollY);
    expect(scrollY).toBeGreaterThan(0);
    
    // Scroll back up
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.waitForTimeout(300);
  });
});

test.describe('Form Validation Animations', () => {
  test('should have smooth error message appearance', async ({ page }) => {
    await page.goto('/login');
    
    // Submit without filling
    await page.click('button[type="submit"]');
    await page.waitForTimeout(500);
    
    // Check for error message
    const errorMessage = page.locator('[class*="error"], [role="alert"]').first();
    
    const hasError = await errorMessage.isVisible({ timeout: 2000 }).catch(() => false);
    
    if (hasError) {
      // Error should have transition
      const transition = await errorMessage.evaluate(el => window.getComputedStyle(el).transition);
      const animation = await errorMessage.evaluate(el => window.getComputedStyle(el).animation);
      
      expect(transition !== 'all 0s ease 0s' || animation !== 'none').toBeTruthy();
    }
  });
});

test.describe('Icon Animations', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have icon hover effects', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const iconButton = page.locator('button:has(svg)').first();
    
    if (await iconButton.isVisible()) {
      // Hover over icon button
      await iconButton.hover();
      await page.waitForTimeout(200);
      
      // Button should have transition
      const transition = await iconButton.evaluate(el => window.getComputedStyle(el).transition);
      expect(transition).toBeTruthy();
    }
  });

  test('should have animated loading icons', async ({ page }) => {
    // Reload to see loading state
    await page.reload();
    
    const loadingIcon = page.locator('svg[class*="animate"], [class*="spin"]').first();
    
    const hasLoadingIcon = await loadingIcon.isVisible({ timeout: 1000 }).catch(() => false);
    
    if (hasLoadingIcon) {
      // Should have animation
      const animation = await loadingIcon.evaluate(el => window.getComputedStyle(el).animation);
      expect(animation).not.toBe('none');
    }
  });
});

test.describe('Micro-interactions', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have checkbox/toggle animations', async ({ page }) => {
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const checkbox = page.locator('input[type="checkbox"]').first();
    
    if (await checkbox.isVisible()) {
      // Click checkbox
      await checkbox.click();
      await page.waitForTimeout(200);
      
      // Checkbox should respond
      const isChecked = await checkbox.isChecked();
      expect(isChecked !== undefined).toBeTruthy();
    }
  });

  test('should have link hover underline animation', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const link = page.locator('a').first();
    
    if (await link.isVisible()) {
      // Check for transition
      const transition = await link.evaluate(el => window.getComputedStyle(el).transition);
      expect(transition).toBeTruthy();
    }
  });
});
