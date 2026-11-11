import type { Completion } from './bridge';

export interface EditorHooks {
  submit(line: string): void;
  complete(line: string, cursor: number): Completion | null;
  showCandidates(line: string, candidates: string[]): void;
  history(): string[];
  clearScreen(): void;
  interrupt(line: string): void;
  hasSelection(): boolean;
}

export interface Editor {
  set(value: string): void;
}

const SAME_TURN_MS = 1;

const toCodePoints = (value: string, offset: number): number => Array.from(value.slice(0, offset)).length;

const toUnits = (value: string, cursor: number): number => Array.from(value).slice(0, cursor).join('').length;

export const attachEditor = (input: HTMLInputElement, hooks: EditorHooks): Editor => {
  let index: number | null = null;
  let draft = '';
  let escaped = false;
  let composing = false;
  let composedAt: number | null = null;

  const set = (value: string, cursor = value.length): void => {
    input.value = value;
    input.setSelectionRange(cursor, cursor);
  };

  const cursor = (): number => input.selectionStart ?? input.value.length;

  const complete = (): void => {
    const value = input.value;
    const result = hooks.complete(value, toCodePoints(value, cursor()));
    if (result === null) {
      return;
    }
    set(result.line, toUnits(result.line, result.cursor));
    if (result.candidates.length > 1) {
      hooks.showCandidates(result.line, result.candidates);
    }
  };

  const recall = (step: number): boolean => {
    const list = hooks.history();
    if (list.length === 0 || (index === null && step > 0)) {
      return false;
    }
    if (index === null) {
      draft = input.value;
      index = list.length;
    }
    index += step;
    if (index >= list.length) {
      index = null;
      set(draft);
      return true;
    }
    index = Math.max(0, index);
    set(list[index]);
    return true;
  };

  const control = (key: string): boolean => {
    const value = input.value;
    const at = cursor();
    switch (key) {
      case 'c':
        if (input.selectionStart !== input.selectionEnd || hooks.hasSelection()) {
          return false;
        }
        index = null;
        set('');
        hooks.interrupt(value);
        return true;
      case 'l':
        hooks.clearScreen();
        return true;
      case 'a':
        set(value, 0);
        return true;
      case 'e':
        set(value);
        return true;
      case 'u':
        set(value.slice(at), 0);
        return true;
      case 'k':
        set(value.slice(0, at));
        return true;
      default:
        return false;
    }
  };

  input.addEventListener('compositionstart', () => {
    composing = true;
    composedAt = null;
  });

  input.addEventListener('compositionend', event => {
    composing = false;
    composedAt = event.timeStamp;
  });

  input.addEventListener('keydown', event => {
    if (event.isComposing || event.keyCode === 229) {
      return;
    }
    const endedThisTurn = composedAt !== null && Math.abs(event.timeStamp - composedAt) < SAME_TURN_MS;
    const plainEnter = event.key === 'Enter' && event.keyCode === 13;
    if (endedThisTurn || (composing && !plainEnter)) {
      return;
    }
    composedAt = null;
    const afterEscape = escaped;
    escaped = event.key === 'Escape';

    if (event.key === 'Enter') {
      event.preventDefault();
      const line = input.value;
      index = null;
      set('');
      hooks.submit(line);
      return;
    }
    if (event.key === 'Tab') {
      if (afterEscape || (event.shiftKey && input.value === '')) {
        return;
      }
      event.preventDefault();
      if (!event.shiftKey) {
        complete();
      }
      return;
    }
    if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
      if (recall(event.key === 'ArrowUp' ? -1 : 1)) {
        event.preventDefault();
      }
      return;
    }
    if (event.ctrlKey && !event.metaKey && !event.altKey && control(event.key.toLowerCase())) {
      event.preventDefault();
    }
  });

  return { set: value => set(value) };
};
