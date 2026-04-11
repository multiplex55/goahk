import type { inspect } from './wailsjs/wailsjs/go/models';
import {
  ActivateWindow,
  ClearHighlight,
  GetNodeChildren,
  GetNodeDetails,
  GetTreeRoot,
  HighlightNode,
  InspectWindow,
  RefreshWindows,
  SelectNode,
  ToggleFollowCursor
} from './wailsjs/wailsjs/go/main/ViewerApp';
import { InspectBindings } from './store/inspectStore';

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

const call = async <T>(fn: () => Promise<T>, fallbackMessage: string): Promise<T> => {
  try {
    return await fn();
  } catch (err) {
    throw toError(err, fallbackMessage);
  }
};

export function createInspectBindings(): InspectBindings {
  return {
    RefreshWindows: async (req: inspect.RefreshWindowsRequest) => {
      const response = await call(() => RefreshWindows(req), 'Failed to refresh windows');
      return { windows: Array.isArray(response?.windows) ? response.windows : [] };
    },
    InspectWindow: async (req: inspect.InspectWindowRequest) => {
      const response = await call(() => InspectWindow(req), 'Failed to inspect window');
      return { window: response?.window, rootNodeID: response?.rootNodeID };
    },
    GetTreeRoot: async (req: inspect.GetTreeRootRequest) => {
      const response = await call(() => GetTreeRoot(req), 'Failed to load tree root');
      if (!response?.root) {
        throw new Error('Tree root response was empty');
      }
      return { root: response.root };
    },
    GetNodeChildren: async (req: inspect.GetNodeChildrenRequest) => {
      const response = await call(() => GetNodeChildren(req), 'Failed to load node children');
      return {
        parentNodeID: response?.parentNodeID ?? req.nodeID,
        children: Array.isArray(response?.children) ? response.children : []
      };
    },
    SelectNode: async (req: inspect.SelectNodeRequest) => {
      const response = await call(() => SelectNode(req), 'Failed to select node');
      if (!response?.selected) {
        throw new Error('Selected node response was empty');
      }
      return { selected: response.selected };
    },
    GetNodeDetails: async (req: inspect.GetNodeDetailsRequest) => {
      const response = await call(() => GetNodeDetails(req), 'Failed to load node details');
      return {
        windowInfo: response?.windowInfo,
        properties: Array.isArray(response?.properties) ? response.properties : [],
        patterns: Array.isArray(response?.patterns) ? response.patterns : [],
        statusText: response?.statusText,
        bestSelector: response?.bestSelector,
        path: Array.isArray(response?.path) ? response.path : []
      };
    },
    HighlightNode: async (req: inspect.HighlightNodeRequest) => {
      const response = await call(() => HighlightNode(req), 'Failed to highlight node');
      return { highlighted: !!response?.highlighted };
    },
    ClearHighlight: async (req: inspect.ClearHighlightRequest = {}) => {
      const response = await call(() => ClearHighlight(req), 'Failed to clear highlight');
      return { cleared: !!response?.cleared };
    },
    ToggleFollowCursor: async (req: inspect.ToggleFollowCursorRequest) => {
      const response = await call(() => ToggleFollowCursor(req), 'Failed to toggle follow cursor');
      return { enabled: !!response?.enabled };
    },
    ActivateWindow: async (req: inspect.ActivateWindowRequest) => {
      const response = await call(() => ActivateWindow(req), 'Failed to activate window');
      return { activated: !!response?.activated };
    }
  };
}
