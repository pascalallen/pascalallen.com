import './styles.css';
import { loadShell } from './bridge';
import { initialTheme, recordVisit } from './effects';
import env, { EnvKey } from './env';
import { createTerminal } from './terminal';

const prefersLight = window.matchMedia?.('(prefers-color-scheme: light)').matches ?? false;
document.documentElement.setAttribute(
  'data-theme',
  initialTheme(() => window.localStorage, prefersLight)
);

const root = document.getElementById('terminal');

if (root !== null) {
  const useFake =
    env(EnvKey.APP_ENV) === 'local' && new URLSearchParams(window.location.search).get('shell') === 'fake';
  const shell = loadShell({
    runtimeSrc: root.getAttribute('data-runtime-src') ?? '/static/assets/wasm_exec.js',
    wasmSrc: root.getAttribute('data-wasm-src') ?? '/static/assets/shell.wasm',
    devShellSrc: useFake ? '/static/assets/dev-shell.js' : undefined
  });
  createTerminal({ win: window, shell });
  if (window.matchMedia?.('(pointer: fine)').matches) {
    document.getElementById('input')?.focus({ preventScroll: true });
  }
  recordVisit(
    (url, init) => window.fetch(url, init),
    () => window.sessionStorage,
    window.navigator
  );
}
