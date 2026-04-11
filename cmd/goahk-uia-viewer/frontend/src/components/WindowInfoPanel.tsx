import PatternPanel from './PatternPanel';
import PropertyGrid, { NotifyFn } from './PropertyGrid';
import SelectorPanel from './SelectorPanel';
import Splitter from './Splitter';
import useSplitter from '../hooks/useSplitter';
import { NodeDetailsView, PatternAction, PropertyItem } from '../types';

type WindowInfoPanelProps = {
  windowTitle: string;
  properties: PropertyItem[];
  patternActions: PatternAction[];
  details?: NodeDetailsView;
  onInvokePattern: (id: string, payload?: string) => Promise<void> | void;
  onNotify?: NotifyFn;
  enableMiddleSplitter?: boolean;
};

export default function WindowInfoPanel({
  windowTitle,
  properties,
  patternActions,
  details,
  onInvokePattern,
  onNotify,
  enableMiddleSplitter = false
}: WindowInfoPanelProps) {
  const patternSectionSplitter = useSplitter({
    storageKey: 'goahk:uiviewer:details-props-height',
    defaultSizePx: 280,
    minSizePx: 140,
    maxSizePx: 640
  });

  const propertiesAndPatterns = enableMiddleSplitter ? (
    <div className="details-middle-split" style={{ gridTemplateRows: `${patternSectionSplitter.sizePx}px 6px minmax(120px, 1fr)` }}>
      <div className="details-properties-scroll">
        <PropertyGrid properties={properties} element={details?.element} onNotify={onNotify} />
      </div>
      <Splitter
        ariaLabel="Resize properties and patterns"
        orientation="horizontal"
        valueNow={patternSectionSplitter.sizePx}
        valueMin={140}
        valueMax={640}
        onDrag={patternSectionSplitter.applyDelta}
      />
      <div className="details-patterns-scroll">
        <PatternPanel actions={patternActions} onInvokePattern={onInvokePattern} onNotify={onNotify} />
      </div>
    </div>
  ) : (
    <>
      <PropertyGrid properties={properties} element={details?.element} onNotify={onNotify} />
      <PatternPanel actions={patternActions} onInvokePattern={onInvokePattern} onNotify={onNotify} />
    </>
  );

  return (
    <section className="pane" aria-label="window info panel">
      <h2>Window Info</h2>
      <p className="window-info">{windowTitle}</p>
      {propertiesAndPatterns}
      <SelectorPanel selector={details?.bestSelector ?? ''} selectorPath={details?.selectorPath} onNotify={onNotify} />
    </section>
  );
}
