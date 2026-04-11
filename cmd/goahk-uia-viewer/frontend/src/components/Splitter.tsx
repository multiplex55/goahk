import { KeyboardEvent, MouseEvent, PointerEvent, useEffect, useMemo, useRef, useState } from 'react';

type SplitterProps = {
  onDrag: (deltaPx: number) => void;
  ariaLabel: string;
  orientation?: 'vertical' | 'horizontal';
  valueNow?: number;
  valueMin?: number;
  valueMax?: number;
  keyboardStepPx?: number;
};

export default function Splitter({
  onDrag,
  ariaLabel,
  orientation = 'vertical',
  valueNow,
  valueMin,
  valueMax,
  keyboardStepPx = 16
}: SplitterProps) {
  const [dragging, setDragging] = useState(false);
  const pointerIdRef = useRef<number | null>(null);
  const startPosRef = useRef(0);

  useEffect(() => {
    if (typeof document === 'undefined') {
      return;
    }
    const className = orientation === 'vertical' ? 'splitter-global-drag-col' : 'splitter-global-drag-row';
    if (dragging) {
      document.body.classList.add(className);
    }
    return () => {
      document.body.classList.remove(className);
    };
  }, [dragging, orientation]);

  const axis = orientation === 'vertical' ? 'x' : 'y';

  const className = useMemo(
    () => ['splitter', orientation === 'horizontal' ? 'splitter-horizontal' : 'splitter-vertical', dragging ? 'splitter-dragging' : ''].filter(Boolean).join(' '),
    [dragging, orientation]
  );

  const readPointerPos = (event: Pick<globalThis.PointerEvent, 'clientX' | 'clientY'>) => {
    const value = axis === 'x' ? event.clientX : event.clientY;
    return Number.isFinite(value) ? value : 0;
  };

  const onPointerDown = (event: PointerEvent<HTMLDivElement>) => {
    pointerIdRef.current = event.pointerId;
    startPosRef.current = readPointerPos(event);
    setDragging(true);
    if (typeof event.currentTarget.setPointerCapture === 'function') {
      event.currentTarget.setPointerCapture(event.pointerId);
    }
  };

  const onPointerMove = (event: PointerEvent<HTMLDivElement>) => {
    if (pointerIdRef.current !== event.pointerId) {
      return;
    }
    const currentPos = readPointerPos(event);
    const delta = currentPos - startPosRef.current;
    if (delta === 0) {
      return;
    }
    startPosRef.current = currentPos;
    onDrag(delta);
  };


  const onMouseDown = (event: MouseEvent<HTMLDivElement>) => {
    pointerIdRef.current = -1;
    startPosRef.current = axis === 'x' ? event.clientX : event.clientY;
    setDragging(true);
  };

  const onMouseMove = (event: MouseEvent<HTMLDivElement>) => {
    if (pointerIdRef.current !== -1) {
      return;
    }
    const currentPos = axis === 'x' ? event.clientX : event.clientY;
    const delta = currentPos - startPosRef.current;
    if (delta === 0) {
      return;
    }
    startPosRef.current = currentPos;
    onDrag(delta);
  };

  const stopDragging = (pointerId?: number) => {
    if (pointerIdRef.current !== null && (pointerId === undefined || pointerIdRef.current === pointerId)) {
      pointerIdRef.current = null;
      setDragging(false);
    }
  };

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    const positiveKey = orientation === 'vertical' ? 'ArrowRight' : 'ArrowDown';
    const negativeKey = orientation === 'vertical' ? 'ArrowLeft' : 'ArrowUp';

    if (event.key === positiveKey) {
      event.preventDefault();
      onDrag(keyboardStepPx);
      return;
    }

    if (event.key === negativeKey) {
      event.preventDefault();
      onDrag(-keyboardStepPx);
      return;
    }

    if (event.key === 'PageUp') {
      event.preventDefault();
      onDrag(-keyboardStepPx * 4);
      return;
    }

    if (event.key === 'PageDown') {
      event.preventDefault();
      onDrag(keyboardStepPx * 4);
    }
  };

  return (
    <div
      role="separator"
      aria-orientation={orientation}
      aria-label={ariaLabel}
      aria-valuenow={valueNow}
      aria-valuemin={valueMin}
      aria-valuemax={valueMax}
      tabIndex={0}
      className={className}
      onPointerDown={onPointerDown}
      onPointerMove={onPointerMove}
      onPointerUp={(event) => stopDragging(event.pointerId)}
      onPointerCancel={(event) => stopDragging(event.pointerId)}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={() => stopDragging(-1)}
      onMouseLeave={() => stopDragging(-1)}
      onBlur={() => stopDragging()}
      onKeyDown={onKeyDown}
    />
  );
}
