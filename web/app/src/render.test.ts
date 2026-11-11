import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { renderBlock, renderEcho, renderLine, renderResult, renderSpan } from './render';
import { block, dom, ORIGIN, result, span } from './test-dom';

describe('render', () => {
  it('keeps markup in span text as text and never creates elements from it', () => {
    const doc = dom().window.document;
    const el = renderSpan(doc, span('<img src=x onerror=alert(1)>'));

    assert.equal(el.querySelector('img'), null);
    assert.equal(el.children.length, 0);
    assert.equal(el.textContent, '<img src=x onerror=alert(1)>');
  });

  it('renders a javascript href as plain text without a link', () => {
    const doc = dom().window.document;
    const el = renderSpan(doc, span('click me', { style: 'accent', href: 'javascript:alert(1)' }));

    assert.equal(el.tagName, 'SPAN');
    assert.equal(el.hasAttribute('href'), false);
    assert.equal(el.textContent, 'click me');
  });

  it('renders a data href as plain text without a link', () => {
    const doc = dom().window.document;
    const el = renderSpan(doc, span('x', { href: 'data:text/html,<script>alert(1)</script>' }));

    assert.equal(el.tagName, 'SPAN');
  });

  it('renders https, http and mailto hrefs as links that open in a new tab without an opener', () => {
    const doc = dom().window.document;
    for (const href of ['https://github.com/pascalallen', 'http://example.com/', 'mailto:me@pascalallen.com']) {
      const el = renderSpan(doc, span(href, { style: 'accent', href }));

      assert.equal(el.tagName, 'A');
      assert.equal(el.getAttribute('href'), href);
      assert.equal(el.getAttribute('target'), '_blank');
      assert.equal(el.getAttribute('rel'), 'noopener noreferrer');
      assert.equal(el.className, 's-accent');
    }
  });

  it('renders a run span as a button carrying the command', () => {
    const doc = dom().window.document;
    const el = renderSpan(doc, span('about.md', { run: 'cat about.md' }));

    assert.equal(el.tagName, 'BUTTON');
    assert.equal(el.getAttribute('type'), 'button');
    assert.equal(el.getAttribute('data-run'), 'cat about.md');
    assert.equal(el.className, 'run s-plain');
    assert.equal(el.textContent, 'about.md');
  });

  it('maps styles to classes and unknown styles to plain', () => {
    const doc = dom().window.document;

    assert.equal(renderSpan(doc, span('x', { style: 'heading' })).className, 's-heading');
    assert.equal(renderSpan(doc, span('x', { style: 'evil" onclick="x' })).className, 's-plain');
  });

  it('marks whitespace-only column gaps so narrow screens can stack them', () => {
    const doc = dom().window.document;

    assert.equal(renderSpan(doc, span('   ')).className, 's-plain gap');
    assert.equal(renderSpan(doc, span(' ')).className, 's-plain');
  });

  it('gives unindented prose lines the measure class and listing rows the full width', () => {
    const doc = dom().window.document;
    const classOf = (spans: ReturnType<typeof span>[]): string => renderLine(doc, spans).className;

    assert.equal(classOf([span('hello '), span('link', { href: 'https://example.com' })]), 'line line-prose');
    assert.equal(classOf([span('a'), span('b', { run: 'cat b' })]), 'line');
    assert.equal(classOf([span('a'), span('   '), span('b')]), 'line');
    assert.equal(classOf([]), 'line');
  });

  it('turns leading spaces into a capped indent class and removes them from the text', () => {
    const doc = dom().window.document;
    const two = renderLine(doc, [span('  Go, PHP')]);
    const deep = renderLine(doc, [span('            x')]);
    const gapFirst = renderLine(doc, [span('    '), span('tap', { run: 'cat x' })]);

    assert.equal(two.className, 'line indent-2');
    assert.equal(two.textContent, 'Go, PHP');
    assert.equal(deep.className, 'line indent-8');
    assert.equal(gapFirst.className, 'line indent-4');
    assert.equal(gapFirst.children.length, 1);
    assert.equal(gapFirst.textContent, 'tap');
  });

  it('refuses a media src on another origin', () => {
    const doc = dom().window.document;
    for (const src of [
      'https://evil.example/cam.jpg',
      '//evil.example/cam.jpg',
      '/\\evil.example/cam.jpg',
      'javascript:alert(1)'
    ]) {
      const el = renderBlock(doc, block([[span('caption')]], { kind: 'media', src, alt: 'cam' }), ORIGIN);

      assert.equal(el.querySelector('img'), null, src);
      assert.equal(el.textContent, 'caption');
    }
  });

  it('renders a same-origin media path as an image with its alt text', () => {
    const doc = dom().window.document;
    const el = renderBlock(
      doc,
      block([[span('live', { style: 'dim' })]], { kind: 'media', src: '/api/v1/camera/stream', alt: 'Live camera' }),
      ORIGIN
    );
    const img = el.querySelector('img');

    assert.equal(el.className, 'block block-media');
    assert.equal(img?.getAttribute('src'), '/api/v1/camera/stream');
    assert.equal(img?.getAttribute('alt'), 'Live camera');
    assert.equal(img?.className, 'media');
  });

  it('renders blocks as lines of spans and empty lines as blank rows', () => {
    const doc = dom().window.document;
    const els = renderResult(
      doc,
      result([
        block([[span('a'), span('b', { style: 'bold' })], []]),
        block([[span('no', { style: 'error' })]], { kind: 'error' })
      ]),
      ORIGIN
    );

    assert.equal(els.length, 2);
    assert.equal(els[0].className, 'block block-text');
    assert.equal(els[0].children.length, 2);
    assert.equal(els[0].children[0].className, 'line line-prose');
    assert.equal(els[0].children[0].children.length, 2);
    assert.equal(els[0].children[1].children.length, 0);
    assert.equal(els[1].className, 'block block-error');
  });

  it('treats an unknown block kind as text', () => {
    const doc = dom().window.document;

    assert.equal(renderBlock(doc, block([[span('x')]], { kind: 'nope' }), ORIGIN).className, 'block block-text');
  });

  it('echoes a command with its prompt as text', () => {
    const doc = dom().window.document;
    const el = renderEcho(doc, 'visitor@pascalallen.com:~$', '<b>ls</b>');

    assert.equal(el.className, 'line echo');
    assert.equal(el.querySelector('.prompt')?.textContent, 'visitor@pascalallen.com:~$');
    assert.equal(el.querySelector('.cmd')?.textContent, '<b>ls</b>');
    assert.equal(el.querySelector('b'), null);
  });
});
