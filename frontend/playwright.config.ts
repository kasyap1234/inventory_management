import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for inventory management system tests
 * Supports UI tests, integration tests, and API tests
 */
export default defineConfig({
  testDir: './tests',
  testIgnore: ['**/unit/**', '**/integration/**'],

  // Maximum time one test can run
  timeout: 30 * 1000,

  // Global timeout for entire test run
  globalTimeout: 10 * 60 * 1000,

  // Run tests in files in parallel
  fullyParallel: true,

  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,

  // Reduce retries
  retries: process.env.CI ? 1 : 0,

  // Increase workers on CI for parallel execution
  workers: process.env.CI ? 2 : undefined,

  // Reporter to use
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['list']
  ],

  // Shared settings for all the projects
  use: {
    // Base URL for tests
    baseURL: process.env.NEXT_PUBLIC_FRONTEND_URL || 'http://localhost:3000',

    // Collect trace when retrying the failed test
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Reduce timeouts
    actionTimeout: 10 * 1000,

    // Navigation timeout
    navigationTimeout: 15 * 1000,
  },

  // Only run Chromium in CI
  projects: process.env.CI ? [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ] : [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    // Mobile viewports
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  // Run your local dev server before starting the tests
  webServer: {
    command: 'bun run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 60 * 1000,
  },
});
