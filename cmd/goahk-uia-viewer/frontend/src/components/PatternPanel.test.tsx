import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import PatternPanel from './PatternPanel';

describe('PatternPanel', () => {
  afterEach(() => cleanup());
  it('disables unsupported actions and prompts for payload actions', async () => {
    const onInvokePattern = vi.fn().mockResolvedValue(undefined);
    const onNotify = vi.fn();
    const promptSpy = vi.spyOn(window, 'prompt').mockReturnValue('new text');
    render(
      <PatternPanel
        actions={[
          { id: 'invoke', label: 'Invoke', supported: true },
          { id: 'set-value', label: 'SetValue', supported: true, requiresInput: true },
          { id: 'select', label: 'Select', supported: false }
        ]}
        onInvokePattern={onInvokePattern}
        onNotify={onNotify}
      />
    );

    expect(screen.getByRole('button', { name: 'Select' })).toBeDisabled();

    const setValueButton = screen.getByRole('button', { name: 'SetValue' });
    expect(setValueButton).toBeEnabled();

    await act(async () => {
      fireEvent.click(setValueButton);
    });
    expect(promptSpy).toHaveBeenCalled();
    expect(onInvokePattern).toHaveBeenCalledWith('set-value', 'new text');
    expect(onNotify).toHaveBeenCalledWith('SetValue succeeded', 'success');
    promptSpy.mockRestore();
  });

  it('shows error notification when action execution fails', async () => {
    const onInvokePattern = vi.fn().mockRejectedValue(new Error('boom'));
    const onNotify = vi.fn();

    render(
      <PatternPanel
        actions={[{ id: 'invoke', label: 'Invoke', supported: true }]}
        onInvokePattern={onInvokePattern}
        onNotify={onNotify}
      />
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Invoke' }));
    });

    expect(onNotify).toHaveBeenCalledWith('Invoke failed', 'error');
  });

  it('does not prompt for non-input actions and sends no payload', async () => {
    const onInvokePattern = vi.fn().mockResolvedValue(undefined);
    const promptSpy = vi.spyOn(window, 'prompt').mockReturnValue('unused');
    render(
      <PatternPanel
        actions={[{ id: 'toggle', label: 'Toggle', supported: true, requiresInput: false }]}
        onInvokePattern={onInvokePattern}
      />
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Toggle' }));
    });
    expect(promptSpy).not.toHaveBeenCalled();
    expect(onInvokePattern).toHaveBeenCalledWith('toggle');
    promptSpy.mockRestore();
  });
});
