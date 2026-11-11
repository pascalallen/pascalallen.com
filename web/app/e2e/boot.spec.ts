import { readdirSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { expect, lastEntry, open, ready, run, test } from './fixtures';

const writing = resolve(__dirname, '../../../internal/pascalallen/domain/content/home/writing');

const publicationURLs = (): string[] =>
  readdirSync(writing)
    .filter(name => name.endsWith('.md'))
    .map(name => {
      const match = readFileSync(resolve(writing, name), 'utf8').match(/https:\/\/\S+/);
      if (match === null) {
        throw new Error(`${name} has no link`);
      }
      return match[0];
    });

test.describe('without JavaScript', () => {
  test.use({ javaScriptEnabled: false });

  test('the boot screen, chips and the plain link are server-rendered @smoke', async ({ page }) => {
    await page.goto('/');

    await expect(page.locator('#output .entry-boot')).toContainText('Pascal Allen');
    await expect(page.locator('#terminal')).toHaveAttribute('data-state', 'loading');
    await expect(page.locator('.chips a.chip[data-run="help"]')).toHaveAttribute('href', '/#help');
    await expect(page.locator('#output .entry-boot a.run[data-run="about"]')).toHaveAttribute('href', '/#about');

    await page.locator('.chips a[href="/plain"]').click();

    await expect(page).toHaveURL(/\/plain$/);
    await expect(page.locator('main')).toContainText('Plain text version');
  });

  test('the plain view lists every publication', async ({ page }) => {
    await page.goto('/plain');

    const urls = publicationURLs();
    expect(urls.length).toBeGreaterThanOrEqual(18);
    for (const url of urls) {
      await expect(page.locator(`main a[href="${url}"]`).first()).toBeAttached();
    }
  });
});

test.describe('hand-over from the pre-render', () => {
  test('the live shell adopts the boot screen without re-rendering it @smoke', async ({ page, request }) => {
    const html = await (await request.get('/')).text();
    const serverPrompt = html.match(/<label class="prompt" id="prompt" for="input">([^<]*)<\/label>/)?.[1];

    await open(page);

    await expect(page.locator('#output > .entry')).toHaveCount(1);
    await expect(page.locator('#output > .entry-boot[data-prerendered]')).toHaveCount(1);
    await expect(page.locator('#output .reveal')).toHaveCount(0);
    expect(await page.locator('#prompt').textContent()).toBe(serverPrompt);
  });

  test('every host asset is requested with its content hash and the module only once', async ({ page }) => {
    const assets: string[] = [];
    page.on('request', request => {
      const url = new URL(request.url());
      if (url.pathname.startsWith('/static/assets/')) {
        assets.push(`${url.pathname}${url.search}`);
      }
    });

    await open(page);
    await page.waitForLoadState('networkidle');

    for (const name of ['theme.js', 'app.css', 'app.js', 'wasm_exec.js', 'shell.wasm']) {
      const requested = assets.filter(path => path.startsWith(`/static/assets/${name}`));
      expect(requested.length, name).toBeGreaterThan(0);
      expect(
        requested.every(path => /\?v=[0-9a-f]{12}$/.test(path)),
        name
      ).toBe(true);
    }
    expect(assets.filter(path => path.startsWith('/static/assets/shell.wasm'))).toHaveLength(1);
  });

  test('a pre-rendered run link runs its command in place', async ({ page }) => {
    await open(page);

    await page.locator('#output .entry-boot a.run[data-run="about"]').click();

    await expect(page).toHaveURL(/#about$/);
    await expect(lastEntry(page).locator('.echo .cmd')).toHaveText('about');
    await expect(page.locator('#output > .entry')).toHaveCount(2);
  });

  test('lines entered before the shell is ready run in order afterwards', async ({ page }) => {
    let release: () => void = () => undefined;
    const held = new Promise<void>(done => {
      release = done;
    });
    await page.route('**/shell.wasm*', async route => {
      await held;
      await route.continue();
    });
    await page.goto('/');
    await expect(page.locator('#terminal')).toHaveAttribute('data-state', 'loading');
    await page.waitForFunction(() => document.readyState === 'complete');

    await run(page, 'whoami');
    await run(page, 'pwd');
    release();
    await ready(page);

    await expect(page.locator('#output .echo .cmd')).toHaveText(['whoami', 'pwd']);
    await expect(lastEntry(page)).toContainText('/home/visitor');
  });

  test('an unknown path boots the terminal with noindex', async ({ page }) => {
    const response = await page.goto('/no/such/page');

    expect(response?.status()).toBe(200);
    await expect(page.locator('meta[name="robots"]')).toHaveAttribute('content', /noindex/);
    await ready(page);
    await run(page, 'whoami');
    await expect(lastEntry(page)).toContainText('visitor');
  });
});

test.describe('when the shell cannot load', () => {
  test.use({ allowedConsole: [/Failed to load resource/] });

  test('the page explains, links to the plain view and disables the input @smoke', async ({ page }) => {
    await page.route('**/shell.wasm*', route => route.abort());

    await page.goto('/');

    await expect(page.locator('#terminal')).toHaveAttribute('data-state', 'failed');
    await expect(page.locator('#output .entry-boot')).toContainText('Pascal Allen');
    const notice = lastEntry(page).locator('.block-error');
    await expect(notice).toContainText('terminal could not start');
    await expect(notice.locator('a[href="/plain"]')).toBeVisible();
    await expect(page.locator('#input')).toBeDisabled();

    await notice.locator('a[href="/plain"]').click();
    await expect(page).toHaveURL(/\/plain$/);
  });

  test('a tapped chip or run link lands on the plain page instead of doing nothing', async ({ page }) => {
    await page.route('**/shell.wasm*', route => route.abort());

    await page.goto('/');
    await expect(page.locator('#terminal')).toHaveAttribute('data-state', 'failed');
    await page.locator('.chips .chip[data-run="about"]').click();

    await expect(page).toHaveURL(/\/plain$/);
    await expect(page.locator('main')).toContainText('Plain text version');
  });

  test('a run link in the boot screen lands on the plain page', async ({ page }) => {
    await page.route('**/shell.wasm*', route => route.abort());

    await page.goto('/');
    await expect(page.locator('#terminal')).toHaveAttribute('data-state', 'failed');
    await page.locator('#output .entry-boot a.run[data-run="projects"]').click();

    await expect(page).toHaveURL(/\/plain$/);
  });
});
