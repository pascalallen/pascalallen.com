import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { describe, it } from 'node:test';

const css = readFileSync('src/styles.css', 'utf8');
const template = readFileSync('../template/index.tmpl', 'utf8');

describe('mobile input sizing', () => {
  it('keeps the prompt line at 16px or more on coarse pointers so iOS does not zoom', () => {
    const block = /@media \(pointer: coarse\)\s*\{[^}]*\.prompt-line\s*\{[^}]*font-size:\s*max\(16px, 1em\)/.exec(css);

    assert.notEqual(block, null);
  });

  it('lets the input inherit the prompt line font size', () => {
    const rule = /\.input\s*\{[^}]*\}/.exec(css)?.[0] ?? '';

    assert.match(rule, /font:\s*inherit/);
  });

  it('does not disable pinch zoom in the viewport meta', () => {
    const meta = /<meta name="viewport"[^>]*>/.exec(template)?.[0] ?? '';

    assert.notEqual(meta, '');
    assert.doesNotMatch(meta, /maximum-scale|user-scalable/);
  });
});
