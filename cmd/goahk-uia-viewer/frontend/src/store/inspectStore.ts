export type InspectProperty = {
  name: string;
  value: string | null;
  status: 'ok' | 'unsupported';
  group: 'identity' | 'semantics' | 'state' | 'geometry' | 'relation';
};

export type InspectPattern = { name: string; payloadSchema?: string; supported?: boolean };

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
  inspectionMode: 'UIA_TREE' | 'WINDOW_TREE';
  fallbackState?: {
    activeMode: 'UIA_TREE' | 'WINDOW_TREE';
    fallbackUsed: boolean;
    failureStage?: string;
    guidanceText?: string;
  };
  windows: InspectWindow[];
  selectedWindowID: string;
  selectedNodeID: string;
  selectedPath: InspectTreeNode[];
  nodesByID: Record<string, InspectTreeNode>;
  childrenByParentID: Record<string, string[]>;
  expandedByID: Record<string, boolean>;
  childrenLoadedByID: Record<string, boolean>;
  properties: InspectProperty[];
  patterns: InspectPattern[];
  statusText: string;
  selectorText: string;
  nodeDetails?: NodeDetailsResponse;
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

const formatStageFailure = (stage: string, err: unknown) => {
  const reason = err instanceof Error ? err.message : String(err);
  return {
    stage,
    statusText: stage,
    errorText: `${stage}: ${reason}`
  };
};

type NodeDetailsResponse = {
  windowInfo?: WindowInfoDetails;
  element?: ElementDetails;
  properties: InspectProperty[];
  patterns: InspectPattern[];
  statusText?: string;
  bestSelector?: string;
  path?: InspectTreeNode[];
  selectorPath?: {
    bestSelector?: Selector;
    fullPath?: InspectTreeNode[];
    selectorSuggestions?: SelectorCandidate[];
  };
};

export type InspectBindings = {
  RefreshWindows(req: { filter: string; visibleOnly: boolean; titleOnly: boolean }): Promise<{ windows: InspectWindow[] }>;
  InspectWindow(req: InspectWindowRequest & { mode?: 'UIA_TREE' | 'WINDOW_TREE' }): Promise<{ window?: InspectWindow; rootNodeID?: string }>;
  GetTreeRoot(req: {
    hwnd: string;
    refresh?: boolean;
    mode?: 'UIA_TREE' | 'WINDOW_TREE';
  }): Promise<{
    root: InspectTreeNode;
    state?: { activeMode?: 'UIA_TREE' | 'WINDOW_TREE'; fallbackUsed?: boolean; failureStage?: string; guidanceText?: string };
  }>;
  GetNodeChildren(req: { nodeID: string }): Promise<{ parentNodeID: string; children: InspectTreeNode[] }>;
  SelectNode(req: { nodeID: string }): Promise<{ selected: InspectTreeNode }>;
  GetNodeDetails(req: { nodeID: string }): Promise<NodeDetailsResponse>;
  CopyBestSelector(req: { nodeID: string }): Promise<{ selector: string; clipboardUpdated: boolean }>;
  InvokePattern(req: { nodeID: string; action: string; payload?: Record<string, unknown> }): Promise<{ invoked: boolean; action: string; nodeID: string; result?: string }>;
  HighlightNode(req: { nodeID: string }): Promise<{ highlighted: boolean }>;
  ClearHighlight?(req?: Record<string, never>): Promise<{ cleared: boolean }>;
  ToggleFollowCursor?(req: { enabled: boolean }): Promise<{ enabled: boolean }>;
  ActivateWindow?(req: { hwnd: string }): Promise<{ activated: boolean }>;
};

export type InspectStore = {
  getState: () => InspectStoreState;
  subscribe: (listener: (state: InspectStoreState) => void) => () => void;
  setFilterInput: (value: string) => void;
  setInspectionMode: (value: 'UIA_TREE' | 'WINDOW_TREE') => void;
  setFollowCursor: (value: boolean) => Promise<void>;
  setVisibleOnly: (value: boolean) => void;
  setTitleOnly: (value: boolean) => void;
  setActivateOnSelect: (value: boolean) => void;
  refreshWindows: () => Promise<void>;
  selectWindow: (windowID: string) => Promise<void>;
  selectNode: (nodeID: string) => Promise<void>;
  expandNode: (nodeID: string, opts?: { refresh?: boolean }) => Promise<void>;
  invokePatternAction: (action: string, payloadInput?: string) => Promise<void>;
  copyBestSelector: () => Promise<{ selector: string; clipboardUpdated: boolean }>;
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
  const ROOT_PARENT_ID = '__root__';
  const debounceMs = opts?.debounceMs ?? 200;
  const followCursorDebounceMs = opts?.followCursorDebounceMs ?? 80;
  const schedule = opts?.schedule ?? ((cb, delay) => setTimeout(cb, delay));
  const cancel = opts?.cancel ?? clearTimeout;

  let state: InspectStoreState = {
    inspectionMode: 'UIA_TREE',
    fallbackState: undefined,
    windows: [],
    selectedWindowID: '',
    selectedNodeID: '',
    selectedPath: [],
    nodesByID: {},
    childrenByParentID: {},
    expandedByID: {},
    childrenLoadedByID: {},
    properties: [],
    patterns: [],
    statusText: 'Ready',
    selectorText: '',
    nodeDetails: undefined,
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
  const listeners = new Set<(nextState: InspectStoreState) => void>();

  const setState = (patch: Partial<InspectStoreState>) => {
    state = { ...state, ...patch };
    listeners.forEach((listener) => listener(state));
  };

  const upsertNode = (node: InspectTreeNode) => {
    setState({ nodesByID: { ...state.nodesByID, [node.nodeID]: { ...state.nodesByID[node.nodeID], ...node } } });
  };

  const getDescendantIDs = (parentID: string): string[] => {
    const descendants: string[] = [];
    const stack = [...(state.childrenByParentID[parentID] ?? [])];
    while (stack.length) {
      const current = stack.pop()!;
      descendants.push(current);
      stack.push(...(state.childrenByParentID[current] ?? []));
    }
    return descendants;
  };

  const invalidateBranchCache = (parentID: string) => {
    const descendants = getDescendantIDs(parentID);
    const nextNodesByID = { ...state.nodesByID };
    const nextChildrenByParentID = { ...state.childrenByParentID };
    const nextChildrenLoadedByID = { ...state.childrenLoadedByID };
    const nextExpandedByID = { ...state.expandedByID };

    for (const descendantID of descendants) {
      delete nextNodesByID[descendantID];
      delete nextChildrenByParentID[descendantID];
      delete nextChildrenLoadedByID[descendantID];
      nextExpandedByID[descendantID] = false;
    }

    delete nextChildrenByParentID[parentID];
    nextChildrenLoadedByID[parentID] = false;

    setState({
      nodesByID: nextNodesByID,
      childrenByParentID: nextChildrenByParentID,
      childrenLoadedByID: nextChildrenLoadedByID,
      expandedByID: nextExpandedByID
    });
  };

  const getVisibleNodeIDs = (): string[] => {
    const ordered: string[] = [];
    const walk = (parentID: string) => {
      for (const childID of state.childrenByParentID[parentID] ?? []) {
        ordered.push(childID);
        if (state.expandedByID[childID]) {
          walk(childID);
        }
      }
    };
    walk(ROOT_PARENT_ID);
    return ordered;
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
      errorText: '',
      statusText: 'Loading window...'
    });

    try {
      if (state.activateOnSelect && bindings.ActivateWindow) {
        await bindings.ActivateWindow({ hwnd: windowID });
        if (token !== windowSelectionToken) {
          return;
        }
      }

      let inspectResp: { window?: InspectWindow; rootNodeID?: string };
      try {
        inspectResp = await bindings.InspectWindow({ hwnd: windowID, mode: state.inspectionMode });
        if (token !== windowSelectionToken) {
          return;
        }
      } catch (err) {
        throw formatStageFailure('Failed InspectWindow', err);
      }

      let rootResp: {
        root: InspectTreeNode;
        state?: { activeMode?: 'UIA_TREE' | 'WINDOW_TREE'; fallbackUsed?: boolean; failureStage?: string; guidanceText?: string };
      };
      try {
        rootResp = await bindings.GetTreeRoot({ hwnd: windowID, refresh: true, mode: state.inspectionMode });
        if (token !== windowSelectionToken) {
          return;
        }
      } catch (err) {
        throw formatStageFailure('Failed GetTreeRoot', err);
      }

      const rootNode = rootResp.root;
      const seededNodesByID: Record<string, InspectTreeNode> = { [rootNode.nodeID]: rootNode };
      const seededChildrenByParentID: Record<string, string[]> = { [ROOT_PARENT_ID]: [rootNode.nodeID] };
      const seededChildrenLoadedByID: Record<string, boolean> = {};
      const seededExpandedByID: Record<string, boolean> = { [rootNode.nodeID]: false };

      const selectedNodeID = inspectResp.rootNodeID || rootNode.nodeID;
      const seededPath: InspectTreeNode[] = [rootNode];

      let details: NodeDetailsResponse;
      try {
        details = await bindings.GetNodeDetails({ nodeID: selectedNodeID });
        if (token !== windowSelectionToken) {
          return;
        }
      } catch (err) {
        throw formatStageFailure('Failed GetNodeDetails', err);
      }
      const patterns = details.patterns ?? [];
      const properties = details.properties ?? [];

      if (token !== windowSelectionToken) {
        return;
      }

      setState({
        inspectionMode: rootResp.state?.activeMode ?? state.inspectionMode,
        fallbackState: rootResp.state,
        selectedNodeID,
        loadingWindow: false,
        nodesByID: seededNodesByID,
        childrenByParentID: seededChildrenByParentID,
        childrenLoadedByID: seededChildrenLoadedByID,
        expandedByID: seededExpandedByID,
        properties,
        patterns,
        selectorText: details.bestSelector ?? '',
        nodeDetails: details,
        selectedPath: details.path ?? seededPath,
        statusText: details.statusText ?? `Selected window ${windowID}`
      });
    } catch (err) {
      if (token !== windowSelectionToken) {
        return;
      }
      setState({
        loadingWindow: false,
        errorText: typeof err === 'object' && err !== null && 'errorText' in err ? String((err as { errorText: string }).errorText) : (err instanceof Error ? err.message : String(err)),
        nodesByID: {},
        childrenByParentID: {},
        expandedByID: {},
        childrenLoadedByID: {},
        selectedNodeID: '',
        selectedPath: [],
        properties: [],
        patterns: [],
        selectorText: '',
        fallbackState: undefined,
        nodeDetails: undefined,
        statusText: typeof err === 'object' && err !== null && 'statusText' in err ? String((err as { statusText: string }).statusText) : 'Failed to load window'
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
        nodeDetails: details,
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
        nodeDetails: undefined,
        statusText: 'Failed to select node'
      });
    }
  };

  const expandNode = async (nodeID: string, opts?: { refresh?: boolean }) => {
    const refresh = opts?.refresh ?? false;
    const isExpanded = !!state.expandedByID[nodeID];
    if (isExpanded && !refresh) {
      setState({ expandedByID: { ...state.expandedByID, [nodeID]: false } });
      return;
    }

    if (refresh) {
      invalidateBranchCache(nodeID);
    }

    setState({ expandedByID: { ...state.expandedByID, [nodeID]: true } });
    if (state.childrenLoadedByID[nodeID] || state.loadingChildren[nodeID]) {
      return;
    }

    setState({ loadingChildren: { ...state.loadingChildren, [nodeID]: true }, errorText: '' });

    try {
      const resp = await bindings.GetNodeChildren({ nodeID });
      if (resp.parentNodeID !== nodeID) {
        return;
      }
      const nextNodesByID = { ...state.nodesByID };
      const childIDs: string[] = [];
      for (const child of resp.children) {
        const childID = child.nodeID;
        childIDs.push(childID);
        nextNodesByID[childID] = { ...nextNodesByID[childID], ...child, parentNodeID: nodeID };
      }
      setState({
        nodesByID: nextNodesByID,
        childrenByParentID: { ...state.childrenByParentID, [nodeID]: childIDs },
        childrenLoadedByID: { ...state.childrenLoadedByID, [nodeID]: true },
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

  const refreshNodeDetails = async (nodeID: string) => {
    const details = await bindings.GetNodeDetails({ nodeID });
    setState({
      properties: details.properties ?? [],
      patterns: details.patterns ?? [],
      selectorText: details.bestSelector ?? '',
      nodeDetails: details,
      selectedPath: details.path ?? state.selectedPath
    });
  };

  const refreshChildBranch = async (nodeID: string) => {
    const resp = await bindings.GetNodeChildren({ nodeID });
    if (resp.parentNodeID !== nodeID) {
      return;
    }
    const nextNodesByID = { ...state.nodesByID };
    const childIDs: string[] = [];
    for (const child of resp.children) {
      const childID = child.nodeID;
      childIDs.push(childID);
      nextNodesByID[childID] = { ...nextNodesByID[childID], ...child, parentNodeID: nodeID };
    }
    setState({
      nodesByID: nextNodesByID,
      childrenByParentID: { ...state.childrenByParentID, [nodeID]: childIDs },
      childrenLoadedByID: { ...state.childrenLoadedByID, [nodeID]: true }
    });
  };

  const invokePatternAction = async (action: string, payloadInput?: string) => {
    const nodeID = state.selectedNodeID;
    if (!nodeID) {
      return;
    }
    const normalizedAction = action === 'set-value' ? 'setValue' : action;
    const payload = payloadInput?.trim();

    setState({ errorText: '' });
    try {
      await bindings.InvokePattern({
        nodeID,
        action: normalizedAction,
        payload: payload ? { value: payload } : undefined
      });

      const isMutatingAction = ['toggle', 'expand', 'collapse', 'set-value', 'setValue'].includes(action);
      if (isMutatingAction) {
        await Promise.all([refreshNodeDetails(nodeID), refreshChildBranch(nodeID)]);
      } else {
        await refreshNodeDetails(nodeID);
      }

      setState({ statusText: payload ? `Executed ${action} with payload` : `Executed ${action}` });
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setState({ statusText: message, errorText: message });
      throw err;
    }
  };

  const copyBestSelector = async (): Promise<{ selector: string; clipboardUpdated: boolean }> => {
    const nodeID = state.selectedNodeID;
    if (!nodeID) {
      return { selector: '', clipboardUpdated: false };
    }
    setState({ errorText: '' });
    const resp = await bindings.CopyBestSelector({ nodeID });
    const selector = resp.selector || state.nodeDetails?.bestSelector || state.selectorText;
    setState({
      selectorText: selector,
      statusText: selector ? (resp.clipboardUpdated ? 'Selector copied (backend)' : 'Selector ready to copy') : 'No selector available'
    });
    return { selector, clipboardUpdated: resp.clipboardUpdated };
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
    const visibleNodeIDs = getVisibleNodeIDs();
    if (!visibleNodeIDs.length) {
      return;
    }
    const currentIndex = Math.max(0, visibleNodeIDs.findIndex((nodeID) => nodeID === state.selectedNodeID));
    const nextIndex = Math.min(visibleNodeIDs.length - 1, Math.max(0, currentIndex + delta));
    const nextNodeID = visibleNodeIDs[nextIndex];
    if (nextNodeID && nextNodeID !== state.selectedNodeID) {
      await selectNode(nextNodeID);
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
    invokePatternAction,
    copyBestSelector,
    setFilterInput,
    setInspectionMode: (value) => setState({ inspectionMode: value, fallbackState: undefined }),
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
import { ElementDetails, Selector, SelectorCandidate, WindowInfoDetails } from '../types';
