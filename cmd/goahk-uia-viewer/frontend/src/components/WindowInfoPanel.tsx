import PatternPanel from './PatternPanel';
import PropertyGrid, { NotifyFn } from './PropertyGrid';
import SelectorPanel from './SelectorPanel';
import Splitter from './Splitter';
import useSplitter from '../hooks/useSplitter';
import { PatternAction, PropertyItem } from '../types';

type WindowInfoPanelProps = {
  windowTitle: string;
  properties: PropertyItem[];
  patternActions: PatternAction[];
  selectorText: string;
  onInvokePattern: (id: string, payload?: string) => Promise<void> | void;
  onNotify?: NotifyFn;
  enableMiddleSplitter?: boolean;
};

export default function WindowInfoPanel({
  windowTitle,
  properties,
  patternActions,
  selectorText,
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
        <PropertyGrid properties={properties} onNotify={onNotify} />
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
      <PropertyGrid properties={properties} onNotify={onNotify} />
      <PatternPanel actions={patternActions} onInvokePattern={onInvokePattern} onNotify={onNotify} />
    </>
  );

  return (
    <section className="pane" aria-label="window info panel">
      <h2>Window Info</h2>
      <p className="window-info">{windowTitle}</p>
      {propertiesAndPatterns}
      <SelectorPanel selector={selectorText} onNotify={onNotify} />
    </section>
  );
}
