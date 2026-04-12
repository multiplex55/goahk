import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App';

const mockStore = {
  getState: vi.fn(() => ({
    windows: [],
    inspectionMode: 'UIA_TREE',
    fallbackState: undefined,
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
  })),
  subscribe: vi.fn(() => vi.fn()),
  refreshWindows: vi.fn().mockResolvedValue(undefined),
  selectWindow: vi.fn().mockResolvedValue(undefined),
  selectNode: vi.fn().mockResolvedValue(undefined),
  expandNode: vi.fn().mockResolvedValue(undefined),
  invokePatternAction: vi.fn().mockResolvedValue(undefined),
  copyBestSelector: vi.fn().mockResolvedValue({ selector: '', clipboardUpdated: false }),
  setFilterInput: vi.fn(),
  setInspectionMode: vi.fn(),
  setFollowCursor: vi.fn().mockResolvedValue(undefined),
  setVisibleOnly: vi.fn(),
  setTitleOnly: vi.fn(),
  setActivateOnSelect: vi.fn(),
  applyBridgeEvent: vi.fn(),
  selectNextWindow: vi.fn().mockResolvedValue(undefined),
  selectPreviousWindow: vi.fn().mockResolvedValue(undefined),
  selectNextTreeNode: vi.fn().mockResolvedValue(undefined),
  selectPreviousTreeNode: vi.fn().mockResolvedValue(undefined)
};

vi.mock('./bindings', () => ({
  createInspectBindings: vi.fn(() => ({}))
}));

vi.mock('./store/inspectStore', () => ({
  createInspectStore: vi.fn(() => mockStore)
}));

describe('App boot flow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('app mount triggers initial refresh once', async () => {
    render(<App />);

    await waitFor(() => {
      expect(mockStore.refreshWindows).toHaveBeenCalledTimes(1);
    });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('shows fallback banner and supports switch action', () => {
    mockStore.getState.mockReturnValueOnce({
      ...mockStore.getState(),
      inspectionMode: 'UIA_TREE',
      fallbackState: {
        activeMode: 'WINDOW_TREE',
        fallbackUsed: true,
        failureStage: 'ResolveWindowRoot',
        guidanceText: 'UIA tree is unavailable.'
      }
    });

    render(<App />);
    expect(screen.getByRole('alert')).toHaveTextContent('UIA tree is unavailable.');
    fireEvent.click(screen.getByRole('button', { name: 'Switch to Window Tree' }));
    expect(mockStore.setInspectionMode).toHaveBeenCalledWith('WINDOW_TREE');
  });
});
