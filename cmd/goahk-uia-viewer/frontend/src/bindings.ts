import { InspectBindings } from './store/inspectStore';

type ViewerApi = {
  RefreshWindows?: (req: { filter: string; visibleOnly: boolean; titleOnly: boolean }) => Promise<{ windows?: unknown[] }>;
  InspectWindow?: (req: { hwnd: string }) => Promise<{ window?: unknown; rootNodeID?: string }>;
  GetTreeRoot?: (req: { hwnd: string; refresh?: boolean }) => Promise<{ root?: unknown }>;
  GetNodeChildren?: (req: { nodeID: string }) => Promise<{ parentNodeID?: string; children?: unknown[] }>;
  SelectNode?: (req: { nodeID: string }) => Promise<{ selected?: unknown }>;
  GetNodeDetails?: (req: { nodeID: string }) => Promise<{
    windowInfo?: unknown;
    properties?: unknown[];
    patterns?: unknown[];
    statusText?: string;
    bestSelector?: string;
    path?: unknown[];
  }>;
  HighlightNode?: (req: { nodeID: string }) => Promise<{ highlighted?: boolean }>;
  ClearHighlight?: (req?: Record<string, never>) => Promise<{ cleared?: boolean }>;
  ToggleFollowCursor?: (req: { enabled: boolean }) => Promise<{ enabled?: boolean }>;
  ActivateWindow?: (req: { hwnd: string }) => Promise<{ activated?: boolean }>;
};

type WailsWindow = Window & {
  go?: {
    main?: {
      ViewerApp?: ViewerApi;
    };
  };
};

const getViewerApi = (): ViewerApi => {
  const api = (window as WailsWindow).go?.main?.ViewerApp;
  if (!api) {
    throw new Error('Viewer backend is unavailable');
  }
  return api;
};

const toError = (err: unknown, fallback: string): Error => {
  if (err instanceof Error && err.message) {
    return err;
  }
  if (typeof err === 'string' && err.trim()) {
    return new Error(err);
  }
  if (err && typeof err === 'object' && 'message' in err && typeof (err as { message: unknown }).message === 'string') {
    return new Error((err as { message: string }).message);
  }
  return new Error(fallback);
};

const call = async <T>(name: string, fn: (() => Promise<T> | undefined) | undefined, fallbackMessage: string): Promise<T> => {
  if (!fn) {
    throw new Error(`${name} is not available in backend`);
  }
  try {
    const result = fn();
    if (!result) {
      throw new Error(`${name} is not available in backend`);
    }
    return await result;
  } catch (err) {
    throw toError(err, fallbackMessage);
  }
};

export function createInspectBindings(): InspectBindings {
  return {
    RefreshWindows: async (req) => {
      const api = getViewerApi();
      const response = await call('RefreshWindows', () => api.RefreshWindows?.(req), 'Failed to refresh windows');
      return { windows: Array.isArray(response?.windows) ? (response.windows as any[]) : [] };
    },
    InspectWindow: async (req) => {
      const api = getViewerApi();
      const response = await call('InspectWindow', () => api.InspectWindow?.(req), 'Failed to inspect window');
      return { window: response?.window as any, rootNodeID: response?.rootNodeID };
    },
    GetTreeRoot: async (req) => {
      const api = getViewerApi();
      const response = await call('GetTreeRoot', () => api.GetTreeRoot?.(req), 'Failed to load tree root');
      if (!response?.root) {
        throw new Error('Tree root response was empty');
      }
      return { root: response.root as any };
    },
    GetNodeChildren: async (req) => {
      const api = getViewerApi();
      const response = await call('GetNodeChildren', () => api.GetNodeChildren?.(req), 'Failed to load node children');
      return {
        parentNodeID: response?.parentNodeID ?? req.nodeID,
        children: Array.isArray(response?.children) ? (response.children as any[]) : []
      };
    },
    SelectNode: async (req) => {
      const api = getViewerApi();
      const response = await call('SelectNode', () => api.SelectNode?.(req), 'Failed to select node');
      if (!response?.selected) {
        throw new Error('Selected node response was empty');
      }
      return { selected: response.selected as any };
    },
    GetNodeDetails: async (req) => {
      const api = getViewerApi();
      const response = await call('GetNodeDetails', () => api.GetNodeDetails?.(req), 'Failed to load node details');
      return {
        windowInfo: response?.windowInfo as any,
        properties: Array.isArray(response?.properties) ? (response.properties as any[]) : [],
        patterns: Array.isArray(response?.patterns) ? (response.patterns as any[]) : [],
        statusText: response?.statusText,
        bestSelector: response?.bestSelector,
        path: Array.isArray(response?.path) ? (response.path as any[]) : []
      };
    },
    HighlightNode: async (req) => {
      const api = getViewerApi();
      const response = await call('HighlightNode', () => api.HighlightNode?.(req), 'Failed to highlight node');
      return { highlighted: !!response?.highlighted };
    },
    ClearHighlight: async (req) => {
      const api = getViewerApi();
      const response = await call('ClearHighlight', () => api.ClearHighlight?.(req), 'Failed to clear highlight');
      return { cleared: !!response?.cleared };
    },
    ToggleFollowCursor: async (req) => {
      const api = getViewerApi();
      const response = await call('ToggleFollowCursor', () => api.ToggleFollowCursor?.(req), 'Failed to toggle follow cursor');
      return { enabled: !!response?.enabled };
    },
    ActivateWindow: async (req) => {
      const api = getViewerApi();
      const response = await call('ActivateWindow', () => api.ActivateWindow?.(req), 'Failed to activate window');
      return { activated: !!response?.activated };
    }
  };
}
