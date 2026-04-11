import { useEffect, useState } from 'react';
import { statusCopySource } from '../copySource';

type StatusBarProps = {
  statusText: string;
  errorText: string;
  preferStageFailure?: boolean;
  path: string;
  selector: string;
  hasDetails?: boolean;
  onCopySelector?: () => Promise<{ selector: string; clipboardUpdated: boolean }>;
};

export default function StatusBar({ statusText, errorText, preferStageFailure = false, path, selector, hasDetails = false, onCopySelector }: StatusBarProps) {
  const [toastMessage, setToastMessage] = useState('');
  const status = preferStageFailure ? (errorText || statusText) : (statusText || errorText);

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

  async function copySelector() {
    try {
      if (onCopySelector) {
        const result = await onCopySelector();
        const selectorText = result.selector || selector;
        if (!selectorText) {
          setToastMessage(hasDetails ? 'No selector available' : 'Selector unavailable');
          return;
        }
        if (!result.clipboardUpdated) {
          await navigator.clipboard.writeText(statusCopySource(selectorText));
        }
        setToastMessage('Selector copied');
        return;
      }
      await copyValue(selector, 'Selector');
    } catch {
      setToastMessage('Failed to copy selector');
    }
  }

  return (
    <div className="footer-status">
      <span>{status}</span>
      <button type="button" className="status-copy" onClick={() => void copyValue(path, 'Path')}>
        <code>{path}</code>
      </button>
      <button type="button" className="status-copy" onClick={() => void copySelector()}>
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
