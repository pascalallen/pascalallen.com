import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { applyStoredTheme } from './theme';
import { dom, memoryStorage, throwingStorage } from './test-dom';

describe('stored theme before first paint', () => {
  it('sets data-theme from a stored known theme', () => {
    const doc = dom().window.document;
    const storage = memoryStorage();
    storage.setItem('theme', 'amber');

    applyStoredTheme(doc, () => storage);

    assert.equal(doc.documentElement.getAttribute('data-theme'), 'amber');
  });

  it('leaves data-theme unset when nothing is stored', () => {
    const doc = dom().window.document;

    applyStoredTheme(doc, () => memoryStorage());

    assert.equal(doc.documentElement.hasAttribute('data-theme'), false);
  });

  it('ignores a stored value that is not a theme', () => {
    const doc = dom().window.document;
    const storage = memoryStorage();
    storage.setItem('theme', 'x" onload="alert(1)');

    applyStoredTheme(doc, () => storage);

    assert.equal(doc.documentElement.hasAttribute('data-theme'), false);
  });

  it('survives storage that throws', () => {
    const doc = dom().window.document;

    assert.doesNotThrow(() => applyStoredTheme(doc, () => throwingStorage()));
    assert.equal(doc.documentElement.hasAttribute('data-theme'), false);
  });
});
