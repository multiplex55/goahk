import { useEffect, useMemo, useState } from 'react';
import FooterControls, { FooterState } from './components/FooterControls';
import TreePane from './components/TreePane';
import WindowDetailsPane from './components/WindowDetailsPane';
import WindowListPane from './components/WindowListPane';
import ThreeColumnLayout from './layout/ThreeColumnLayout';
import { createInspectBindings } from './bindings';
import { createInspectStore, InspectBridgeEvent, InspectStore } from './store/inspectStore';

function pathToText(path: { name?: string; nodeID: string }[]): string {
  return path.map((item) => item.name || item.nodeID).join(' > ');
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
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [leftWidthPx, setLeftWidthPx] = useState(300);
  const [middleWidthPx, setMiddleWidthPx] = useState(420);

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
    filter: snapshot.filter,
    status: snapshot.errorText || snapshot.statusText,
    path: snapshot.selectorText || pathToText(snapshot.selectedPath)
  };

  return (
    <div className="app-shell">
      <ThreeColumnLayout
        leftWidthPx={leftWidthPx}
        middleWidthPx={middleWidthPx}
        onResize={({ leftWidthPx: nextLeft, middleWidthPx: nextMiddle }) => {
          setLeftWidthPx(nextLeft);
          setMiddleWidthPx(nextMiddle);
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
          <WindowDetailsPane
            windowTitle={snapshot.windows.find((window) => window.hwnd === snapshot.selectedWindowID)?.title ?? 'Unknown Window'}
            properties={snapshot.properties}
            patternActions={snapshot.patterns.map((pattern) => ({
              id: pattern.name,
              label: pattern.name,
              requiresInput: !!pattern.payloadSchema,
              supported: true
            }))}
            onInvokePattern={(id) => {
              setSnapshot((current) => ({ ...current, statusText: `Pattern ${id} is not wired yet` }));
            }}
          />
        }
        right={
          <TreePane
            rootNodes={snapshot.treeNodes.map((node) => ({ id: node.nodeID, name: node.name ?? node.nodeID, hasChildren: node.hasChildren }))}
            expandedNodeIds={expandedNodes}
            selectedNodeId={snapshot.selectedNodeID}
            onSelectNode={(id) => {
              void store.selectNode(id);
            }}
            onToggleNode={(id) => {
              setExpandedNodes((current) => {
                const next = new Set(current);
                if (next.has(id)) {
                  next.delete(id);
                } else {
                  next.add(id);
                  void store.expandNode(id);
                }
                return next;
              });
            }}
          />
        }
      />

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
    </div>
  );
}
