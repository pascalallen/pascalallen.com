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
