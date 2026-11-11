import type { Block, Line, Result, Span } from './bridge';

const STYLES = new Set(['plain', 'bold', 'dim', 'accent', 'success', 'warning', 'error', 'heading', 'path']);

const LINK_SCHEMES = new Set(['http:', 'https:', 'mailto:']);

const BLOCK_KINDS = new Set(['text', 'error', 'media']);

export const safeHref = (href: string): string | null => {
  try {
    const url = new URL(href);
    return LINK_SCHEMES.has(url.protocol) ? url.href : null;
  } catch {
    return null;
  }
};

export const safeMediaSrc = (src: string, origin: string): string | null => {
  try {
    const url = new URL(src, origin);
    return url.origin === origin ? url.pathname + url.search : null;
  } catch {
    return null;
  }
};

const styleClass = (span: Span): string => {
  const base = `s-${STYLES.has(span.style) ? span.style : 'plain'}`;
  return /^ {2,}$/.test(span.text) ? `${base} gap` : base;
};

export const renderSpan = (doc: Document, span: Span): HTMLElement => {
  const href = span.href === '' ? null : safeHref(span.href);
  let el: HTMLElement;
  if (href !== null) {
    el = doc.createElement('a');
    el.setAttribute('href', href);
    el.setAttribute('target', '_blank');
    el.setAttribute('rel', 'noopener noreferrer');
    el.className = styleClass(span);
  } else if (span.run !== '') {
    el = doc.createElement('button');
    el.setAttribute('type', 'button');
    el.setAttribute('data-run', span.run);
    el.className = `run ${styleClass(span)}`;
  } else {
    el = doc.createElement('span');
    el.className = styleClass(span);
  }
  el.textContent = span.text;
  return el;
};

export const renderLine = (doc: Document, line: Line): HTMLElement => {
  const el = doc.createElement('div');
  el.className = 'line';
  for (const span of line) {
    el.appendChild(renderSpan(doc, span));
  }
  return el;
};

export const renderBlock = (doc: Document, block: Block, origin: string): HTMLElement => {
  const kind = BLOCK_KINDS.has(block.kind) ? block.kind : 'text';
  const el = doc.createElement('div');
  el.className = `block block-${kind}`;
  if (kind === 'media') {
    const src = safeMediaSrc(block.src, origin);
    if (src !== null) {
      const img = doc.createElement('img');
      img.className = 'media';
      img.setAttribute('src', src);
      img.setAttribute('alt', block.alt);
      el.appendChild(img);
    }
  }
  for (const line of block.lines) {
    el.appendChild(renderLine(doc, line));
  }
  return el;
};

export const renderResult = (doc: Document, result: Result, origin: string): HTMLElement[] =>
  result.blocks.map(block => renderBlock(doc, block, origin));

export const renderEcho = (doc: Document, prompt: string, line: string): HTMLElement => {
  const el = doc.createElement('div');
  el.className = 'line echo';
  const promptEl = doc.createElement('span');
  promptEl.className = 'prompt';
  promptEl.textContent = prompt;
  const cmd = doc.createElement('span');
  cmd.className = 'cmd';
  cmd.textContent = line;
  el.append(promptEl, doc.createTextNode(' '), cmd);
  return el;
};

export const renderEntry = (doc: Document): HTMLElement => {
  const el = doc.createElement('div');
  el.className = 'entry';
  return el;
};
