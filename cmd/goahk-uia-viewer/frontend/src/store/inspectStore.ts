export type InspectProperty = { name: string; value: string };

export type InspectPattern = { name: string; payloadSchema?: string };

export type InspectTreeNode = {
  nodeID: string;
  name?: string;
  controlType?: string;
  className?: string;
  hasChildren: boolean;
  parentNodeID?: string;
};

export type InspectWindow = {
  hwnd: string;
  title: string;
  processName?: string;
  className?: string;
  processID?: number;
};

export type InspectWindowRequest = {
  hwnd: string;
};

export type FollowCursorBridgeEvent = {
  type: 'follow-cursor';
  eventID?: number;
  windowID?: string;
  element: InspectTreeNode;
  path?: InspectTreeNode[];
};

export type SelectionBridgeEvent = {
  type: 'selection-changed';
  eventID?: number;
  windowID?: string;
  selectedNodeID: string;
  path?: InspectTreeNode[];
};

export type InspectBridgeEvent = FollowCursorBridgeEvent | SelectionBridgeEvent;

export type InspectStoreState = {
  windows: InspectWindow[];
  selectedWindowID: string;
  selectedNodeID: string;
  selectedPath: InspectTreeNode[];
  treeNodes: InspectTreeNode[];
  properties: InspectProperty[];
  patterns: InspectPattern[];
  statusText: string;
  selectorText: string;
  filter: string;
  followCursor: boolean;
  followCursorBusy: boolean;
  visibleOnly: boolean;
  titleOnly: boolean;
  activateOnSelect: boolean;
  loadingWindows: boolean;
  loadingWindow: boolean;
  loadingNode: boolean;
  loadingChildren: Record<string, boolean>;
  errorText: string;
};

type NodeDetailsResponse = {
  windowInfo?: InspectWindow;
  properties: InspectProperty[];
  patterns: InspectPattern[];
  statusText?: string;
  bestSelector?: string;
  path?: InspectTreeNode[];
};

export type InspectBindings = {
  RefreshWindows(req: { filter: string; visibleOnly: boolean; titleOnly: boolean }): Promise<{ windows: InspectWindow[] }>;
  InspectWindow(req: InspectWindowRequest): Promise<{ window?: InspectWindow; rootNodeID?: string }>;
  GetTreeRoot(req: { hwnd: string; refresh?: boolean }): Promise<{ root: InspectTreeNode }>;
  GetNodeChildren(req: { nodeID: string }): Promise<{ parentNodeID: string; children: InspectTreeNode[] }>;
  SelectNode(req: { nodeID: string }): Promise<{ selected: InspectTreeNode }>;
  GetNodeDetails(req: { nodeID: string }): Promise<NodeDetailsResponse>;
  HighlightNode(req: { nodeID: string }): Promise<{ highlighted: boolean }>;
  ClearHighlight?(req?: Record<string, never>): Promise<{ cleared: boolean }>;
  ToggleFollowCursor?(req: { enabled: boolean }): Promise<{ enabled: boolean }>;
  ActivateWindow?(req: { hwnd: string }): Promise<{ activated: boolean }>;
};

export type InspectStore = {
  getState: () => InspectStoreState;
  subscribe: (listener: (state: InspectStoreState) => void) => () => void;
  setFilterInput: (value: string) => void;
  setFollowCursor: (value: boolean) => Promise<void>;
  setVisibleOnly: (value: boolean) => void;
  setTitleOnly: (value: boolean) => void;
  setActivateOnSelect: (value: boolean) => void;
  refreshWindows: () => Promise<void>;
  selectWindow: (windowID: string) => Promise<void>;
  selectNode: (nodeID: string) => Promise<void>;
  expandNode: (nodeID: string) => Promise<void>;
  applyBridgeEvent: (event: InspectBridgeEvent) => void;
  selectNextWindow: () => Promise<void>;
  selectPreviousWindow: () => Promise<void>;
  selectNextTreeNode: () => Promise<void>;
  selectPreviousTreeNode: () => Promise<void>;
};

export function createInspectStore(
  bindings: InspectBindings,
  opts?: {
    debounceMs?: number;
    followCursorDebounceMs?: number;
    schedule?: (cb: () => void, delay: number) => ReturnType<typeof setTimeout>;
    cancel?: (timer: ReturnType<typeof setTimeout>) => void;
  }
): InspectStore {
  const debounceMs = opts?.debounceMs ?? 200;
  const followCursorDebounceMs = opts?.followCursorDebounceMs ?? 80;
  const schedule = opts?.schedule ?? ((cb, delay) => setTimeout(cb, delay));
  const cancel = opts?.cancel ?? clearTimeout;

  let state: InspectStoreState = {
    windows: [],
    selectedWindowID: '',
    selectedNodeID: '',
    selectedPath: [],
    treeNodes: [],
    properties: [],
    patterns: [],
    statusText: 'Ready',
    selectorText: '',
    filter: '',
    followCursor: false,
    followCursorBusy: false,
    visibleOnly: true,
    titleOnly: false,
    activateOnSelect: false,
    loadingWindows: false,
    loadingWindow: false,
    loadingNode: false,
    loadingChildren: {},
    errorText: ''
  };

  let pendingFilterTimer: ReturnType<typeof setTimeout> | undefined;
  let pendingFollowEventTimer: ReturnType<typeof setTimeout> | undefined;
  let pendingFollowEvent: FollowCursorBridgeEvent | undefined;
  let windowsToken = 0;
  let windowSelectionToken = 0;
  let nodeSelectionToken = 0;
  let lastAppliedBridgeEventID = 0;
  const childrenLoaded = new Set<string>();
  const listeners = new Set<(nextState: InspectStoreState) => void>();

  const setState = (patch: Partial<InspectStoreState>) => {
    state = { ...state, ...patch };
    listeners.forEach((listener) => listener(state));
  };

  const upsertNode = (node: InspectTreeNode) => {
    const idx = state.treeNodes.findIndex((existing) => existing.nodeID === node.nodeID);
    if (idx === -1) {
      state = { ...state, treeNodes: [...state.treeNodes, node] };
      listeners.forEach((listener) => listener(state));
      return;
    }

    const next = [...state.treeNodes];
    next[idx] = { ...next[idx], ...node };
    state = { ...state, treeNodes: next };
    listeners.forEach((listener) => listener(state));
  };

  const applyFollowCursorEvent = async (event: FollowCursorBridgeEvent) => {
    if (event.windowID && state.selectedWindowID && event.windowID !== state.selectedWindowID) {
      return;
    }
    if (event.eventID && event.eventID <= lastAppliedBridgeEventID) {
      return;
    }

    upsertNode(event.element);
    const selectedNodeID = event.element.nodeID;
    setState({
      selectedNodeID,
      selectedPath: event.path ?? [],
      statusText: `Following cursor: ${event.element.name || selectedNodeID}`
    });

    lastAppliedBridgeEventID = event.eventID ?? lastAppliedBridgeEventID;
    await selectNode(selectedNodeID);
  };

  const refreshWindows = async () => {
    const token = ++windowsToken;
    setState({ loadingWindows: true, errorText: '' });
    await bindings.ClearHighlight?.({});

    try {
      const resp = await bindings.RefreshWindows({
        filter: state.filter,
        visibleOnly: state.visibleOnly,
        titleOnly: state.titleOnly
      });

      if (token !== windowsToken) {
        return;
      }

      const windows = resp.windows ?? [];
      const selectedStillPresent = windows.some((window) => window.hwnd === state.selectedWindowID);
      setState({
        windows,
        selectedWindowID: selectedStillPresent ? state.selectedWindowID : '',
        loadingWindows: false,
        statusText: `Loaded ${windows.length} windows`
      });
    } catch (err) {
      if (token !== windowsToken) {
        return;
      }
      setState({
        loadingWindows: false,
        errorText: err instanceof Error ? err.message : String(err),
        statusText: 'Failed to refresh windows'
      });
    }
  };

  const selectWindow = async (windowID: string) => {
    const token = ++windowSelectionToken;
    await bindings.ClearHighlight?.({});
    setState({
      selectedWindowID: windowID,
      loadingWindow: true,
      selectedNodeID: '',
      selectedPath: [],
      treeNodes: [],
      properties: [],
      patterns: [],
      selectorText: '',
      errorText: '',
      statusText: 'Loading window...'
    });

    try {
      if (state.activateOnSelect && bindings.ActivateWindow) {
        await bindings.ActivateWindow({ hwnd: windowID });
      }

      const inspectResp = await bindings.InspectWindow({ hwnd: windowID });
      const rootResp = await bindings.GetTreeRoot({ hwnd: windowID, refresh: true });
      const rootNode = rootResp.root;
      upsertNode(rootNode);

      const selectedNodeID = inspectResp.rootNodeID || rootNode.nodeID;
      setState({ selectedNodeID, selectedPath: [rootNode] });

      const details = await bindings.GetNodeDetails({ nodeID: selectedNodeID });
      const patterns = details.patterns ?? [];
      const properties = details.properties ?? [];

      if (token !== windowSelectionToken) {
        return;
      }

      setState({
        selectedNodeID,
        loadingWindow: false,
        properties,
        patterns,
        selectorText: details.bestSelector ?? '',
        selectedPath: details.path ?? state.selectedPath,
        statusText: details.statusText ?? `Selected window ${windowID}`
      });
    } catch (err) {
      if (token !== windowSelectionToken) {
        return;
      }
      setState({
        loadingWindow: false,
        errorText: err instanceof Error ? err.message : String(err),
        selectedNodeID: '',
        selectedPath: [],
        properties: [],
        patterns: [],
        selectorText: '',
        statusText: 'Failed to load window'
      });
    }
  };

  const selectNode = async (nodeID: string) => {
    const token = ++nodeSelectionToken;
    setState({ selectedNodeID: nodeID, loadingNode: true, errorText: '' });

    try {
      const selectedResp = await bindings.SelectNode({ nodeID });
      upsertNode(selectedResp.selected);
      const [details] = await Promise.all([bindings.GetNodeDetails({ nodeID }), bindings.HighlightNode({ nodeID })]);

      if (token !== nodeSelectionToken) {
        return;
      }

      setState({
        properties: details.properties ?? [],
        patterns: details.patterns ?? [],
        selectorText: details.bestSelector ?? '',
        selectedPath: details.path ?? state.selectedPath,
        statusText: details.statusText ?? `Selected node ${nodeID}`,
        loadingNode: false
      });
    } catch (err) {
      if (token !== nodeSelectionToken) {
        return;
      }
      setState({
        loadingNode: false,
        errorText: err instanceof Error ? err.message : String(err),
        selectedNodeID: '',
        selectedPath: [],
        properties: [],
        patterns: [],
        selectorText: '',
        statusText: 'Failed to select node'
      });
    }
  };

  const expandNode = async (nodeID: string) => {
    if (childrenLoaded.has(nodeID) || state.loadingChildren[nodeID]) {
      return;
    }

    setState({ loadingChildren: { ...state.loadingChildren, [nodeID]: true }, errorText: '' });

    try {
      const resp = await bindings.GetNodeChildren({ nodeID });
      if (resp.parentNodeID !== nodeID) {
        return;
      }

      const merged = [...state.treeNodes];
      for (const child of resp.children) {
        const idx = merged.findIndex((existing) => existing.nodeID === child.nodeID);
        if (idx === -1) {
          merged.push({ ...child, parentNodeID: nodeID });
        } else {
          merged[idx] = { ...merged[idx], ...child, parentNodeID: nodeID };
        }
      }

      childrenLoaded.add(nodeID);
      setState({
        treeNodes: merged,
        loadingChildren: { ...state.loadingChildren, [nodeID]: false },
        statusText: `Expanded ${nodeID}`
      });
    } catch (err) {
      setState({
        loadingChildren: { ...state.loadingChildren, [nodeID]: false },
        errorText: err instanceof Error ? err.message : String(err),
        statusText: 'Failed to expand node'
      });
    }
  };

  const setFilterInput = (value: string) => {
    setState({ filter: value });
    if (pendingFilterTimer) {
      cancel(pendingFilterTimer);
    }

    pendingFilterTimer = schedule(() => {
      void refreshWindows();
    }, debounceMs);
  };

  const setFollowCursor = async (value: boolean) => {
    setState({ followCursorBusy: true, errorText: '' });
    try {
      if (bindings.ToggleFollowCursor) {
        const resp = await bindings.ToggleFollowCursor({ enabled: value });
        setState({ followCursor: resp.enabled, followCursorBusy: false, statusText: resp.enabled ? 'Follow cursor enabled' : 'Follow cursor disabled' });
      } else {
        setState({ followCursor: value, followCursorBusy: false });
      }
    } catch (err) {
      setState({
        followCursorBusy: false,
        errorText: err instanceof Error ? err.message : String(err),
        statusText: 'Failed to toggle follow cursor'
      });
    }
  };

  const applyBridgeEvent = (event: InspectBridgeEvent) => {
    if (event.type === 'selection-changed') {
      if (event.windowID && state.selectedWindowID && event.windowID !== state.selectedWindowID) {
        return;
      }
      if (event.eventID && event.eventID <= lastAppliedBridgeEventID) {
        return;
      }

      lastAppliedBridgeEventID = event.eventID ?? lastAppliedBridgeEventID;
      if (!event.selectedNodeID) {
        void bindings.ClearHighlight?.({});
      }
      setState({ selectedNodeID: event.selectedNodeID, selectedPath: event.path ?? state.selectedPath, statusText: `Selected ${event.selectedNodeID}` });
      return;
    }

    pendingFollowEvent = event;
    if (pendingFollowEventTimer) {
      return;
    }

    pendingFollowEventTimer = schedule(() => {
      pendingFollowEventTimer = undefined;
      const nextEvent = pendingFollowEvent;
      pendingFollowEvent = undefined;
      if (nextEvent) {
        void applyFollowCursorEvent(nextEvent);
      }
    }, followCursorDebounceMs);
  };

  const selectWindowByDelta = async (delta: number) => {
    if (!state.windows.length) {
      return;
    }
    const currentIndex = Math.max(0, state.windows.findIndex((window) => window.hwnd === state.selectedWindowID));
    const nextIndex = Math.min(state.windows.length - 1, Math.max(0, currentIndex + delta));
    const nextWindow = state.windows[nextIndex];
    if (nextWindow && nextWindow.hwnd !== state.selectedWindowID) {
      await selectWindow(nextWindow.hwnd);
    }
  };

  const selectTreeNodeByDelta = async (delta: number) => {
    if (!state.treeNodes.length) {
      return;
    }
    const currentIndex = Math.max(0, state.treeNodes.findIndex((node) => node.nodeID === state.selectedNodeID));
    const nextIndex = Math.min(state.treeNodes.length - 1, Math.max(0, currentIndex + delta));
    const nextNode = state.treeNodes[nextIndex];
    if (nextNode && nextNode.nodeID !== state.selectedNodeID) {
      await selectNode(nextNode.nodeID);
    }
  };

  return {
    getState: () => state,
    subscribe: (listener) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    refreshWindows,
    selectWindow,
    selectNode,
    expandNode,
    setFilterInput,
    setFollowCursor,
    setVisibleOnly: (value) => setState({ visibleOnly: value }),
    setTitleOnly: (value) => setState({ titleOnly: value }),
    setActivateOnSelect: (value) => setState({ activateOnSelect: value }),
    applyBridgeEvent,
    selectNextWindow: () => selectWindowByDelta(1),
    selectPreviousWindow: () => selectWindowByDelta(-1),
    selectNextTreeNode: () => selectTreeNodeByDelta(1),
    selectPreviousTreeNode: () => selectTreeNodeByDelta(-1)
  };
}
