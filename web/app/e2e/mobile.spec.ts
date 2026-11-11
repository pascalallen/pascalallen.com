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
