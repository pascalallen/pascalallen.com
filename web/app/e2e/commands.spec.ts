import { expect, lastEntry, open, run, test } from './fixtures';

test('help lists the commands @smoke', async ({ page }) => {
  await open(page);

  await run(page, 'help');

  const entry = lastEntry(page);
  await expect(entry.locator('.echo .cmd')).toHaveText('help');
  for (const name of ['cat', 'ls', 'repos', 'theme']) {
    await expect(entry).toContainText(name);
  }
  await expect(page.locator('#input')).toHaveValue('');
});

test('ls, cd and cat walk the tree and render safe links @smoke', async ({ page }) => {
  await open(page);

  await run(page, 'ls');
  await expect(lastEntry(page)).toContainText('writing');

  await run(page, 'cd writing');
  await expect(page.locator('#prompt')).toHaveText('visitor@pascalallen.com:~/writing$');

  await run(page, 'cat 01-run-a-go-web-server-with-apple-container.md');
  const link = lastEntry(page).locator('a[href^="https://medium.com/"]');
  await expect(link).toHaveCount(1);
  await expect(link).toHaveAttribute('target', '_blank');
  await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
});

test('Tab completes a single match and lists several', async ({ page }) => {
  await open(page);
  const input = page.locator('#input');

  await input.fill('whoa');
  await input.press('Tab');
  await expect(input).toHaveValue('whoami ');

  await input.fill('c');
  await input.press('Tab');
  const listing = lastEntry(page);
  for (const name of ['cat', 'cd', 'clear']) {
    await expect(listing).toContainText(name);
  }
  await expect(input).toHaveValue(/^c/);
});

test('Up and Down walk history and keep the draft', async ({ page }) => {
  await open(page);
  const input = page.locator('#input');
  await run(page, 'whoami');
  await run(page, 'pwd');

  await input.fill('dra');
  await input.press('ArrowUp');
  await expect(input).toHaveValue('pwd');
  await input.press('ArrowUp');
  await expect(input).toHaveValue('whoami');
  await input.press('ArrowDown');
  await expect(input).toHaveValue('pwd');
  await input.press('ArrowDown');
  await expect(input).toHaveValue('dra');
});

test('Ctrl-L clears the screen', async ({ page }) => {
  await open(page);
  await run(page, 'whoami');

  await page.locator('#input').press('Control+l');

  await expect(page.locator('#output > *')).toHaveCount(0);
});

test('a chip runs its command @smoke', async ({ page }) => {
  await open(page);

  await page.locator('.chips .chip[data-run="about"]').click();

  await expect(lastEntry(page).locator('.echo .cmd')).toHaveText('about');
  await expect(page).toHaveURL(/#about$/);
});

test.describe('deep links', () => {
  test('a read-only command runs on load', async ({ page }) => {
    await open(page, '/#cat%20about.md');

    await expect(lastEntry(page).locator('.echo .cmd')).toHaveText('cat about.md');
    await expect(page.locator('#input')).toHaveValue('');
  });

  test('theme is only placed in the input', async ({ page }) => {
    await open(page, '/#theme%20amber');

    await expect(page.locator('#input')).toHaveValue('theme amber');
    await expect(page.locator('#output .echo')).toHaveCount(0);
    await expect(page.locator('html')).not.toHaveAttribute('data-theme', 'amber');
  });

  test('open is only placed in the input and opens nothing', async ({ page, context }) => {
    const pages: unknown[] = [];
    context.on('page', p => pages.push(p));

    await open(page, '/#open%20https%3A%2F%2Fexample.com');

    await expect(page.locator('#input')).toHaveValue('open https://example.com');
    await page.waitForTimeout(300);
    expect(pages).toHaveLength(0);
  });
});

test('markup typed into echo stays text', async ({ page }) => {
  await open(page);

  await run(page, 'echo <img src=x onerror=alert(1)>');

  await expect(lastEntry(page).locator('.echo .cmd')).toHaveText('echo <img src=x onerror=alert(1)>');
  await expect(page.locator('#output img')).toHaveCount(0);

  await run(page, 'echo "<img src=x onerror=alert(1)>"');

  await expect(lastEntry(page).locator('.block')).toContainText('<img src=x onerror=alert(1)>');
  await expect(page.locator('#output img')).toHaveCount(0);
});
