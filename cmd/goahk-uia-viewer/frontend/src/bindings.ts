import type { inspect } from './wailsjs/wailsjs/go/models';
import * as ViewerApp from './wailsjs/wailsjs/go/main/ViewerApp';
import { InspectBindings } from './store/inspectStore';
import { ElementDetails, Selector, SelectorCandidate, SelectorResolution, WindowInfoDetails } from './types';

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

const call = async (fn: () => Promise<any>, fallbackMessage: string): Promise<any> => {
  try {
    return await fn();
  } catch (err) {
    throw toError(err, fallbackMessage);
  }
};

const normalizeNode = <T extends { nodeID?: string; nodeId?: string }>(node: T): T & { nodeID: string; nodeId: string } => {
  const resolved = node?.nodeId ?? node?.nodeID ?? '';
  return { ...node, nodeID: resolved, nodeId: resolved };
};

const normalizeMode = (mode?: string): 'UIA_TREE' | 'WINDOW_TREE' | undefined => {
  if (mode === 'WINDOW_TREE') return 'WINDOW_TREE';
  if (mode === 'UIA_TREE') return 'UIA_TREE';
  return undefined;
};

const normalizeState = (state?: { activeMode?: string; fallbackUsed?: boolean; failureStage?: string; guidanceText?: string }) =>
  state
    ? {
        activeMode: normalizeMode(state.activeMode),
        fallbackUsed: !!state.fallbackUsed,
        failureStage: state.failureStage,
        guidanceText: state.guidanceText
      }
    : undefined;

const normalizeDiagnostics = (diag?: { stage?: string; errorCode?: string; hresult?: string; message?: string; fallbackMode?: string; privilegeHint?: string }) =>
  diag
    ? {
        stage: diag.stage,
        errorCode: diag.errorCode,
        hresult: diag.hresult,
        message: diag.message,
        fallbackMode: normalizeMode(diag.fallbackMode),
        privilegeHint: diag.privilegeHint
      }
    : undefined;

type PatternErrorClass = 'not_supported' | 'invalid_input' | 'transient_state' | 'access_denied';
type NormalizedPatternError = { class: PatternErrorClass; code: string; message: string; retryable: boolean };

const normalizePatternErrorClass = (value?: string): PatternErrorClass => {
  switch (value) {
    case 'not_supported':
    case 'invalid_input':
    case 'transient_state':
    case 'access_denied':
      return value;
    default:
      return 'transient_state';
  }
};

const normalizePatternError = (err?: { class?: string; code?: string; message?: string; retryable?: boolean }): NormalizedPatternError | undefined =>
  err
    ? {
        class: normalizePatternErrorClass(err.class),
        code: err.code ?? '',
        message: err.message ?? '',
        retryable: !!err.retryable
      }
    : undefined;

export function createInspectBindings(): InspectBindings {
  return {
    RefreshWindows: async (req: inspect.RefreshWindowsRequest) => {
      const response = await call(() => (ViewerApp as any).RefreshWindows(req), 'Failed to refresh windows');
      return { windows: Array.isArray(response?.windows) ? response.windows : [] };
    },
    InspectWindow: async (req: inspect.InspectWindowRequest) => {
      const response = await call(() => (ViewerApp as any).InspectWindow(req), 'Failed to inspect window');
      return { window: response?.window, rootNodeID: response?.rootNodeID };
    },
    GetTreeRoot: async (req: inspect.GetTreeRootRequest) => {
      const response = await call(() => (ViewerApp as any).GetTreeRoot(req), 'Failed to load tree root');
      if (!response?.root) {
        throw new Error('Tree root response was empty');
      }
      return { root: normalizeNode(response.root), state: normalizeState(response?.state), diagnostics: normalizeDiagnostics(response?.diagnostics) };
    },
    GetNodeChildren: async (req: inspect.GetNodeChildrenRequest) => {
      const response = await call(() => (ViewerApp as any).GetNodeChildren(req), 'Failed to load node children');
      return {
        parentNodeID: response?.parentNodeID ?? req.nodeID,
        children: Array.isArray(response?.children) ? response.children.map((child: any) => normalizeNode(child)) : []
      };
    },
    SelectNode: async (req: inspect.SelectNodeRequest) => {
      const response = await call(() => (ViewerApp as any).SelectNode(req), 'Failed to select node');
      if (!response?.selected) {
        throw new Error('Selected node response was empty');
      }
      return { selected: normalizeNode(response.selected) };
    },
    GetNodeDetails: async (req: inspect.GetNodeDetailsRequest) => {
      const response = await call(() => (ViewerApp as any).GetNodeDetails(req), 'Failed to load node details');
      const dto = (response ?? {}) as unknown as {
        windowInfo?: WindowInfoDetails;
        element?: ElementDetails;
        properties?: { name: string; value: string | null; status?: 'ok' | 'unsupported'; group?: 'identity' | 'semantics' | 'state' | 'geometry' | 'relation' }[];
        patterns?: {
          name: string;
          displayName?: string;
          payloadSchema?: string;
          supported?: boolean;
          preconditions?: { name: string; satisfied: boolean; reason?: string }[];
        }[];
        statusText?: string;
        bestSelector?: string;
        path?: { nodeID: string; hasChildren: boolean; name?: string; parentNodeID?: string; controlType?: string; localizedControlType?: string; displayLabel?: string }[];
        selectorPath?: {
          bestSelector?: Selector;
          fullPath?: { nodeID: string; hasChildren: boolean; name?: string; parentNodeID?: string; controlType?: string; localizedControlType?: string; displayLabel?: string }[];
          selectorSuggestions?: SelectorCandidate[];
        };
        selectorOptions?: SelectorResolution;
        accPath?: string;
      };
      return {
        windowInfo: dto.windowInfo,
        element: dto.element,
        properties: Array.isArray(dto.properties)
          ? dto.properties.map((property) => ({
              name: property.name,
              value: property.value ?? null,
              status: property.status === 'unsupported' ? 'unsupported' : 'ok',
              group: property.group ?? 'semantics'
            }))
          : [],
        patterns: Array.isArray(dto.patterns) ? dto.patterns : [],
        statusText: dto.statusText,
        bestSelector: dto.bestSelector,
        path: Array.isArray(dto.path) ? dto.path : [],
        selectorPath: dto.selectorPath,
        selectorOptions: dto.selectorOptions,
        accPath: dto.accPath
      };
    },
    HighlightNode: async (req: inspect.HighlightNodeRequest) => {
      const response = await call(() => (ViewerApp as any).HighlightNode(req), 'Failed to highlight node');
      return { highlighted: !!response?.highlighted };
    },
    InvokePattern: async (req: inspect.InvokePatternRequest) => {
      const response = await call(() => (ViewerApp as any).InvokePattern(req), 'Failed to invoke pattern action');
      return {
        invoked: !!response?.invoked,
        action: response?.action ?? req.action,
        nodeID: response?.nodeID ?? req.nodeID,
        result: response?.result,
        error: normalizePatternError(response?.error)
      };
    },
    ClearHighlight: async (req: inspect.ClearHighlightRequest = {}) => {
      const response = await call(() => (ViewerApp as any).ClearHighlight(req), 'Failed to clear highlight');
      return { cleared: !!response?.cleared };
    },
    ToggleFollowCursor: async (req: inspect.ToggleFollowCursorRequest) => {
      const response = await call(() => (ViewerApp as any).ToggleFollowCursor(req), 'Failed to toggle follow cursor');
      return { enabled: !!response?.enabled };
    },
    PauseFollowCursor: async () => {
      const response = await call(() => ((ViewerApp as any).PauseFollowCursor?.({}) ?? Promise.resolve({})), 'Failed to pause follow cursor');
      return { paused: !!response?.paused };
    },
    ResumeFollowCursor: async () => {
      const response = await call(() => ((ViewerApp as any).ResumeFollowCursor?.({}) ?? Promise.resolve({})), 'Failed to resume follow cursor');
      return { paused: !!response?.paused };
    },
    LockFollowCursor: async (req: { nodeID?: string }) => {
      const response = await call(() => ((ViewerApp as any).LockFollowCursor?.(req) ?? Promise.resolve({})), 'Failed to lock follow cursor');
      return { locked: !!response?.locked, nodeID: response?.nodeID };
    },
    UnlockFollowCursor: async () => {
      const response = await call(() => ((ViewerApp as any).UnlockFollowCursor?.({}) ?? Promise.resolve({})), 'Failed to unlock follow cursor');
      return { locked: !!response?.locked };
    },
    ActivateWindow: async (req: inspect.ActivateWindowRequest) => {
      const response = await call(() => (ViewerApp as any).ActivateWindow(req), 'Failed to activate window');
      return { activated: !!response?.activated };
    },
    RefreshTreeRoot: async (req: { hwnd: string; mode?: 'UIA_TREE' | 'WINDOW_TREE' }) => {
      const response = await call(() => ((ViewerApp as any).RefreshTreeRoot?.(req) ?? Promise.resolve({})), 'Failed to refresh root');
      if (!response?.root) throw new Error('Tree root response was empty');
      return { root: normalizeNode(response.root), state: normalizeState(response?.state), diagnostics: normalizeDiagnostics(response?.diagnostics) };
    },
    RefreshNodeChildren: async (req: { nodeID: string }) => {
      const response = await call(() => ((ViewerApp as any).RefreshNodeChildren?.(req) ?? Promise.resolve({})), 'Failed to refresh children');
      return {
        parentNodeID: response?.parentNodeID ?? req.nodeID,
        children: Array.isArray(response?.children) ? response.children.map((child: any) => normalizeNode(child)) : []
      };
    },
    RefreshNodeDetails: async (req: { nodeID: string }) => {
      const response = await call(() => ((ViewerApp as any).RefreshNodeDetails?.(req) ?? Promise.resolve({})), 'Failed to refresh details');
      return { details: (response?.details ?? {}) as any };
    },
    GetDiagnostics: async () => {
      const response = await call(() => ((ViewerApp as any).GetDiagnostics?.({}) ?? Promise.resolve({})), 'Failed to read diagnostics');
      return { diagnostics: normalizeDiagnostics(response?.diagnostics) };
    },
    CopyBestSelector: async (req: inspect.CopyBestSelectorRequest) => {
      const response = await call(() => (ViewerApp as any).CopyBestSelector(req), 'Failed to copy selector');
      return {
        selector: response?.selector ?? '',
        clipboardUpdated: !!response?.clipboardUpdated
      };
    }
  };
}
