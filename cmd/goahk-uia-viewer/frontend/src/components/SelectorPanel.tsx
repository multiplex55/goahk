import { statusCopySource } from '../copySource';
import { NotifyFn } from './PropertyGrid';

type SelectorPanelProps = {
  selector: string;
  onNotify?: NotifyFn;
};

export default function SelectorPanel({ selector, onNotify }: SelectorPanelProps) {
  const hasSelector = selector.trim().length > 0;

  async function copySelector() {
    try {
      await navigator.clipboard.writeText(statusCopySource(selector));
      onNotify?.('Selector copied', 'success');
    } catch {
      onNotify?.('Failed to copy selector', 'error');
    }
  }

  return (
    <section aria-label="selector panel">
      <h3>Best Selector</h3>
      <div className="selector-row">
        <code>{hasSelector ? selector : 'No selector available'}</code>
        <button type="button" disabled={!hasSelector} onClick={() => void copySelector()}>
          Copy Selector
        </button>
      </div>
    </section>
  );
}
