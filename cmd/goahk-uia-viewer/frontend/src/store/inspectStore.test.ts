import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createInspectStore, InspectBindings } from './inspectStore';

function makeBindings(overrides?: Partial<InspectBindings>): InspectBindings {
  return {
    RefreshWindows: vi.fn().mockResolvedValue({ windows: [{ hwnd: 'w1', title: 'Notepad' }] }),
    InspectWindow: vi.fn().mockResolvedValue({ rootNodeID: 'root-1' }),
    GetTreeRoot: vi.fn().mockResolvedValue({ root: { nodeID: 'root-1', name: 'Root', hasChildren: true } }),
    GetNodeChildren: vi.fn().mockResolvedValue({
      parentNodeID: 'root-1',
      children: [{ nodeID: 'child-1', name: 'Child', hasChildren: false }]
    }),
    SelectNode: vi.fn().mockImplementation(async ({ nodeID }) => ({
      selected: { nodeID, name: `Node ${nodeID}`, hasChildren: false }
    })),
    GetNodeDetails: vi.fn().mockImplementation(async ({ nodeID }) => ({
      properties: [{ name: 'AutomationId', value: nodeID }],
      patterns: [{ name: 'Invoke' }],
      statusText: `Details ${nodeID}`
    })),
    HighlightNode: vi.fn().mockResolvedValue({ highlighted: true }),
    ActivateWindow: vi.fn().mockResolvedValue({ activated: true }),
    ...overrides
  };
}

describe('inspectStore', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it('Flow 1: refreshes windows using current filters and repopulates state', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);
    store.setFilterInput('note');
    store.setVisibleOnly(false);
    store.setTitleOnly(true);

    await vi.runAllTimersAsync();

    expect(bindings.RefreshWindows).toHaveBeenCalledWith({
      filter: 'note',
      visibleOnly: false,
      titleOnly: true
    });
    expect(store.getState().windows).toEqual([{ hwnd: 'w1', title: 'Notepad' }]);
    expect(store.getState().statusText).toContain('Loaded');
  });

  it('Flow 2: selecting a window optionally activates and bootstraps root, properties, patterns', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);
    store.setActivateOnSelect(true);

    await store.selectWindow('w1');

    expect(bindings.ActivateWindow).toHaveBeenCalledWith({ hwnd: 'w1' });
    expect(bindings.InspectWindow).toHaveBeenCalledWith({ hwnd: 'w1' });
    expect(bindings.GetTreeRoot).toHaveBeenCalledWith({ hwnd: 'w1', refresh: true });
    expect(store.getState().treeNodes.map((node) => node.nodeID)).toContain('root-1');
    expect(store.getState().selectedNodeID).toBe('root-1');
    expect(store.getState().properties[0].value).toBe('root-1');
    expect(store.getState().patterns[0].name).toBe('Invoke');
  });

  it('Flow 3: selecting a node refreshes properties, patterns, status, and highlight', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);

    await store.selectNode('node-22');

    expect(bindings.SelectNode).toHaveBeenCalledWith({ nodeID: 'node-22' });
    expect(bindings.HighlightNode).toHaveBeenCalledWith({ nodeID: 'node-22' });
    expect(store.getState().selectedNodeID).toBe('node-22');
    expect(store.getState().properties).toEqual([{ name: 'AutomationId', value: 'node-22' }]);
    expect(store.getState().statusText).toBe('Details node-22');
  });

  it('Flow 4: expanding a node lazily loads children and reuses cache', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);

    await store.expandNode('root-1');
    await store.expandNode('root-1');

    expect(bindings.GetNodeChildren).toHaveBeenCalledTimes(1);
    expect(store.getState().treeNodes.map((node) => node.nodeID)).toContain('child-1');
  });

  it('debounces filter input so rapid typing performs one refresh', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings, { debounceMs: 250 });

    store.setFilterInput('n');
    store.setFilterInput('no');
    store.setFilterInput('not');
    await vi.advanceTimersByTimeAsync(249);
    expect(bindings.RefreshWindows).toHaveBeenCalledTimes(0);

    await vi.advanceTimersByTimeAsync(1);
    expect(bindings.RefreshWindows).toHaveBeenCalledTimes(1);
    expect(bindings.RefreshWindows).toHaveBeenLastCalledWith({
      filter: 'not',
      visibleOnly: true,
      titleOnly: false
    });
  });

  it('ignores stale window selection response when rapid selection changes occur', async () => {
    let resolveSlow: ((value: { properties: { name: string; value: string }[]; patterns: { name: string }[]; statusText: string }) => void) | undefined;
    const bindings = makeBindings({
      GetNodeDetails: vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise((resolve) => {
              resolveSlow = resolve as typeof resolveSlow;
            })
        )
        .mockResolvedValueOnce({ properties: [{ name: 'AutomationId', value: 'root-2' }], patterns: [{ name: 'Focus' }], statusText: 'Details root-2' }),
      InspectWindow: vi
        .fn()
        .mockResolvedValueOnce({ rootNodeID: 'root-1' })
        .mockResolvedValueOnce({ rootNodeID: 'root-2' }),
      GetTreeRoot: vi
        .fn()
        .mockResolvedValueOnce({ root: { nodeID: 'root-1', hasChildren: true } })
        .mockResolvedValueOnce({ root: { nodeID: 'root-2', hasChildren: true } })
    });
    const store = createInspectStore(bindings);

    const first = store.selectWindow('w1');
    const second = store.selectWindow('w2');

    await second;
    resolveSlow?.({ properties: [{ name: 'AutomationId', value: 'root-1' }], patterns: [{ name: 'Invoke' }], statusText: 'Details root-1' });
    await first;

    expect(store.getState().selectedWindowID).toBe('w2');
    expect(store.getState().selectedNodeID).toBe('root-2');
    expect(store.getState().properties[0].value).toBe('root-2');
  });

  it('surfaces backend failures into error state', async () => {
    const bindings = makeBindings({
      GetNodeChildren: vi.fn().mockRejectedValue(new Error('network down'))
    });
    const store = createInspectStore(bindings);

    await store.expandNode('root-1');

    expect(store.getState().errorText).toBe('network down');
    expect(store.getState().statusText).toBe('Failed to expand node');
  });
});
