import { fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';
import Splitter from './Splitter';

function SplitterHarness() {
  const [width, setWidth] = useState(300);
  const min = 200;
  const max = 500;

  return (
    <>
      <output data-testid="width">{width}</output>
      <Splitter
        ariaLabel="Resize harness"
        onDrag={(delta) => setWidth((current) => Math.max(min, Math.min(max, current + delta)))}
      />
    </>
  );
}

describe('Splitter', () => {
  it('updates width on drag and enforces min/max constraints', () => {
    render(<SplitterHarness />);
    const splitter = screen.getByRole('separator', { name: 'Resize harness' });

    fireEvent.mouseDown(splitter, { pointerId: 1, clientX: 100 });
    fireEvent.mouseMove(splitter, { pointerId: 1, clientX: 220 });
    expect(screen.getByTestId('width')).toHaveTextContent('420');

    fireEvent.mouseMove(splitter, { pointerId: 1, clientX: 520 });
    expect(screen.getByTestId('width')).toHaveTextContent('500');

    fireEvent.mouseMove(splitter, { pointerId: 1, clientX: -1000 });
    expect(screen.getByTestId('width')).toHaveTextContent('200');
  });
});
