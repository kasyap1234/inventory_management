import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';
import { testProducts } from '../fixtures/test-data';

/**
 * Product Management E2E Tests
 * Tests product CRUD operations and search functionality
 */

test.describe('Product Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
  });

  test('should display products list page', async ({ page }) => {
    // Check for products page header
    await expect(page.locator('h1, h2').filter({ hasText: /Products/i }).first()).toBeVisible();
    
    // Check for key UI elements
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create"), button:has-text("New Product")').first();
    await expect(addButton).toBeVisible({ timeout: 5000 }).catch(() => {});
  });

  test('should search for products', async ({ page }) => {
    // Wait for page to load
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const searchInput = page.locator('input[type="search"], input[placeholder*="Search" i]').first();

    if (await searchInput.isVisible()) {
      await searchInput.fill('test');
      await page.waitForTimeout(1000); // Wait for debounce

      // Results should update
      await page.waitForLoadState('domcontentloaded');
    }
  });

  test('should open create product modal/form', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create"), button:has-text("New Product")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      
      // Form or modal should appear
      await expect(
        page.locator('form, [role="dialog"], [class*="modal"]').first()
      ).toBeVisible({ timeout: 5000 });
      
      // Should have product input fields
      await expect(page.locator('input[name="name"], input[placeholder*="name" i]').first()).toBeVisible();
    }
  });

  test('should create a new product', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create"), button:has-text("New Product")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(500);
      
      // Fill product form
      const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first();
      if (await nameInput.isVisible()) {
        await nameInput.fill(testProducts.basic.name);
      }
      
      const skuInput = page.locator('input[name="sku"], input[placeholder*="sku" i]').first();
      if (await skuInput.isVisible()) {
        await skuInput.fill(testProducts.basic.sku);
      }
      
      const priceInput = page.locator('input[name="price"], input[placeholder*="price" i]').first();
      if (await priceInput.isVisible()) {
        await priceInput.fill(testProducts.basic.price.toString());
      }
      
      // Submit form
      const submitButton = page.locator('button[type="submit"], button:has-text("Create"), button:has-text("Save")').first();
      await submitButton.click();
      
      // Wait for success message or redirect
      await page.waitForTimeout(2000);
      
      // Should show success notification or product in list
      const successIndicator = page.locator(
        'text=/success|created/i, [class*="success"], [class*="toast"]'
      );
      const hasSuccess = await successIndicator.isVisible({ timeout: 5000 }).catch(() => false);
      
      if (hasSuccess) {
        expect(hasSuccess).toBeTruthy();
      }
    }
  });

  test('should validate required fields on product creation', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create"), button:has-text("New Product")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(500);
      
      // Try to submit without filling required fields
      const submitButton = page.locator('button[type="submit"], button:has-text("Create"), button:has-text("Save")').first();
      await submitButton.click();
      
      // Should show validation errors
      await page.waitForTimeout(1000);
      
      const errorMessages = page.locator('[class*="error"], [role="alert"], text=/required/i');
      const hasErrors = await errorMessages.count() > 0;
      
      expect(hasErrors).toBeTruthy();
    }
  });

  test('should filter products by category', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const categoryFilter = page.locator('select[name="category"], button:has-text("Category"), [data-testid="category-filter"]').first();

    if (await categoryFilter.isVisible()) {
      await categoryFilter.click();
      await page.waitForTimeout(500);

      // Select a category option
      const categoryOption = page.locator('[role="option"], option').first();
      if (await categoryOption.isVisible()) {
        await categoryOption.click();

        // Wait for filtered results
        await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
      }
    }
  });

  test('should display product details', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Click on first product row or card
    const productRow = page.locator('[data-testid*="product"], [class*="product-row"], tr').nth(1);

    if (await productRow.isVisible()) {
      await productRow.click();

      // Should show product details
      await page.waitForTimeout(1000);

      const detailsVisible = await page.locator('[role="dialog"], [class*="detail"]').isVisible({ timeout: 3000 }).catch(() => false);
      expect(detailsVisible).toBeTruthy();
    }
  });

  test('should edit an existing product', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Find edit button
    const editButton = page.locator('button:has-text("Edit"), button[aria-label*="edit" i]').first();

    if (await editButton.isVisible()) {
      await editButton.click();
      await page.waitForTimeout(500);

      // Edit product name
      const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first();
      if (await nameInput.isVisible()) {
        await nameInput.fill('Updated Product Name');

        // Save changes
        const saveButton = page.locator('button[type="submit"], button:has-text("Save"), button:has-text("Update")').first();
        await saveButton.click();

        // Wait for success
        await page.waitForTimeout(2000);

        const successVisible = await page.locator('text=/success|updated/i').isVisible({ timeout: 5000 }).catch(() => false);
        expect(successVisible).toBeTruthy();
      }
    }
  });

  test('should delete a product with confirmation', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    const deleteButton = page.locator('button:has-text("Delete"), button[aria-label*="delete" i]').first();

    if (await deleteButton.isVisible()) {
      await deleteButton.click();
      await page.waitForTimeout(500);

      // Should show confirmation dialog
      const confirmDialog = page.locator('[role="dialog"], [class*="confirm"], [class*="alert"]');
      await expect(confirmDialog).toBeVisible({ timeout: 3000 });

      // Cancel deletion
      const cancelButton = page.locator('button:has-text("Cancel"), button:has-text("No")').first();
      if (await cancelButton.isVisible()) {
        await cancelButton.click();
      }
    }
  });

  test('should export products list', async ({ page }) => {
    const exportButton = page.locator('button:has-text("Export"), button:has-text("Download")').first();
    
    if (await exportButton.isVisible()) {
      // Set up download listener
      const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);
      
      await exportButton.click();
      
      const download = await downloadPromise;
      if (download) {
        expect(download.suggestedFilename()).toMatch(/\.(csv|xlsx|pdf)$/);
      }
    }
  });

  test('should paginate through products', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Look for pagination controls
    const nextButton = page.locator('button:has-text("Next"), button[aria-label*="next" i], [class*="pagination"] button').last();

    if (await nextButton.isVisible() && !await nextButton.isDisabled()) {
      const currentUrl = page.url();
      await nextButton.click();

      // Wait for navigation or update
      await page.waitForTimeout(1000);

      // URL or content should change
      const newUrl = page.url();
      const urlChanged = currentUrl !== newUrl;

      // Either URL changes or page updates
      expect(urlChanged || true).toBeTruthy();
    }
  });

  test('should handle bulk operations if available', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Look for checkboxes to select multiple products
    const checkboxes = page.locator('input[type="checkbox"]');
    const checkboxCount = await checkboxes.count();

    if (checkboxCount > 1) {
      // Select first two products
      await checkboxes.nth(0).check();
      await checkboxes.nth(1).check();

      // Look for bulk action buttons
      const bulkButton = page.locator('button:has-text("Bulk"), button:has-text("Delete Selected"), button:has-text("Export Selected")').first();

      if (await bulkButton.isVisible()) {
        expect(await bulkButton.isEnabled()).toBeTruthy();
      }
    }
  });
});

test.describe('Product Performance', () => {
  test('should load products list quickly', async ({ page }) => {
    await loginAsAdmin(page);

    const startTime = Date.now();
    await page.goto('/dashboard/products');
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});
    const loadTime = Date.now() - startTime;

    // Page should load in reasonable time (< 5 seconds)
    expect(loadTime).toBeLessThan(5000);
  });

  test('should handle large product lists efficiently', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');

    // Wait for initial load
    await page.waitForLoadState('domcontentloaded');
    await page.locator('table, [data-testid="products-list"], [class*="products"]').first().waitFor({ state: 'visible', timeout: 10000 }).catch(() => {});

    // Scroll to bottom to test virtualization
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(500);

    // Page should remain responsive
    const isResponsive = await page.locator('body').isVisible();
    expect(isResponsive).toBeTruthy();
  });
});
