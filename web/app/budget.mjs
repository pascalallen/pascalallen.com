import { readFileSync } from 'node:fs';
import { gzipSync } from 'node:zlib';

const budgets = [
  { file: '../static/assets/app.js', limit: 30 * 1024 },
  { file: '../static/assets/app.css', limit: 10 * 1024 },
  { file: '../static/assets/theme.js', limit: 512 }
];

let failed = false;
for (const { file, limit } of budgets) {
  const raw = readFileSync(new URL(file, import.meta.url));
  const gzip = gzipSync(raw, { level: 9 }).length;
  const ok = gzip <= limit;
  failed = failed || !ok;
  console.log(`${ok ? 'ok  ' : 'FAIL'} ${file} raw ${raw.length} B, gzip ${gzip} B, budget ${limit} B`);
}

if (failed) {
  process.exit(1);
}
