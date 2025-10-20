# Frontend Test Suite

Comprehensive test suite for the AgroMart Inventory Management System frontend application using Playwright.

## Overview

This test suite includes:
- **E2E Tests**: End-to-end user flow tests
- **Integration Tests**: API and UI integration tests
- **Accessibility Tests**: WCAG compliance tests
- **Performance Tests**: Load time and runtime performance tests
- **Responsive Tests**: Multi-device and responsive design tests

## Test Structure

```
tests/
├── e2e/                          # End-to-end tests
│   ├── auth.spec.ts              # Authentication flows
│   ├── dashboard.spec.ts         # Dashboard functionality
│   ├── products.spec.ts          # Product management
│   ├── inventory.spec.ts         # Inventory & orders
│   ├── analytics.spec.ts         # Analytics & reporting
│   ├── responsive.spec.ts        # Responsive design
│   └── error-handling.spec.ts    # Error handling
├── integration/                  # Integration tests
│   ├── api-ui-integration.spec.ts # API-UI sync tests
│   ├── accessibility.spec.ts     # Accessibility tests
│   └── performance.spec.ts       # Performance tests
└── fixtures/                     # Test utilities
    ├── test-data.ts              # Test data and constants
    ├── auth-helpers.ts           # Auth helper functions
    └── api-helpers.ts            # API helper functions
```

## Running Tests

### Run all tests
```bash
bun run test
```

### Run specific test suite
```bash
bun run test:e2e           # E2E tests only
bun run test:integration   # Integration tests only
bun run test:auth          # Auth tests only
bun run test:ui            # UI tests in headed mode
```

### Run tests in specific browser
```bash
bun run test --project=chromium
bun run test --project=firefox
bun run test --project=webkit
```

### Run tests in headed mode (see browser)
```bash
bun run test:ui
```

### Debug tests
```bash
bun run test:debug
```

### Generate HTML report
```bash
bun run test:report
```

## Prerequisites

### 1. Install Playwright Browsers
```bash
bunx playwright install
```

### 2. Set Environment Variables
Create a `.env.test` file:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
```

### 3. Start Backend Server
Ensure the Go backend server is running:
```bash
cd .. && ./run_start.sh
```

### 4. Prepare Test Data
Create test users in the database:
- Admin user: admin@test.com / Admin@123456
- Manager user: manager@test.com / Manager@123456
- Regular user: user@test.com / User@123456

## Test Coverage

### Authentication Tests (auth.spec.ts)
- ✅ Login with valid/invalid credentials
- ✅ Signup flow
- ✅ Logout functionality
- ✅ Password reset
- ✅ MFA flow
- ✅ Session persistence
- ✅ Security (token storage, session expiration)

### Dashboard Tests (dashboard.spec.ts)
- ✅ Dashboard rendering
- ✅ Navigation between pages
- ✅ Metrics and analytics cards
- ✅ User profile menu
- ✅ Mobile responsive navigation
- ✅ Charts and graphs loading
- ✅ Notifications

### Product Management Tests (products.spec.ts)
- ✅ Product list display
- ✅ Search and filter
- ✅ Create new product
- ✅ Edit existing product
- ✅ Delete product with confirmation
- ✅ Form validation
- ✅ Pagination
- ✅ Bulk operations
- ✅ Export functionality

### Inventory Tests (inventory.spec.ts)
- ✅ Inventory levels display
- ✅ Stock adjustments
- ✅ Low stock alerts
- ✅ Warehouse filtering
- ✅ Inventory history
- ✅ Inventory valuation
- ✅ Order management
- ✅ Order status updates

### Analytics Tests (analytics.spec.ts)
- ✅ Analytics dashboard
- ✅ Sales trends charts
- ✅ Top products display
- ✅ Date range filtering
- ✅ Revenue breakdown
- ✅ Export reports
- ✅ Real-time data updates

### Integration Tests (api-ui-integration.spec.ts)
- ✅ API to UI data sync
- ✅ UI to API data sync
- ✅ Real-time updates
- ✅ End-to-end workflows
- ✅ Concurrent operations
- ✅ Error handling
- ✅ Session management

### Accessibility Tests (accessibility.spec.ts)
- ✅ Keyboard navigation
- ✅ Screen reader support
- ✅ ARIA labels and roles
- ✅ Form labels
- ✅ Focus management
- ✅ Color contrast
- ✅ Semantic HTML
- ✅ Alt text for images

### Performance Tests (performance.spec.ts)
- ✅ Page load times
- ✅ Network optimization
- ✅ Resource caching
- ✅ Bundle sizes
- ✅ Runtime performance
- ✅ API response times
- ✅ Memory leak detection

### Responsive Tests (responsive.spec.ts)
- ✅ Mobile layout (iPhone, Android)
- ✅ Tablet layout (iPad)
- ✅ Desktop layout (1920x1080)
- ✅ Touch-friendly buttons
- ✅ Responsive tables
- ✅ Orientation changes
- ✅ Cross-browser compatibility

### Error Handling Tests (error-handling.spec.ts)
- ✅ Network errors
- ✅ API errors (401, 404, 500)
- ✅ Form validation errors
- ✅ Error boundaries
- ✅ Timeout handling
- ✅ Loading states
- ✅ Error recovery

## Best Practices

1. **Test Isolation**: Each test should be independent and not rely on other tests
2. **Clean Up**: Always clean up test data after tests complete
3. **Realistic Data**: Use realistic test data that mimics production scenarios
4. **Wait Strategies**: Use proper wait strategies instead of fixed timeouts when possible
5. **Assertions**: Make specific, meaningful assertions
6. **Page Objects**: Use helper functions to avoid code duplication

## CI/CD Integration

Tests can be run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Playwright tests
  run: |
    bunx playwright install --with-deps
    bun run test
```

## Debugging Failed Tests

### View test report
```bash
bun run test:report
```

### Run specific failing test
```bash
bun run test tests/e2e/auth.spec.ts:10
```

### Enable trace viewer
```bash
bunx playwright show-trace trace.zip
```

### Take screenshots on failure
Screenshots are automatically saved to `test-results/` on failure.

## Writing New Tests

1. Create a new spec file in the appropriate directory
2. Import required helpers from `fixtures/`
3. Use descriptive test names
4. Follow existing patterns for consistency
5. Add comments for complex test logic
6. Clean up test data in `afterEach` or `afterAll` hooks

Example:
```typescript
import { test, expect } from '@playwright/test';
import { loginAsAdmin } from '../fixtures/auth-helpers';

test.describe('My Feature', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/my-feature');
  });

  test('should do something', async ({ page }) => {
    // Test implementation
  });
});
```

## Troubleshooting

### Tests timing out
- Increase timeout in `playwright.config.ts`
- Check backend server is running
- Verify network connectivity

### Tests failing on CI but passing locally
- Ensure all dependencies are installed
- Check environment variables
- Verify database state

### Flaky tests
- Use proper wait strategies
- Avoid hardcoded timeouts
- Check for race conditions

## Contributing

When adding new features, please:
1. Add corresponding tests
2. Ensure all tests pass
3. Update this README if needed
4. Follow existing test patterns

## Support

For questions or issues:
1. Check test logs and screenshots
2. Review Playwright documentation
3. Check existing tests for examples
