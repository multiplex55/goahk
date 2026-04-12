import { PatternAction } from '../types';
import { NotifyFn } from './PropertyGrid';

type PatternPanelProps = {
  actions: PatternAction[];
  onInvokePattern: (id: string, payload?: string) => Promise<void> | void;
  onNotify?: NotifyFn;
};

export default function PatternPanel({ actions, onInvokePattern, onNotify }: PatternPanelProps) {
  async function handleInvoke(action: PatternAction) {
    if (!action.supported) {
      return;
    }
    try {
      if (action.requiresInput) {
        const payload = window.prompt(`Enter value for ${action.label}`)?.trim();
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
          const disabled = action.supported === false;
          return (
            <div key={action.id} className={`pattern-row ${disabled ? 'disabled' : ''}`.trim()}>
              <button type="button" disabled={disabled} onClick={() => void handleInvoke(action)}>
                {action.label}
              </button>
            </div>
          );
        })}
      </div>
    </section>
  );
}
