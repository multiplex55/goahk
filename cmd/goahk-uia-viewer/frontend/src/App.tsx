import { useEffect, useMemo, useState } from 'react';
import useSplitter from './hooks/useSplitter';
import FooterControls, { FooterState } from './components/FooterControls';
import StatusBar from './components/StatusBar';
import TreePane from './components/TreePane';
import WindowInfoPanel from './components/WindowInfoPanel';
import WindowListPane from './components/WindowListPane';
import ThreeColumnLayout from './layout/ThreeColumnLayout';
import { createInspectBindings } from './bindings';
import { createInspectStore, InspectBridgeEvent, InspectStore } from './store/inspectStore';

function pathToText(path: { name?: string; nodeID: string }[]): string {
  return path.map((item) => item.name || item.nodeID).join(' > ');
}

function isStageSpecificFailure(text: string): boolean {
  return /^Failed (InspectWindow|GetTreeRoot|GetNodeDetails)/.test(text);
}

function subscribeFollowCursor(store: InspectStore): (() => void) | undefined {
  const runtime = (window as Window & { runtime?: { EventsOn?: (name: string, cb: (payload: unknown) => void) => void; EventsOff?: (name: string) => void } }).runtime;
  if (!runtime?.EventsOn || !runtime?.EventsOff) {
    return undefined;
  }

  const onFollowCursor = (payload: unknown) => {
    store.applyBridgeEvent({ type: 'follow-cursor', ...(payload as Omit<InspectBridgeEvent, 'type'>) } as InspectBridgeEvent);
  };

  runtime.EventsOn('inspect:follow-cursor', onFollowCursor);
  return () => runtime.EventsOff?.('inspect:follow-cursor');
}

export default function App() {
  const store = useMemo(() => createInspectStore(createInspectBindings()), []);
  const [snapshot, setSnapshot] = useState(store.getState());
  const leftSplitter = useSplitter({
    storageKey: 'goahk:uiviewer:left-width',
    defaultSizePx: 300,
    minSizePx: 220,
    maxSizePx: 720
  });
  const middleSplitter = useSplitter({
    storageKey: 'goahk:uiviewer:middle-width',
    defaultSizePx: 420,
    minSizePx: 220,
    maxSizePx: 720
  });

  useEffect(() => store.subscribe(setSnapshot), [store]);

  useEffect(() => {
    void store.refreshWindows();
    const unsubscribe = subscribeFollowCursor(store);
    return () => unsubscribe?.();
  }, [store]);

  const footerState: FooterState = {
    visibleOnly: snapshot.visibleOnly,
    titleOnly: snapshot.titleOnly,
    activateWindow: snapshot.activateOnSelect,
    filter: snapshot.filter
  };

  const backendPath = snapshot.nodeDetails?.path?.length ? pathToText(snapshot.nodeDetails.path) : '';
  const fallbackPath = snapshot.selectedPath.length ? pathToText(snapshot.selectedPath) : '[path:fallback] unavailable';
  const footerPath = backendPath || fallbackPath;
  const footerSelector = snapshot.nodeDetails?.bestSelector || snapshot.selectorText;
  const footerStatusText = snapshot.nodeDetails?.statusText || snapshot.statusText;

  return (
    <div className="app-shell">
      <ThreeColumnLayout
        leftWidthPx={leftSplitter.sizePx}
        middleWidthPx={middleSplitter.sizePx}
        onResize={({ leftWidthPx: nextLeft, middleWidthPx: nextMiddle }) => {
          leftSplitter.setSizePx(nextLeft);
          middleSplitter.setSizePx(nextMiddle);
        }}
        left={
          <WindowListPane
            windows={snapshot.windows.map((window) => ({
              id: window.hwnd,
              title: window.title,
              processName: window.processName ?? ''
            }))}
            selectedWindowId={snapshot.selectedWindowID}
            onSelectWindow={(id) => {
              void store.selectWindow(id);
            }}
          />
        }
        middle={
          <WindowInfoPanel
            windowTitle={snapshot.windows.find((window) => window.hwnd === snapshot.selectedWindowID)?.title ?? 'Unknown Window'}
            properties={snapshot.properties}
            details={{
              bestSelector: snapshot.selectorText,
              windowInfo: snapshot.nodeDetails?.windowInfo,
              element: snapshot.nodeDetails?.element,
              selectorPath: snapshot.nodeDetails?.selectorPath
            }}
            patternActions={snapshot.patterns.map((pattern) => ({
              id: pattern.name,
              label: pattern.name,
              requiresInput: !!pattern.payloadSchema,
              supported: pattern.supported !== false
            }))}
            onInvokePattern={async (id, payload) => {
              await store.invokePatternAction(id, payload);
            }}
            enableMiddleSplitter
          />
        }
        right={
          <TreePane
            nodesByID={Object.fromEntries(
              Object.entries(snapshot.nodesByID).map(([id, node]) => [id, { id, name: node.name ?? node.nodeID, hasChildren: node.hasChildren }])
            )}
            childrenByParentID={snapshot.childrenByParentID}
            expandedByID={snapshot.expandedByID}
            selectedNodeId={snapshot.selectedNodeID}
            onSelectNode={(id) => {
              void store.selectNode(id);
            }}
            onToggleNode={(id) => {
              void store.expandNode(id);
            }}
          />
        }
      />

      <footer className="footer-controls">
        <FooterControls
          state={footerState}
          onRefresh={() => {
            void store.refreshWindows();
          }}
          onToggleVisible={(value) => store.setVisibleOnly(value)}
          onToggleTitle={(value) => store.setTitleOnly(value)}
          onToggleActivate={(value) => store.setActivateOnSelect(value)}
          onChangeFilter={(value) => store.setFilterInput(value)}
        />
        <StatusBar
          statusText={footerStatusText}
          errorText={snapshot.errorText}
          preferStageFailure={isStageSpecificFailure(snapshot.errorText) || isStageSpecificFailure(snapshot.statusText)}
          path={footerPath}
          selector={footerSelector}
          hasDetails={!!snapshot.nodeDetails}
          onCopySelector={() => store.copyBestSelector()}
        />
      </footer>
    </div>
  );
}
