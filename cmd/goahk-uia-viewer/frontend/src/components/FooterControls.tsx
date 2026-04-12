import { ChangeEvent } from 'react';

export type FooterState = {
  visibleOnly: boolean;
  titleOnly: boolean;
  activateWindow: boolean;
  filter: string;
  followCursor: boolean;
  followCursorPaused: boolean;
  followCursorLocked: boolean;
};

type FooterControlsProps = {
  state: FooterState;
  onRefresh: () => void;
  onToggleVisible: (value: boolean) => void;
  onToggleTitle: (value: boolean) => void;
  onToggleActivate: (value: boolean) => void;
  onChangeFilter: (value: string) => void;
  onToggleFollowCursor: (value: boolean) => void;
  onPauseFollowCursor: () => void;
  onResumeFollowCursor: () => void;
  onLockFollowCursor: () => void;
  onUnlockFollowCursor: () => void;
  onRefreshRoot: () => void;
  onRefreshChildren: () => void;
  onRefreshDetails: () => void;
};

function onCheckboxChange(handler: (value: boolean) => void) {
  return (event: ChangeEvent<HTMLInputElement>) => handler(event.target.checked);
}

export default function FooterControls({
  state,
  onRefresh,
  onToggleVisible,
  onToggleTitle,
  onToggleActivate,
  onChangeFilter,
  onToggleFollowCursor,
  onPauseFollowCursor,
  onResumeFollowCursor,
  onLockFollowCursor,
  onUnlockFollowCursor,
  onRefreshRoot,
  onRefreshChildren,
  onRefreshDetails
}: FooterControlsProps) {
  return (
    <div className="footer-actions">
      <button type="button" onClick={onRefresh}>
        Refresh
      </button>
      <button type="button" onClick={onRefreshRoot}>
        Refresh Root
      </button>
      <button type="button" onClick={onRefreshChildren}>
        Refresh Children
      </button>
      <button type="button" onClick={onRefreshDetails}>
        Refresh Details
      </button>
      <label>
        <input type="checkbox" checked={state.visibleOnly} onChange={onCheckboxChange(onToggleVisible)} /> Visible
      </label>
      <label>
        <input type="checkbox" checked={state.titleOnly} onChange={onCheckboxChange(onToggleTitle)} /> Title
      </label>
      <label>
        <input type="checkbox" checked={state.activateWindow} onChange={onCheckboxChange(onToggleActivate)} /> Activate
      </label>
      <label>
        <input type="checkbox" checked={state.followCursor} onChange={onCheckboxChange(onToggleFollowCursor)} /> Follow Cursor
      </label>
      <button type="button" onClick={state.followCursorPaused ? onResumeFollowCursor : onPauseFollowCursor}>
        {state.followCursorPaused ? 'Resume Follow' : 'Pause Follow'}
      </button>
      <button type="button" onClick={state.followCursorLocked ? onUnlockFollowCursor : onLockFollowCursor}>
        {state.followCursorLocked ? 'Unlock' : 'Lock'}
      </button>
      <input
        aria-label="window filter"
        placeholder="Filter windows"
        value={state.filter}
        onChange={(event) => onChangeFilter(event.target.value)}
      />
    </div>
  );
}
