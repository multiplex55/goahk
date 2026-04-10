import PatternPanel from './PatternPanel';
import PropertyGrid, { NotifyFn } from './PropertyGrid';
import SelectorPanel from './SelectorPanel';
import { PatternAction, PropertyItem } from '../types';

type WindowInfoPanelProps = {
  windowTitle: string;
  properties: PropertyItem[];
  patternActions: PatternAction[];
  selectorText: string;
  onInvokePattern: (id: string, payload?: string) => Promise<void> | void;
  onNotify?: NotifyFn;
};

export default function WindowInfoPanel({
  windowTitle,
  properties,
  patternActions,
  selectorText,
  onInvokePattern,
  onNotify
}: WindowInfoPanelProps) {
  return (
    <section className="pane" aria-label="window info panel">
      <h2>Window Info</h2>
      <p className="window-info">{windowTitle}</p>
      <PropertyGrid properties={properties} onNotify={onNotify} />
      <PatternPanel actions={patternActions} onInvokePattern={onInvokePattern} onNotify={onNotify} />
      <SelectorPanel selector={selectorText} onNotify={onNotify} />
    </section>
  );
}
