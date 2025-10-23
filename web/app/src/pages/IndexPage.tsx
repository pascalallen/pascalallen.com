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
    <div
      id="index-page"
      className="index-page"
      style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 16 }}>
        <div
          id="wasm-terminal"
          style={{
            width: '100%',
            maxWidth: 900,
            height: 500,
            background: '#0b0f14',
            color: '#cfe3ff',
            borderRadius: 8,
            border: '1px solid #1e2733',
            fontFamily: 'SFMono-Regular,Consolas,Menlo,monospace',
            fontSize: 14,
            display: 'flex',
            flexDirection: 'column',
            boxShadow: '0 10px 30px rgba(0,0,0,0.3)'
          }}>
          <div style={{ padding: '8px 12px', borderBottom: '1px solid #1e2733', background: '#0f141b' }}>
            <span style={{ color: '#68a0ff' }}>Go+WASM</span> <span style={{ color: '#6dd5a3' }}>Terminal</span>
          </div>
          <div ref={scrollRef} style={{ flex: 1, overflowY: 'auto', padding: '12px' }}>
            {lines.map((line, i) => (
              <div key={i} style={{ whiteSpace: 'pre-wrap' }}>
                {line}
              </div>
            ))}
            {ready && (
              <div style={{ display: 'flex' }}>
                <div style={{ whiteSpace: 'pre' }}>{prompt}</div>
                <input
                  ref={inputRef}
                  value={input}
                  onChange={e => setInput(e.target.value)}
                  onKeyDown={onKeyDown}
                  spellCheck={false}
                  autoCapitalize="off"
                  autoCorrect="off"
                  style={{
                    flex: 1,
                    background: 'transparent',
                    border: 'none',
                    outline: 'none',
                    color: '#cfe3ff',
                    font: 'inherit'
                  }}
                />
              </div>
            )}
          </div>
          <div
            style={{
              padding: '6px 12px',
              borderTop: '1px solid #1e2733',
              background: '#0f141b',
              fontSize: 12,
              color: '#6b7c93'
            }}>
            {ready ? 'Type help, clear, ls, cd, cat, echo, mkdir, touch, rm' : 'Initializing...'}
          </div>
        </div>
      </div>
    </div>
  );
};

export default IndexPage;
