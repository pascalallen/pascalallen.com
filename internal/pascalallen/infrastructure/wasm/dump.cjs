'use strict';

const fs = require('node:fs');
const path = require('node:path');

const [wasmPath, execPath] = process.argv.slice(2);
if (!wasmPath || !execPath) {
    console.error('usage: node dump.cjs <shell.wasm> <wasm_exec.js>');
    process.exit(2);
}

globalThis.require = require;
globalThis.fs = fs;
globalThis.path = path;
globalThis.crypto ??= require('node:crypto');
globalThis.performance ??= require('node:perf_hooks').performance;
require(path.resolve(execPath));

const SKIP = new Set(['date', 'camera']);
const LIMIT = 400;

function canonical(v) {
    if (Array.isArray(v)) return v.map(canonical);
    if (v && typeof v === 'object') {
        return Object.fromEntries(
            Object.keys(v)
                .sort()
                .map((k) => [k, canonical(v[k])])
        );
    }
    return v;
}

function runs(result) {
    return result.blocks.flatMap((b) => b.lines.flatMap((l) => l.map((s) => s.run).filter((r) => r !== '')));
}

async function main() {
    const ready = new Promise((resolve) => {
        globalThis.__shellReady = resolve;
    });
    const go = new Go();
    const { instance } = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);
    go.run(instance);
    await ready;

    const sh = globalThis.__shell;
    const log = [];
    const queue = ['help', 'ls -la', 'tree', 'history', 'pwd', 'whoami', 'echo "a b" \'c\' ; echo d && echo e', 'cd nope', 'cat', 'ls | wc', 'foo'];
    const boot = sh.boot();
    log.push(['boot', boot]);
    queue.push(...runs(boot));
    const seen = new Set();
    while (queue.length > 0 && seen.size < LIMIT) {
        const line = queue.shift();
        if (seen.has(line) || SKIP.has(line.trim())) continue;
        seen.add(line);
        const r = sh.exec(line);
        log.push([line, r]);
        queue.push(...runs(r));
    }
    for (const name of sh.commands().map((c) => c.name)) {
        if (!SKIP.has(name)) log.push(['man ' + name, sh.exec('man ' + name)]);
    }
    for (const [line, cursor] of [['ca', 2], ['cat a', 5], ['cd ', 3], ['theme ', 6], ['ls ~/pr', 7], ['ls; ca', 6], ['zzz', 3], ['é', 1]]) {
        log.push(['complete ' + line, sh.complete(line, cursor)]);
    }
    const items = [{ name: 'a', description: 'b', url: 'https://example.com', meta: 'c' }, { name: 'é‮', description: '', url: 'mailto:x@y.z', meta: '' }];
    log.push(['deliver', sh.deliver('repos', items)]);
    log.push(['deliver packages', sh.deliver('packages', items)]);
    log.push(['deliverError', sh.deliverError('packages')]);
    log.push(['deliver bad feed', sh.deliver('zzz', items)]);
    log.push(['prompt', sh.prompt()]);
    log.push(['history', sh.history()]);
    log.push(['commands', sh.commands()]);
    log.push(['long', sh.exec('echo ' + 'x'.repeat(5000))]);
    log.push(['many', sh.exec(Array(40).fill('pwd').join(';'))]);

    process.stdout.write(JSON.stringify(canonical(log), null, 1) + '\n');
    process.exit(0);
}

main().catch((err) => {
    console.error(err);
    process.exit(1);
});
