import type { Result, Shell } from './bridge';
import {
  applyTheme,
  CAMERA_STREAM,
  cameraAvailable,
  fetchFeed,
  OpenImpl,
  openUrl,
  readTheme,
  StorageAccess,
  THEMES
} from './effects';
import { attachEditor } from './input';
import { renderEcho, renderEntry, renderResult } from './render';

export interface TerminalOptions {
  win: Window;
  shell: Promise<Shell>;
  fetchImpl?: typeof fetch;
  openImpl?: OpenImpl;
  storage?: StorageAccess;
  fetchTimeoutMs?: number;
}

export interface Terminal {
  ready: Promise<void>;
}

const READ_ONLY = new Set([
  'help',
  'man',
  'ls',
  'cd',
  'pwd',
  'cat',
  'tree',
  'head',
  'tail',
  'grep',
  'whoami',
  'date',
  'about',
  'resume',
  'skills',
  'experience',
  'projects',
  'writing',
  'contact',
  'education'
]);

const SIMPLE_LINE = /^[A-Za-z0-9 ._/~@:=,+-]+$/;

const isReadOnly = (line: string): boolean => {
  const trimmed = line.trim();
  return SIMPLE_LINE.test(trimmed) && READ_ONLY.has(trimmed.split(/\s+/)[0]);
};

const REVEAL_LINES = 24;

const byId = <T extends HTMLElement>(doc: Document, id: string): T => {
  const el = doc.getElementById(id);
  if (el === null) {
    throw new Error(`terminal: #${id} missing`);
  }
  return el as T;
};

const plain = (text: string, style = 'plain') => ({ text, style, href: '', run: '' });

const CAMERA_OFFLINE: Result = {
  blocks: [{ kind: 'error', lines: [[plain('camera: the camera is offline right now', 'error')]], src: '', alt: '' }],
  effects: []
};

const failureNotice = (doc: Document): HTMLElement => {
  const block = doc.createElement('div');
  block.className = 'block block-error';
  const line = doc.createElement('div');
  line.className = 'line';
  const text = doc.createElement('span');
  text.className = 's-warning';
  text.textContent = 'The terminal could not start. Read the ';
  const link = doc.createElement('a');
  link.className = 's-accent';
  link.setAttribute('href', '/plain');
  link.textContent = 'plain text version';
  const end = doc.createElement('span');
  end.className = 's-warning';
  end.textContent = '.';
  line.append(text, link, end);
  block.append(line);
  return block;
};

export const createTerminal = (options: TerminalOptions): Terminal => {
  const { win } = options;
  const doc = win.document;
  const root = byId(doc, 'terminal');
  const scroll = byId(doc, 'scroll');
  const output = byId(doc, 'output');
  const input = byId<HTMLInputElement>(doc, 'input');
  const prompt = byId(doc, 'prompt');
  const fetchImpl: typeof fetch = options.fetchImpl ?? ((url, init) => win.fetch(url, init));
  const openImpl: OpenImpl = options.openImpl ?? ((url, target, features) => win.open(url, target, features));
  const storage: StorageAccess = options.storage ?? (() => win.localStorage);
  const fetchTimeoutMs = options.fetchTimeoutMs ?? 8000;
  const origin = win.location.origin;

  let shell: Shell | null = null;
  let failed = false;
  const queue: string[] = [];

  const atBottom = (): boolean => scroll.scrollHeight - scroll.scrollTop - scroll.clientHeight < 8;

  const follow = (force: boolean): void => {
    if (force) {
      scroll.scrollTop = scroll.scrollHeight;
    }
  };

  const reveal = (entry: HTMLElement): void => {
    entry.classList.add('reveal');
    entry.querySelectorAll<HTMLElement>('.block > .line').forEach((line, i) => {
      if (i < REVEAL_LINES) {
        line.style.setProperty('--i', String(i));
      }
    });
  };

  const append = (target: HTMLElement, result: Result): void => {
    target.append(...renderResult(doc, result, origin));
    reveal(target);
  };

  const entryWithEcho = (line: string): HTMLElement => {
    const entry = renderEntry(doc);
    entry.append(renderEcho(doc, shell?.prompt() ?? prompt.textContent ?? '', line));
    output.append(entry);
    return entry;
  };

  const deliverInto = (target: HTMLElement, feed: string): void => {
    const sh = shell;
    if (sh === null) {
      return;
    }
    fetchFeed(feed, fetchImpl, fetchTimeoutMs)
      .then(
        items => sh.deliver(feed, items),
        () => sh.deliverError(feed)
      )
      .catch(() => sh.deliverError(feed))
      .then(result => {
        const wasAtBottom = atBottom();
        append(target, result);
        follow(wasAtBottom);
      })
      .catch(() => undefined);
  };

  const showWhenLive = (target: HTMLElement, result: Result): void => {
    if (!result.blocks.some(block => block.kind === 'media' && block.src === CAMERA_STREAM)) {
      append(target, result);
      return;
    }
    cameraAvailable(fetchImpl, fetchTimeoutMs).then(available => {
      const wasAtBottom = atBottom();
      append(target, available ? result : CAMERA_OFFLINE);
      follow(wasAtBottom);
    });
  };

  const run = (line: string, user: boolean): void => {
    if (failed) {
      return;
    }
    if (shell === null) {
      queue.push(line);
      return;
    }
    let target = entryWithEcho(line);
    const result = shell.exec(line);
    for (const effect of result.effects) {
      switch (effect.kind) {
        case 'clear':
          if (user) {
            output.replaceChildren();
            target = renderEntry(doc);
            output.append(target);
          }
          break;
        case 'open':
          if (user) {
            openUrl(effect.arg, openImpl);
          }
          break;
        case 'theme':
          if (user) {
            applyTheme(doc, effect.arg, storage);
          }
          break;
        case 'fetch':
          if (user) {
            deliverInto(target, effect.arg);
          }
          break;
      }
    }
    showWhenLive(target, result);
    if (!target.hasChildNodes()) {
      target.remove();
    }
    prompt.textContent = shell.prompt();
    if (user && line.trim() !== '') {
      const base = `${win.location.pathname}${win.location.search}`;
      win.history.replaceState(null, '', isReadOnly(line) ? `${base}#${encodeURIComponent(line)}` : base);
    }
    follow(true);
  };

  const editor = attachEditor(input, {
    submit: line => run(line, true),
    complete: (line, cursor) => (shell === null ? null : shell.complete(line, cursor)),
    showCandidates: (line, candidates) => {
      const entry = entryWithEcho(line);
      append(entry, {
        blocks: [
          {
            kind: 'text',
            lines: [candidates.flatMap((c, i) => (i === 0 ? [plain(c)] : [plain('  '), plain(c)]))],
            src: '',
            alt: ''
          }
        ],
        effects: []
      });
      follow(true);
    },
    history: () => (shell === null ? [] : shell.history()),
    clearScreen: () => output.replaceChildren(),
    interrupt: line => {
      entryWithEcho(`${line}^C`);
      follow(true);
    },
    hasSelection: () => (win.getSelection()?.toString() ?? '') !== ''
  });

  const deepLink = (): void => {
    const raw = win.location.hash.slice(1);
    let line: string;
    try {
      line = decodeURIComponent(raw).trim();
    } catch {
      return;
    }
    if (line === '') {
      return;
    }
    if (isReadOnly(line)) {
      run(line, false);
    } else if (input.value === '') {
      editor.set(line);
    }
  };

  let pointer = '';
  root.addEventListener('pointerdown', event => {
    pointer = event.pointerType;
  });

  root.addEventListener('click', event => {
    const target = event.target as Element | null;
    const runEl = target?.closest?.('[data-run]');
    if (runEl) {
      const mouse = event as MouseEvent;
      const modified = mouse.metaKey || mouse.ctrlKey || mouse.shiftKey || mouse.altKey || mouse.button !== 0;
      if (runEl.tagName === 'A' && (failed || modified)) {
        return;
      }
      event.preventDefault();
      run(runEl.getAttribute('data-run') ?? '', true);
      if (pointer !== 'touch') {
        input.focus({ preventScroll: true });
      }
      pointer = '';
      return;
    }
    if (target?.closest?.('a, button, input') || (win.getSelection()?.toString() ?? '') !== '') {
      return;
    }
    input.focus({ preventScroll: true });
  });

  const viewport = win.visualViewport;
  if (viewport) {
    const sync = (): void => {
      doc.documentElement.style.setProperty('--vvh', `${viewport.height}px`);
      doc.documentElement.style.setProperty('--vvt', `${viewport.offsetTop}px`);
      if (doc.activeElement === input) {
        follow(true);
      }
    };
    viewport.addEventListener('resize', sync);
    viewport.addEventListener('scroll', sync);
    sync();
  }

  const ready = options.shell.then(
    sh => {
      shell = sh;
      root.setAttribute('data-state', 'ready');
      if (output.querySelector('[data-prerendered]') === null) {
        const entry = renderEntry(doc);
        entry.classList.add('entry-boot');
        output.append(entry);
        append(entry, sh.boot());
      }
      prompt.textContent = sh.prompt();
      const active = readTheme(storage) ?? doc.documentElement.getAttribute('data-theme');
      if (active !== null && THEMES.includes(active)) {
        sh.setTheme(active);
      }
      deepLink();
      for (const line of queue.splice(0)) {
        run(line, true);
      }
    },
    () => {
      failed = true;
      root.setAttribute('data-state', 'failed');
      input.disabled = true;
      doc.querySelectorAll('a[data-run]').forEach(el => el.setAttribute('href', '/plain'));
      doc.querySelectorAll('button[data-run]').forEach(el => {
        el.setAttribute('disabled', '');
        el.setAttribute('title', 'The terminal could not start. Use the plain text version.');
      });
      const entry = renderEntry(doc);
      entry.append(failureNotice(doc));
      output.append(entry);
      reveal(entry);
    }
  );

  return { ready };
};
