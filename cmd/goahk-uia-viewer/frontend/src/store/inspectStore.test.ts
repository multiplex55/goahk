import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createInspectStore, InspectBindings } from './inspectStore';

function makeBindings(overrides?: Partial<InspectBindings>): InspectBindings {
  return {
    RefreshWindows: vi.fn().mockResolvedValue({ windows: [{ hwnd: 'w1', title: 'Notepad', processName: 'notepad.exe', processID: 10 }] }),
    InspectWindow: vi.fn().mockResolvedValue({ window: { hwnd: 'w1', title: 'Notepad', processName: 'notepad.exe' }, rootNodeID: 'root-1' }),
    GetTreeRoot: vi.fn().mockResolvedValue({ root: { nodeID: 'root-1', name: 'Root', hasChildren: true } }),
    GetNodeChildren: vi.fn().mockResolvedValue({
      parentNodeID: 'root-1',
      children: [{ nodeID: 'child-1', name: 'Child', hasChildren: false }]
    }),
    SelectNode: vi.fn().mockImplementation(async ({ nodeID }) => ({
      selected: { nodeID, name: `Node ${nodeID}`, hasChildren: false }
    })),
    GetNodeDetails: vi.fn().mockImplementation(async ({ nodeID }) => ({
      windowInfo: { hwnd: 'w1', title: 'Notepad' },
      properties: [{ name: 'AutomationId', value: nodeID }],
      patterns: [{ name: 'Invoke' }],
      statusText: `Details ${nodeID}`,
      bestSelector: `#${nodeID}`,
      path: [{ nodeID: 'root-1', hasChildren: true }, { nodeID, hasChildren: false }]
    })),
    InvokePattern: vi.fn().mockImplementation(async ({ nodeID, action }) => ({
      invoked: true,
      action,
      nodeID
    })),
    CopyBestSelector: vi.fn().mockResolvedValue({ selector: '#copied', clipboardUpdated: false }),
    HighlightNode: vi.fn().mockResolvedValue({ highlighted: true }),
    ClearHighlight: vi.fn().mockResolvedValue({ cleared: true }),
    ToggleFollowCursor: vi.fn().mockImplementation(async ({ enabled }) => ({ enabled })),
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
    expect(bindings.ClearHighlight).toHaveBeenCalled();
    expect(store.getState().windows[0].processName).toBe('notepad.exe');
    expect(store.getState().statusText).toContain('Loaded');
  });



  it('startup refresh path succeeds with one-arg RefreshWindows contract', async () => {
    const bindings = makeBindings({
      RefreshWindows: vi.fn().mockImplementation(async (req: { filter: string; visibleOnly: boolean; titleOnly: boolean }) => ({
        windows: [{ hwnd: `w-${req.filter || 'all'}`, title: 'Ready' }]
      }))
    });
    const store = createInspectStore(bindings);

    await store.refreshWindows();

    expect(bindings.RefreshWindows).toHaveBeenCalledTimes(1);
    expect((bindings.RefreshWindows as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.RefreshWindows as any).mock.calls[0][0]).toEqual({ filter: '', visibleOnly: true, titleOnly: false });
    expect(store.getState().windows[0].hwnd).toBe('w-all');
    expect(store.getState().statusText).toContain('Loaded 1 windows');
  });

  it('store sends exactly one request object per backend call path', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings, { followCursorDebounceMs: 1 });

    await store.refreshWindows();
    await store.selectWindow('w1');
    await store.selectNode('node-42');
    await store.expandNode('root-1');
    await store.setFollowCursor(true);
    store.applyBridgeEvent({ type: 'selection-changed', selectedNodeID: '' });

    expect((bindings.ClearHighlight as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.RefreshWindows as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.InspectWindow as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.GetTreeRoot as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.GetNodeDetails as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.SelectNode as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.HighlightNode as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.GetNodeChildren as any).mock.calls[0]).toHaveLength(1);
    expect((bindings.ToggleFollowCursor as any).mock.calls[0]).toHaveLength(1);
  });

  it('Flow 2: selecting a window optionally activates and bootstraps root, properties, patterns', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);
    store.setActivateOnSelect(true);

    await store.selectWindow('w1');

    expect(bindings.ClearHighlight).toHaveBeenCalled();
    expect(bindings.ActivateWindow).toHaveBeenCalledWith({ hwnd: 'w1' });
    expect(bindings.InspectWindow).toHaveBeenCalledWith({ hwnd: 'w1' });
    const inspectWindowArg = (bindings.InspectWindow as any).mock.calls[0][0];
    expect(inspectWindowArg).toEqual({ hwnd: 'w1' });
    expect('refresh' in inspectWindowArg).toBe(false);
    expect(bindings.GetTreeRoot).toHaveBeenCalledWith({ hwnd: 'w1', refresh: true });
    expect(Object.keys(store.getState().nodesByID)).toContain('root-1');
    expect(store.getState().selectedNodeID).toBe('root-1');
    expect(store.getState().selectedPath.map((node) => node.nodeID)).toEqual(['root-1', 'root-1']);
    expect(store.getState().properties[0].value).toBe('root-1');
    expect(store.getState().patterns[0].name).toBe('Invoke');
    expect(store.getState().selectorText).toBe('#root-1');
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
    expect(store.getState().selectorText).toBe('#node-22');
  });

  it('prioritizes backend copy selector result and status semantics', async () => {
    const bindings = makeBindings({
      CopyBestSelector: vi.fn().mockResolvedValue({ selector: '#backend', clipboardUpdated: true })
    });
    const store = createInspectStore(bindings);
    await store.selectNode('node-22');

    const resp = await store.copyBestSelector();

    expect(bindings.CopyBestSelector).toHaveBeenCalledWith({ nodeID: 'node-22' });
    expect(resp).toEqual({ selector: '#backend', clipboardUpdated: true });
    expect(store.getState().selectorText).toBe('#backend');
    expect(store.getState().statusText).toBe('Selector copied (backend)');
  });

  it('Flow 4: expanding a node lazily loads children and reuses cache', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);

    await store.expandNode('root-1');
    await store.expandNode('root-1');
    await store.expandNode('root-1');

    expect(bindings.GetNodeChildren).toHaveBeenCalledTimes(1);
    expect(store.getState().childrenByParentID['root-1']).toEqual(['child-1']);
  });

  it('repeated expand/collapse cycles fetch children only once per parent', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);

    await store.expandNode('root-1'); // expand + fetch
    await store.expandNode('root-1'); // collapse
    await store.expandNode('root-1'); // expand from cache

    expect(bindings.GetNodeChildren).toHaveBeenCalledTimes(1);
    expect(store.getState().expandedByID['root-1']).toBe(true);
    expect(store.getState().childrenLoadedByID['root-1']).toBe(true);
  });

  it('refresh invalidates only requested branch cache and preserves sibling cache', async () => {
    const bindings = makeBindings({
      GetNodeChildren: vi
        .fn()
        .mockResolvedValueOnce({
          parentNodeID: 'root-1',
          children: [
            { nodeID: 'child-a', name: 'Child A', hasChildren: true },
            { nodeID: 'child-b', name: 'Child B', hasChildren: true }
          ]
        })
        .mockResolvedValueOnce({
          parentNodeID: 'child-a',
          children: [{ nodeID: 'leaf-a1', name: 'Leaf A1', hasChildren: false }]
        })
        .mockResolvedValueOnce({
          parentNodeID: 'child-b',
          children: [{ nodeID: 'leaf-b1', name: 'Leaf B1', hasChildren: false }]
        })
        .mockResolvedValueOnce({
          parentNodeID: 'child-a',
          children: [{ nodeID: 'leaf-a2', name: 'Leaf A2', hasChildren: false }]
        })
    });
    const store = createInspectStore(bindings);

    await store.expandNode('root-1');
    await store.expandNode('child-a');
    await store.expandNode('child-b');
    await store.expandNode('child-a', { refresh: true });
    await store.expandNode('child-b'); // collapse
    await store.expandNode('child-b'); // expand from cache

    expect(bindings.GetNodeChildren).toHaveBeenCalledTimes(4);
    expect(store.getState().childrenByParentID['child-a']).toEqual(['leaf-a2']);
    expect(store.getState().childrenByParentID['child-b']).toEqual(['leaf-b1']);
    expect(store.getState().childrenLoadedByID['child-a']).toBe(true);
    expect(store.getState().childrenLoadedByID['child-b']).toBe(true);
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
      GetNodeDetails: vi.fn().mockImplementation(async ({ nodeID }) => {
        if (nodeID === 'root-1') {
          return await new Promise((resolve) => {
            resolveSlow = resolve as typeof resolveSlow;
          });
        }
        return { properties: [{ name: 'AutomationId', value: 'root-2' }], patterns: [{ name: 'Focus' }], statusText: 'Details root-2', path: [{ nodeID: 'root-2', hasChildren: true }] };
      }),
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

  it('stale window failure cannot clobber a newer successful selection', async () => {
    let rejectSlow: ((err: unknown) => void) | undefined;
    const bindings = makeBindings({
      GetNodeDetails: vi.fn().mockImplementation(async ({ nodeID }) => {
        if (nodeID === 'root-1') {
          return await new Promise((_, reject) => {
            rejectSlow = reject;
          });
        }
        return {
          properties: [{ name: 'AutomationId', value: 'root-2' }],
          patterns: [{ name: 'Value' }],
          statusText: 'Details root-2',
          bestSelector: '#root-2',
          path: [{ nodeID: 'root-2', hasChildren: true }]
        };
      }),
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

    rejectSlow?.(new Error('first failed late'));
    await first;

    expect(store.getState().selectedWindowID).toBe('w2');
    expect(store.getState().selectedNodeID).toBe('root-2');
    expect(store.getState().properties).toEqual([{ name: 'AutomationId', value: 'root-2' }]);
    expect(store.getState().patterns).toEqual([{ name: 'Value' }]);
    expect(store.getState().selectorText).toBe('#root-2');
    expect(store.getState().errorText).toBe('');
  });

  it('backend error path sets error state and clears stale selection', async () => {
    const bindings = makeBindings({
      SelectNode: vi.fn().mockRejectedValue(new Error('network down'))
    });
    const store = createInspectStore(bindings);
    await store.selectWindow('w1');

    await store.selectNode('root-1');

    expect(store.getState().errorText).toBe('network down');
    expect(store.getState().statusText).toBe('Failed to select node');
    expect(store.getState().selectedNodeID).toBe('');
    expect(store.getState().properties).toEqual([]);
    expect(store.getState().patterns).toEqual([]);
  });

  it.each([
    {
      name: 'InspectWindow failure',
      overrides: {
        InspectWindow: vi.fn().mockRejectedValue(new Error('inspect boom'))
      },
      expectedStatus: 'Failed InspectWindow',
      expectedError: 'Failed InspectWindow: inspect boom'
    },
    {
      name: 'GetTreeRoot failure',
      overrides: {
        GetTreeRoot: vi.fn().mockRejectedValue(new Error('tree boom'))
      },
      expectedStatus: 'Failed GetTreeRoot',
      expectedError: 'Failed GetTreeRoot: tree boom'
    },
    {
      name: 'GetNodeDetails failure',
      overrides: {
        GetNodeDetails: vi.fn().mockRejectedValue(new Error('details boom'))
      },
      expectedStatus: 'Failed GetNodeDetails',
      expectedError: 'Failed GetNodeDetails: details boom'
    }
  ])('window bootstrap stage failure annotates status/error: $name', async ({ overrides, expectedStatus, expectedError }) => {
    const bindings = makeBindings(overrides);
    const store = createInspectStore(bindings);

    await store.selectWindow('w1');

    expect(store.getState().statusText).toBe(expectedStatus);
    expect(store.getState().errorText).toBe(expectedError);
    expect(store.getState().selectedNodeID).toBe('');
    expect(store.getState().selectedPath).toEqual([]);
    expect(store.getState().nodesByID).toEqual({});
    expect(store.getState().childrenByParentID).toEqual({});
    expect(store.getState().properties).toEqual([]);
    expect(store.getState().patterns).toEqual([]);
    expect(store.getState().selectorText).toBe('');
  });

  it('applies bridge follow-cursor events and updates selection/tree/highlight path', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings, { followCursorDebounceMs: 10 });

    await store.selectWindow('w1');
    store.applyBridgeEvent({
      type: 'follow-cursor',
      eventID: 1,
      windowID: 'w1',
      element: { nodeID: 'node-99', name: 'Hovered', hasChildren: false },
      path: [{ nodeID: 'root-1', hasChildren: true }, { nodeID: 'node-99', hasChildren: false }]
    });

    await vi.advanceTimersByTimeAsync(10);

    expect(store.getState().selectedNodeID).toBe('node-99');
    expect(store.getState().selectedPath.map((item) => item.nodeID)).toEqual(['root-1', 'node-99']);
    expect(bindings.SelectNode).toHaveBeenCalledWith({ nodeID: 'node-99' });
    expect(bindings.HighlightNode).toHaveBeenCalledWith({ nodeID: 'node-99' });
  });

  it('toggle follow-cursor is idempotent and reflects backend enabled state', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);

    await store.setFollowCursor(true);
    await store.setFollowCursor(true);
    await store.setFollowCursor(false);

    expect(bindings.ToggleFollowCursor).toHaveBeenNthCalledWith(1, { enabled: true });
    expect(bindings.ToggleFollowCursor).toHaveBeenNthCalledWith(2, { enabled: true });
    expect(bindings.ToggleFollowCursor).toHaveBeenNthCalledWith(3, { enabled: false });
    expect(store.getState().followCursor).toBe(false);
  });

  it('throttles/debounces rapid follow-cursor events to latest selection', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings, { followCursorDebounceMs: 50 });
    await store.selectWindow('w1');

    store.applyBridgeEvent({ type: 'follow-cursor', eventID: 2, windowID: 'w1', element: { nodeID: 'n1', hasChildren: false } });
    store.applyBridgeEvent({ type: 'follow-cursor', eventID: 3, windowID: 'w1', element: { nodeID: 'n2', hasChildren: false } });
    store.applyBridgeEvent({ type: 'follow-cursor', eventID: 4, windowID: 'w1', element: { nodeID: 'n3', hasChildren: false } });

    await vi.advanceTimersByTimeAsync(50);

    expect(bindings.SelectNode).toHaveBeenCalledWith({ nodeID: 'n3' });
    expect(store.getState().selectedNodeID).toBe('n3');
  });

  it('suppresses stale follow-cursor events from old window after window switch', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings, { followCursorDebounceMs: 10 });

    await store.selectWindow('w1');
    await store.selectWindow('w2');

    store.applyBridgeEvent({ type: 'follow-cursor', eventID: 10, windowID: 'w1', element: { nodeID: 'old', hasChildren: false } });
    await vi.advanceTimersByTimeAsync(10);

    expect(store.getState().selectedWindowID).toBe('w2');
    expect(store.getState().selectedNodeID).not.toBe('old');
  });

  it('selection-changed deselection clears native highlight', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);

    store.applyBridgeEvent({ type: 'selection-changed', eventID: 11, selectedNodeID: '' });

    expect(bindings.ClearHighlight).toHaveBeenCalled();
    expect(store.getState().selectedNodeID).toBe('');
  });

  it('invokes backend pattern action with normalized payload and updates success status', async () => {
    const bindings = makeBindings();
    const store = createInspectStore(bindings);
    await store.selectNode('node-22');

    await store.invokePatternAction('set-value', 'new text');

    expect(bindings.InvokePattern).toHaveBeenCalledWith({
      nodeID: 'node-22',
      action: 'setValue',
      payload: { value: 'new text' }
    });
    expect(store.getState().statusText).toBe('Executed set-value with payload');
    expect(store.getState().errorText).toBe('');
  });

  it('refreshes details and child branch after mutating pattern actions and surfaces backend errors', async () => {
    const bindings = makeBindings({
      GetNodeChildren: vi.fn().mockResolvedValue({
        parentNodeID: 'node-22',
        children: [{ nodeID: 'mutated-child', name: 'Mutated Child', hasChildren: false }]
      })
    });
    const store = createInspectStore(bindings);
    await store.selectNode('node-22');

    await store.invokePatternAction('toggle');

    expect(bindings.GetNodeDetails).toHaveBeenCalledWith({ nodeID: 'node-22' });
    expect(bindings.GetNodeChildren).toHaveBeenLastCalledWith({ nodeID: 'node-22' });
    expect(store.getState().childrenByParentID['node-22']).toEqual(['mutated-child']);

    (bindings.InvokePattern as any).mockRejectedValueOnce(new Error('backend explode'));
    await expect(store.invokePatternAction('toggle')).rejects.toThrow('backend explode');
    expect(store.getState().statusText).toBe('backend explode');
    expect(store.getState().errorText).toBe('backend explode');
  });
});
