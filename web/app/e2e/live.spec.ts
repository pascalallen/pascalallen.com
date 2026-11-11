import { expect, lastEntry, open, ready, run, test } from './fixtures';
import { MARKUP_NAME } from './fixture-data';

test('repos lists repositories from the server with links @smoke', async ({ page }) => {
  await open(page);

  await run(page, 'repos');

  const entry = lastEntry(page);
  const link = entry.locator('a[href="https://github.com/pascalallen/fixture-queue"]');
  await expect(link).toHaveText('fixture-queue');
  await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  await expect(entry).toContainText('Go, 7 stars');
});

test('a feed item name with markup stays text', async ({ page }) => {
  await open(page);

  await run(page, 'repos');

  const entry = lastEntry(page);
  await expect(entry.locator('a[href="https://github.com/pascalallen/markup"]')).toHaveText(MARKUP_NAME);
  await expect(entry).toContainText('<script>alert(2)</script>');
  await expect(page.locator('#output img, #output script')).toHaveCount(0);
});

test('packages lists packages from the server with links', async ({ page }) => {
  await open(page);

  await run(page, 'packages');

  const entry = lastEntry(page);
  await expect(entry.locator('a[href="https://www.npmjs.com/package/@pascalallen/fixture-forms"]')).toHaveText(
    '@pascalallen/fixture-forms'
  );
  await expect(entry).toContainText('v1.2.3');
});

test.describe('when the upstream fails', () => {
  test.use({ allowedConsole: [/Failed to load resource: the server responded with a status of 502/] });

  test('repos and packages show the shell error text', async ({ page, brokenURL }) => {
    await open(page, `${brokenURL}/`);

    await run(page, 'repos');
    const failure = () => lastEntry(page).locator('.block-error .line').first();
    await expect(failure()).toHaveText('repos: could not load the list right now');

    await run(page, 'packages');
    await expect(failure()).toHaveText('packages: could not load the list right now');
  });
});

test('camera checks the status first and reports the camera offline', async ({ page }) => {
  const streams: string[] = [];
  page.on('request', request => {
    if (request.url().includes('/api/v1/camera/stream')) {
      streams.push(request.url());
    }
  });
  await open(page);

  const status = page.waitForResponse(response => response.url().endsWith('/api/v1/camera/status'));
  await run(page, 'camera');
  await status;

  await expect(lastEntry(page).locator('.block-error')).toContainText('camera is offline');
  await expect(page.locator('#output img')).toHaveCount(0);
  expect(streams).toEqual([]);
});

test('open opens a new tab when the visitor runs it', async ({ page, context }) => {
  await context.route('https://example.com/**', route =>
    route.fulfill({ status: 200, contentType: 'text/html', body: '<title>stub</title>' })
  );
  await open(page);

  const opened = context.waitForEvent('page');
  await run(page, 'open https://example.com');

  expect((await opened).url()).toBe('https://example.com/');
});

test.describe('theme', () => {
  test.use({ allowedConsole: [/Failed to load resource/] });

  test('a chosen theme persists across reload without a flash and agrees with the shell @smoke', async ({ page }) => {
    await open(page);
    await run(page, 'theme amber');
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'amber');

    await page.route('**/app.js*', route => route.abort());
    await page.reload();
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'amber');
    await page.unroute('**/app.js*');

    await page.reload();
    await ready(page);
    await run(page, 'theme');
    const marked = lastEntry(page).locator('.line', { hasText: '* ' });
    await expect(marked).toHaveCount(1);
    await expect(marked).toContainText('amber');
  });

  test.describe('with a light color scheme and nothing stored', () => {
    test.use({ colorScheme: 'light' });

    test('the shell reports the light theme the page shows', async ({ page }) => {
      await open(page);

      await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
      await run(page, 'theme');
      await expect(lastEntry(page).locator('.line', { hasText: '* ' })).toContainText('light');
    });
  });
});

test('the visit is recorded once per session', async ({ page }) => {
  const visits: string[] = [];
  page.on('request', request => {
    if (request.url().endsWith('/api/v1/visits') && request.method() === 'POST') {
      visits.push(request.url());
    }
  });

  await open(page);
  await page.reload();
  await ready(page);
  await page.waitForTimeout(200);

  expect(visits).toHaveLength(1);
});
