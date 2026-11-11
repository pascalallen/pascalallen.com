'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const [wasmPath, execPath] = process.argv.slice(2);
if (!wasmPath || !execPath) {
    console.error('usage: node smoke.cjs <shell.wasm> <wasm_exec.js>');
    process.exit(2);
}

globalThis.require = require;
globalThis.fs = fs;
globalThis.path = path;
globalThis.crypto ??= require('node:crypto');
globalThis.performance ??= require('node:perf_hooks').performance;
require(path.resolve(execPath));

const STYLES = ['plain', 'bold', 'dim', 'accent', 'success', 'warning', 'error', 'heading', 'path'];
const BLOCKS = ['text', 'error', 'media'];
const EFFECTS = ['clear', 'open', 'fetch', 'theme'];

function assertResult(r) {
    assert.deepEqual(Object.keys(r).sort(), ['blocks', 'effects']);
    assert.ok(Array.isArray(r.blocks));
    assert.ok(Array.isArray(r.effects));
    for (const b of r.blocks) {
        assert.deepEqual(Object.keys(b).sort(), ['alt', 'kind', 'lines', 'src']);
        assert.ok(BLOCKS.includes(b.kind), b.kind);
        assert.ok(Array.isArray(b.lines));
        for (const line of b.lines) {
            assert.ok(Array.isArray(line));
            for (const s of line) {
                assert.deepEqual(Object.keys(s).sort(), ['href', 'run', 'style', 'text']);
                assert.ok(STYLES.includes(s.style), s.style);
                for (const k of ['text', 'href', 'run']) assert.equal(typeof s[k], 'string');
            }
        }
    }
    for (const e of r.effects) {
        assert.deepEqual(Object.keys(e).sort(), ['arg', 'kind']);
        assert.ok(EFFECTS.includes(e.kind), e.kind);
        assert.equal(typeof e.arg, 'string');
    }
}

const text = (r) => r.blocks.flatMap((b) => b.lines.map((l) => l.map((s) => s.text).join(''))).join('\n');

async function main() {
    const logged = [];
    const originalError = console.error;
    console.error = (...args) => logged.push(args.join(' '));
    const ready = new Promise((resolve) => {
        globalThis.__shellReady = () => {
            resolve();
            throw new Error('ready hook failed');
        };
    });
    const go = new Go();
    const { instance } = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);
    go.run(instance);
    await ready;

    const sh = globalThis.__shell;
    assert.equal(logged.length, 1, 'a throwing __shellReady is logged once');
    assert.match(logged[0], /__shellReady/);
    console.error = originalError;
    for (const name of ['boot', 'exec', 'complete', 'prompt', 'history', 'commands', 'deliver', 'deliverError']) {
        assert.equal(typeof sh[name], 'function', name);
    }

    const boot = sh.boot();
    assertResult(boot);
    assert.ok(boot.blocks.length > 0);
    assert.ok(boot.blocks.some((b) => b.lines.some((l) => l.some((s) => s.run !== ''))));

    const help = sh.exec('help');
    assertResult(help);
    assert.match(text(help), /cat/);

    const about = sh.exec('cat about.md');
    assertResult(about);
    assert.ok(about.blocks.some((b) => b.lines.some((l) => l.some((s) => s.style === 'heading'))));

    const missing = sh.exec('cat nope');
    assertResult(missing);
    assert.equal(missing.blocks[0].kind, 'error');

    const fetch = sh.exec('repos');
    assertResult(fetch);
    assert.deepEqual(fetch.effects, [{ kind: 'fetch', arg: 'repos' }]);

    const theme = sh.exec('theme amber');
    assert.deepEqual(theme.effects, [{ kind: 'theme', arg: 'amber' }]);

    const clear = sh.exec('clear');
    assert.deepEqual(clear.effects, [{ kind: 'clear', arg: '' }]);

    const c = sh.complete('ca', 2);
    assert.deepEqual(Object.keys(c).sort(), ['candidates', 'cursor', 'line']);
    assert.equal(typeof c.line, 'string');
    assert.equal(typeof c.cursor, 'number');
    assert.ok(Array.isArray(c.candidates));
    assert.ok(c.candidates.includes('cat'));

    assert.equal(sh.prompt(), 'visitor@pascalallen.com:~$');
    assert.deepEqual(sh.history().slice(0, 2), ['help', 'cat about.md']);

    const commands = sh.commands();
    assert.ok(commands.length > 0);
    assert.deepEqual(Object.keys(commands[0]).sort(), ['name', 'summary', 'synopsis']);

    const items = [{ name: 'pgqueue', description: 'Postgres-backed job queue', url: 'https://github.com/pascalallen/pgqueue', meta: 'Go' }];
    const delivered = sh.deliver('repos', items);
    assertResult(delivered);
    assert.equal(delivered.blocks[0].kind, 'text');
    assert.match(text(delivered), /pgqueue/);
    assert.ok(delivered.blocks[0].lines.some((l) => l.some((s) => s.href === items[0].url)));

    const failed = sh.deliverError('repos');
    assertResult(failed);
    assert.equal(failed.blocks[0].kind, 'error');

    for (const bad of [
        () => sh.exec(),
        () => sh.exec(1),
        () => sh.exec('a', 'b'),
        () => sh.complete('ca'),
        () => sh.complete('ca', 'x'),
        () => sh.complete('ca', -1),
        () => sh.boot(1),
        () => sh.deliver('repos', 'x'),
        () => sh.deliver('repos', [{ name: 1 }]),
        () => sh.deliverError(),
    ]) {
        const r = bad();
        assertResult(r);
        assert.equal(r.blocks.length, 1);
        assert.equal(r.blocks[0].kind, 'error');
        assert.equal(r.effects.length, 0);
    }

    const item = (extra) => ({ name: 'n', description: 'd', url: 'https://example.com', meta: 'm', ...extra });
    const hostile = [
        [{ get name() { throw 1; }, description: 'd', url: 'u', meta: 'm' }],
        [new Proxy({}, { ownKeys() { throw 1; } })],
        [new Proxy({}, { getPrototypeOf() { throw 1; } })],
        [Object.defineProperty(item(), 'name', { get: () => 'accessor', enumerable: true })],
        [Object.create({ name: 'inherited', description: 'd', url: 'u', meta: 'm' })],
        [new Date()],
        [new (class Item { constructor() { Object.assign(this, item()); } })()],
    ];
    for (const items of hostile) {
        const r = sh.deliver('repos', items);
        assertResult(r);
        assert.equal(r.blocks.length, 1);
        assert.equal(r.blocks[0].kind, 'error');
        assert.equal(sh.prompt(), 'visitor@pascalallen.com:~$');
    }
    const throwingArray = [];
    Object.defineProperty(throwingArray, '0', { get() { throw 1; }, enumerable: true });
    const rr = sh.deliver('repos', throwingArray);
    assert.equal(rr.blocks[0].kind, 'error');
    const hostileLine = sh.exec(new Proxy({}, { getPrototypeOf() { throw 1; } }));
    assert.equal(hostileLine.blocks[0].kind, 'error');
    assert.equal(sh.deliver('repos', [item()]).blocks[0].kind, 'text');
    assert.equal(sh.deliver('repos', [Object.assign(Object.create(null), item())]).blocks[0].kind, 'text');

    const cyclic = {};
    cyclic.self = cyclic;
    const r = sh.deliver('repos', [cyclic]);
    assert.equal(r.blocks[0].kind, 'error');

    assert.equal(sh.prompt(), 'visitor@pascalallen.com:~$');
    console.log('wasm smoke test passed');
    process.exit(0);
}

main().catch((err) => {
    console.error(err);
    process.exit(1);
});
