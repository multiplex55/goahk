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
    const onToggleFollowCursor = vi.fn();
    const onPauseFollowCursor = vi.fn();
    const onResumeFollowCursor = vi.fn();
    const onLockFollowCursor = vi.fn();
    const onUnlockFollowCursor = vi.fn();
    const onRefreshRoot = vi.fn();
    const onRefreshChildren = vi.fn();
    const onRefreshDetails = vi.fn();

    render(
      <FooterControls
        state={{
          visibleOnly: false,
          titleOnly: false,
          activateWindow: false,
          filter: '',
          followCursor: false,
          followCursorPaused: false,
          followCursorLocked: false
        }}
        onRefresh={onRefresh}
        onToggleVisible={onToggleVisible}
        onToggleTitle={onToggleTitle}
        onToggleActivate={onToggleActivate}
        onChangeFilter={onChangeFilter}
        onToggleFollowCursor={onToggleFollowCursor}
        onPauseFollowCursor={onPauseFollowCursor}
        onResumeFollowCursor={onResumeFollowCursor}
        onLockFollowCursor={onLockFollowCursor}
        onUnlockFollowCursor={onUnlockFollowCursor}
        onRefreshRoot={onRefreshRoot}
        onRefreshChildren={onRefreshChildren}
        onRefreshDetails={onRefreshDetails}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    fireEvent.click(screen.getByLabelText(/Visible/));
    fireEvent.click(screen.getByLabelText(/Title/));
    fireEvent.click(screen.getByLabelText(/Activate/));
    fireEvent.click(screen.getByLabelText(/Follow Cursor/));
    fireEvent.click(screen.getByRole('button', { name: 'Pause Follow' }));
    fireEvent.click(screen.getByRole('button', { name: 'Lock' }));
    fireEvent.change(screen.getByLabelText('window filter'), { target: { value: 'note' } });

    expect(onRefresh).toHaveBeenCalledOnce();
    expect(onToggleVisible).toHaveBeenCalledWith(true);
    expect(onToggleTitle).toHaveBeenCalledWith(true);
    expect(onToggleActivate).toHaveBeenCalledWith(true);
    expect(onToggleFollowCursor).toHaveBeenCalledWith(true);
    expect(onPauseFollowCursor).toHaveBeenCalledOnce();
    expect(onLockFollowCursor).toHaveBeenCalledOnce();
    expect(onChangeFilter).toHaveBeenCalledWith('note');
  });
});
