import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import type { JSDOM } from 'jsdom';
import type { Shell } from './bridge';
import { createTerminal, TerminalOptions } from './terminal';
import { createFakeShell } from './fake-shell';
import { dom, key, memoryStorage, spyShell, tick, type Spy } from './test-dom';

interface Setup {
  jsdom: JSDOM;
  doc: Document;
  input: HTMLInputElement;
  output: HTMLElement;
  shell: Spy;
  opened: unknown[][];
  ready: Promise<void>;
}

const setup = (
  options: { output?: string; hash?: string; extra?: Partial<TerminalOptions>; shell?: Promise<Shell> } = {}
): Setup => {
  const jsdom = dom(options.output ?? '', options.hash ?? '');
  const doc = jsdom.window.document;
  const shell = spyShell();
  const opened: unknown[][] = [];
  const storage = memoryStorage();
  const terminal = createTerminal({
    win: jsdom.window as unknown as Window,
    shell: options.shell ?? Promise.resolve(shell),
    openImpl: (...args: unknown[]) => {
      opened.push(args);
      return null;
    },
    storage: () => storage,
    fetchImpl: (async () => {
      throw new Error('no network in tests');
    }) as unknown as typeof fetch,
    ...options.extra
  });
  return {
    jsdom,
    doc,
    input: doc.getElementById('input') as HTMLInputElement,
    output: doc.getElementById('output') as HTMLElement,
    shell,
    opened,
    ready: terminal.ready
  };
};

const type = (s: Setup, value: string): void => {
  s.input.value = value;
  s.input.setSelectionRange(value.length, value.length);
};

const enter = (s: Setup, value: string): void => {
  type(s, value);
  key(s.jsdom, s.input, 'Enter');
};

const execs = (s: Setup): string[] => s.shell.calls.filter(c => c.startsWith('exec:'));

describe('boot', () => {
  it('renders boot when the output has no prerendered children', async () => {
    const s = setup();
    await s.ready;

    assert.deepEqual(s.shell.calls, ['boot']);
    assert.equal(s.output.querySelectorAll('.entry').length, 1);
    assert.match(s.output.textContent ?? '', /pascalallen\.com/);
  });

  it('keeps prerendered output and does not render boot again', async () => {
    const prerendered =
      '<div class="entry entry-boot" data-prerendered=""><div class="block block-text"><div class="line"><span class="s-plain">from the server</span></div></div></div>';
    const s = setup({ output: prerendered });
    await s.ready;

    assert.equal(s.shell.calls.includes('boot'), false);
    assert.equal(s.output.querySelectorAll('.entry').length, 1);
    assert.equal(s.output.textContent, 'from the server');
  });

  it('shows a failure state when the shell cannot load', async () => {
    const s = setup({ shell: Promise.reject(new Error('no module')) });
    await s.ready;

    assert.equal(s.doc.getElementById('terminal')?.getAttribute('data-state'), 'failed');
    const notice = s.output.querySelector('.entry:last-child .block-error') as HTMLElement;
    assert.equal(notice.textContent, 'The terminal could not start. Read the plain text version.');
    const link = notice.querySelector('a') as HTMLAnchorElement;
    assert.equal(link.getAttribute('href'), '/plain');
    assert.equal(link.textContent, 'plain text version');
    assert.equal(link.hasAttribute('target'), false);
  });

  it('disables the input when the shell cannot load', async () => {
    const s = setup({ shell: Promise.reject(new Error('no module')) });
    await s.ready;

    assert.equal(s.input.disabled, true);
  });

  it('keeps the input enabled once the shell is ready', async () => {
    const s = setup();
    await s.ready;

    assert.equal(s.input.disabled, false);
  });

  it('marks the terminal ready once the shell resolves', async () => {
    const s = setup();
    await s.ready;

    assert.equal(s.doc.getElementById('terminal')?.getAttribute('data-state'), 'ready');
  });
});

describe('running a command', () => {
  it('echoes the line with its prompt, renders the result and clears the input', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'whoami');

    const entry = s.output.lastElementChild as HTMLElement;
    assert.equal(entry.querySelector('.echo .prompt')?.textContent, 'visitor@pascalallen.com:~$');
    assert.equal(entry.querySelector('.echo .cmd')?.textContent, 'whoami');
    assert.match(entry.textContent ?? '', /visitor$/);
    assert.equal(s.input.value, '');
  });

  it('updates the hash with replaceState', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'cat about.md');

    assert.equal(s.jsdom.window.location.hash, '#cat%20about.md');
  });

  it('runs a tapped command button as if typed', async () => {
    const s = setup();
    await s.ready;
    const button = s.output.querySelector('[data-run="help"]') as HTMLElement;
    button.click();

    assert.deepEqual(execs(s), ['exec:help']);
    assert.equal(s.doc.activeElement, s.input);
  });

  it('runs a prerendered run link instead of navigating', async () => {
    const prerendered =
      '<div class="entry entry-boot" data-prerendered=""><div class="block block-text"><div class="line"><a class="run s-accent" href="/#about" data-run="about">about</a></div></div></div>';
    const s = setup({ output: prerendered });
    await s.ready;
    const link = s.output.querySelector('a[data-run]') as HTMLElement;
    const event = new s.jsdom.window.MouseEvent('click', { bubbles: true, cancelable: true });
    link.dispatchEvent(event);

    assert.equal(event.defaultPrevented, true);
    assert.deepEqual(execs(s), ['exec:about']);
  });

  it('leaves a modifier click on a run link to the browser', async () => {
    const prerendered =
      '<div class="entry entry-boot" data-prerendered=""><div class="block block-text"><div class="line"><a class="run s-accent" href="/#about" data-run="about">about</a></div></div></div>';
    const s = setup({ output: prerendered });
    await s.ready;
    const link = s.output.querySelector('a[data-run]') as HTMLElement;
    for (const init of [{ metaKey: true }, { ctrlKey: true }, { shiftKey: true }, { altKey: true }, { button: 1 }]) {
      const event = new s.jsdom.window.MouseEvent('click', { bubbles: true, cancelable: true, ...init });
      link.dispatchEvent(event);

      assert.equal(event.defaultPrevented, false, JSON.stringify(init));
    }
    assert.deepEqual(execs(s), []);
  });

  it('still runs a modifier click on a run button', async () => {
    const s = setup();
    await s.ready;
    const button = s.output.querySelector('[data-run="help"]') as HTMLElement;
    button.dispatchEvent(new s.jsdom.window.MouseEvent('click', { bubbles: true, cancelable: true, shiftKey: true }));

    assert.deepEqual(execs(s), ['exec:help']);
  });

  it('points run links and chips at the plain page and disables run buttons when the shell cannot load', async () => {
    const prerendered =
      '<div class="entry entry-boot" data-prerendered=""><div class="block block-text"><div class="line"><a class="run s-accent" href="/#about" data-run="about">about</a><button type="button" class="run" data-run="help">help</button></div></div></div>';
    const s = setup({ output: prerendered, shell: Promise.reject(new Error('no module')) });
    await s.ready;

    const chip = s.doc.querySelector('.chips a.chip[data-run]') as HTMLElement;
    assert.equal(chip.getAttribute('href'), '/plain');
    const link = s.output.querySelector('a[data-run]') as HTMLElement;
    assert.equal(link.getAttribute('href'), '/plain');
    const event = new s.jsdom.window.MouseEvent('click', { bubbles: true, cancelable: true });
    link.dispatchEvent(event);
    assert.equal(event.defaultPrevented, false);
    const button = s.output.querySelector('button[data-run]') as HTMLButtonElement;
    assert.equal(button.disabled, true);
    assert.match(button.getAttribute('title') ?? '', /plain text version/);
  });

  it('empties the log on a clear effect', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'whoami');
    enter(s, 'clear');

    assert.equal(s.output.textContent, '');
  });

  it('opens a link from a user-initiated run', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'open https://example.com');

    assert.deepEqual(s.opened, [['https://example.com/', '_blank', 'noopener']]);
  });

  it('applies and persists a theme effect', async () => {
    const storage = memoryStorage();
    const s = setup({ extra: { storage: () => storage } });
    await s.ready;
    enter(s, 'theme amber');

    assert.equal(s.doc.documentElement.getAttribute('data-theme'), 'amber');
    assert.equal(storage.getItem('theme'), 'amber');
  });
});

describe('camera in the terminal', () => {
  const statusFetch = (body: unknown, urls: string[], status = 200): typeof fetch =>
    (async (url: string) => {
      urls.push(url);
      return new Response(JSON.stringify(body), { status });
    }) as unknown as typeof fetch;

  it('checks the camera status and embeds the stream when it is available', async () => {
    const urls: string[] = [];
    const s = setup({ extra: { fetchImpl: statusFetch({ status: 'success', data: { available: true } }, urls) } });
    await s.ready;
    enter(s, 'camera');
    const entry = s.output.lastElementChild as HTMLElement;
    assert.equal(entry.querySelector('img'), null);
    await tick(5);

    assert.deepEqual(urls, ['/api/v1/camera/status']);
    assert.equal(entry.querySelector('img.media')?.getAttribute('src'), '/api/v1/camera/stream');
  });

  it('prints an offline message and no stream when the camera is unavailable', async () => {
    const urls: string[] = [];
    const s = setup({ extra: { fetchImpl: statusFetch({ status: 'success', data: { available: false } }, urls) } });
    await s.ready;
    enter(s, 'camera');
    const entry = s.output.lastElementChild as HTMLElement;
    await tick(5);

    assert.equal(entry.querySelector('img'), null);
    assert.equal(entry.querySelector('.block-error')?.textContent, 'camera: the camera is offline right now');
  });

  it('prints the offline message when the status check fails', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'camera');
    const entry = s.output.lastElementChild as HTMLElement;
    await tick(5);

    assert.equal(entry.querySelector('img'), null);
    assert.equal(entry.querySelector('.block-error')?.textContent, 'camera: the camera is offline right now');
  });

  it('prints the offline message when the status endpoint answers with an error', async () => {
    const urls: string[] = [];
    const s = setup({ extra: { fetchImpl: statusFetch({ status: 'success', data: { available: true } }, urls, 503) } });
    await s.ready;
    enter(s, 'camera');
    const entry = s.output.lastElementChild as HTMLElement;
    await tick(5);

    assert.equal(entry.querySelector('img'), null);
  });
});

describe('fetch effect in the terminal', () => {
  it('delivers fetched items to the shell and renders them in the same entry', async () => {
    const fetchImpl = (async () =>
      new Response(
        JSON.stringify({
          status: 'success',
          data: {
            items: [
              { name: 'pgqueue', description: 'queue', url: 'https://github.com/pascalallen/pgqueue', meta: 'Go' }
            ]
          }
        }),
        { status: 200 }
      )) as unknown as typeof fetch;
    const s = setup({ extra: { fetchImpl } });
    await s.ready;
    enter(s, 'repos');
    const entry = s.output.lastElementChild as HTMLElement;
    await tick(5);

    assert.ok(s.shell.calls.includes('deliver:repos:pgqueue'));
    assert.equal(entry.querySelector('a')?.getAttribute('href'), 'https://github.com/pascalallen/pgqueue');
  });

  it('delivers an error when the request fails', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'packages');
    await tick(5);

    assert.ok(s.shell.calls.includes('deliverError:packages'));
    assert.match(s.output.textContent ?? '', /could not load the list/);
  });
});

describe('completion', () => {
  it('replaces the word with a single match', async () => {
    const s = setup();
    await s.ready;
    type(s, 'wh');
    const event = key(s.jsdom, s.input, 'Tab');

    assert.equal(event.defaultPrevented, true);
    assert.equal(s.input.value, 'whoami ');
    assert.equal(s.input.selectionStart, 7);
  });

  it('prints all candidates like bash for several matches', async () => {
    const s = setup();
    await s.ready;
    type(s, 'ca');
    key(s.jsdom, s.input, 'Tab');

    const entry = s.output.lastElementChild as HTMLElement;
    assert.equal(entry.querySelector('.echo .cmd')?.textContent, 'ca');
    assert.match(entry.textContent ?? '', /camera\s+cat/);
    assert.equal(s.input.value, 'ca');
    assert.deepEqual(execs(s), []);
  });

  it('lets Tab leave the terminal only on an empty line with Shift or after Escape', async () => {
    const s = setup();
    await s.ready;

    type(s, '');
    assert.equal(key(s.jsdom, s.input, 'Tab', { shiftKey: true }).defaultPrevented, false);
    type(s, 'wh');
    assert.equal(key(s.jsdom, s.input, 'Tab', { shiftKey: true }).defaultPrevented, true);
    type(s, '');
    assert.equal(key(s.jsdom, s.input, 'Tab').defaultPrevented, true);
    type(s, 'whoami');
    key(s.jsdom, s.input, 'Escape');
    assert.equal(key(s.jsdom, s.input, 'Tab').defaultPrevented, false);
    assert.equal(key(s.jsdom, s.input, 'Tab').defaultPrevented, true);
  });
});

describe('history', () => {
  it('walks history with Up and Down and restores the draft', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'whoami');
    enter(s, 'ls');
    type(s, 'ec');

    key(s.jsdom, s.input, 'ArrowUp');
    assert.equal(s.input.value, 'ls');
    key(s.jsdom, s.input, 'ArrowUp');
    assert.equal(s.input.value, 'whoami');
    key(s.jsdom, s.input, 'ArrowUp');
    assert.equal(s.input.value, 'whoami');
    key(s.jsdom, s.input, 'ArrowDown');
    assert.equal(s.input.value, 'ls');
    key(s.jsdom, s.input, 'ArrowDown');
    assert.equal(s.input.value, 'ec');
  });
});

describe('line editing', () => {
  it('abandons the line on Ctrl-C', async () => {
    const s = setup();
    await s.ready;
    type(s, 'cat abo');
    key(s.jsdom, s.input, 'c', { ctrlKey: true });

    assert.equal(s.input.value, '');
    assert.equal(s.output.lastElementChild?.querySelector('.echo .cmd')?.textContent, 'cat abo^C');
    assert.deepEqual(execs(s), []);
  });

  it('leaves Ctrl-C to copy when text in the input is selected', async () => {
    const s = setup();
    await s.ready;
    type(s, 'cat about.md');
    s.input.setSelectionRange(0, 3);
    const event = key(s.jsdom, s.input, 'c', { ctrlKey: true });

    assert.equal(event.defaultPrevented, false);
    assert.equal(s.input.value, 'cat about.md');
  });

  it('clears the screen on Ctrl-L without running a command', async () => {
    const s = setup();
    await s.ready;
    type(s, 'who');
    key(s.jsdom, s.input, 'l', { ctrlKey: true });

    assert.equal(s.output.children.length, 0);
    assert.equal(s.input.value, 'who');
  });

  it('supports Ctrl-A, Ctrl-E, Ctrl-U and Ctrl-K', async () => {
    const s = setup();
    await s.ready;

    type(s, 'cat about.md');
    key(s.jsdom, s.input, 'a', { ctrlKey: true });
    assert.equal(s.input.selectionStart, 0);
    key(s.jsdom, s.input, 'e', { ctrlKey: true });
    assert.equal(s.input.selectionStart, 12);

    s.input.setSelectionRange(4, 4);
    key(s.jsdom, s.input, 'k', { ctrlKey: true });
    assert.equal(s.input.value, 'cat ');

    type(s, 'cat about.md');
    s.input.setSelectionRange(4, 4);
    key(s.jsdom, s.input, 'u', { ctrlKey: true });
    assert.equal(s.input.value, 'about.md');
    assert.equal(s.input.selectionStart, 0);
  });

  it('leaves Ctrl-W to the browser', async () => {
    const s = setup();
    await s.ready;
    type(s, 'cat about.md');
    const event = key(s.jsdom, s.input, 'w', { ctrlKey: true });

    assert.equal(event.defaultPrevented, false);
    assert.equal(s.input.value, 'cat about.md');
  });
});

describe('before the shell is ready', () => {
  it('keeps lines entered early and runs them once the shell is ready', async () => {
    let resolve: (shell: Shell) => void = () => undefined;
    const pending = new Promise<Shell>(r => {
      resolve = r;
    });
    const s = setup({ shell: pending });

    enter(s, 'whoami');
    enter(s, 'echo hi');
    assert.equal(s.output.querySelectorAll('.echo').length, 0);
    assert.equal(s.input.value, '');

    resolve(s.shell);
    await s.ready;

    assert.deepEqual(execs(s), ['exec:whoami', 'exec:echo hi']);
    assert.equal(s.output.querySelectorAll('.echo').length, 2);
  });
});

describe('theme sync', () => {
  it('tells the shell the stored theme once it is ready', async () => {
    const storage = memoryStorage();
    storage.setItem('theme', 'amber');
    const s = setup({ extra: { storage: () => storage } });
    await s.ready;

    assert.ok(s.shell.calls.includes('setTheme:amber'));
  });

  it('tells the shell the theme already on the page when nothing is stored', async () => {
    const jsdomPage = dom();
    jsdomPage.window.document.documentElement.setAttribute('data-theme', 'light');
    const shell = spyShell();
    await createTerminal({
      win: jsdomPage.window as unknown as Window,
      shell: Promise.resolve(shell),
      storage: () => memoryStorage()
    }).ready;

    assert.ok(shell.calls.includes('setTheme:light'));
  });

  it('does not call setTheme when no theme is stored', async () => {
    const s = setup();
    await s.ready;

    assert.equal(
      s.shell.calls.some(c => c.startsWith('setTheme:')),
      false
    );
  });

  it('ignores a stored theme that is not a known theme', async () => {
    const storage = memoryStorage();
    storage.setItem('theme', 'neon');
    const s = setup({ extra: { storage: () => storage } });
    await s.ready;

    assert.equal(
      s.shell.calls.some(c => c.startsWith('setTheme:')),
      false
    );
  });
});

describe('deep links', () => {
  it('runs an allow-listed read-only command from the hash', async () => {
    const s = setup({ hash: '#help' });
    await s.ready;

    assert.deepEqual(execs(s), ['exec:help']);
    assert.equal(s.input.value, '');
  });

  it('only pre-fills open, never opening a tab', async () => {
    const s = setup({ hash: '#open%20https%3A%2F%2Fexample.com' });
    await s.ready;

    assert.deepEqual(execs(s), []);
    assert.deepEqual(s.opened, []);
    assert.equal(s.input.value, 'open https://example.com');
  });

  it('only pre-fills the live commands', async () => {
    for (const command of ['repos', 'packages', 'camera']) {
      const s = setup({ hash: `#${command}` });
      await s.ready;

      assert.deepEqual(execs(s), [], command);
      assert.equal(s.input.value, command);
    }
  });

  it('only pre-fills a line that chains another command', async () => {
    const s = setup({ hash: '#help%3B%20repos' });
    await s.ready;

    assert.deepEqual(execs(s), []);
    assert.equal(s.input.value, 'help; repos');
  });

  it('never opens a tab or fetches from a deep-linked run even when the shell asks', async () => {
    const base = createFakeShell();
    let fetched = false;
    const shell = spyShell({
      ...base,
      exec: line =>
        line === 'help'
          ? {
              blocks: [],
              effects: [
                { kind: 'open', arg: 'https://example.com/' },
                { kind: 'fetch', arg: 'repos' }
              ]
            }
          : base.exec(line)
    });
    const s = setup({
      hash: '#help',
      shell: Promise.resolve(shell),
      extra: {
        fetchImpl: (async () => {
          fetched = true;
          throw new Error('no');
        }) as unknown as typeof fetch
      }
    });
    await s.ready;
    await tick(5);

    assert.deepEqual(shell.calls, ['boot', 'exec:help']);
    assert.deepEqual(s.opened, []);
    assert.equal(fetched, false);
  });

  const themeOf = (s: Setup): string | null => s.doc.documentElement.getAttribute('data-theme');

  const prefilled = async (hash: string, value: string): Promise<void> => {
    const storage = memoryStorage();
    const s = setup({ hash, extra: { storage: () => storage } });
    await s.ready;

    assert.deepEqual(execs(s), [], hash);
    assert.equal(themeOf(s), null, hash);
    assert.equal(storage.getItem('theme'), null, hash);
    assert.equal(s.input.value, value, hash);
    assert.equal(s.output.querySelectorAll('.entry-boot').length, 1, hash);
  };

  it('only pre-fills theme, changing and storing nothing', async () => {
    await prefilled('#theme%20amber', 'theme amber');
  });

  it('only pre-fills clear and leaves the boot screen', async () => {
    await prefilled('#clear', 'clear');
  });

  it('only pre-fills a read-only command chained with and', async () => {
    await prefilled('#help%20%26%26%20theme%20amber', 'help && theme amber');
  });

  it('only pre-fills a read-only command chained with a semicolon', async () => {
    await prefilled('#help%3Btheme%20amber', 'help;theme amber');
  });

  it('only pre-fills a read-only command chained with a pipe', async () => {
    await prefilled('#help%20%7C%20theme%20amber', 'help | theme amber');
  });

  it('does not run a line containing a newline', async () => {
    const s = setup({ hash: '#help%0Atheme' });
    await s.ready;

    assert.deepEqual(execs(s), []);
    assert.equal(s.doc.documentElement.getAttribute('data-theme'), null);
  });

  it('only pre-fills an unknown first word', async () => {
    await prefilled('#frobnicate', 'frobnicate');
  });

  it('only pre-fills history, echo and exit', async () => {
    for (const command of ['history', 'echo hi', 'exit']) {
      await prefilled(`#${encodeURIComponent(command)}`, command);
    }
  });

  it('runs every read-only command', async () => {
    for (const command of ['man help', 'ls', 'cd projects', 'pwd', 'whoami', 'date', 'skills', 'education']) {
      const s = setup({ hash: `#${encodeURIComponent(command)}` });
      await s.ready;

      assert.deepEqual(execs(s), [`exec:${command}`], command);
    }
  });

  it('ignores theme and clear effects from a deep-linked run', async () => {
    const base = createFakeShell();
    const shell = spyShell({
      ...base,
      exec: line =>
        line === 'help'
          ? {
              blocks: [],
              effects: [
                { kind: 'theme', arg: 'amber' },
                { kind: 'clear', arg: '' }
              ]
            }
          : base.exec(line)
    });
    const storage = memoryStorage();
    const s = setup({ hash: '#help', shell: Promise.resolve(shell), extra: { storage: () => storage } });
    await s.ready;

    assert.equal(themeOf(s), null);
    assert.equal(storage.getItem('theme'), null);
    assert.equal(s.output.querySelectorAll('.entry-boot').length, 1);
  });

  it('keeps a read-only command in the hash after a user run', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'ls');

    assert.equal(s.jsdom.window.location.hash, '#ls');
  });

  it('keeps theme out of the hash when typed by hand', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'theme amber');

    assert.equal(themeOf(s), 'amber');
    assert.equal(s.jsdom.window.location.hash, '');
  });

  it('clears an earlier hash when a non-read-only command is typed', async () => {
    const s = setup();
    await s.ready;
    enter(s, 'ls');
    enter(s, 'theme amber');

    assert.equal(s.jsdom.window.location.hash, '');
  });

  it('ignores a malformed hash', async () => {
    const s = setup({ hash: '#%E0%A4%A' });
    await s.ready;

    assert.deepEqual(execs(s), []);
    assert.equal(s.input.value, '');
  });
});

describe('focus', () => {
  it('focuses the input when the terminal is clicked', async () => {
    const s = setup();
    await s.ready;
    (s.doc.getElementById('scroll') as HTMLElement).click();

    assert.equal(s.doc.activeElement, s.input);
  });
});

describe('input method composition', () => {
  const fire = (s: Setup, type: string): void => {
    s.input.dispatchEvent(new s.jsdom.window.CompositionEvent(type, { bubbles: true }));
  };

  it('ignores Enter while a composition is active', async () => {
    const s = setup();
    await s.ready;
    type(s, 'ni');
    fire(s, 'compositionstart');
    key(s.jsdom, s.input, 'Enter');

    assert.deepEqual(execs(s), []);
    assert.equal(s.input.value, 'ni');
  });

  it('ignores the Enter that confirms a composition right after compositionend', async () => {
    const s = setup();
    await s.ready;
    type(s, 'ni');
    at(s, 'compositionstart', 100);
    at(s, 'compositionend', 200);
    at(s, 'keydown', 200, { key: 'Enter' });

    assert.deepEqual(execs(s), []);
    assert.equal(s.input.value, 'ni');
  });

  const at = (s: Setup, type: string, timeStamp: number, init: KeyboardEventInit = {}): Event => {
    const event =
      type === 'keydown'
        ? new s.jsdom.window.KeyboardEvent('keydown', { bubbles: true, cancelable: true, ...init })
        : new s.jsdom.window.CompositionEvent(type, { bubbles: true });
    Object.defineProperty(event, 'timeStamp', { value: timeStamp });
    s.input.dispatchEvent(event);
    return event;
  };

  it('ignores the keyCode 13 Enter that ended the composition in the same turn', async () => {
    const s = setup();
    await s.ready;
    type(s, 'ni');
    at(s, 'compositionstart', 100);
    at(s, 'compositionend', 200);
    at(s, 'keydown', 200, { key: 'Enter', keyCode: 13 });

    assert.deepEqual(execs(s), []);
    assert.equal(s.input.value, 'ni');
  });

  it('runs a keyCode 13 Enter that arrives before compositionend', async () => {
    const s = setup();
    await s.ready;
    type(s, 'help');
    at(s, 'compositionstart', 100);
    const event = at(s, 'keydown', 200, { key: 'Enter', keyCode: 13 });
    at(s, 'compositionend', 201);

    assert.equal(event.defaultPrevented, true);
    assert.deepEqual(execs(s), ['exec:help']);
  });

  it('runs a keyCode 13 Enter from a later turn even before any timer has fired', async () => {
    const s = setup();
    await s.ready;
    type(s, 'help');
    at(s, 'compositionstart', 100);
    at(s, 'compositionend', 200);
    at(s, 'keydown', 240, { key: 'Enter', keyCode: 13 });

    assert.deepEqual(execs(s), ['exec:help']);
  });

  it('still ignores a keyCode 13 Enter the browser marks as composing', async () => {
    const s = setup();
    await s.ready;
    type(s, 'ni');
    at(s, 'compositionstart', 100);
    at(s, 'keydown', 200, { key: 'Enter', keyCode: 13, isComposing: true });

    assert.deepEqual(execs(s), []);
  });

  it('accepts Enter once the tick after compositionend has passed', async () => {
    const s = setup();
    await s.ready;
    type(s, 'help');
    fire(s, 'compositionstart');
    fire(s, 'compositionend');
    await tick(5);
    key(s.jsdom, s.input, 'Enter');

    assert.deepEqual(execs(s), ['exec:help']);
  });

  it('ignores Enter, Tab and arrows reported with keyCode 229', async () => {
    const s = setup();
    await s.ready;
    type(s, 'he');
    for (const name of ['Enter', 'Tab', 'ArrowUp', 'ArrowDown']) {
      const event = key(s.jsdom, s.input, name, { keyCode: 229 });
      assert.equal(event.defaultPrevented, false, name);
    }

    assert.deepEqual(execs(s), []);
    assert.equal(s.input.value, 'he');
  });
});
