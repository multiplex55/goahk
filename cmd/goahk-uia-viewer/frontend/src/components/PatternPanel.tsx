import { useState } from 'react';
import { PatternAction } from '../types';

type PatternPanelProps = {
  actions: PatternAction[];
  onInvokePattern: (id: string, payload?: string) => void;
};

export default function PatternPanel({ actions, onInvokePattern }: PatternPanelProps) {
  const [payloadByActionId, setPayloadByActionId] = useState<Record<string, string>>({});

  function handleInvoke(action: PatternAction) {
    if (!action.supported) {
      return;
    }
    if (action.requiresInput) {
      const payload = payloadByActionId[action.id]?.trim();
      if (!payload) {
        return;
      }
      onInvokePattern(action.id, payload);
      return;
    }
    onInvokePattern(action.id);
  }

  return (
    <section aria-label="pattern panel">
      <h3>Patterns</h3>
      <div className="pattern-actions">
        {actions.map((action) => {
          const payload = payloadByActionId[action.id] ?? '';
          const disabled = action.supported === false || (action.requiresInput && payload.trim().length === 0);
          return (
            <div
              key={action.id}
              className={`pattern-row ${disabled ? 'disabled' : ''}`.trim()}
              onDoubleClick={() => handleInvoke(action)}
            >
              <button type="button" disabled={disabled} onClick={() => handleInvoke(action)}>
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
