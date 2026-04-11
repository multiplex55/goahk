import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import useSplitter, { clampSize, restorePersistedSize } from './useSplitter';

describe('useSplitter helpers', () => {
  it('clamps values to min and max', () => {
    expect(clampSize(100, 200, 500)).toBe(200);
    expect(clampSize(250, 200, 500)).toBe(250);
    expect(clampSize(900, 200, 500)).toBe(500);
  });

  it('sanitizes persisted values and rejects invalid numbers', () => {
    expect(restorePersistedSize('250', 200, 500)).toBe(250);
    expect(restorePersistedSize('999', 200, 500)).toBe(500);
    expect(restorePersistedSize('-42', 200, 500)).toBe(200);
    expect(restorePersistedSize('nope', 200, 500)).toBeNull();
    expect(restorePersistedSize(null, 200, 500)).toBeNull();
  });
});

describe('useSplitter', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('restores persisted size and persists updates', () => {
    window.localStorage.setItem('splitter:test', '450');

    const { result } = renderHook(() =>
      useSplitter({
        storageKey: 'splitter:test',
        defaultSizePx: 300,
        minSizePx: 200,
        maxSizePx: 500
      })
    );

    expect(result.current.sizePx).toBe(450);

    act(() => {
      result.current.applyDelta(100);
    });

    expect(result.current.sizePx).toBe(500);
    expect(window.localStorage.getItem('splitter:test')).toBe('500');
  });

  it('falls back to default when persisted value is invalid', () => {
    window.localStorage.setItem('splitter:invalid', 'bad-value');

    const { result } = renderHook(() =>
      useSplitter({
        storageKey: 'splitter:invalid',
        defaultSizePx: 320,
        minSizePx: 200,
        maxSizePx: 500
      })
    );

    expect(result.current.sizePx).toBe(320);
    expect(window.localStorage.getItem('splitter:invalid')).toBe('320');
  });
});
