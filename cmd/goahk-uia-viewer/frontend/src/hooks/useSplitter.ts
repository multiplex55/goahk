import { useCallback, useEffect, useState } from 'react';

export function clampSize(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

export function restorePersistedSize(rawValue: string | null, min: number, max: number): number | null {
  if (!rawValue) {
    return null;
  }
  const parsed = Number(rawValue);
  if (!Number.isFinite(parsed)) {
    return null;
  }
  return clampSize(parsed, min, max);
}

type UseSplitterOptions = {
  storageKey: string;
  defaultSizePx: number;
  minSizePx: number;
  maxSizePx: number;
};

export default function useSplitter({ storageKey, defaultSizePx, minSizePx, maxSizePx }: UseSplitterOptions) {
  const [sizePx, setSizePx] = useState(() => {
    const fallback = clampSize(defaultSizePx, minSizePx, maxSizePx);
    if (typeof window === 'undefined') {
      return fallback;
    }

    const restored = restorePersistedSize(window.localStorage.getItem(storageKey), minSizePx, maxSizePx);
    return restored ?? fallback;
  });

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(storageKey, String(sizePx));
  }, [sizePx, storageKey]);

  const setClampedSize = useCallback(
    (next: number | ((current: number) => number)) => {
      setSizePx((current) => {
        const candidate = typeof next === 'function' ? next(current) : next;
        return clampSize(candidate, minSizePx, maxSizePx);
      });
    },
    [maxSizePx, minSizePx]
  );

  const applyDelta = useCallback(
    (deltaPx: number) => {
      setClampedSize((current) => current + deltaPx);
    },
    [setClampedSize]
  );

  return {
    sizePx,
    setSizePx: setClampedSize,
    applyDelta
  };
}
