import React, { ReactElement, useEffect, useRef, useState } from 'react';

type HistoryEntry = {
  id: number;
  cmd: string;
  output: string;
};

const TerminalBio = (): ReactElement => {
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [input, setInput] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);
  const bodyRef = useRef<HTMLDivElement>(null);
  const keyRef = useRef(0);

  useEffect(() => {
    if (bodyRef.current) {
      bodyRef.current.scrollTop = bodyRef.current.scrollHeight;
    }
  }, [history]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>): void => {
    if (e.key !== 'Enter') return;
    const cmd = input.trim();
    if (!cmd) return;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    if (typeof (window as any).executeCommand !== 'function') {
      setHistory(prev => [
        ...prev,
        { id: keyRef.current++, cmd, output: 'terminal initializing, please try again...' }
      ]);
      setInput('');
      return;
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const result: string = (window as any).executeCommand(cmd);
    if (result === '__CLEAR__') {
      setHistory([]);
    } else {
      setHistory(prev => [...prev, { id: keyRef.current++, cmd, output: result }]);
    }
    setInput('');
  };

  const focusInput = (): void => {
    inputRef.current?.focus();
  };

  return (
    <div className="terminal-window">
      <div className="terminal-title-bar">bash &mdash; 80&times;24</div>
      <div className="terminal-body" ref={bodyRef} onClick={focusInput}>
        {history.map(entry => (
          <React.Fragment key={entry.id}>
            <div className="terminal-line">
              <span className="terminal-prompt">user@pascalallen:~$&nbsp;</span>
              {entry.cmd}
            </div>
            <div className="terminal-output">{entry.output}</div>
          </React.Fragment>
        ))}
        <div className="terminal-line">
          <span className="terminal-prompt">user@pascalallen:~$&nbsp;</span>
          <input
            ref={inputRef}
            className="terminal-input"
            aria-label="terminal command input"
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            autoFocus
            spellCheck={false}
            autoComplete="off"
          />
        </div>
      </div>
    </div>
  );
};

export default TerminalBio;
