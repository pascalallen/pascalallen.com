import AxeBuilder from '@axe-core/playwright';
import { expect, open, test } from './fixtures';

const BUDGET = 250 * 1024;

const FIRST_LOAD = [
  { name: 'HTML', match: (path: string) => path === '/' },
  { name: 'CSS', match: (path: string) => path === '/static/assets/app.css' },
  { name: 'theme JS', match: (path: string) => path === '/static/assets/theme.js' },
  { name: 'app JS', match: (path: string) => path === '/static/assets/app.js' },
  { name: 'runtime JS', match: (path: string) => path === '/static/assets/wasm_exec.js' },
  { name: 'WASM', match: (path: string) => path === '/static/assets/shell.wasm' }
];

for (const path of ['/', '/plain']) {
  test(`${path} has no serious or critical accessibility violations`, async ({ page }) => {
    if (path === '/') {
      await open(page);
    } else {
      await page.goto(path);
    }

    const results = await new AxeBuilder({ page }).analyze();
    const severe = results.violations
      .filter(v => v.impact === 'serious' || v.impact === 'critical')
      .map(v => `${v.id}: ${v.nodes.map(n => n.target.join(' ')).join(', ')}`);

    expect(severe).toEqual([]);
  });
}

test.describe('budgets', () => {
  test.skip(({ browserName }) => browserName !== 'chromium', 'transfer sizes and layout shift are read from Chromium');

  test('the first load transfers under 250 KB with brotli', async ({ page }, testInfo) => {
    const sizes = new Map<string, { bytes: number; encoding: string }>();
    page.on('requestfinished', async request => {
      const url = new URL(request.url());
      const entry = FIRST_LOAD.find(item => item.match(url.pathname));
      if (entry === undefined) {
        return;
      }
      const response = await request.response();
      const measured = await request.sizes();
      const previous = sizes.get(entry.name)?.bytes ?? 0;
      sizes.set(entry.name, {
        bytes: previous + measured.responseBodySize + measured.responseHeadersSize,
        encoding: (await response?.headerValue('content-encoding')) ?? 'identity'
      });
    });

    await open(page);
    await page.waitForLoadState('networkidle');

    const rows = FIRST_LOAD.map(item => ({
      asset: item.name,
      ...(sizes.get(item.name) ?? { bytes: 0, encoding: '-' })
    }));
    const total = rows.reduce((sum, row) => sum + row.bytes, 0);
    const table = [...rows, { asset: 'total', bytes: total, encoding: '' }]
      .map(row => `${row.asset.padEnd(12)} ${String(row.bytes).padStart(8)} B  ${row.encoding}`)
      .join('\n');
    console.log(`first-load transfer\n${table}`);
    await testInfo.attach('transfer-sizes', { body: table, contentType: 'text/plain' });

    expect(sizes.get('WASM')?.encoding).toBe('br');
    expect(sizes.get('runtime JS')?.encoding).toBe('br');
    expect(total).toBeLessThan(BUDGET);
  });

  test('the shell becomes ready without layout shift', async ({ page }, testInfo) => {
    await page.addInitScript(() => {
      const w = window as unknown as { __readyAt?: number };
      new MutationObserver(() => {
        if (w.__readyAt === undefined && document.getElementById('terminal')?.dataset.state === 'ready') {
          w.__readyAt = performance.now();
        }
      }).observe(document, { attributes: true, subtree: true, attributeFilter: ['data-state'] });
    });

    await open(page);
    await page.waitForTimeout(500);

    const shift = await page.evaluate(
      () =>
        new Promise<number>(done => {
          let sum = 0;
          new PerformanceObserver(list => {
            for (const entry of list.getEntries() as unknown as { value: number; hadRecentInput: boolean }[]) {
              if (!entry.hadRecentInput) {
                sum += entry.value;
              }
            }
          }).observe({ type: 'layout-shift', buffered: true });
          setTimeout(() => done(sum), 100);
        })
    );
    const timing = await page.evaluate(() => {
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      const readyAt = (window as unknown as { __readyAt?: number }).__readyAt ?? NaN;
      return { dcl: nav.domContentLoadedEventEnd, ready: readyAt };
    });
    const line = `DOMContentLoaded ${timing.dcl.toFixed(0)} ms, shell ready ${timing.ready.toFixed(0)} ms, interactive after DOMContentLoaded ${(timing.ready - timing.dcl).toFixed(0)} ms, layout shift ${shift.toFixed(4)}`;
    console.log(line);
    await testInfo.attach('timing', { body: line, contentType: 'text/plain' });

    expect(shift).toBeLessThan(0.02);
  });
});
