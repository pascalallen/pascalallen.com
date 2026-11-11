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
  test('the first load transfers under 250 KB, requests the module once and compresses every text asset @smoke', async ({
    page,
    browserName
  }, testInfo) => {
    const encodings = new Map<string, string>();
    const requests = new Map<string, number>();
    page.on('request', request => {
      const entry = FIRST_LOAD.find(item => item.match(new URL(request.url()).pathname));
      if (entry !== undefined) {
        requests.set(entry.name, (requests.get(entry.name) ?? 0) + 1);
      }
    });
    page.on('response', async response => {
      const entry = FIRST_LOAD.find(item => item.match(new URL(response.url()).pathname));
      if (entry !== undefined) {
        encodings.set(entry.name, (await response.headerValue('content-encoding')) ?? 'identity');
      }
    });

    await open(page);
    await page.waitForLoadState('networkidle');

    const transferred = await page.evaluate(() =>
      performance.getEntriesByType('resource').map(entry => ({
        path: new URL(entry.name).pathname,
        bytes: (entry as PerformanceResourceTiming).transferSize
      }))
    );
    const nav = await page.evaluate(
      () => (performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming).transferSize
    );
    const rows = FIRST_LOAD.map(item => ({
      asset: item.name,
      bytes:
        item.name === 'HTML'
          ? nav
          : transferred.filter(entry => item.match(entry.path)).reduce((sum, entry) => sum + entry.bytes, 0),
      requests: requests.get(item.name) ?? 0,
      encoding: encodings.get(item.name) ?? '-'
    }));
    const total = rows.reduce((sum, row) => sum + row.bytes, 0);
    const table = [...rows, { asset: 'total', bytes: total, requests: 0, encoding: '' }]
      .map(
        row =>
          `${row.asset.padEnd(12)} ${String(row.bytes).padStart(8)} B  ${String(row.requests).padStart(2)} req  ${row.encoding}`
      )
      .join('\n');
    console.log(`first-load transfer (${browserName})\n${table}`);
    await testInfo.attach('transfer-sizes', { body: table, contentType: 'text/plain' });

    expect(requests.get('WASM'), 'shell.wasm is requested exactly once per page load').toBe(1);
    expect(transferred.filter(entry => entry.path === '/static/assets/shell.wasm')).toHaveLength(1);
    const compressed = browserName === 'chromium' ? ['br'] : ['br', 'gzip'];
    for (const name of ['CSS', 'theme JS', 'app JS', 'runtime JS', 'WASM']) {
      expect(compressed, `${name} is compressed`).toContain(encodings.get(name));
    }
    if (browserName === 'chromium') {
      expect(encodings.get('app JS')).toBe('br');
    }
    expect(total).toBeLessThan(BUDGET);
  });

  test('the shell becomes ready without layout shift', async ({ page, browserName }, testInfo) => {
    test.skip(browserName !== 'chromium', 'layout shift entries are read from Chromium');
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
