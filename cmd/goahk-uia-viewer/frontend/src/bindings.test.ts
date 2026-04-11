import { describe, expect, it, vi, beforeEach } from 'vitest';

const viewerAppFns = vi.hoisted(() => ({
  RefreshWindows: vi.fn(),
  InspectWindow: vi.fn(),
  GetTreeRoot: vi.fn(),
  GetNodeChildren: vi.fn(),
  SelectNode: vi.fn(),
  GetNodeDetails: vi.fn(),
  HighlightNode: vi.fn(),
  ClearHighlight: vi.fn(),
  ToggleFollowCursor: vi.fn(),
  ActivateWindow: vi.fn()
}));

vi.mock('./wailsjs/wailsjs/go/main/ViewerApp', () => viewerAppFns);

import { createInspectBindings } from './bindings';

describe('createInspectBindings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('forwards exactly one request object to every backend API', async () => {
    viewerAppFns.RefreshWindows.mockResolvedValue({ windows: [] });
    viewerAppFns.InspectWindow.mockResolvedValue({});
    viewerAppFns.GetTreeRoot.mockResolvedValue({ root: { nodeID: 'root', hasChildren: true } });
    viewerAppFns.GetNodeChildren.mockResolvedValue({ parentNodeID: 'n1', children: [] });
    viewerAppFns.SelectNode.mockResolvedValue({ selected: { nodeID: 'n1', hasChildren: false } });
    viewerAppFns.GetNodeDetails.mockResolvedValue({ properties: [], patterns: [], path: [] });
    viewerAppFns.HighlightNode.mockResolvedValue({ highlighted: true });
    viewerAppFns.ClearHighlight.mockResolvedValue({ cleared: true });
    viewerAppFns.ToggleFollowCursor.mockResolvedValue({ enabled: true });
    viewerAppFns.ActivateWindow.mockResolvedValue({ activated: true });

    const bindings = createInspectBindings();

    const reqs = {
      refresh: { filter: 'note', visibleOnly: true, titleOnly: false },
      inspect: { hwnd: '0x1' },
      root: { hwnd: '0x1', refresh: true },
      children: { nodeID: 'n1' },
      select: { nodeID: 'n1' },
      details: { nodeID: 'n1' },
      highlight: { nodeID: 'n1' },
      clear: {},
      follow: { enabled: true },
      activate: { hwnd: '0x1' }
    };

    await bindings.RefreshWindows(reqs.refresh);
    await bindings.InspectWindow(reqs.inspect);
    await bindings.GetTreeRoot(reqs.root);
    await bindings.GetNodeChildren(reqs.children);
    await bindings.SelectNode(reqs.select);
    await bindings.GetNodeDetails(reqs.details);
    await bindings.HighlightNode(reqs.highlight);
    await bindings.ClearHighlight?.(reqs.clear);
    await bindings.ToggleFollowCursor?.(reqs.follow);
    await bindings.ActivateWindow?.(reqs.activate);

    expect(viewerAppFns.RefreshWindows.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.RefreshWindows).toHaveBeenCalledWith(reqs.refresh);
    expect(viewerAppFns.InspectWindow.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.InspectWindow).toHaveBeenCalledWith(reqs.inspect);
    expect(viewerAppFns.GetTreeRoot.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.GetTreeRoot).toHaveBeenCalledWith(reqs.root);
    expect(viewerAppFns.GetNodeChildren.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.GetNodeChildren).toHaveBeenCalledWith(reqs.children);
    expect(viewerAppFns.SelectNode.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.SelectNode).toHaveBeenCalledWith(reqs.select);
    expect(viewerAppFns.GetNodeDetails.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.GetNodeDetails).toHaveBeenCalledWith(reqs.details);
    expect(viewerAppFns.HighlightNode.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.HighlightNode).toHaveBeenCalledWith(reqs.highlight);
    expect(viewerAppFns.ClearHighlight.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.ClearHighlight).toHaveBeenCalledWith(reqs.clear);
    expect(viewerAppFns.ToggleFollowCursor.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.ToggleFollowCursor).toHaveBeenCalledWith(reqs.follow);
    expect(viewerAppFns.ActivateWindow.mock.calls[0]).toHaveLength(1);
    expect(viewerAppFns.ActivateWindow).toHaveBeenCalledWith(reqs.activate);
  });

  it('applies fallback behavior for missing/invalid payloads', async () => {
    viewerAppFns.RefreshWindows.mockResolvedValue(undefined);
    viewerAppFns.InspectWindow.mockResolvedValue(undefined);
    viewerAppFns.GetNodeChildren.mockResolvedValue({ parentNodeID: undefined, children: null });
    viewerAppFns.GetNodeDetails.mockResolvedValue({ properties: null, patterns: 'bad', path: null });
    viewerAppFns.HighlightNode.mockResolvedValue({});
    viewerAppFns.ClearHighlight.mockResolvedValue({});
    viewerAppFns.ToggleFollowCursor.mockResolvedValue({});
    viewerAppFns.ActivateWindow.mockResolvedValue({});

    const bindings = createInspectBindings();

    await expect(bindings.RefreshWindows({ filter: '', visibleOnly: true, titleOnly: false })).resolves.toEqual({ windows: [] });
    await expect(bindings.InspectWindow({ hwnd: '0x1' })).resolves.toEqual({ window: undefined, rootNodeID: undefined });
    await expect(bindings.GetNodeChildren({ nodeID: 'n1' })).resolves.toEqual({ parentNodeID: 'n1', children: [] });
    await expect(bindings.GetNodeDetails({ nodeID: 'n1' })).resolves.toEqual({
      windowInfo: undefined,
      properties: [],
      patterns: [],
      statusText: undefined,
      bestSelector: undefined,
      path: []
    });
    await expect(bindings.HighlightNode({ nodeID: 'n1' })).resolves.toEqual({ highlighted: false });
    await expect(bindings.ClearHighlight?.({})).resolves.toEqual({ cleared: false });
    await expect(bindings.ToggleFollowCursor?.({ enabled: true })).resolves.toEqual({ enabled: false });
    await expect(bindings.ActivateWindow?.({ hwnd: '0x1' })).resolves.toEqual({ activated: false });
  });

  it('preserves existing hard-fail behavior for missing required payloads', async () => {
    viewerAppFns.GetTreeRoot.mockResolvedValue({});
    viewerAppFns.SelectNode.mockResolvedValue({});

    const bindings = createInspectBindings();

    await expect(bindings.GetTreeRoot({ hwnd: '0x1', refresh: true })).rejects.toThrow('Tree root response was empty');
    await expect(bindings.SelectNode({ nodeID: 'n1' })).rejects.toThrow('Selected node response was empty');
  });
});
