import type { Page } from '@playwright/test';
import { expect, lastEntry, open, run, test } from './fixtures';

const lineHeightOf = (page: Page): Promise<number> =>
  page.locator('#input').evaluate(el => parseFloat(getComputedStyle(el).lineHeight));

test.describe('a 1440px desktop', () => {
  test.use({ viewport: { width: 1440, height: 900 } });

  test('ls -l writing rows do not wrap and an about paragraph is capped near 80ch', async ({ page }) => {
    await open(page);
    await run(page, 'ls -l writing');
    const lineHeight = await lineHeightOf(page);
    const heights = await lastEntry(page)
      .locator('.block > .line')
      .evaluateAll(lines => lines.map(line => line.getBoundingClientRect().height));
    expect(heights.length).toBeGreaterThan(5);
    for (const height of heights) {
      expect(height).toBeLessThan(lineHeight * 1.5);
    }

    await run(page, 'about');
    const measured = await lastEntry(page)
      .locator('.line-prose')
      .last()
      .evaluate(line => {
        const probe = document.createElement('span');
        probe.style.setProperty('position', 'absolute');
        probe.style.setProperty('white-space', 'nowrap');
        probe.textContent = '0'.repeat(80);
        line.appendChild(probe);
        const eighty = probe.getBoundingClientRect().width;
        probe.remove();
        return { width: line.getBoundingClientRect().width, eighty };
      });
    expect(measured.width).toBeLessThanOrEqual(measured.eighty + 2);
    expect(measured.width).toBeGreaterThan(measured.eighty / 2);
  });

  test('skills rows are not clipped by the prose measure', async ({ page }) => {
    await open(page);
    await run(page, 'skills');
    const rows = lastEntry(page).locator('.block > .line');
    const clipped = await rows.evaluateAll(
      lines =>
        lines.filter(line => !line.classList.contains('line-prose') && line.scrollWidth > line.clientWidth + 1).length
    );
    expect(clipped).toBe(0);
    const indented = await lastEntry(page).locator('.block > .line[class*="indent-"]').count();
    expect(indented).toBeGreaterThan(3);
    expect(await lastEntry(page).locator('.line-prose').count()).toBeLessThan(await rows.count());
  });
});

test.describe('a 390px touch screen', () => {
  test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

  test('listed names never break inside a name', async ({ page }) => {
    await open(page);
    const input = page.locator('#input');
    await input.fill('cat ');
    await input.press('Tab');
    const lineHeight = await lineHeightOf(page);
    const items = lastEntry(page).locator('.s-path');
    expect(await items.count()).toBeGreaterThan(3);
    const heights = await items.evaluateAll(els => els.map(el => el.getBoundingClientRect().height));
    for (const height of heights) {
      expect(height).toBeLessThan(lineHeight * 1.5);
    }
    const rows = await items.evaluateAll(
      els => new Set(els.map(el => Math.round(el.getBoundingClientRect().top))).size
    );
    expect(rows).toBeGreaterThan(1);
  });

  test('experience entries share one left edge and wrapped lines keep their indent', async ({ page }) => {
    await open(page);
    await run(page, 'experience');
    const indented = lastEntry(page).locator('.line[class*="indent-"]');
    expect(await indented.count()).toBeGreaterThanOrEqual(6);

    const edges = await indented.evaluateAll(lines =>
      lines.map(line => ({
        left: Math.round((line.firstElementChild ?? line).getBoundingClientRect().left),
        padding: parseFloat(getComputedStyle(line).paddingLeft),
        startsWithSpace: (line.textContent ?? '').startsWith(' ')
      }))
    );
    for (const edge of edges) {
      expect(edge.padding).toBeGreaterThan(0);
      expect(edge.startsWithSpace).toBe(false);
    }
    const entryLefts = edges.filter(edge => edge.padding === edges[0].padding).map(edge => edge.left);
    expect(new Set(entryLefts).size).toBe(1);
  });

  test('a wrapped bullet line in resume hangs its continuation rows under the text after the dash', async ({
    page
  }) => {
    await open(page);
    await run(page, 'resume');
    const lineHeight = await lineHeightOf(page);
    const hang = await lastEntry(page)
      .locator('.line.bullet')
      .evaluateAll((lines, lh) => {
        for (const line of lines) {
          if (line.getBoundingClientRect().height < lh * 1.5) {
            continue;
          }
          const walker = document.createTreeWalker(line, NodeFilter.SHOW_TEXT);
          const node = walker.nextNode() as Text;
          const rect = (offset: number): DOMRect => {
            const range = document.createRange();
            range.setStart(node, offset);
            range.setEnd(node, offset + 1);
            return range.getBoundingClientRect();
          };
          const dash = rect(0);
          const text = rect(2);
          let wrapped: DOMRect | null = null;
          for (let i = 3; i < node.length; i++) {
            const r = rect(i);
            if (r.top > dash.top + lh / 2) {
              wrapped = r;
              break;
            }
          }
          if (wrapped) {
            return { dashLeft: dash.left, textLeft: text.left, wrappedLeft: wrapped.left };
          }
        }
        return null;
      }, lineHeight);
    expect(hang).not.toBeNull();
    expect(hang!.textLeft).toBeGreaterThan(hang!.dashLeft + 1);
    expect(hang!.wrappedLeft).toBeGreaterThan(hang!.dashLeft + 1);
    expect(Math.abs(hang!.wrappedLeft - hang!.textLeft)).toBeLessThan(2);
  });
});
