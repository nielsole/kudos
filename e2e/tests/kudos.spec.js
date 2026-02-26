// @ts-check
const { test, expect } = require('@playwright/test');
const { execSync } = require('child_process');

const TEST_URL = 'e2e-test.example.com/test';
const BACKEND_URL = 'http://localhost:8082';
const REDIS_PORT = 6380;

// Complete a full hover gesture that lasts longer than the widget's duration
// (100 ms in the fixture) so that kudos are actually given.
async function giveKudos(page) {
  const kudos = page.locator('.kudos');
  await kudos.hover();
  await page.waitForTimeout(300);
  await expect(kudos).toHaveClass(/finish/, { timeout: 1000 });
}

test.describe('Kudos widget – end-to-end', () => {
  test.beforeEach(async ({ page }) => {
    // Wipe the test URL's Redis key so each test starts with count = 0
    try {
      execSync(
        `redis-cli -p ${REDIS_PORT} del "${TEST_URL}"`,
        { stdio: 'ignore' }
      );
    } catch (_) {}

    await page.goto('/e2e-fixture.html');
  });

  // ------------------------------------------------------------------
  // Test 1: Initial state
  // ------------------------------------------------------------------
  test('widget renders with zero count initially', async ({ page }) => {
    const kudos = page.locator('.kudos');

    await expect(kudos).toBeVisible();
    // No kudos given yet – should not be in finished or active state
    await expect(kudos).not.toHaveClass(/finish/);
    await expect(kudos).not.toHaveClass(/active/);
  });

  // ------------------------------------------------------------------
  // Test 2: Hovering triggers the active animation class
  // ------------------------------------------------------------------
  test('hovering triggers the active class', async ({ page }) => {
    const kudos = page.locator('.kudos');

    await kudos.hover();
    // The widget adds `active` synchronously on mouseover
    await expect(kudos).toHaveClass(/active/, { timeout: 500 });
  });

  // ------------------------------------------------------------------
  // Test 3: Moving out before the timer fires cancels the gesture
  // ------------------------------------------------------------------
  test('moving out before timer cancels the gesture', async ({ page }) => {
    const kudos = page.locator('.kudos');

    await kudos.hover();
    await expect(kudos).toHaveClass(/active/, { timeout: 500 });

    // Move the mouse away from the widget
    await page.mouse.move(0, 0);

    await expect(kudos).not.toHaveClass(/active/, { timeout: 500 });
    await expect(kudos).not.toHaveClass(/finish/);
  });

  // ------------------------------------------------------------------
  // Test 4: Completing the hover gesture transitions to the finish state
  //         and fires a POST request to the backend
  // ------------------------------------------------------------------
  test('full hover gives kudos and transitions to finish state', async ({ page }) => {
    const kudos = page.locator('.kudos');

    // Intercept the POST that the widget fires after the gesture completes
    const postPromise = page.waitForRequest(
      (req) => req.url().includes('localhost:8082') && req.method() === 'POST'
    );

    await kudos.hover();
    // Wait well beyond the 100 ms duration used in the fixture
    await page.waitForTimeout(300);

    // The POST must have been fired with the correct URL parameter
    const postReq = await postPromise;
    expect(postReq.url()).toContain(encodeURIComponent(TEST_URL));

    // Widget transitions: active → (removed) finish
    await expect(kudos).not.toHaveClass(/active/);
    await expect(kudos).toHaveClass(/finish/, { timeout: 500 });
  });

  // ------------------------------------------------------------------
  // Test 5: Backend count increments in Redis
  // ------------------------------------------------------------------
  test('backend count increments in Redis after kudos', async ({ page, request }) => {
    // Verify starting count is 0
    const before = await request.get(
      `${BACKEND_URL}/?url=${encodeURIComponent(TEST_URL)}`
    );
    expect((await before.text()).trim()).toBe('0');

    await giveKudos(page);

    // Redis count should now be 1
    const after = await request.get(
      `${BACKEND_URL}/?url=${encodeURIComponent(TEST_URL)}`
    );
    expect((await after.text()).trim()).toBe('1');
  });

  // ------------------------------------------------------------------
  // Test 6: Reloading the page shows the finish state (localStorage + backend)
  // ------------------------------------------------------------------
  test('reloading after kudos shows finish state', async ({ page }) => {
    await giveKudos(page);

    // Reload – localStorage is preserved within the same browser context.
    // The widget reads localStorage on init and transitions straight to finish.
    await page.reload();

    await expect(page.locator('.kudos')).toHaveClass(/finish/, { timeout: 2000 });
  });

  // ------------------------------------------------------------------
  // Test 7: CSS transform is applied when hovering over a finished widget
  // ------------------------------------------------------------------
  test('hovering a finished widget applies the CSS scale transform', async ({ page }) => {
    await giveKudos(page);

    // First verify the CSS rule is correctly defined in the loaded stylesheet.
    // This is a reliable check independent of hover-state timing.
    const ruleTransform = await page.evaluate(() => {
      for (const sheet of document.styleSheets) {
        try {
          for (const rule of sheet.cssRules) {
            if (rule instanceof CSSStyleRule && rule.selectorText === '.kudos.finish:hover') {
              return rule.style.transform;
            }
          }
        } catch (_) {}
      }
      return null;
    });
    expect(ruleTransform).toContain('scale(1.25, 1.25)');
    expect(ruleTransform).toContain('translateY');

    // Then verify that hovering the finished widget produces the scaled transform.
    // The .kudos.finish transition for transform is 0.5625s; wait for it to settle.
    const kudos = page.locator('.kudos');
    await kudos.hover();
    await page.waitForTimeout(700);

    const computedTransform = await page.evaluate(() => {
      const el = document.querySelector('.kudos.finish');
      return el ? getComputedStyle(el).transform : 'none';
    });
    // scale(1.25, 1.25) is represented in a CSS matrix as matrix(1.25, 0, 0, 1.25, ...)
    expect(computedTransform).toMatch(/matrix\(1\.25/);
  });

  // ------------------------------------------------------------------
  // Test 8: Widget loads directly in finish state when localStorage is set
  // ------------------------------------------------------------------
  test('widget starts in finish state when localStorage marks URL as voted', async ({ page }) => {
    // Register an init script that sets localStorage before the widget JS runs.
    // Since page.addInitScript() runs on every subsequent navigation, and
    // beforeEach already navigated once (without this script), we need a reload.
    await page.addInitScript((url) => {
      window.localStorage.setItem(`kudos:saved:${url}`, '1');
    }, TEST_URL);

    await page.reload();

    // The widget reads localStorage in its constructor and skips to finish state
    await expect(page.locator('.kudos')).toHaveClass(/finish/, { timeout: 2000 });
  });
});
