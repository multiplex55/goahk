import { statusCopySource } from '../copySource';
import { Selector, SelectorCandidate } from '../types';
import { NotifyFn } from './PropertyGrid';

type SelectorPanelProps = {
  selector: string;
  selectorPath?: {
    bestSelector?: Selector;
    fullPath?: { nodeID: string; name?: string }[];
    selectorSuggestions?: SelectorCandidate[];
  };
  onNotify?: NotifyFn;
};

export default function SelectorPanel({ selector, selectorPath, onNotify }: SelectorPanelProps) {
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
      {selectorPath?.fullPath?.length ? (
        <p>Path: {selectorPath.fullPath.map((item) => item.name || item.nodeID).join(' > ')}</p>
      ) : null}
      {selectorPath?.selectorSuggestions?.length ? (
        <ul>
          {selectorPath.selectorSuggestions.map((candidate) => (
            <li key={`${candidate.rank}-${candidate.source ?? 'selector'}`}>
              #{candidate.rank} {candidate.source ?? 'selector'}
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}
