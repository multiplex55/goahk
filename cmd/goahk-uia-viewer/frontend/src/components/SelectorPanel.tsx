import { statusCopySource } from '../copySource';
import { useState } from 'react';
import { Selector, SelectorCandidate, SelectorResolution } from '../types';
import { NotifyFn } from './PropertyGrid';

type SelectorPanelProps = {
  selector: string;
  selectorPath?: {
    bestSelector?: Selector;
    fullPath?: { nodeID: string; name?: string; controlType?: string; localizedControlType?: string; displayLabel?: string }[];
    selectorSuggestions?: SelectorCandidate[];
  };
  selectorOptions?: SelectorResolution;
  onNotify?: NotifyFn;
  accPath?: string;
};

function selectorToText(selector?: Selector): string {
  if (!selector) {
    return '';
  }
  const parts: string[] = [];
  if (selector.automationId) parts.push(`automationId=${selector.automationId}`);
  if (selector.name) parts.push(`name=${selector.name}`);
  if (selector.controlType) parts.push(`controlType=${selector.controlType}`);
  if (selector.className) parts.push(`className=${selector.className}`);
  if (selector.frameworkId) parts.push(`frameworkId=${selector.frameworkId}`);
  return parts.join(';');
}

function pathSegment(item: { nodeID: string; displayLabel?: string; localizedControlType?: string; controlType?: string; name?: string }): string {
  return item.localizedControlType || item.controlType || item.displayLabel || item.name || item.nodeID;
}

export default function SelectorPanel({ selector, selectorPath, selectorOptions, onNotify, accPath }: SelectorPanelProps) {
  const [tab, setTab] = useState<'uia' | 'acc'>('uia');
  const hasSelector = selector.trim().length > 0;
  const best = selectorOptions?.best;
  const alternates = selectorOptions?.alternates ?? selectorPath?.selectorSuggestions?.slice(1) ?? [];

  async function copyText(value: string, success: string) {
    try {
      await navigator.clipboard.writeText(statusCopySource(value));
      onNotify?.(success, 'success');
    } catch {
      onNotify?.('Failed to copy selector', 'error');
    }
  }

  return (
    <section aria-label="selector panel">
      <div className="selector-tabs">
        <button type="button" onClick={() => setTab('uia')} aria-pressed={tab === 'uia'}>
          UIA
        </button>
        <button type="button" onClick={() => setTab('acc')} aria-pressed={tab === 'acc'}>
          ACC/MSAA
        </button>
      </div>
      {tab === 'acc' ? (
        <div className="selector-row">
          <code>{accPath?.trim() ? accPath : 'No ACC/MSAA path available'}</code>
          <button type="button" disabled={!accPath?.trim()} onClick={() => void copyText(accPath || '', 'ACC path copied')}>
            Copy ACC
          </button>
        </div>
      ) : null}
      {tab === 'uia' ? (
        <>
      <h3>Best Selector</h3>
      <div className="selector-row">
        <code>{hasSelector ? selector : 'No selector available'}</code>
        <button type="button" disabled={!hasSelector} onClick={() => void copyText(selector, 'Selector copied')}>
          Copy Selector
        </button>
      </div>
      {best?.rationale ? <p>Rationale: {best.rationale}</p> : null}
      {selectorPath?.fullPath?.length ? <p>Path: {selectorPath.fullPath.map((item) => pathSegment(item)).join(' > ')}</p> : null}
      {alternates.length ? (
        <ul>
          {alternates.map((candidate) => {
            const text = selectorToText(candidate.selector);
            return (
              <li key={`${candidate.rank}-${candidate.source ?? 'selector'}`}>
                <strong>#{candidate.rank}</strong> {text || candidate.source || 'selector'}
                {candidate.score ? ` (score: ${candidate.score})` : ''}
                {candidate.rationale ? ` — ${candidate.rationale}` : ''}{' '}
                <button type="button" onClick={() => void copyText(text, `Selector #${candidate.rank} copied`)} disabled={!text}>
                  Copy
                </button>
              </li>
            );
          })}
        </ul>
      ) : null}
        </>
      ) : null}
    </section>
  );
}
