import { expect, Page, test as base } from '@playwright/test';

export interface Options {
  allowedConsole: RegExp[];
}

const BROWSER_NOTICES = [/^Viewport argument key "interactive-widget" not recognized and ignored\.$/];

interface Fixtures {
  brokenURL: string;
  guard: void;
}

export const test = base.extend<Options & Fixtures>({
  allowedConsole: [[], { option: true }],
  baseURL: async ({}, use) => {
    await use(process.env.E2E_BASE_URL);
  },
  brokenURL: async ({}, use) => {
    await use(process.env.E2E_BROKEN_URL ?? '');
  },
  guard: [
    async ({ context, allowedConsole, browserName }, use) => {
      const problems: string[] = [];
      await context.exposeBinding('__reportCsp', (_source, report: string) => {
        problems.push(`csp: ${report}`);
      });
      await context.addInitScript(() => {
        document.addEventListener('securitypolicyviolation', event => {
          (window as unknown as { __reportCsp(r: string): void }).__reportCsp(
            `${event.violatedDirective} ${event.blockedURI}`
          );
        });
      });
      const watch = (page: Page): void => {
        page.on('console', message => {
          const text = message.text();
          if (/content security policy/i.test(text)) {
            problems.push(`csp console: ${text}`);
          } else if (browserName === 'chromium' && /preload/i.test(text)) {
            problems.push(`preload: ${text}`);
          } else if (
            message.type() === 'error' &&
            ![...BROWSER_NOTICES, ...allowedConsole].some(pattern => pattern.test(text))
          ) {
            problems.push(`console error: ${text}`);
          }
        });
        page.on('pageerror', error => problems.push(`page error: ${error.message}`));
        page.on('dialog', dialog => {
          problems.push(`dialog: ${dialog.message()}`);
          dialog.dismiss().catch(() => undefined);
        });
      };
      context.pages().forEach(watch);
      context.on('page', watch);
      await use();
      expect(problems, 'no console errors, page errors, dialogs or CSP violations').toEqual([]);
    },
    { auto: true }
  ]
});

export { expect };

export const ready = async (page: Page): Promise<void> => {
  await expect(page.locator('#terminal')).toHaveAttribute('data-state', 'ready');
};

export const open = async (page: Page, path = '/'): Promise<void> => {
  await page.goto(path);
  await ready(page);
};

export const run = async (page: Page, line: string): Promise<void> => {
  const input = page.locator('#input');
  await input.fill(line);
  await input.press('Enter');
};

export const lastEntry = (page: Page) => page.locator('#output > .entry').last();
