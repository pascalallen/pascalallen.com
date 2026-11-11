import type { Block, CommandInfo, Completion, FeedItem, Line, Result, Shell, Span } from './bridge';

const span = (text: string, extra: Partial<Span> = {}): Span => ({ text, style: 'plain', href: '', run: '', ...extra });

const text = (...lines: Line[]): Block => ({ kind: 'text', lines, src: '', alt: '' });

const error = (message: string): Block => ({
  kind: 'error',
  lines: [[span(message, { style: 'error' })]],
  src: '',
  alt: ''
});

const result = (blocks: Block[], effects: Result['effects'] = []): Result => ({ blocks, effects });

const files: Record<string, string[]> = {
  'about.md': [
    '# About',
    'Software engineer and computer scientist in Austin, Texas.',
    'I build Go services, MCP servers and the tools around them.',
    'Writing: https://medium.com/@pascal.allen88'
  ],
  'contact.md': ['# Contact', 'mail: mailto:me@pascalallen.com', 'code: https://github.com/pascalallen'],
  'resume.md': ['# Resume', 'Go, distributed systems, DDD, CQRS, hexagonal architecture.']
};

const directories = ['projects/', 'writing/'];

const visible: CommandInfo[] = [
  { name: 'help', synopsis: 'help [command]', summary: 'list commands' },
  { name: 'about', synopsis: 'about', summary: 'who I am' },
  { name: 'cat', synopsis: 'cat <file>...', summary: 'print files' },
  { name: 'clear', synopsis: 'clear', summary: 'clear the screen' },
  { name: 'contact', synopsis: 'contact', summary: 'how to reach me' },
  { name: 'echo', synopsis: 'echo [text]...', summary: 'print text' },
  { name: 'history', synopsis: 'history', summary: 'list history' },
  { name: 'ls', synopsis: 'ls [path]', summary: 'list files' },
  { name: 'whoami', synopsis: 'whoami', summary: 'print the user' },
  { name: 'theme', synopsis: 'theme [name]', summary: 'switch theme' },
  { name: 'open', synopsis: 'open <url>', summary: 'open a link' },
  { name: 'repos', synopsis: 'repos', summary: 'live GitHub repositories' },
  { name: 'packages', synopsis: 'packages', summary: 'live npm packages' },
  { name: 'camera', synopsis: 'camera', summary: 'live camera' }
];

const themes = ['dark', 'light', 'amber', 'green'];

const urlPattern = /(https?:\/\/\S+|mailto:\S+)/;

const renderFileLine = (line: string): Line => {
  if (line.startsWith('# ')) {
    return [span(line.slice(2), { style: 'heading' })];
  }
  return line
    .split(urlPattern)
    .filter(part => part !== '')
    .map(part => (urlPattern.test(part) ? span(part, { style: 'accent', href: part }) : span(part)));
};

const chips = (names: string[]): Line =>
  names.flatMap((name, i) => {
    const button = span(name, { style: 'accent', run: name });
    return i === 0 ? [button] : [span('  '), button];
  });

const commonPrefix = (words: string[]): string =>
  words.reduce((prefix, word) => {
    let i = 0;
    while (i < prefix.length && prefix[i] === word[i]) {
      i++;
    }
    return prefix.slice(0, i);
  });

export const createFakeShell = (): Shell => {
  const history: string[] = [];

  const cat = (args: string[]): Result => {
    if (args.length === 0) {
      return result([error('cat: missing file operand')]);
    }
    const blocks: Block[] = [];
    for (const name of args) {
      const body = files[name.replace(/^~\//, '')];
      blocks.push(
        body === undefined ? error(`cat: ${name}: No such file or directory`) : text(...body.map(renderFileLine))
      );
    }
    return result(blocks);
  };

  const run = (name: string, args: string[]): Result => {
    switch (name) {
      case '':
        return result([]);
      case 'help':
        return result([
          text(
            [span('Commands', { style: 'heading' })],
            ...visible.map(c => [
              span(c.name, { style: 'accent', run: c.name }),
              span('  '),
              span(c.summary, { style: 'dim' })
            ])
          )
        ]);
      case 'about':
        return cat(['about.md']);
      case 'contact':
        return cat(['contact.md']);
      case 'cat':
        return cat(args);
      case 'ls':
        return result([
          text([
            ...Object.keys(files).flatMap(f => [span(f, { run: `cat ${f}` }), span('  ')]),
            ...directories.flatMap((d, i) => [
              span(d, { style: 'path' }),
              ...(i < directories.length - 1 ? [span('  ')] : [])
            ])
          ])
        ]);
      case 'echo':
        return result([text([span(args.join(' '))])]);
      case 'whoami':
        return result([text([span('visitor')])]);
      case 'history':
        return result([text(...history.map((h, i) => [span(`${String(i + 1).padStart(5)}  ${h}`)]))]);
      case 'clear':
        return result([], [{ kind: 'clear', arg: '' }]);
      case 'theme':
        if (args.length === 0) {
          return result([text([span(`themes: ${themes.join(' ')}`)])]);
        }
        if (!themes.includes(args[0])) {
          return result([error(`theme: ${args[0]}: unknown theme`)]);
        }
        return result(
          [text([span('theme set to '), span(args[0], { style: 'success' })])],
          [{ kind: 'theme', arg: args[0] }]
        );
      case 'open':
        if (!/^(https?:|mailto:)/.test(args[0] ?? '')) {
          return result([error('open: only http, https and mailto links')]);
        }
        return result(
          [text([span('opening '), span(args[0], { style: 'accent', href: args[0] })])],
          [{ kind: 'open', arg: args[0] }]
        );
      case 'repos':
      case 'packages':
        return result([text([span(`fetching ${name}...`, { style: 'dim' })])], [{ kind: 'fetch', arg: name }]);
      case 'camera':
        return result([
          {
            kind: 'media',
            lines: [[span('live from a Raspberry Pi; type clear to close it', { style: 'dim' })]],
            src: '/api/v1/camera/stream',
            alt: 'Live camera stream from a Raspberry Pi'
          }
        ]);
      default:
        return result([error(`bash: ${name}: command not found`)]);
    }
  };

  const complete = (line: string, cursor: number): Completion => {
    const chars = Array.from(line);
    const before = chars.slice(0, cursor).join('');
    const after = chars.slice(cursor).join('');
    const start = before.lastIndexOf(' ') + 1;
    const word = before.slice(start);
    const pool = start === 0 ? visible.map(c => c.name) : [...Object.keys(files), ...directories];
    const matches = pool.filter(p => p.startsWith(word)).sort();
    if (matches.length === 0) {
      return { line, cursor, candidates: [] };
    }
    const done = matches.length === 1 ? matches[0] + (matches[0].endsWith('/') ? '' : ' ') : commonPrefix(matches);
    const head = before.slice(0, start) + done;
    return {
      line: head + after,
      cursor: Array.from(head).length,
      candidates: matches.length === 1 ? [] : matches
    };
  };

  return {
    boot: () =>
      result([
        text(
          [span('pascalallen.com', { style: 'heading' })],
          [span('Go engineer. Type '), span('help', { style: 'accent', run: 'help' }), span(' or tap a command.')],
          [],
          chips(['about', 'projects', 'writing', 'repos', 'contact'])
        )
      ]),
    exec: line => {
      const trimmed = line.trim();
      if (trimmed !== '' && history[history.length - 1] !== trimmed) {
        history.push(trimmed);
      }
      const segments = trimmed.split(';').map(s =>
        s
          .trim()
          .split(/\s+/)
          .filter(w => w !== '')
      );
      const out: Result = result([]);
      for (const words of segments) {
        const r = run(words[0] ?? '', words.slice(1));
        if (r.effects.some(e => e.kind === 'clear')) {
          out.blocks = [];
        }
        out.blocks.push(...r.blocks);
        out.effects.push(...r.effects);
      }
      return out;
    },
    complete,
    prompt: () => 'visitor@pascalallen.com:~$',
    history: () => [...history],
    setTheme: () => undefined,
    commands: () => visible.map(c => ({ ...c })),
    deliver: (feed: string, items: FeedItem[]) =>
      result([
        text(
          [span(feed === 'repos' ? 'Repositories on GitHub' : 'Packages on npm', { style: 'heading' })],
          ...items.flatMap(item => [
            [span(item.name, { style: 'accent', href: item.url }), span(`  ${item.meta}`, { style: 'dim' })],
            [span(`  ${item.description}`)]
          ])
        )
      ]),
    deliverError: (feed: string) =>
      result([
        {
          kind: 'error',
          lines: [
            [span(`${feed}: could not load the list right now`, { style: 'error' })],
            [span('try again: '), span(feed, { style: 'accent', run: feed })]
          ],
          src: '',
          alt: ''
        }
      ])
  };
};
