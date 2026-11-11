export const THEMES = ['dark', 'light', 'amber', 'green'];

export const THEME_KEY = 'theme';

export type StorageAccess = () => Storage;

export const readTheme = (storage: StorageAccess): string | null => {
  try {
    const stored = storage().getItem(THEME_KEY);
    return stored !== null && THEMES.includes(stored) ? stored : null;
  } catch {
    return null;
  }
};

export const applyStoredTheme = (doc: Document, storage: StorageAccess): void => {
  const stored = readTheme(storage);
  if (stored !== null) {
    doc.documentElement.setAttribute('data-theme', stored);
  }
};
