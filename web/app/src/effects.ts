import type { FeedItem } from './bridge';
import { safeHref } from './render';
import { readTheme, StorageAccess, THEME_KEY, THEMES } from './theme';

export { readTheme, THEME_KEY, THEMES };
export type { StorageAccess };

const VISIT_KEY = 'visited';

const FEEDS: Record<string, string> = {
  repos: '/api/v1/repos',
  packages: '/api/v1/packages'
};

export const CAMERA_STREAM = '/api/v1/camera/stream';

const CAMERA_STATUS = '/api/v1/camera/status';

export type OpenImpl = (url: string, target: string, features: string) => unknown;

const read = (storage: StorageAccess, key: string): string | null => {
  try {
    return storage().getItem(key);
  } catch {
    return null;
  }
};

const write = (storage: StorageAccess, key: string, value: string): void => {
  try {
    storage().setItem(key, value);
  } catch {
    return;
  }
};

export const applyTheme = (doc: Document, name: string, storage: StorageAccess): boolean => {
  if (!THEMES.includes(name)) {
    return false;
  }
  doc.documentElement.setAttribute('data-theme', name);
  write(storage, THEME_KEY, name);
  return true;
};

export const initialTheme = (storage: StorageAccess, prefersLight: boolean): string =>
  readTheme(storage) ?? (prefersLight ? 'light' : 'dark');

export const cameraAvailable = async (fetchImpl: typeof fetch, timeoutMs: number): Promise<boolean> => {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const response = await fetchImpl(CAMERA_STATUS, {
      signal: controller.signal,
      headers: { Accept: 'application/json' }
    });
    if (!response.ok) {
      return false;
    }
    const body = (await response.json()) as { status?: unknown; data?: { available?: unknown } };
    return body.status === 'success' && body.data?.available === true;
  } catch {
    return false;
  } finally {
    clearTimeout(timer);
  }
};

export const openUrl = (url: string, open: OpenImpl): boolean => {
  const href = safeHref(url);
  if (href === null) {
    return false;
  }
  open(href, '_blank', 'noopener');
  return true;
};

const field = (value: unknown): string =>
  typeof value === 'string' ? value : typeof value === 'number' ? String(value) : '';

export const fetchFeed = async (feed: string, fetchImpl: typeof fetch, timeoutMs: number): Promise<FeedItem[]> => {
  const url = Object.prototype.hasOwnProperty.call(FEEDS, feed) ? FEEDS[feed] : undefined;
  if (url === undefined) {
    throw new Error(`unknown feed ${feed}`);
  }
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const response = await fetchImpl(url, { signal: controller.signal, headers: { Accept: 'application/json' } });
    if (!response.ok) {
      throw new Error(`${feed}: HTTP ${response.status}`);
    }
    const body = (await response.json()) as { status?: unknown; data?: { items?: unknown } };
    if (body.status !== 'success' || !Array.isArray(body.data?.items)) {
      throw new Error(`${feed}: unexpected response`);
    }
    return body.data.items.map((item: Record<string, unknown>) => ({
      name: field(item?.name),
      description: field(item?.description),
      url: field(item?.url),
      meta: field(item?.meta)
    }));
  } finally {
    clearTimeout(timer);
  }
};

export const recordVisit = (
  fetchImpl: typeof fetch,
  storage: StorageAccess,
  nav: { language: string; userAgent: string }
): void => {
  if (read(storage, VISIT_KEY) !== null) {
    return;
  }
  write(storage, VISIT_KEY, '1');
  try {
    fetchImpl('/api/v1/visits', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ language: nav.language, userAgent: nav.userAgent }),
      keepalive: true
    }).catch(() => undefined);
  } catch {
    return;
  }
};
