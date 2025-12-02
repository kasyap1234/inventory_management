import { test, expect, request } from '@playwright/test';
import { testUsers, selectors } from '../fixtures/test-data';
import { login, logout, ensureLoggedOut, waitForAuthToken } from '../fixtures/auth-helpers';

/**
 * Invitation Flow E2E Tests
 * Tests the full invitation lifecycle:
 * 1. Admin creates invitation via API
 * 2. User receives invitation (simulated by getting token)
 * 3. User accepts invitation
 * 4. User can login
 */

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1';

test.describe('Invitation Flow', () => {
    test.beforeEach(async ({ page }) => {
        await ensureLoggedOut(page);
    });

    test('should successfully invite and register a new user', async ({ page, request }) => {
        // 1. Login as Admin to get token
        await login(page, testUsers.admin.email, testUsers.admin.password);

        // Get auth token from local storage
        const token = await waitForAuthToken(page);
        expect(token).toBeTruthy();

        // 2. Create Invitation via API
        // First, we need a Role ID. Let's fetch roles.
        const rolesResponse = await request.get(`${API_URL}/roles`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Cookie': `auth_token=${token}` // Try both header and cookie
            }
        });

        expect(rolesResponse.ok()).toBeTruthy();
        const rolesData = await rolesResponse.json();
        const roleId = rolesData.roles?.[0]?.id; // Pick the first role
        expect(roleId).toBeTruthy();

        // Create the invitation
        const newEmail = `invited-${Date.now()}@example.com`;
        const inviteResponse = await request.post(`${API_URL}/invitations`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Cookie': `auth_token=${token}`
            },
            data: {
                email: newEmail,
                role_id: roleId,
                permissions: []
            }
        });

        expect(inviteResponse.ok()).toBeTruthy();
        const inviteData = await inviteResponse.json();
        const inviteToken = inviteData.token;
        expect(inviteToken).toBeTruthy();

        // 3. Logout Admin
        await logout(page);

        // 4. Navigate to Invitation Link
        await page.goto(`/invitations/${inviteToken}`);

        // Verify Invitation Page
        await expect(page.locator('text=Accept Invitation')).toBeVisible();
        await expect(page.locator(`input[value="${newEmail}"]`)).toBeVisible();

        // 5. Fill Registration Form
        const newUserPassword = 'NewUserPass123!';
        await page.fill('input[id="firstName"]', 'Invited');
        await page.fill('input[id="lastName"]', 'User');
        await page.fill('input[id="password"]', newUserPassword);
        await page.fill('input[id="confirmPassword"]', newUserPassword);

        // Submit
        await page.click('button:has-text("Create Account")');

        // 6. Verify Redirection to Login
        await expect(page).toHaveURL(/\/login/);

        // Check for success message if applicable
        // await expect(page.locator('text=Invitation accepted')).toBeVisible();

        // 7. Login with New User
        await page.fill(selectors.auth.emailInput, newEmail);
        await page.fill(selectors.auth.passwordInput, newUserPassword);
        await page.click(selectors.auth.loginButton);

        // 8. Verify Dashboard Access
        await expect(page).toHaveURL(/\/dashboard/);
        await expect(page.locator('text=Dashboard')).toBeVisible();
    });

    test('should show error for invalid invitation token', async ({ page }) => {
        await page.goto('/invitations/invalid-token-12345');

        // Verify Error Message
        await expect(page.locator('text=Invalid Invitation')).toBeVisible();
        await expect(page.locator('text=This invitation is invalid or has expired')).toBeVisible();

        // Verify Link to Login
        await expect(page.locator('button:has-text("Go to Login")')).toBeVisible();
    });
});
