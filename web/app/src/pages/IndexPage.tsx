import React, { ReactElement, useEffect, useRef, useState } from 'react';

// A lightweight terminal UI that talks to Go WASM via window.TerminalAPI

type TerminalAPI = {
  init: () => string | Promise<string>;
  handleInput: (line: string) => string | Promise<string>;
  getPrompt: () => string | Promise<string>;
  reset: () => void | Promise<void>;
};

declare global {
  interface Window {
    TerminalAPI?: TerminalAPI;
    terminalReady?: Promise<boolean>;
  }
}

const IndexPage = (): ReactElement => {
  const [ready, setReady] = useState(false);
  const [prompt, setPrompt] = useState('');
  const [lines, setLines] = useState<string[]>(['Loading WASM terminal...']);
  const [input, setInput] = useState('');
  const historyRef = useRef<string[]>([]);
  const historyIdxRef = useRef<number>(-1);
  const inputRef = useRef<HTMLInputElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const waitForWasm = (): Promise<void> => {
      return new Promise<void>((resolve, reject) => {
        const start = Date.now();
        const timeoutMs = 15000;
        const tick = async () => {
          try {
            if (window.terminalReady) {
              const ok = await window.terminalReady;
              if (ok) return resolve();
            }
            if (window.TerminalAPI) return resolve();
            if (Date.now() - start > timeoutMs) return reject(new Error('WASM init timeout'));
            setTimeout(tick, 50);
          } catch (e) {
            reject(e as Error);
          }
        };
        tick();
      });
    };

    const boot = async () => {
      try {
        await waitForWasm();
        if (!window.TerminalAPI) {
          setLines(prev => [...prev, 'Failed to load Terminal API. Ensure wasm.wasm and wasm.js are available.']);
          return;
        }
        const p = await window.TerminalAPI.init();
        setPrompt(p);
        setLines(['Welcome to the WebAssembly shell (Go + WASM).', "Type 'help' to see available commands."]);
        setReady(true);
        // focus input when ready
        setTimeout(() => inputRef.current?.focus(), 0);
      } catch (e) {
        setLines(prev => [...prev, `Error initializing terminal: ${e}`]);
      }
    };
    boot();
  }, []);

  useEffect(() => {
    // keep view scrolled to bottom
    scrollRef.current && (scrollRef.current.scrollTop = scrollRef.current.scrollHeight);
  }, [lines, prompt]);

  const runCommand = async (cmd: string) => {
    if (!window.TerminalAPI) return;
    // record history
    if (cmd.trim()) {
      historyRef.current.unshift(cmd);
      historyIdxRef.current = -1;
    }
    const header = `${prompt}${cmd}`;
    const output = await window.TerminalAPI.handleInput(cmd);
    if (output === '__CLEAR__') {
      setLines([]);
    } else if (output && output.length > 0) {
      setLines(prev => [...prev, header, ...output.split('\n')]);
    } else {
      setLines(prev => [...prev, header]);
    }
    const p = await window.TerminalAPI.getPrompt();
    setPrompt(p);
    setInput('');
  };

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      runCommand(input);
      return;
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      const nextIdx = Math.min(historyRef.current.length - 1, historyIdxRef.current + 1);
      historyIdxRef.current = nextIdx;
      setInput(historyRef.current[nextIdx] ?? input);
      return;
    }
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      const nextIdx = Math.max(-1, historyIdxRef.current - 1);
      historyIdxRef.current = nextIdx;
      setInput(nextIdx === -1 ? '' : (historyRef.current[nextIdx] ?? ''));
      return;
    }
    if (e.key === 'c' && (e.ctrlKey || e.metaKey)) {
      // Ctrl+C clears current line
      e.preventDefault();
      setInput('');
      return;
    }
  };

  return (
    <div id="index-page" className="index-page">
      <div className="index-page__content">
        <div id="wasm-terminal" className="wasm-terminal">
          <div className="wasm-terminal__header">
            <span className="wasm-terminal__title wasm-terminal__title--blue">Go+WASM</span>{' '}
            <span className="wasm-terminal__title wasm-terminal__title--green">Terminal</span>
          </div>
          <div ref={scrollRef} className="wasm-terminal__body">
            {lines.map((line, i) => (
              <div key={i} className="wasm-terminal__line">
                {line}
              </div>
            ))}
            {ready && (
              <div className="wasm-terminal__input-row">
                <div className="wasm-terminal__prompt">{prompt}</div>
                <input
                  ref={inputRef}
                  value={input}
                  onChange={e => setInput(e.target.value)}
                  onKeyDown={onKeyDown}
                  spellCheck={false}
                  autoCapitalize="off"
                  autoCorrect="off"
                  className="wasm-terminal__input"
                />
              </div>
            )}
          </div>
          <div className="wasm-terminal__footer">
            {ready ? 'Type help, clear, ls, cd, cat, echo, mkdir, touch, rm' : 'Initializing...'}
          </div>
        </div>
      </div>
    </div>
  );
};

export default IndexPage;
