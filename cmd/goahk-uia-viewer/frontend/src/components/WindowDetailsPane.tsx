import PatternPanel from './PatternPanel';
import PropertyGrid from './PropertyGrid';
import { PatternAction, PropertyItem } from '../types';

type WindowDetailsPaneProps = {
  windowTitle: string;
  properties: PropertyItem[];
  patternActions: PatternAction[];
  onInvokePattern: (id: string) => void;
};

export default function WindowDetailsPane({
  windowTitle,
  properties,
  patternActions,
  onInvokePattern
}: WindowDetailsPaneProps) {
  return (
    <section className="pane" aria-label="window details pane">
      <h2>Window Info</h2>
      <p className="window-info">{windowTitle}</p>
      <PropertyGrid properties={properties} />
      <PatternPanel actions={patternActions} onInvokePattern={onInvokePattern} />
    </section>
  );
}
