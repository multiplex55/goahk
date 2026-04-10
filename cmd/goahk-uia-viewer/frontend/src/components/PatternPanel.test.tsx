import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import PatternPanel from './PatternPanel';

describe('PatternPanel', () => {
  it('dispatches action on button click and row double-click', () => {
    const onInvokePattern = vi.fn();
    render(
      <PatternPanel
        actions={[{ id: 'invoke', label: 'Invoke', supported: true }]}
        onInvokePattern={onInvokePattern}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Invoke' }));
    fireEvent.doubleClick(screen.getByText('Invoke').closest('.pattern-row')!);

    expect(onInvokePattern).toHaveBeenNthCalledWith(1, 'invoke');
    expect(onInvokePattern).toHaveBeenNthCalledWith(2, 'invoke');
  });

  it('requires payload input for SetValue before enabling action', () => {
    const onInvokePattern = vi.fn();
    render(
      <PatternPanel
        actions={[{ id: 'set-value', label: 'SetValue', supported: true, requiresInput: true }]}
        onInvokePattern={onInvokePattern}
      />
    );

    const button = screen.getByRole('button', { name: 'SetValue' });
    expect(button).toBeDisabled();

    fireEvent.change(screen.getByLabelText('SetValue payload'), { target: { value: 'new text' } });
    expect(button).toBeEnabled();

    fireEvent.click(button);
    expect(onInvokePattern).toHaveBeenCalledWith('set-value', 'new text');
  });
});
