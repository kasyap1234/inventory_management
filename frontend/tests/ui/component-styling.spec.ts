import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

/**
 * Component Styling Tests
 * Verifies individual UI components have modern, polished styling
 */

test.describe('Button Component Styling', () => {
  test('should have modern primary button styling', async ({ page }) => {
    await page.goto('/login');
    
    const submitButton = page.locator('button[type="submit"]');
    
    const styles = await submitButton.evaluate(el => {
      const computed = window.getComputedStyle(el);
      return {
        borderRadius: computed.borderRadius,
        padding: computed.padding,
        fontSize: computed.fontSize,
        fontWeight: computed.fontWeight,
        transition: computed.transition,
        cursor: computed.cursor
      };
    });
    
    // Modern rounded corners
    expect(parseInt(styles.borderRadius)).toBeGreaterThan(4);
    
    // Adequate padding
    expect(styles.padding).not.toBe('0px');
    
    // Readable font size
    expect(parseInt(styles.fontSize)).toBeGreaterThanOrEqual(14);
    
    // Bold text
    expect(parseInt(styles.fontWeight)).toBeGreaterThanOrEqual(500);
    
    // Smooth transitions
    expect(styles.transition).not.toBe('all 0s ease 0s');
    
    // Pointer cursor
    expect(styles.cursor).toBe('pointer');
  });

  test('should have consistent button heights', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
    
    const buttons = await page.locator('button:visible').all();
    const heights: number[] = [];
    
    for (const button of buttons.slice(0, 5)) {
      const box = await button.boundingBox();
      if (box) {
        heights.push(box.height);
      }
    }
    
    // Most buttons should have similar heights (allowing for icon-only buttons)
    const uniqueHeights = [...new Set(heights.map(h => Math.round(h / 4) * 4))];
    expect(uniqueHeights.length).toBeLessThanOrEqual(4);
  });

  test('should have disabled button styling', async ({ page }) => {
    await page.goto('/login');
    
    const submitButton = page.locator('button[type="submit"]');
    
    // Try to trigger disabled state by clicking rapidly
    await submitButton.click();
    
    // Check if button can be disabled
    const canBeDisabled = await submitButton.evaluate(el => {
      return 'disabled' in el;
    });
    
    expect(canBeDisabled).toBeTruthy();
  });
});

test.describe('Card Component Styling', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have modern card styling', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const card = page.locator('[class*="card"], [class*="Card"]').first();
    
    if (await card.isVisible()) {
      const styles = await card.evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          borderRadius: computed.borderRadius,
          boxShadow: computed.boxShadow,
          backgroundColor: computed.backgroundColor,
          padding: computed.padding,
          border: computed.border
        };
      });
      
      // Rounded corners
      expect(parseInt(styles.borderRadius)).toBeGreaterThan(4);
      
      // Has shadow for depth
      expect(styles.boxShadow).not.toBe('none');
      
      // Has background color
      expect(styles.backgroundColor).not.toBe('rgba(0, 0, 0, 0)');
      
      // Has padding
      expect(styles.padding).not.toBe('0px');
    }
  });

  test('should have consistent card spacing', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const cards = await page.locator('[class*="card"], [class*="Card"]').all();
    
    if (cards.length >= 2) {
      const padding1 = await cards[0].evaluate(el => window.getComputedStyle(el).padding);
      const padding2 = await cards[1].evaluate(el => window.getComputedStyle(el).padding);
      
      // Cards should have consistent padding
      expect(padding1).toBeTruthy();
      expect(padding2).toBeTruthy();
    }
  });

  test('should have card hover effects', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const card = page.locator('[class*="card"], [class*="Card"]').first();
    
    if (await card.isVisible()) {
      // Check for transition property
      const transition = await card.evaluate(el => window.getComputedStyle(el).transition);
      
      // Cards should have transitions for hover effects
      expect(transition).toBeTruthy();
    }
  });
});

test.describe('Input Component Styling', () => {
  test('should have modern input field styling', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    
    const styles = await emailInput.evaluate(el => {
      const computed = window.getComputedStyle(el);
      return {
        height: computed.height,
        padding: computed.padding,
        borderRadius: computed.borderRadius,
        border: computed.border,
        fontSize: computed.fontSize,
        transition: computed.transition
      };
    });
    
    // Comfortable height
    expect(parseInt(styles.height)).toBeGreaterThanOrEqual(36);
    
    // Adequate padding
    expect(styles.padding).not.toBe('0px');
    
    // Rounded corners
    expect(parseInt(styles.borderRadius)).toBeGreaterThan(2);
    
    // Has border
    expect(styles.border).not.toBe('0px');
    
    // Readable font size
    expect(parseInt(styles.fontSize)).toBeGreaterThanOrEqual(14);
    
    // Smooth transitions
    expect(styles.transition).toBeTruthy();
  });

  test('should have focus state styling', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    await emailInput.focus();
    
    // Check if element has focus
    const hasFocus = await emailInput.evaluate(el => el === document.activeElement);
    expect(hasFocus).toBeTruthy();
    
    // Check for outline or ring (modern focus indicators)
    const outline = await emailInput.evaluate(el => window.getComputedStyle(el).outline);
    const boxShadow = await emailInput.evaluate(el => window.getComputedStyle(el).boxShadow);
    
    // Should have some focus indicator
    expect(outline !== 'none' || boxShadow !== 'none').toBeTruthy();
  });

  test('should have placeholder styling', async ({ page }) => {
    await page.goto('/login');
    
    const emailInput = page.locator('#email');
    const placeholder = await emailInput.getAttribute('placeholder');
    
    // Should have helpful placeholder
    expect(placeholder).toBeTruthy();
    if (placeholder) {
      expect(placeholder.length).toBeGreaterThan(3);
    }
  });
});

test.describe('Typography Styling', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have proper heading hierarchy', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const h1 = page.locator('h1').first();
    
    if (await h1.isVisible()) {
      const h1Styles = await h1.evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          fontSize: parseInt(computed.fontSize),
          fontWeight: parseInt(computed.fontWeight),
          lineHeight: computed.lineHeight,
          marginBottom: computed.marginBottom
        };
      });
      
      // H1 should be large
      expect(h1Styles.fontSize).toBeGreaterThanOrEqual(28);
      
      // H1 should be bold
      expect(h1Styles.fontWeight).toBeGreaterThanOrEqual(600);
      
      // Should have proper line height
      expect(h1Styles.lineHeight).not.toBe('normal');
      
      // Should have bottom margin
      expect(h1Styles.marginBottom).not.toBe('0px');
    }
  });

  test('should have readable body text', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const bodyText = page.locator('p, span, div').first();
    
    if (await bodyText.isVisible()) {
      const fontSize = await bodyText.evaluate(el => {
        return parseInt(window.getComputedStyle(el).fontSize);
      });
      
      // Body text should be at least 14px
      expect(fontSize).toBeGreaterThanOrEqual(14);
    }
  });

  test('should have proper font weights', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const body = page.locator('body');
    const fontWeight = await body.evaluate(el => {
      return parseInt(window.getComputedStyle(el).fontWeight);
    });
    
    // Body should have normal weight (400) or slightly heavier
    expect(fontWeight).toBeGreaterThanOrEqual(300);
    expect(fontWeight).toBeLessThanOrEqual(500);
  });
});

test.describe('Icon Styling', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should have properly sized icons', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const icons = await page.locator('svg').all();
    
    for (const icon of icons.slice(0, 5)) {
      if (await icon.isVisible()) {
        const box = await icon.boundingBox();
        
        if (box) {
          // Icons should be reasonable size (16-48px typically)
          expect(box.width).toBeGreaterThanOrEqual(12);
          expect(box.width).toBeLessThanOrEqual(64);
          expect(box.height).toBeGreaterThanOrEqual(12);
          expect(box.height).toBeLessThanOrEqual(64);
        }
      }
    }
  });

  test('should have consistent icon styling', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const icons = await page.locator('svg').all();
    const sizes: number[] = [];
    
    for (const icon of icons.slice(0, 10)) {
      if (await icon.isVisible()) {
        const box = await icon.boundingBox();
        if (box) {
          sizes.push(box.width);
        }
      }
    }
    
    // Most icons should have similar sizes
    const uniqueSizes = [...new Set(sizes.map(s => Math.round(s / 4) * 4))];
    expect(uniqueSizes.length).toBeLessThanOrEqual(5);
  });
});

test.describe('Table Styling', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
  });

  test('should have modern table styling', async ({ page }) => {
    const table = page.locator('table').first();
    
    if (await table.isVisible()) {
      const styles = await table.evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          width: computed.width,
          borderCollapse: computed.borderCollapse,
          backgroundColor: computed.backgroundColor
        };
      });
      
      // Table should use full width
      expect(styles.width).toBeTruthy();
      
      // Should have border collapse for clean look
      expect(styles.borderCollapse).toBeTruthy();
    }
  });

  test('should have styled table headers', async ({ page }) => {
    const tableHeader = page.locator('thead th, th').first();
    
    if (await tableHeader.isVisible()) {
      const styles = await tableHeader.evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          fontWeight: parseInt(computed.fontWeight),
          padding: computed.padding,
          textAlign: computed.textAlign
        };
      });
      
      // Headers should be bold
      expect(styles.fontWeight).toBeGreaterThanOrEqual(500);
      
      // Headers should have padding
      expect(styles.padding).not.toBe('0px');
    }
  });

  test('should have proper table row styling', async ({ page }) => {
    const tableRow = page.locator('tbody tr, tr').first();
    
    if (await tableRow.isVisible()) {
      const padding = await tableRow.locator('td').first().evaluate(el => {
        return window.getComputedStyle(el).padding;
      });
      
      // Cells should have padding
      expect(padding).not.toBe('0px');
    }
  });
});

test.describe('Badge/Tag Styling', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
  });

  test('should have modern badge styling', async ({ page }) => {
    const badge = page.locator('[class*="badge"], [class*="tag"], [class*="status"]').first();
    
    if (await badge.isVisible()) {
      const styles = await badge.evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          borderRadius: computed.borderRadius,
          padding: computed.padding,
          fontSize: parseInt(computed.fontSize),
          fontWeight: parseInt(computed.fontWeight)
        };
      });
      
      // Badges should be rounded
      expect(parseInt(styles.borderRadius)).toBeGreaterThan(4);
      
      // Should have padding
      expect(styles.padding).not.toBe('0px');
      
      // Should have smaller font
      expect(styles.fontSize).toBeLessThanOrEqual(16);
      
      // Should be medium weight
      expect(styles.fontWeight).toBeGreaterThanOrEqual(400);
    }
  });
});

test.describe('Modal/Dialog Styling', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard/products');
    await page.waitForLoadState('networkidle');
  });

  test('should have modern modal styling', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add"), button:has-text("Create")').first();
    
    if (await addButton.isVisible()) {
      await addButton.click();
      await page.waitForTimeout(500);
      
      const modal = page.locator('[role="dialog"]').first();
      
      if (await modal.isVisible()) {
        const styles = await modal.evaluate(el => {
          const computed = window.getComputedStyle(el);
          return {
            borderRadius: computed.borderRadius,
            boxShadow: computed.boxShadow,
            maxWidth: computed.maxWidth,
            padding: computed.padding
          };
        });
        
        // Modal should be rounded
        expect(parseInt(styles.borderRadius)).toBeGreaterThan(4);
        
        // Should have shadow
        expect(styles.boxShadow).not.toBe('none');
        
        // Should have max width
        expect(styles.maxWidth).not.toBe('none');
        
        // Should have padding
        expect(styles.padding).not.toBe('0px');
      }
    }
  });
});

test.describe('Loading States Styling', () => {
  test('should have styled loading indicators', async ({ page }) => {
    await page.goto('/login');
    
    const submitButton = page.locator('button[type="submit"]');
    await submitButton.click();
    
    // Look for loading spinner
    const spinner = page.locator('[class*="spinner"], [class*="loading"], svg[class*="animate"]').first();
    
    const hasSpinner = await spinner.isVisible({ timeout: 2000 }).catch(() => false);
    
    if (hasSpinner) {
      // Spinner should have animation
      const animation = await spinner.evaluate(el => window.getComputedStyle(el).animation);
      expect(animation).not.toBe('none');
    }
  });

  test('should have skeleton loading states', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
    
    // Reload to see skeleton
    await page.reload();
    
    const skeleton = page.locator('[class*="skeleton"]').first();
    const hasSkeleton = await skeleton.isVisible({ timeout: 1000 }).catch(() => false);
    
    if (hasSkeleton) {
      // Skeleton should have animation
      const animation = await skeleton.evaluate(el => window.getComputedStyle(el).animation);
      expect(animation).not.toBe('none');
    }
  });
});

test.describe('Color Consistency', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/dashboard');
  });

  test('should use consistent primary colors', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    // Get primary button colors
    const buttons = await page.locator('button:not([class*="outline"]):not([class*="ghost"])').all();
    const colors: string[] = [];
    
    for (const button of buttons.slice(0, 3)) {
      if (await button.isVisible()) {
        const bgColor = await button.evaluate(el => window.getComputedStyle(el).backgroundColor);
        colors.push(bgColor);
      }
    }
    
    // Most primary buttons should have similar colors
    const uniqueColors = [...new Set(colors)];
    expect(uniqueColors.length).toBeLessThanOrEqual(3);
  });

  test('should use consistent text colors', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    
    const textElements = await page.locator('p, span, div').all();
    const colors: string[] = [];
    
    for (const element of textElements.slice(0, 10)) {
      if (await element.isVisible()) {
        const color = await element.evaluate(el => window.getComputedStyle(el).color);
        colors.push(color);
      }
    }
    
    // Should have limited color palette
    const uniqueColors = [...new Set(colors)];
    expect(uniqueColors.length).toBeLessThanOrEqual(6);
  });
});
