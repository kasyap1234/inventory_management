import { Page, expect } from '@playwright/test';
import { testUsers, selectors } from './test-data';

/**
 * Authentication helper functions for Playwright tests
 */

export async function login(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.fill(selectors.auth.emailInput, email);
  await page.fill(selectors.auth.passwordInput, password);
  await page.click(selectors.auth.loginButton);
  
  // Wait for navigation to dashboard or MFA page
  await page.waitForURL(/\/(dashboard|mfa)/, { timeout: 10000 });
}

export async function loginAsAdmin(page: Page) {
  await login(page, testUsers.admin.email, testUsers.admin.password);
}

export async function loginAsManager(page: Page) {
  await login(page, testUsers.manager.email, testUsers.manager.password);
}

export async function loginAsUser(page: Page) {
  await login(page, testUsers.user.email, testUsers.user.password);
}

export async function logout(page: Page) {
  // Click on profile menu or settings
  const profileButton = page.locator('button:has-text("Profile"), button:has-text("Settings"), [data-testid="user-menu"]').first();
  if (await profileButton.isVisible({ timeout: 5000 })) {
    await profileButton.click();
  }
  
  // Click logout
  const logoutButton = page.locator('button:has-text("Logout"), button:has-text("Sign out"), a:has-text("Logout")').first();
  await logoutButton.click();
  
  // Wait for redirect to login
  await page.waitForURL('/login', { timeout: 10000 });
}

export async function ensureLoggedOut(page: Page) {
  await page.context().clearCookies();
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
}

export async function signup(page: Page, userData: {
  email: string;
  password: string;
  name?: string;
  phoneNumber?: string;
}) {
  await page.goto('/signup');
  
  await page.fill('input[type="email"]', userData.email);
  await page.fill('input[type="password"]', userData.password);
  
  if (userData.name) {
    const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]');
    if (await nameInput.isVisible()) {
      await nameInput.fill(userData.name);
    }
  }
  
  if (userData.phoneNumber) {
    const phoneInput = page.locator('input[name="phone"], input[placeholder*="phone" i]');
    if (await phoneInput.isVisible()) {
      await phoneInput.fill(userData.phoneNumber);
    }
  }
  
  await page.click('button[type="submit"]');
}

export async function waitForAuthToken(page: Page): Promise<string | null> {
  return await page.evaluate(() => {
    return localStorage.getItem('access_token') || localStorage.getItem('token');
  });
}

export async function setAuthToken(page: Page, token: string) {
  await page.evaluate((t) => {
    localStorage.setItem('access_token', t);
  }, token);
}

export async function isAuthenticated(page: Page): Promise<boolean> {
  const token = await waitForAuthToken(page);
  return !!token;
}
