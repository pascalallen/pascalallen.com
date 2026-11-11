import { JSDOM } from 'jsdom';
import type { Block, Result, Shell, Span } from './bridge';
import { createFakeShell } from './fake-shell';

export const ORIGIN = 'https://pascalallen.test';

export const page = (output = ''): string => `<!DOCTYPE html><html><body>
<div class="terminal" id="terminal" data-state="loading" data-wasm-src="/static/assets/shell.wasm">
  <div class="scroll" id="scroll">
    <div class="output" id="output" role="log" aria-live="polite" aria-label="Terminal output">${output}</div>
    <div class="prompt-line">
      <label class="prompt" id="prompt" for="input">visitor@pascalallen.com:~$</label>
      <input class="input" id="input" type="text" autocapitalize="off" autocomplete="off" autocorrect="off" spellcheck="false">
    </div>
  </div>
  <nav class="chips" aria-label="Commands"><a class="chip" href="/#help" data-run="help">help</a></nav>
</div>
</body></html>`;

export const dom = (output = '', hash = ''): JSDOM =>
  new JSDOM(page(output), { url: `${ORIGIN}/${hash}`, pretendToBeVisual: true });

export const span = (text: string, extra: Partial<Span> = {}): Span => ({
  text,
  style: 'plain',
  href: '',
  run: '',
  ...extra
});

export const block = (lines: Span[][], extra: Partial<Block> = {}): Block => ({
  kind: 'text',
  lines,
  src: '',
  alt: '',
  ...extra
});

export const result = (blocks: Block[], effects: Result['effects'] = []): Result => ({ blocks, effects });

export interface Spy extends Shell {
  calls: string[];
}

export const spyShell = (base: Shell = createFakeShell()): Spy => {
  const calls: string[] = [];
  return {
    calls,
    boot: () => {
      calls.push('boot');
      return base.boot();
    },
    exec: line => {
      calls.push(`exec:${line}`);
      return base.exec(line);
    },
    complete: (line, cursor) => base.complete(line, cursor),
    prompt: () => base.prompt(),
    history: () => base.history(),
    commands: () => base.commands(),
    deliver: (feed, items) => {
      calls.push(`deliver:${feed}:${items.map(i => i.name).join(',')}`);
      return base.deliver(feed, items);
    },
    deliverError: feed => {
      calls.push(`deliverError:${feed}`);
      return base.deliverError(feed);
    }
  };
};

export const key = (jsdom: JSDOM, target: Element, k: string, init: KeyboardEventInit = {}): KeyboardEvent => {
  const event = new jsdom.window.KeyboardEvent('keydown', { key: k, bubbles: true, cancelable: true, ...init });
  target.dispatchEvent(event);
  return event;
};

export const tick = (ms = 0): Promise<void> => new Promise(resolve => setTimeout(resolve, ms));

export const memoryStorage = (): Storage => {
  const data = new Map<string, string>();
  return {
    get length() {
      return data.size;
    },
    clear: () => data.clear(),
    getItem: k => data.get(k) ?? null,
    key: i => Array.from(data.keys())[i] ?? null,
    removeItem: k => {
      data.delete(k);
    },
    setItem: (k, v) => {
      data.set(k, String(v));
    }
  };
};

export const throwingStorage = (): Storage => {
  const fail = (): never => {
    throw new Error('SecurityError');
  };
  return {
    length: 0,
    clear: fail,
    getItem: fail,
    key: fail,
    removeItem: fail,
    setItem: fail
  };
};
