import { useState } from 'react';
import FooterControls, { FooterState } from './components/FooterControls';
import TreePane from './components/TreePane';
import WindowDetailsPane from './components/WindowDetailsPane';
import WindowListPane from './components/WindowListPane';
import ThreeColumnLayout from './layout/ThreeColumnLayout';
import { PatternAction, PropertyItem, TreeNode, WindowItem } from './types';

const windows: WindowItem[] = [
  { id: '1', title: 'Settings', processName: 'SystemSettings.exe' },
  { id: '2', title: 'Notepad', processName: 'notepad.exe' }
];

const properties: PropertyItem[] = [
  { name: 'AutomationId', value: 'MainWindow' },
  { name: 'ControlType', value: 'Window' }
];

const patternActions: PatternAction[] = [
  { id: 'invoke', label: 'Invoke' },
  { id: 'focus', label: 'SetFocus' }
];

const rootNodes: TreeNode[] = [
  { id: 'root', name: 'Desktop', hasChildren: true },
  { id: 'child-window', name: 'Settings', hasChildren: true }
];

export default function App() {
  const [selectedWindowId, setSelectedWindowId] = useState(windows[0].id);
  const [expandedNodes, setExpandedNodes] = useState(new Set<string>(['root']));
  const [leftWidthPx, setLeftWidthPx] = useState(300);
  const [middleWidthPx, setMiddleWidthPx] = useState(420);
  const [footerState, setFooterState] = useState<FooterState>({
    visibleOnly: true,
    titleOnly: false,
    activateWindow: false,
    filter: '',
    status: 'Ready',
    path: 'Desktop > Settings'
  });

  return (
    <div className="app-shell">
      <ThreeColumnLayout
        leftWidthPx={leftWidthPx}
        middleWidthPx={middleWidthPx}
        onResize={({ leftWidthPx: nextLeft, middleWidthPx: nextMiddle }) => {
          setLeftWidthPx(nextLeft);
          setMiddleWidthPx(nextMiddle);
        }}
        left={<WindowListPane windows={windows} selectedWindowId={selectedWindowId} onSelectWindow={setSelectedWindowId} />}
        middle={
          <WindowDetailsPane
            windowTitle={windows.find((window) => window.id === selectedWindowId)?.title ?? 'Unknown Window'}
            properties={properties}
            patternActions={patternActions}
            onInvokePattern={(id) => setFooterState((current) => ({ ...current, status: `Invoked: ${id}` }))}
          />
        }
        right={
          <TreePane
            rootNodes={rootNodes}
            expandedNodeIds={expandedNodes}
            onToggleNode={(id) => {
              setExpandedNodes((current) => {
                const next = new Set(current);
                if (next.has(id)) {
                  next.delete(id);
                } else {
                  next.add(id);
                }
                return next;
              });
            }}
          />
        }
      />

      <FooterControls
        state={footerState}
        onRefresh={() => setFooterState((current) => ({ ...current, status: 'Refreshed windows' }))}
        onToggleVisible={(value) => setFooterState((current) => ({ ...current, visibleOnly: value }))}
        onToggleTitle={(value) => setFooterState((current) => ({ ...current, titleOnly: value }))}
        onToggleActivate={(value) => setFooterState((current) => ({ ...current, activateWindow: value }))}
        onChangeFilter={(value) => setFooterState((current) => ({ ...current, filter: value }))}
      />
    </div>
  );
}
