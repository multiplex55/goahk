import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import FooterControls from './FooterControls';

describe('FooterControls', () => {
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
