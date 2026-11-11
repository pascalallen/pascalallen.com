import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { describe, it } from 'node:test';

const css = readFileSync('src/styles.css', 'utf8');
const template = readFileSync('../template/index.tmpl', 'utf8') + readFileSync('../template/head.tmpl', 'utf8');

describe('mobile input sizing', () => {
  it('sets the terminal body text to 16px on coarse pointers so prompt and output match and iOS does not zoom', () => {
    const block = /@media \(pointer: coarse\)\s*\{[^}]*\bbody\s*\{[^}]*font-size:\s*16px/.exec(css);

    assert.notEqual(block, null);
  });

  it('does not size the prompt line separately from the output', () => {
    assert.doesNotMatch(css, /\.prompt-line\s*\{[^}]*font-size/);
  });

  it('never turns gap spans into row breaks', () => {
    assert.doesNotMatch(css, /\.gap\s*\{[^}]*display:\s*block/);
  });

  it('gives run buttons a 40px tap target through padding on coarse pointers', () => {
    const rule = /@media \(pointer: coarse\)\s*\{[^@]*?\.run\s*\{[^}]*\}/.exec(css)?.[0] ?? '';

    assert.match(rule, /display:\s*inline-block/);
    assert.match(rule, /padding-block:/);
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
