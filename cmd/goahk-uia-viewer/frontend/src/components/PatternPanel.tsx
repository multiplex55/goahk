import { useState } from 'react';
import { PatternAction } from '../types';
import { NotifyFn } from './PropertyGrid';

type PatternPanelProps = {
  actions: PatternAction[];
  onInvokePattern: (id: string, payload?: string) => Promise<void> | void;
  onNotify?: NotifyFn;
};

export default function PatternPanel({ actions, onInvokePattern, onNotify }: PatternPanelProps) {
  const [payloadByActionId, setPayloadByActionId] = useState<Record<string, string>>({});

  async function handleInvoke(action: PatternAction) {
    if (!action.supported) {
      return;
    }
    try {
      if (action.requiresInput) {
        const payload = payloadByActionId[action.id]?.trim();
        if (!payload) {
          return;
        }
        await onInvokePattern(action.id, payload);
      } else {
        await onInvokePattern(action.id);
      }
      onNotify?.(`${action.label} succeeded`, 'success');
    } catch {
      onNotify?.(`${action.label} failed`, 'error');
    }
  }

  return (
    <section aria-label="pattern panel">
      <h3>Patterns</h3>
      <div className="pattern-actions">
        {actions.map((action) => {
          const payload = payloadByActionId[action.id] ?? '';
          const disabled = action.supported === false || (action.requiresInput && payload.trim().length === 0);
          return (
            <div key={action.id} className={`pattern-row ${disabled ? 'disabled' : ''}`.trim()}>
              <button type="button" disabled={disabled} onClick={() => void handleInvoke(action)}>
                {action.label}
              </button>
              {action.requiresInput ? (
                <input
                  aria-label={`${action.label} payload`}
                  placeholder="Enter value"
                  value={payload}
                  onChange={(event) =>
                    setPayloadByActionId((current) => ({ ...current, [action.id]: event.target.value }))
                  }
                />
              ) : null}
            </div>
          );
        })}
      </div>
    </section>
  );
}
