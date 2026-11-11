import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { brotliCompressSync, constants, gzipSync } from 'node:zlib';

const NAMES = ['app.js', 'app.css', 'theme.js', 'wasm_exec.js'];
const dir = process.argv[2] ?? fileURLToPath(new URL('../static/assets/', import.meta.url));

let written = 0;
for (const name of NAMES) {
  const file = join(dir, name);
  if (!existsSync(file)) {
    continue;
  }
  const raw = readFileSync(file);
  const gz = gzipSync(raw, { level: 9 });
  const br = brotliCompressSync(raw, {
    params: { [constants.BROTLI_PARAM_QUALITY]: 11, [constants.BROTLI_PARAM_SIZE_HINT]: raw.length }
  });
  writeFileSync(`${file}.gz`, gz);
  writeFileSync(`${file}.br`, br);
  written += 1;
  console.log(`${name.padEnd(14)} raw ${raw.length} B, gzip ${gz.length} B, brotli ${br.length} B`);
}

if (written === 0) {
  console.error(`compress: none of ${NAMES.join(', ')} found in ${dir}`);
  process.exit(1);
}
