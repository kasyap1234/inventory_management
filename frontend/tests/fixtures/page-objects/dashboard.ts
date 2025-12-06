import { expect, Locator, Page } from '@playwright/test';

export class DashboardPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard');
  }

  async openInventory() {
    const link = this.page.locator('a[href*="/inventory"], button:has-text("Inventory")').first();
    await link.click();
  }

  async openOrders() {
    const link = this.page.locator('a[href*="/orders"], button:has-text("Orders")').first();
    await link.click();
  }

  async openReports() {
    const link = this.page.locator('a[href*="/reports"], button:has-text("Reports")').first();
    if (await link.isVisible()) {
      await link.click();
    } else {
      await this.page.goto('/dashboard/reports');
    }
  }

  async openAnalytics() {
    const link = this.page.locator('a[href*="/analytics"], button:has-text("Analytics")').first();
    if (await link.isVisible()) {
      await link.click();
    } else {
      await this.page.goto('/dashboard/analytics');
    }
  }

  async assertLoaded() {
    await expect(this.page.locator('text=/Dashboard|Overview/i').first()).toBeVisible();
    await expect(this.page.locator('nav, [role="navigation"]').first()).toBeVisible();
  }
}

export class InventoryPage {
  constructor(private readonly page: Page) {}

  private rowForProduct(name: string): Locator {
    return this.page.locator('tr, [data-testid*="inventory"], [data-row-key]').filter({
      hasText: name,
    }).first();
  }

  async search(name: string) {
    const searchInput = this.page.locator('input[type="search"], input[placeholder*="search" i]').first();
    if (await searchInput.isVisible()) {
      await searchInput.fill(name);
      await this.page.keyboard.press('Enter');
    }
    await this.page.waitForLoadState('networkidle');
  }

  async expectProductVisible(name: string) {
    await expect(this.rowForProduct(name)).toBeVisible({ timeout: 15000 });
  }

  async readQuantity(name: string) {
    const row = this.rowForProduct(name);
    await expect(row).toBeVisible();
    const qtyCell = row.locator('td:has-text(/\\d/), [data-testid*="qty"], [class*="qty"]').first();
    const text = (await qtyCell.textContent()) || '0';
    const numeric = parseInt(text.replace(/[^\d-]/g, ''), 10);
    return Number.isFinite(numeric) ? numeric : 0;
  }

  async expectQuantityAtLeast(name: string, expected: number) {
    const qty = await this.readQuantity(name);
    expect(qty).toBeGreaterThanOrEqual(expected);
  }
}

export class OrdersPage {
  constructor(private readonly page: Page) {}

  private rowForOrder(text: string): Locator {
    return this.page.locator('tr, [data-testid*="order"]').filter({ hasText: text }).first();
  }

  async search(term: string) {
    const searchInput = this.page.locator('input[type="search"], input[placeholder*="search" i]').first();
    if (await searchInput.isVisible()) {
      await searchInput.fill(term);
      await this.page.keyboard.press('Enter');
      await this.page.waitForLoadState('networkidle');
    }
  }

  async expectOrderVisible(term: string) {
    await expect(this.rowForOrder(term)).toBeVisible({ timeout: 15000 });
  }
}

export class ReportsPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/reports');
  }

  async expectInventorySummary() {
    await expect(this.page.locator('text=/Inventory Report/i').first()).toBeVisible();
    await expect(this.page.locator('text=/Total Items/i').first()).toBeVisible();
  }

  async expectSalesSummary() {
    await expect(this.page.locator('text=/Sales Report/i').first()).toBeVisible();
    await expect(this.page.locator('text=/Total Orders/i').first()).toBeVisible();
  }
}

