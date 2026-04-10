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
  className?: string;
};

export type InspectStoreState = {
  windows: InspectWindow[];
  selectedWindowID: string;
  selectedNodeID: string;
  treeNodes: InspectTreeNode[];
  properties: InspectProperty[];
  patterns: InspectPattern[];
  statusText: string;
  filter: string;
  followCursor: boolean;
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
  properties: InspectProperty[];
  patterns: InspectPattern[];
  statusText?: string;
};

export type InspectBindings = {
  RefreshWindows(req: { filter: string; visibleOnly: boolean; titleOnly: boolean }): Promise<{ windows: InspectWindow[] }>;
  InspectWindow(req: { hwnd: string }): Promise<{ rootNodeID?: string }>;
  GetTreeRoot(req: { hwnd: string; refresh?: boolean }): Promise<{ root: InspectTreeNode }>;
  GetNodeChildren(req: { nodeID: string }): Promise<{ parentNodeID: string; children: InspectTreeNode[] }>;
  SelectNode(req: { nodeID: string }): Promise<{ selected: InspectTreeNode }>;
  GetNodeDetails(req: { nodeID: string }): Promise<NodeDetailsResponse>;
  HighlightNode(req: { nodeID: string }): Promise<{ highlighted: boolean }>;
  ActivateWindow?(req: { hwnd: string }): Promise<{ activated: boolean }>;
};

export type InspectStore = {
  getState: () => InspectStoreState;
  setFilterInput: (value: string) => void;
  setFollowCursor: (value: boolean) => void;
  setVisibleOnly: (value: boolean) => void;
  setTitleOnly: (value: boolean) => void;
  setActivateOnSelect: (value: boolean) => void;
  refreshWindows: () => Promise<void>;
  selectWindow: (windowID: string) => Promise<void>;
  selectNode: (nodeID: string) => Promise<void>;
  expandNode: (nodeID: string) => Promise<void>;
};

export function createInspectStore(
  bindings: InspectBindings,
  opts?: {
    debounceMs?: number;
    schedule?: (cb: () => void, delay: number) => ReturnType<typeof setTimeout>;
    cancel?: (timer: ReturnType<typeof setTimeout>) => void;
  }
): InspectStore {
  const debounceMs = opts?.debounceMs ?? 200;
  const schedule = opts?.schedule ?? ((cb, delay) => setTimeout(cb, delay));
  const cancel = opts?.cancel ?? clearTimeout;

  let state: InspectStoreState = {
    windows: [],
    selectedWindowID: '',
    selectedNodeID: '',
    treeNodes: [],
    properties: [],
    patterns: [],
    statusText: 'Ready',
    filter: '',
    followCursor: false,
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
  let windowsToken = 0;
  let windowSelectionToken = 0;
  let nodeSelectionToken = 0;
  const childrenLoaded = new Set<string>();

  const setState = (patch: Partial<InspectStoreState>) => {
    state = { ...state, ...patch };
  };

  const upsertNode = (node: InspectTreeNode) => {
    const idx = state.treeNodes.findIndex((existing) => existing.nodeID === node.nodeID);
    if (idx === -1) {
      state = { ...state, treeNodes: [...state.treeNodes, node] };
      return;
    }

    const next = [...state.treeNodes];
    next[idx] = { ...next[idx], ...node };
    state = { ...state, treeNodes: next };
  };

  const refreshWindows = async () => {
    const token = ++windowsToken;
    setState({ loadingWindows: true, errorText: '' });

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
    setState({
      selectedWindowID: windowID,
      loadingWindow: true,
      selectedNodeID: '',
      treeNodes: [],
      properties: [],
      patterns: [],
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
      setState({ selectedNodeID });

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
        statusText: details.statusText ?? `Selected window ${windowID}`
      });
    } catch (err) {
      if (token !== windowSelectionToken) {
        return;
      }
      setState({
        loadingWindow: false,
        errorText: err instanceof Error ? err.message : String(err),
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
      const [details] = await Promise.all([
        bindings.GetNodeDetails({ nodeID }),
        bindings.HighlightNode({ nodeID })
      ]);

      if (token !== nodeSelectionToken) {
        return;
      }

      setState({
        properties: details.properties ?? [],
        patterns: details.patterns ?? [],
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

  return {
    getState: () => state,
    refreshWindows,
    selectWindow,
    selectNode,
    expandNode,
    setFilterInput,
    setFollowCursor: (value) => setState({ followCursor: value }),
    setVisibleOnly: (value) => setState({ visibleOnly: value }),
    setTitleOnly: (value) => setState({ titleOnly: value }),
    setActivateOnSelect: (value) => setState({ activateOnSelect: value })
  };
}
