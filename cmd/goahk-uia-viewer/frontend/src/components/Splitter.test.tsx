import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import Splitter from './Splitter';
import useSplitter from '../hooks/useSplitter';

function SplitterHarness() {
  const splitter = useSplitter({
    storageKey: 'splitter:harness',
    defaultSizePx: 300,
    minSizePx: 200,
    maxSizePx: 500
  });

  return (
    <>
      <output data-testid="width">{splitter.sizePx}</output>
      <Splitter
        ariaLabel="Resize harness"
        valueNow={splitter.sizePx}
        valueMin={200}
        valueMax={500}
        onDrag={splitter.applyDelta}
      />
    </>
  );
}

describe('Splitter', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  afterEach(() => {
    cleanup();
  });

  it('updates width on pointer drag and enforces min/max constraints', () => {
    render(<SplitterHarness />);
    const splitter = screen.getByRole('separator', { name: 'Resize harness' });

    fireEvent.mouseDown(splitter, { clientX: 100 });
    fireEvent.mouseMove(splitter, { clientX: 220 });
    expect(screen.getByTestId('width')).toHaveTextContent('420');

    fireEvent.mouseMove(splitter, { clientX: 520 });
    expect(screen.getByTestId('width')).toHaveTextContent('500');

    fireEvent.mouseMove(splitter, { clientX: -1000 });
    expect(screen.getByTestId('width')).toHaveTextContent('200');

    fireEvent.mouseUp(splitter);
    expect(window.localStorage.getItem('splitter:harness')).toBe('200');
  });

  it('supports keyboard fallback interactions', () => {
    render(<SplitterHarness />);
    const splitter = screen.getByRole('separator', { name: 'Resize harness' });

    splitter.focus();
    fireEvent.keyDown(splitter, { key: 'ArrowRight' });
    fireEvent.keyDown(splitter, { key: 'PageDown' });

    expect(screen.getByTestId('width')).toHaveTextContent('380');
  });

  it('restores persisted width for component interactions', () => {
    window.localStorage.setItem('splitter:harness', '470');

    render(<SplitterHarness />);

    expect(screen.getByTestId('width')).toHaveTextContent('470');
  });
});
