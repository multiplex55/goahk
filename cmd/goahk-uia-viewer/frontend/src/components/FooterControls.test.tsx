import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import FooterControls from './FooterControls';

describe('FooterControls', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('copies status text and shows toast feedback', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText }
    });

    render(
      <FooterControls
        state={{
          visibleOnly: false,
          titleOnly: false,
          activateWindow: false,
          filter: '',
          status: 'Ready',
          path: 'Desktop > Settings'
        }}
        onRefresh={vi.fn()}
        onToggleVisible={vi.fn()}
        onToggleTitle={vi.fn()}
        onToggleActivate={vi.fn()}
        onChangeFilter={vi.fn()}
      />
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Ready' }));
    });
    expect(writeText).toHaveBeenCalledWith('Ready');
    expect(screen.getByRole('status')).toHaveTextContent('Status copied');
  });

  it('updates control state and dispatches handlers', () => {
    const onRefresh = vi.fn();
    const onToggleVisible = vi.fn();
    const onToggleTitle = vi.fn();
    const onToggleActivate = vi.fn();
    const onChangeFilter = vi.fn();

    render(
      <FooterControls
        state={{
          visibleOnly: false,
          titleOnly: false,
          activateWindow: false,
          filter: '',
          status: 'Ready',
          path: 'Desktop'
        }}
        onRefresh={onRefresh}
        onToggleVisible={onToggleVisible}
        onToggleTitle={onToggleTitle}
        onToggleActivate={onToggleActivate}
        onChangeFilter={onChangeFilter}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    fireEvent.click(screen.getByLabelText(/Visible/));
    fireEvent.click(screen.getByLabelText(/Title/));
    fireEvent.click(screen.getByLabelText(/Activate/));
    fireEvent.change(screen.getByLabelText('window filter'), { target: { value: 'note' } });

    expect(onRefresh).toHaveBeenCalledOnce();
    expect(onToggleVisible).toHaveBeenCalledWith(true);
    expect(onToggleTitle).toHaveBeenCalledWith(true);
    expect(onToggleActivate).toHaveBeenCalledWith(true);
    expect(onChangeFilter).toHaveBeenCalledWith('note');
  });
});
