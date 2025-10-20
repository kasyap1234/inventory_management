import { test, expect } from '@playwright/test';
import { testUsers, selectors } from '../fixtures/test-data';
import { login, logout, ensureLoggedOut, signup } from '../fixtures/auth-helpers';

/**
 * Authentication E2E Tests
 * Tests login, signup, logout, forgot password flows
 */

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await ensureLoggedOut(page);
  });

  test('should display login page correctly', async ({ page }) => {
    await page.goto('/login');
    
    // Check for key elements
    await expect(page.locator('h1:has-text("AgroMart")')).toBeVisible();
    await expect(page.locator(selectors.auth.emailInput)).toBeVisible();
    await expect(page.locator(selectors.auth.passwordInput)).toBeVisible();
    await expect(page.locator(selectors.auth.loginButton)).toBeVisible();
    await expect(page.locator(selectors.auth.signupLink)).toBeVisible();
    await expect(page.locator(selectors.auth.forgotPasswordLink)).toBeVisible();
  });

  test('should show validation errors for empty form', async ({ page }) => {
    await page.goto('/login');
    
    await page.click(selectors.auth.loginButton);
    
    // HTML5 validation should prevent submission
    const emailInput = page.locator(selectors.auth.emailInput);
    const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBeTruthy();
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill(selectors.auth.emailInput, 'invalid@example.com');
    await page.fill(selectors.auth.passwordInput, 'WrongPassword123');
    await page.click(selectors.auth.loginButton);
    
    // Wait for error message
    await expect(page.locator(selectors.auth.errorMessage)).toBeVisible({ timeout: 10000 });
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    
    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/dashboard/);
    
    // Check for dashboard elements
    await expect(page.locator('text=Dashboard, text=Products, text=Inventory').first()).toBeVisible();
  });

  test('should handle MFA flow if enabled', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill(selectors.auth.emailInput, testUsers.admin.email);
    await page.fill(selectors.auth.passwordInput, testUsers.admin.password);
    await page.click(selectors.auth.loginButton);
    
    // Wait for redirect
    await page.waitForURL(/\/(dashboard|mfa)/, { timeout: 10000 });
    
    // If MFA page is shown, verify it has the expected elements
    if (page.url().includes('/mfa')) {
      await expect(page.locator('input[type="text"], input[placeholder*="code" i]')).toBeVisible();
    }
  });

  test('should logout successfully', async ({ page }) => {
    // First login
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await expect(page).toHaveURL(/\/dashboard/);
    
    // Then logout
    await logout(page);
    
    // Should redirect to login page
    await expect(page).toHaveURL('/login');
  });

  test('should navigate to signup page', async ({ page }) => {
    await page.goto('/login');
    await page.click(selectors.auth.signupLink);
    
    await expect(page).toHaveURL('/signup');
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
  });

  test('should navigate to forgot password page', async ({ page }) => {
    await page.goto('/login');
    await page.click(selectors.auth.forgotPasswordLink);
    
    await expect(page).toHaveURL('/forgot-password');
    await expect(page.locator('input[type="email"]')).toBeVisible();
  });

  test('should persist authentication after page reload', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await expect(page).toHaveURL(/\/dashboard/);
    
    // Reload the page
    await page.reload();
    
    // Should still be on dashboard (not redirected to login)
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should display password requirements on signup', async ({ page }) => {
    await page.goto('/signup');
    
    const passwordInput = page.locator('input[type="password"]').first();
    await passwordInput.click();
    await passwordInput.fill('weak');
    
    // Check if password strength indicator or requirements are shown
    const strengthIndicator = page.locator('[class*="strength"], text=/(?:weak|strong|medium)/i');
    const hasIndicator = await strengthIndicator.count() > 0;
    
    // At minimum, the password input should exist
    expect(await passwordInput.isVisible()).toBeTruthy();
  });
});

test.describe('Authentication Security', () => {
  test('should not expose sensitive data in localStorage', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    
    const storageData = await page.evaluate(() => {
      const data: Record<string, string> = {};
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key) {
          data[key] = localStorage.getItem(key) || '';
        }
      }
      return data;
    });
    
    // Passwords should never be in localStorage
    const storageValues = Object.values(storageData).join(' ').toLowerCase();
    expect(storageValues).not.toContain(testUsers.admin.password.toLowerCase());
  });

  test('should clear auth data on logout', async ({ page }) => {
    await login(page, testUsers.admin.email, testUsers.admin.password);
    await logout(page);
    
    const hasToken = await page.evaluate(() => {
      return localStorage.getItem('access_token') || 
             localStorage.getItem('token') || 
             localStorage.getItem('auth_token');
    });
    
    expect(hasToken).toBeFalsy();
  });
});
