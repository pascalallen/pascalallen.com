export interface Span {
  text: string;
  style: string;
  href: string;
  run: string;
}

export type Line = Span[];

export interface Block {
  kind: string;
  lines: Line[];
  src: string;
  alt: string;
}

export interface Effect {
  kind: string;
  arg: string;
}

export interface Result {
  blocks: Block[];
  effects: Effect[];
}

export interface Completion {
  line: string;
  cursor: number;
  candidates: string[];
}

export interface CommandInfo {
  name: string;
  synopsis: string;
  summary: string;
}

export interface FeedItem {
  name: string;
  description: string;
  url: string;
  meta: string;
}

export interface Shell {
  boot(): Result;
  exec(line: string): Result;
  complete(line: string, cursor: number): Completion;
  prompt(): string;
  history(): string[];
  commands(): CommandInfo[];
  deliver(feed: string, items: FeedItem[]): Result;
  deliverError(feed: string): Result;
}

interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

interface ShellWindow {
  __shell?: unknown;
  __shellReady?: () => void;
  Go?: new () => GoRuntime;
}

export interface LoadOptions {
  wasmSrc: string;
  devShellSrc?: string;
  timeoutMs?: number;
  win?: Window;
  fetchImpl?: typeof fetch;
}

type Raw = Record<string, unknown>;

const str = (value: unknown): string => (typeof value === 'string' ? value : '');

const list = (value: unknown): unknown[] => (Array.isArray(value) ? value : []);

const record = (value: unknown): Raw => (value !== null && typeof value === 'object' ? (value as Raw) : {});

const toSpan = (value: unknown): Span => {
  const raw = record(value);
  return { text: str(raw.text), style: str(raw.style), href: str(raw.href), run: str(raw.run) };
};

const toBlock = (value: unknown): Block => {
  const raw = record(value);
  return {
    kind: str(raw.kind),
    lines: list(raw.lines).map(line => list(line).map(toSpan)),
    src: str(raw.src),
    alt: str(raw.alt)
  };
};

export const toResult = (value: unknown): Result => {
  const raw = record(value);
  return {
    blocks: list(raw.blocks).map(toBlock),
    effects: list(raw.effects).map(effect => {
      const e = record(effect);
      return { kind: str(e.kind), arg: str(e.arg) };
    })
  };
};

const toCompletion = (value: unknown, line: string, cursor: number): Completion => {
  const raw = record(value);
  return {
    line: typeof raw.line === 'string' ? raw.line : line,
    cursor: typeof raw.cursor === 'number' ? raw.cursor : cursor,
    candidates: list(raw.candidates).map(str)
  };
};

const call = (raw: Raw, name: string, ...args: unknown[]): unknown => {
  const fn = raw[name];
  return typeof fn === 'function' ? fn.apply(raw, args) : undefined;
};

export const wrap = (value: unknown): Shell => {
  const raw = record(value);
  return {
    boot: () => toResult(call(raw, 'boot')),
    exec: line => toResult(call(raw, 'exec', line)),
    complete: (line, cursor) => toCompletion(call(raw, 'complete', line, cursor), line, cursor),
    prompt: () => str(call(raw, 'prompt')),
    history: () => list(call(raw, 'history')).map(str),
    commands: () =>
      list(call(raw, 'commands')).map(item => {
        const c = record(item);
        return { name: str(c.name), synopsis: str(c.synopsis), summary: str(c.summary) };
      }),
    deliver: (feed, items) => toResult(call(raw, 'deliver', feed, items)),
    deliverError: feed => toResult(call(raw, 'deliverError', feed))
  };
};

export const installShell = (shell: Shell, win: Window = window): void => {
  const w = win as unknown as ShellWindow;
  w.__shell = shell;
  w.__shellReady?.();
};

const instantiate = async (src: string, go: GoRuntime, fetchImpl: typeof fetch): Promise<WebAssembly.Instance> => {
  try {
    const streamed = await WebAssembly.instantiateStreaming(fetchImpl(src), go.importObject);
    return streamed.instance;
  } catch {
    const response = await fetchImpl(src);
    if (!response.ok) {
      throw new Error(`shell module: HTTP ${response.status}`);
    }
    const compiled = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
    return compiled.instance;
  }
};

const injectScript = (win: Window, src: string): Promise<void> =>
  new Promise((resolve, reject) => {
    const script = win.document.createElement('script');
    script.src = src;
    script.async = true;
    script.addEventListener('load', () => resolve());
    script.addEventListener('error', () => reject(new Error(`could not load ${src}`)));
    win.document.head.appendChild(script);
  });

export const loadShell = (options: LoadOptions): Promise<Shell> => {
  const win = options.win ?? window;
  const w = win as unknown as ShellWindow;
  const fetchImpl: typeof fetch = options.fetchImpl ?? ((input, init) => win.fetch(input, init));

  return new Promise<Shell>((resolve, reject) => {
    const timer = win.setTimeout(() => reject(new Error('shell module: timed out')), options.timeoutMs ?? 30000);
    const fail = (error: unknown): void => {
      win.clearTimeout(timer);
      reject(error);
    };
    w.__shellReady = () => {
      win.clearTimeout(timer);
      if (w.__shell === undefined) {
        reject(new Error('shell module: ready without a shell'));
        return;
      }
      resolve(wrap(w.__shell));
    };

    if (options.devShellSrc !== undefined) {
      injectScript(win, options.devShellSrc).catch(fail);
      return;
    }

    const Go = w.Go;
    if (typeof Go !== 'function') {
      fail(new Error('shell module: Go runtime missing'));
      return;
    }
    const go = new Go();
    instantiate(options.wasmSrc, go, fetchImpl)
      .then(instance => go.run(instance))
      .then(() => fail(new Error('shell module: exited')))
      .catch(fail);
  });
};
