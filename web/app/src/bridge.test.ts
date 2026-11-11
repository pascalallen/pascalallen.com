import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { installShell, loadShell, wrap } from './bridge';
import { createFakeShell } from './fake-shell';
import { dom } from './test-dom';

const emptyModule = new Uint8Array([0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00]);

const goRuntime = (win: Window, onRun: () => void): void => {
  (win as unknown as { Go: unknown }).Go = class {
    importObject = {};
    run(): Promise<void> {
      onRun();
      return new Promise(() => undefined);
    }
  };
};

const respond =
  (status: number, type: string): typeof fetch =>
  async () =>
    new Response(status === 200 ? emptyModule : null, { status, headers: { 'Content-Type': type } });

describe('bridge', () => {
  it('normalises missing and mistyped fields from the module', () => {
    const shell = wrap({
      exec: () => ({
        blocks: [{ kind: 'text', lines: [[{ text: 'hi', style: 7 }], 'bad'] }],
        effects: [{ kind: 'clear' }]
      }),
      prompt: () => 42
    });

    assert.deepEqual(shell.exec('x'), {
      blocks: [{ kind: 'text', lines: [[{ text: 'hi', style: '', href: '', run: '' }], []], src: '', alt: '' }],
      effects: [{ kind: 'clear', arg: '' }]
    });
    assert.equal(shell.prompt(), '');
    assert.deepEqual(shell.boot(), { blocks: [], effects: [] });
    assert.deepEqual(shell.complete('ca', 2), { line: 'ca', cursor: 2, candidates: [] });
  });

  it('resolves once the module signals readiness', async () => {
    const win = dom().window as unknown as Window;
    goRuntime(win, () => installShell(createFakeShell(), win));

    const shell = await loadShell({
      wasmSrc: '/static/assets/shell.wasm',
      win,
      fetchImpl: respond(200, 'application/wasm')
    });

    assert.equal(shell.prompt(), 'visitor@pascalallen.com:~$');
  });

  it('falls back to a buffered instantiate when streaming is refused', async () => {
    const win = dom().window as unknown as Window;
    goRuntime(win, () => installShell(createFakeShell(), win));

    const shell = await loadShell({ wasmSrc: '/static/assets/shell.wasm', win, fetchImpl: respond(200, 'text/plain') });

    assert.equal(shell.history().length, 0);
  });

  it('rejects when the module is missing', async () => {
    const win = dom().window as unknown as Window;
    goRuntime(win, () => undefined);

    await assert.rejects(
      loadShell({ wasmSrc: '/static/assets/shell.wasm', win, fetchImpl: respond(404, 'text/plain') }),
      /HTTP 404/
    );
  });

  it('rejects when the Go runtime script is missing', async () => {
    const win = dom().window as unknown as Window;

    await assert.rejects(loadShell({ wasmSrc: '/x.wasm', win }), /Go runtime missing/);
  });

  it('rejects when readiness never arrives', async () => {
    const win = dom().window as unknown as Window;
    goRuntime(win, () => undefined);

    await assert.rejects(
      loadShell({ wasmSrc: '/x.wasm', win, timeoutMs: 20, fetchImpl: respond(200, 'application/wasm') }),
      /timed out/
    );
  });
});

describe('runtime script loading', () => {
  const script = (status: number, type: string): typeof fetch =>
    (async (url: string) =>
      String(url).endsWith('.wasm')
        ? new Response(emptyModule, { status: 200, headers: { 'Content-Type': 'application/wasm' } })
        : new Response('x', { status, headers: { 'Content-Type': type } })) as unknown as typeof fetch;

  const inject = (win: Window): string[] => {
    const added: string[] = [];
    const doc = win.document;
    const original = doc.head.appendChild.bind(doc.head);
    doc.head.appendChild = (<T extends Node>(node: T): T => {
      added.push((node as unknown as HTMLScriptElement).src);
      return original(node);
    }) as typeof doc.head.appendChild;
    return added;
  };

  it('rejects without injecting a script when the runtime is missing', async () => {
    const win = dom().window as unknown as Window;
    const added = inject(win);

    await assert.rejects(
      loadShell({
        wasmSrc: '/x.wasm',
        runtimeSrc: '/static/assets/wasm_exec.js',
        win,
        fetchImpl: script(404, 'text/html')
      }),
      /runtime/
    );
    assert.deepEqual(added, []);
  });

  it('rejects without injecting a script when the runtime answers 404 with a script type', async () => {
    const win = dom().window as unknown as Window;
    const added = inject(win);

    await assert.rejects(
      loadShell({
        wasmSrc: '/x.wasm',
        runtimeSrc: '/static/assets/wasm_exec.js',
        win,
        fetchImpl: script(404, 'text/javascript')
      }),
      /HTTP 404/
    );
    assert.deepEqual(added, []);
  });

  it('rejects without injecting a script when the runtime is served as HTML', async () => {
    const win = dom().window as unknown as Window;
    const added = inject(win);

    await assert.rejects(
      loadShell({
        wasmSrc: '/x.wasm',
        runtimeSrc: '/static/assets/wasm_exec.js',
        win,
        fetchImpl: script(200, 'text/html; charset=utf-8')
      }),
      /runtime/
    );
    assert.deepEqual(added, []);
  });

  it('rejects when the runtime script loads but defines no Go', async () => {
    const win = dom().window as unknown as Window;
    const doc = win.document;
    const original = doc.head.appendChild.bind(doc.head);
    doc.head.appendChild = (<T extends Node>(node: T): T => {
      queueMicrotask(() => node.dispatchEvent(new (win as unknown as typeof globalThis).Event('load')));
      return original(node);
    }) as typeof doc.head.appendChild;

    await assert.rejects(
      loadShell({
        wasmSrc: '/x.wasm',
        runtimeSrc: '/static/assets/wasm_exec.js',
        win,
        fetchImpl: script(200, 'text/javascript')
      }),
      /Go runtime missing/
    );
  });

  it('skips loading the runtime when Go is already defined', async () => {
    const win = dom().window as unknown as Window;
    goRuntime(win, () => installShell(createFakeShell(), win));
    const added = inject(win);

    await loadShell({
      wasmSrc: '/static/assets/shell.wasm',
      runtimeSrc: '/static/assets/wasm_exec.js',
      win,
      fetchImpl: respond(200, 'application/wasm')
    });

    assert.deepEqual(added, []);
  });
});

describe('theme sync', () => {
  it('forwards the stored theme to a module that exposes setTheme', () => {
    const seen: string[] = [];
    const shell = wrap({ setTheme: (name: string) => seen.push(name) });

    shell.setTheme('amber');

    assert.deepEqual(seen, ['amber']);
  });

  it('does nothing when the module has no setTheme', () => {
    const shell = wrap({});

    assert.doesNotThrow(() => shell.setTheme('amber'));
  });
});
