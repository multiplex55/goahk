import { useEffect, useState } from 'react';
import { statusCopySource } from '../copySource';

type StatusBarProps = {
  status: string;
  path: string;
  selector: string;
};

export default function StatusBar({ status, path, selector }: StatusBarProps) {
  const [toastMessage, setToastMessage] = useState('');

  useEffect(() => {
    if (!toastMessage) {
      return;
    }
    const timer = window.setTimeout(() => setToastMessage(''), 1800);
    return () => window.clearTimeout(timer);
  }, [toastMessage]);

  async function copyValue(value: string, label: string) {
    try {
      await navigator.clipboard.writeText(statusCopySource(value));
      setToastMessage(`${label} copied`);
    } catch {
      setToastMessage(`Failed to copy ${label.toLowerCase()}`);
    }
  }

  return (
    <div className="footer-status">
      <span>{status}</span>
      <button type="button" className="status-copy" onClick={() => void copyValue(path, 'Path')}>
        <code>{path}</code>
      </button>
      <button type="button" className="status-copy" onClick={() => void copyValue(selector, 'Selector')}>
        Copy selector
      </button>
      {toastMessage ? (
        <div className="status-toast" role="status" aria-live="polite">
          {toastMessage}
        </div>
      ) : null}
    </div>
  );
}
