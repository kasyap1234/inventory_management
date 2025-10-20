# Test Setup Guide

This guide will help you set up and run the test suite for the first time.

## Quick Start

```bash
# 1. Install Playwright browsers
bun run test:install

# 2. Run all tests
bun run test

# 3. View test report
bun run test:report
```

## Detailed Setup

### 1. Install Dependencies

Playwright dependencies are already installed via `package.json`. If you need to reinstall:

```bash
bun install
bun run test:install
```

### 2. Configure Environment

Create a `.env.local` file in the frontend directory:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
```

### 3. Prepare Test Database

The tests require test users to be present in the database. Run these SQL commands:

```sql
-- Create test users (adjust based on your schema)
INSERT INTO users (email, password_hash, role, is_verified) VALUES
  ('admin@test.com', '$hashed_password', 'admin', true),
  ('manager@test.com', '$hashed_password', 'manager', true),
  ('user@test.com', '$hashed_password', 'user', true);
```

**Important**: Hash the passwords using your application's password hashing mechanism.

Test user credentials:
- **Admin**: admin@test.com / Admin@123456
- **Manager**: manager@test.com / Manager@123456
- **User**: user@test.com / User@123456

### 4. Start Required Services

Before running tests, ensure these services are running:

#### Start Backend Server
```bash
cd ..
./run_start.sh
```

#### Start Frontend Dev Server
The test configuration will automatically start the frontend server.
Alternatively, start it manually:
```bash
bun run dev
```

### 5. Run Tests

Run all tests:
```bash
bun run test
```

Run specific test suites:
```bash
bun run test:e2e           # E2E tests
bun run test:integration   # Integration tests
bun run test:auth          # Auth tests only
```

Run tests in UI mode (recommended for development):
```bash
bun run test:ui
```

### 6. View Results

After tests complete, view the HTML report:
```bash
bun run test:report
```

Test results are also saved to:
- `test-results/` - Screenshots, videos, traces
- `playwright-report/` - HTML report
- `test-results/results.json` - JSON report

## Test Data Management

### Cleanup Test Data

The tests automatically clean up most test data. However, you may want to periodically clean up:

```sql
-- Delete test products (careful with production!)
DELETE FROM products WHERE sku LIKE 'TEST-%' OR sku LIKE 'UI-TEST-%';

-- Delete test orders
DELETE FROM orders WHERE created_by IN (
  SELECT id FROM users WHERE email LIKE '%@test.com'
);
```

### Mock Data

Some tests create products, orders, and other entities. These are automatically cleaned up in the `afterAll` hooks, but manual cleanup may be needed if tests fail unexpectedly.

## Troubleshooting

### Tests Timeout

If tests timeout, check:
1. Backend server is running: `curl http://localhost:8080/api/v1/health`
2. Frontend server is running: `curl http://localhost:3000`
3. Database is accessible
4. Network connectivity is stable

Increase timeouts in `playwright.config.ts` if needed:
```typescript
timeout: 90 * 1000, // Increase to 90 seconds
```

### Authentication Failures

If auth tests fail:
1. Verify test users exist in database
2. Check password hashes are correct
3. Verify JWT signing keys match
4. Check CORS settings on backend

### Database Conflicts

If you get duplicate key errors:
1. Clean up test data manually (see above)
2. Use unique identifiers with timestamps
3. Ensure proper cleanup in `afterAll` hooks

### Port Conflicts

If ports 3000 or 8080 are in use:
1. Stop other services using these ports
2. Update `.env.local` with different ports
3. Update `playwright.config.ts` base URL

### Browser Installation

If browsers aren't installed:
```bash
bunx playwright install --with-deps
```

For specific browsers:
```bash
bunx playwright install chromium
bunx playwright install firefox
bunx playwright install webkit
```

### CI/CD Setup

For running tests in CI/CD:

```yaml
# GitHub Actions example
- name: Install dependencies
  run: bun install

- name: Install Playwright Browsers
  run: bunx playwright install --with-deps

- name: Start backend
  run: |
    cd ..
    ./run_start.sh &
    sleep 10

- name: Run tests
  run: bun run test

- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v3
  with:
    name: playwright-report
    path: playwright-report/
```

## Running Specific Tests

### By file
```bash
bunx playwright test tests/e2e/auth.spec.ts
```

### By test name
```bash
bunx playwright test -g "should login successfully"
```

### By project (browser)
```bash
bunx playwright test --project=chromium
bunx playwright test --project=firefox
bunx playwright test --project=webkit
```

### In headed mode (see browser)
```bash
bunx playwright test --headed
```

### With debugging
```bash
bunx playwright test --debug
```

### Step-by-step debugging
```bash
bunx playwright test --debug tests/e2e/auth.spec.ts:25
```

## Test Configuration

Edit `playwright.config.ts` to customize:
- Timeouts
- Number of retries
- Number of workers
- Screenshot/video settings
- Browser configurations
- Base URL

## Performance Testing

For performance tests, ensure:
1. Backend has realistic data volume
2. Network conditions are stable
3. No other heavy processes are running
4. Run tests multiple times for consistency

## Accessibility Testing

The accessibility tests check:
- Keyboard navigation
- Screen reader compatibility
- ARIA labels
- Color contrast
- Semantic HTML

For deeper accessibility auditing, consider:
- axe-core integration
- Lighthouse CI
- Manual testing with screen readers

## Next Steps

1. Review test results
2. Fix any failing tests
3. Add tests for new features
4. Integrate into CI/CD pipeline
5. Set up automated test runs

## Getting Help

- Review test logs: `cat test-results/results.json`
- Check screenshots: `test-results/**/*.png`
- View traces: `bunx playwright show-trace test-results/**/*.zip`
- Read Playwright docs: https://playwright.dev
