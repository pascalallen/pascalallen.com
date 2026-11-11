import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { describe, it } from 'node:test';
import { brotliDecompressSync, gunzipSync } from 'node:zlib';

const NAMES = ['app.js', 'app.css', 'theme.js', 'wasm_exec.js'];

const compress = (dir: string): number => spawnSync('node', ['compress.mjs', dir], { encoding: 'utf8' }).status ?? -1;

describe('asset compression', () => {
  it('writes br and gz siblings that decode to the original for every text asset', () => {
    const dir = mkdtempSync(join(tmpdir(), 'compress-'));
    try {
      for (const name of NAMES) {
        writeFileSync(join(dir, name), `/* ${name} */ ${'payload '.repeat(200)}`);
      }

      assert.equal(compress(dir), 0);

      for (const name of NAMES) {
        const raw = readFileSync(join(dir, name));
        assert.deepEqual(brotliDecompressSync(readFileSync(join(dir, `${name}.br`))), raw, `${name}.br`);
        assert.deepEqual(gunzipSync(readFileSync(join(dir, `${name}.gz`))), raw, `${name}.gz`);
      }
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it('skips an asset that is not built and does not touch the module', () => {
    const dir = mkdtempSync(join(tmpdir(), 'compress-'));
    try {
      writeFileSync(join(dir, 'app.js'), 'x'.repeat(500));
      writeFileSync(join(dir, 'shell.wasm'), 'y'.repeat(500));

      assert.equal(compress(dir), 0);

      assert.equal(existsSync(join(dir, 'app.js.br')), true);
      assert.equal(existsSync(join(dir, 'theme.js.br')), false);
      assert.equal(existsSync(join(dir, 'shell.wasm.br')), false);
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  });

  it('fails when the directory has none of the expected assets', () => {
    const dir = mkdtempSync(join(tmpdir(), 'compress-'));
    try {
      assert.notEqual(compress(dir), 0);
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  });
});
