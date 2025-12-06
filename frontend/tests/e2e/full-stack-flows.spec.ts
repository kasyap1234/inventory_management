import { test, expect } from '@playwright/test';
import { loginAsAdmin, loginAsManager, ensureLoggedOut } from '../fixtures/auth-helpers';
import { TestDataManager } from '../fixtures/data-manager';
import { DashboardPage, InventoryPage, OrdersPage, ReportsPage } from '../fixtures/page-objects/dashboard';

test.describe.configure({ mode: 'serial' });

test.describe('Full-stack flows (frontend ↔ backend)', () => {
  let dataManager: TestDataManager;
  const seeded = {
    productId: '',
    productName: '',
    warehouseId: '',
    distributorId: '',
    inventoryId: '',
    orderId: '',
  };

  test.beforeAll(async ({ request }) => {
    dataManager = await TestDataManager.init(request);

    const warehouse = await dataManager.createWarehouse();
    const distributor = await dataManager.createDistributor();
    const product = await dataManager.createProduct();

    if (!warehouse.id || !distributor.id || !product.id) {
      throw new Error('Failed to seed prerequisites for full-stack tests');
    }

    const inventory = await dataManager.createInventory(product.id, warehouse.id, 25);
    const order = await dataManager.createOrder({
      productId: product.id,
      warehouseId: warehouse.id,
      distributorId: distributor.id,
      quantity: 3,
      unitPrice: 200,
    });

    seeded.productId = product.id;
    seeded.productName = product.payload.name;
    seeded.warehouseId = warehouse.id;
    seeded.distributorId = distributor.id;
    seeded.inventoryId = inventory.id || '';
    seeded.orderId = order.id || '';
  });

  test.afterAll(async () => {
    if (dataManager) {
      await dataManager.cleanup(seeded);
    }
  });

  test('Inventory reflects backend seed and API adjustments', async ({ page }) => {
    const dashboard = new DashboardPage(page);
    const inventoryPage = new InventoryPage(page);

    await loginAsAdmin(page);
    await dashboard.goto();
    await dashboard.openInventory();
    await inventoryPage.search(seeded.productName);
    await inventoryPage.expectProductVisible(seeded.productName);

    const { quantity: quantityBefore } = await dataManager.getInventoryQuantity(seeded.productId);
    await dataManager.adjustInventory(seeded.productId, 5, 'E2E adjustment');

    await page.reload();
    await inventoryPage.search(seeded.productName);
    await inventoryPage.expectQuantityAtLeast(seeded.productName, quantityBefore + 5);
  });

  test('Orders surface backend-created order in the UI', async ({ page }) => {
    const dashboard = new DashboardPage(page);
    const ordersPage = new OrdersPage(page);
    const orderSearchTerm = seeded.orderId ? seeded.orderId.slice(0, 8) : 'Order';

    await loginAsAdmin(page);
    await dashboard.goto();
    await dashboard.openOrders();
    await ordersPage.search(orderSearchTerm);
    await ordersPage.expectOrderVisible(orderSearchTerm);
  });

  test('Reports render sales and inventory summaries from live data', async ({ page }) => {
    const dashboard = new DashboardPage(page);
    const reportsPage = new ReportsPage(page);

    await loginAsAdmin(page);
    await dashboard.goto();
    await dashboard.openReports();
    await reportsPage.expectInventorySummary();
    await reportsPage.expectSalesSummary();
  });

  test('Non-admin role is blocked from admin/RBAC screens', async ({ page }) => {
    await ensureLoggedOut(page);
    await loginAsManager(page);

    await page.goto('/dashboard/rbac');
    const unauthorized = page.locator('text=/unauthorized|forbidden|access denied/i').first();
    const onRbac = page.url().includes('/rbac');

    if (onRbac) {
      await expect(unauthorized).toBeVisible();
    } else {
      await expect(page).not.toHaveURL(/\/admin|\/rbac/);
    }
  });
});

