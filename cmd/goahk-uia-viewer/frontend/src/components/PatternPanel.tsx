import { PatternAction } from '../types';

type PatternPanelProps = {
  actions: PatternAction[];
  onInvokePattern: (id: string) => void;
};

export default function PatternPanel({ actions, onInvokePattern }: PatternPanelProps) {
  return (
    <section aria-label="pattern panel">
      <h3>Patterns</h3>
      <div className="pattern-actions">
        {actions.map((action) => (
          <button key={action.id} type="button" onClick={() => onInvokePattern(action.id)}>
            {action.label}
          </button>
        ))}
      </div>
    </section>
  );
}
