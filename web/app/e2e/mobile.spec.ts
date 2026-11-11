import { expect, lastEntry, open, run, test } from './fixtures';

for (const viewport of [
  { width: 320, height: 568 },
  { width: 390, height: 844 }
]) {
  test.describe(`a ${viewport.width}x${viewport.height} touch screen`, () => {
    test.use({ viewport, isMobile: true, hasTouch: true });

    test('keeps the input visible and legible, never scrolls sideways and taps chips', async ({ page }) => {
      await open(page);
      await run(page, 'projects');

      const input = page.locator('#input');
      await expect(input).toBeInViewport();
      const fontSize = await input.evaluate(el => parseFloat(getComputedStyle(el).fontSize));
      expect(fontSize).toBeGreaterThanOrEqual(16);
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
      expect(overflow).toBeLessThanOrEqual(0);

      await page.locator('.chips .chip[data-run="help"]').tap();
      await expect(lastEntry(page).locator('.echo .cmd')).toHaveText('help');
      await expect(input).toBeInViewport();
    });
  });
}

test.describe('a 390x844 touch screen reading the boot screen', () => {
  test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

  test('wraps the command row inline, matches prompt and output text and keeps 40px tap targets', async ({ page }) => {
    await open(page);

    const rows = await page.locator('#output .entry-boot .line').evaluateAll(lines => {
      const links = lines
        .map(line => [...line.querySelectorAll('a.run')])
        .reduce((best, items) => (items.length > best.length ? items : best), [] as Element[]);
      const tops = new Set(links.map(link => Math.round(link.getBoundingClientRect().top / 4)));
      return { rows: tops.size, count: links.length, heights: links.map(link => link.getBoundingClientRect().height) };
    });
    expect(rows.count).toBeGreaterThanOrEqual(8);
    expect(rows.rows).toBeLessThanOrEqual(3);
    for (const height of rows.heights) {
      expect(height).toBeGreaterThanOrEqual(40);
    }

    const sizes = await page.evaluate(() => ({
      prompt: getComputedStyle(document.getElementById('input') as HTMLElement).fontSize,
      output: getComputedStyle(document.querySelector('#output .line') as HTMLElement).fontSize,
      label: getComputedStyle(document.getElementById('prompt') as HTMLElement).fontSize
    }));
    expect(sizes.prompt).toBe(sizes.output);
    expect(sizes.label).toBe(sizes.output);

    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    expect(overflow).toBeLessThanOrEqual(0);
  });

  test('the more row and ls output wrap inline', async ({ page }) => {
    await open(page);
    await run(page, 'about');
    const more = await lastEntry(page)
      .locator('a.run, button.run')
      .evaluateAll(items => ({
        count: items.length,
        rows: new Set(items.map(item => Math.round(item.getBoundingClientRect().top / 4))).size
      }));
    expect(more.count).toBeGreaterThanOrEqual(4);
    expect(more.rows).toBeLessThan(more.count);

    await run(page, 'ls -a');
    const ls = await lastEntry(page)
      .locator('.block > .line')
      .first()
      .evaluate(line => {
        const items = [...line.querySelectorAll('a, button')];
        return {
          count: items.length,
          rows: new Set(items.map(item => Math.round(item.getBoundingClientRect().top / 4))).size
        };
      });
    expect(ls.count).toBeGreaterThanOrEqual(4);
    expect(ls.rows).toBeLessThan(ls.count);
  });
});

test.describe('with reduced motion', () => {
  test.use({ reducedMotion: 'reduce' });

  test('no animation runs when output appears', async ({ page }) => {
    await open(page);

    await run(page, 'help');
    await expect(lastEntry(page).locator('.echo .cmd')).toHaveText('help');

    const animations = await page.evaluate(() => document.getAnimations().length);
    expect(animations).toBe(0);
    const names = await lastEntry(page)
      .locator('.block > .line')
      .evaluateAll(lines => lines.map(line => getComputedStyle(line).animationName));
    expect(names.every(name => name === 'none')).toBe(true);
  });
});
