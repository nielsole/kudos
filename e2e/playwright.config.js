// @ts-check
const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './tests',
  globalSetup: './global-setup.js',
  globalTeardown: './global-teardown.js',
  use: {
    baseURL: 'http://localhost:3000',
    // Capture screenshots on failure for debugging
    screenshot: 'only-on-failure',
    // Video on first retry
    video: 'on-first-retry',
  },
  // Serve the frontend directory as a static file server
  webServer: {
    command: 'npx http-server ../frontend -p 3000 --cors -c-1 --silent',
    url: 'http://localhost:3000',
    reuseExistingServer: false,
    timeout: 15000,
  },
  // Run tests sequentially to avoid Redis key collisions
  workers: 1,
  // Retry once on CI
  retries: process.env.CI ? 1 : 0,
  reporter: [['list'], ['html', { open: 'never' }]],
});
