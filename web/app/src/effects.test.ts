import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { applyTheme, fetchFeed, initialTheme, openUrl, recordVisit, THEME_KEY } from './effects';
import { dom, memoryStorage, throwingStorage } from './test-dom';

const json = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

describe('fetch effect', () => {
  it('gets the feed endpoint and returns its items as strings', async () => {
    const urls: string[] = [];
    const fetchImpl = (async (url: string) => {
      urls.push(url);
      return json(200, {
        status: 'success',
        data: {
          items: [{ name: 'pgqueue', description: 'queue', url: 'https://github.com/pascalallen/pgqueue', meta: 7 }]
        }
      });
    }) as unknown as typeof fetch;

    const items = await fetchFeed('repos', fetchImpl, 1000);

    assert.deepEqual(urls, ['/api/v1/repos']);
    assert.deepEqual(items, [
      { name: 'pgqueue', description: 'queue', url: 'https://github.com/pascalallen/pgqueue', meta: '7' }
    ]);
  });

  it('maps packages to the packages endpoint', async () => {
    const urls: string[] = [];
    const fetchImpl = (async (url: string) => {
      urls.push(url);
      return json(200, { status: 'success', data: { items: [] } });
    }) as unknown as typeof fetch;

    await fetchFeed('packages', fetchImpl, 1000);

    assert.deepEqual(urls, ['/api/v1/packages']);
  });

  it('rejects an unknown feed without fetching', async () => {
    let called = false;
    const fetchImpl = (async () => {
      called = true;
      return json(200, {});
    }) as unknown as typeof fetch;

    await assert.rejects(fetchFeed('../visits', fetchImpl, 1000));
    assert.equal(called, false);
  });

  it('rejects on an HTTP error', async () => {
    const fetchImpl = (async () =>
      json(502, { status: 'error', message: 'feed unavailable' })) as unknown as typeof fetch;

    await assert.rejects(fetchFeed('repos', fetchImpl, 1000), /HTTP 502/);
  });

  it('rejects a JSend body that is not a success', async () => {
    const fetchImpl = (async () => json(200, { status: 'fail', data: { items: [] } })) as unknown as typeof fetch;

    await assert.rejects(fetchFeed('repos', fetchImpl, 1000));
  });

  it('aborts and rejects when the request times out', async () => {
    let aborted = false;
    const fetchImpl = ((_url: string, init: RequestInit) =>
      new Promise((_resolve, reject) => {
        init.signal?.addEventListener('abort', () => {
          aborted = true;
          reject(new Error('aborted'));
        });
      })) as unknown as typeof fetch;

    await assert.rejects(fetchFeed('repos', fetchImpl, 10));
    assert.equal(aborted, true);
  });
});

describe('theme effect', () => {
  it('sets data-theme on html and persists it', () => {
    const doc = dom().window.document;
    const storage = memoryStorage();

    assert.equal(
      applyTheme(doc, 'amber', () => storage),
      true
    );
    assert.equal(doc.documentElement.getAttribute('data-theme'), 'amber');
    assert.equal(storage.getItem(THEME_KEY), 'amber');
  });

  it('still applies the theme when localStorage throws', () => {
    const doc = dom().window.document;

    assert.equal(
      applyTheme(doc, 'green', () => throwingStorage()),
      true
    );
    assert.equal(
      applyTheme(doc, 'light', () => {
        throw new Error('SecurityError');
      }),
      true
    );
    assert.equal(doc.documentElement.getAttribute('data-theme'), 'light');
  });

  it('ignores an unknown theme', () => {
    const doc = dom().window.document;
    const storage = memoryStorage();

    assert.equal(
      applyTheme(doc, 'neon', () => storage),
      false
    );
    assert.equal(doc.documentElement.hasAttribute('data-theme'), false);
    assert.equal(storage.getItem(THEME_KEY), null);
  });

  it('follows prefers-color-scheme unless a theme is stored', () => {
    const storage = memoryStorage();

    assert.equal(
      initialTheme(() => storage, true),
      'light'
    );
    assert.equal(
      initialTheme(() => storage, false),
      'dark'
    );
    storage.setItem(THEME_KEY, 'amber');
    assert.equal(
      initialTheme(() => storage, true),
      'amber'
    );
    storage.setItem(THEME_KEY, 'neon');
    assert.equal(
      initialTheme(() => storage, false),
      'dark'
    );
    assert.equal(
      initialTheme(() => throwingStorage(), true),
      'light'
    );
  });
});

describe('open effect', () => {
  it('opens http, https and mailto links in a new tab without an opener', () => {
    const opened: unknown[][] = [];
    const open = (...args: unknown[]) => {
      opened.push(args);
      return null;
    };

    assert.equal(openUrl('https://github.com/pascalallen', open), true);
    assert.deepEqual(opened, [['https://github.com/pascalallen', '_blank', 'noopener']]);
  });

  it('refuses other schemes', () => {
    const opened: unknown[] = [];
    const open = (...args: unknown[]) => {
      opened.push(args);
      return null;
    };

    assert.equal(openUrl('javascript:alert(1)', open), false);
    assert.equal(opened.length, 0);
  });
});

describe('visit', () => {
  it('posts language and user agent once per session', async () => {
    const posts: { url: string; init: RequestInit }[] = [];
    const fetchImpl = (async (url: string, init: RequestInit) => {
      posts.push({ url, init });
      return new Response(null, { status: 204 });
    }) as unknown as typeof fetch;
    const storage = memoryStorage();
    const nav = { language: 'en-US', userAgent: 'test-agent' };

    recordVisit(fetchImpl, () => storage, nav);
    recordVisit(fetchImpl, () => storage, nav);

    assert.equal(posts.length, 1);
    assert.equal(posts[0].url, '/api/v1/visits');
    assert.equal(posts[0].init.method, 'POST');
    assert.deepEqual(JSON.parse(String(posts[0].init.body)), { language: 'en-US', userAgent: 'test-agent' });
  });

  it('never throws when the request or storage fails', async () => {
    const fetchImpl = (async () => {
      throw new Error('offline');
    }) as unknown as typeof fetch;

    assert.doesNotThrow(() => recordVisit(fetchImpl, () => throwingStorage(), { language: 'en', userAgent: 'x' }));
  });
});
