import { ChangeEvent, useEffect, useState } from 'react';
import { statusCopySource } from '../copySource';

export type FooterState = {
  visibleOnly: boolean;
  titleOnly: boolean;
  activateWindow: boolean;
  filter: string;
  status: string;
  path: string;
};

type FooterControlsProps = {
  state: FooterState;
  onRefresh: () => void;
  onToggleVisible: (value: boolean) => void;
  onToggleTitle: (value: boolean) => void;
  onToggleActivate: (value: boolean) => void;
  onChangeFilter: (value: string) => void;
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
  onChangeFilter
}: FooterControlsProps) {
  const [toastMessage, setToastMessage] = useState('');

  useEffect(() => {
    if (!toastMessage) {
      return;
    }
    const timer = window.setTimeout(() => setToastMessage(''), 1800);
    return () => window.clearTimeout(timer);
  }, [toastMessage]);

  async function copyToClipboard(value: string, label: string) {
    await navigator.clipboard.writeText(statusCopySource(value));
    setToastMessage(`${label} copied`);
  }

  return (
    <footer className="footer-controls">
      <div className="footer-actions">
        <button type="button" onClick={onRefresh}>
          Refresh
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
        <input
          aria-label="window filter"
          placeholder="Filter windows"
          value={state.filter}
          onChange={(event) => onChangeFilter(event.target.value)}
        />
      </div>
      <div className="footer-status">
        <button type="button" className="status-copy" onClick={() => copyToClipboard(state.status, 'Status')}>
          {state.status}
        </button>
        <button type="button" className="status-copy" onClick={() => copyToClipboard(state.path, 'Path')}>
          <code>{state.path}</code>
        </button>
        {toastMessage ? (
          <div className="status-toast" role="status" aria-live="polite">
            {toastMessage}
          </div>
        ) : null}
      </div>
    </footer>
  );
}
