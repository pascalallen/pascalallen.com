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
  setTheme(name: string): void;
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
  runtimeSrc?: string;
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
    setTheme: name => {
      call(raw, 'setTheme', name);
    },
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

const instantiate = async (pending: Promise<Response>, go: GoRuntime, src: string): Promise<WebAssembly.Instance> => {
  const response = await pending;
  if (!response.ok) {
    throw new Error(`shell module: HTTP ${response.status} for ${src}`);
  }
  try {
    const streamed = await WebAssembly.instantiateStreaming(Promise.resolve(response), go.importObject);
    return streamed.instance;
  } catch {
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

const JS_TYPE = /(java|ecma)script/i;

const loadRuntime = async (win: Window, src: string, fetchImpl: typeof fetch): Promise<void> => {
  const response = await fetchImpl(src);
  if (!response.ok) {
    throw new Error(`shell runtime: HTTP ${response.status}`);
  }
  if (!JS_TYPE.test(response.headers.get('Content-Type') ?? '')) {
    throw new Error('shell runtime: not a script');
  }
  await injectScript(win, src);
};

export const loadShell = (options: LoadOptions): Promise<Shell> => {
  const win = options.win ?? window;
  const w = win as unknown as ShellWindow;
  const fetchImpl: typeof fetch = options.fetchImpl ?? ((input, init) => win.fetch(input, init));

  const pendingModule: Promise<Response> | null =
    options.devShellSrc === undefined ? new Promise<Response>(accept => accept(fetchImpl(options.wasmSrc))) : null;
  pendingModule?.catch(() => undefined);

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

    const ensureRuntime =
      typeof w.Go === 'function' || options.runtimeSrc === undefined
        ? Promise.resolve()
        : loadRuntime(win, options.runtimeSrc, fetchImpl);
    ensureRuntime
      .then(() => {
        const Go = w.Go;
        if (typeof Go !== 'function') {
          throw new Error('shell module: Go runtime missing');
        }
        const go = new Go();
        if (pendingModule === null) {
          throw new Error('shell module: not requested');
        }
        return instantiate(pendingModule, go, options.wasmSrc).then(instance => go.run(instance));
      })
      .then(() => fail(new Error('shell module: exited')))
      .catch(fail);
  });
};
