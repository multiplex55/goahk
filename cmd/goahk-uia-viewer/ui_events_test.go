package main

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"goahk/internal/inspect"
)

type queueMarshaller struct{ ch chan func() }

func (m *queueMarshaller) Queue(fn func()) { m.ch <- fn }

type guardedView struct {
	mu              sync.Mutex
	queued          bool
	busy            []bool
	status          []string
	updates         int
	expanded        []string
	expandCalls     int
	childrenUpdated []string
	selected        []string
	fatal           []string
}

func (v *guardedView) enterQueue()                                        { v.mu.Lock(); v.queued = true; v.mu.Unlock() }
func (v *guardedView) exitQueue()                                         { v.mu.Lock(); v.queued = false; v.mu.Unlock() }
func (v *guardedView) SetBusy(b bool)                                     { v.mu.Lock(); v.busy = append(v.busy, b); v.mu.Unlock() }
func (v *guardedView) SetStatus(s string)                                 { v.mu.Lock(); v.status = append(v.status, s); v.mu.Unlock() }
func (v *guardedView) ShowFatal(s string)                                 { v.mu.Lock(); v.fatal = append(v.fatal, s); v.mu.Unlock() }
func (v *guardedView) UpdateWindowDetails(inspect.GetNodeDetailsResponse) { v.updates++ }
func (v *guardedView) UpdateNodeDetails(inspect.GetNodeDetailsResponse)   { v.updates++ }
func (v *guardedView) UpdateTreeRoot(inspect.TreeNodeDTO)                 { v.updates++ }
func (v *guardedView) UpdateNodeChildren(nodeID string, _ []inspect.TreeNodeDTO) {
	v.updates++
	v.childrenUpdated = append(v.childrenUpdated, nodeID)
}
func (v *guardedView) ExpandTreeNode(nodeID string) {
	v.expandCalls++
	v.expanded = append(v.expanded, nodeID)
}
func (v *guardedView) SelectTreeNode(nodeID string) { v.selected = append(v.selected, nodeID) }

func TestViewerEventAdapter_WindowSelectionPipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnWindowSelected("0x2", false)
	select {
	case fn := <-mq.ch:
		view.enterQueue()
		fn()
		view.exitQueue()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued callback")
	}
	if svc.treeRootCalls < 1 || svc.nodeDetailsCalls < 1 {
		t.Fatalf("pipeline incomplete: %+v", svc)
	}
	if len(view.busy) != 2 || !view.busy[0] || view.busy[1] {
		t.Fatalf("busy toggles = %v, want [true false]", view.busy)
	}
}

func TestOnWindowSelectedUpdatesTreeRootAndDetails(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if view.updates < 3 {
		t.Fatalf("expected tree+window+node updates, got %d", view.updates)
	}
}
func TestOnWindowSelectedDoesNotCallRefreshSelectedNodeDetailsWithEmptyNode(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if svc.nodeDetailsCalls != 1 {
		t.Fatalf("expected details from select+selectNode only, got %d", svc.nodeDetailsCalls)
	}
}

func TestViewerEventAdapter_TreeExpandLazyLoad(t *testing.T) {
	svc := &fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"n1": {{NodeID: "c1"}}}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnTreeExpanded("n1", true)
	if len(mq.ch) != 0 {
		t.Fatal("loaded node should not queue work")
	}
	adapter.OnTreeExpanded("n1", false)
	fn := <-mq.ch
	view.enterQueue()
	fn()
	view.exitQueue()
	if len(svc.nodeChildrenReqs) != 1 {
		t.Fatalf("expected child load")
	}
	if len(view.busy) != 2 || !view.busy[0] || view.busy[1] {
		t.Fatalf("busy toggles = %v, want [true false]", view.busy)
	}
}

func TestViewerEventAdapter_TreeSelectPipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnTreeSelected("node-1")
	fn := <-mq.ch
	view.enterQueue()
	fn()
	view.exitQueue()
	if svc.selectCalls != 1 || svc.nodeDetailsCalls < 1 || svc.highlightCalls != 1 {
		t.Fatalf("select pipeline incomplete: %+v", svc)
	}
	if len(view.busy) != 2 || !view.busy[0] || view.busy[1] {
		t.Fatalf("busy toggles = %v, want [true false]", view.busy)
	}
}

func TestViewerEventAdapter_WindowSelectionStatusSet(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	view.enterQueue()
	fn()
	view.exitQueue()
	if len(view.status) == 0 || !strings.Contains(view.status[len(view.status)-1], "window loaded GetTreeRoot [0x2]: properties=0 patterns=0 children=1; parity:") {
		t.Fatalf("expected success status with counts, got %v", view.status)
	}
}

func TestViewerEventAdapter_WindowSelectionFallbackStatusCopy(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root"},
		rootState: inspect.InspectModeState{
			RequestedMode: inspect.InspectModeUIATree,
			ActiveMode:    inspect.InspectModeHWNDTree,
			FallbackUsed:  true,
			Provider:      "acc",
			Backend:       "hwnd",
			DegradeReason: "UIA tree is unavailable. Switch to ACC/MSAA mode to continue inspecting this window.",
		},
	}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.status) == 0 || !strings.Contains(view.status[len(view.status)-1], "degrade reason: UIA tree is unavailable.") {
		t.Fatalf("expected fallback warning status, got %v", view.status)
	}
}

func TestViewerEventAdapter_WindowSelection_RespectsActivateFlag(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnWindowSelected("0x2", true)
	fn := <-mq.ch
	view.enterQueue()
	fn()
	view.exitQueue()
	if svc.activateCalls != 1 {
		t.Fatalf("expected activate call when requested, got %d", svc.activateCalls)
	}
}

func TestOnWindowSelectedShowsErrorWhenGetTreeRootFails(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, treeRootErr: context.DeadlineExceeded}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.status) == 0 || len(view.fatal) == 0 {
		t.Fatal("expected error status")
	}
}

func TestOnWindowSelectedTransientFailureDoesNotShowFatalModal(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, treeRootErr: inspect.ErrTransientFailure}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.status) == 0 {
		t.Fatal("expected status update")
	}
	if len(view.fatal) != 0 {
		t.Fatalf("transient error should not show fatal modal: %v", view.fatal)
	}
}

func TestOnWindowSelectedHardFailureShowsFatalModal(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, treeRootErr: context.DeadlineExceeded}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.fatal) == 0 {
		t.Fatal("expected fatal modal for hard failure")
	}
}

func TestOnWindowSelectedStillShowsDetailsWhenChildrenFail(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root"},
		nodeChildrenErr:    context.DeadlineExceeded,
	}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if view.updates < 3 {
		t.Fatalf("expected root/details updates despite child error, got updates=%d", view.updates)
	}
	if len(view.expanded) != 0 {
		t.Fatalf("did not expect expansion when children failed, got %v", view.expanded)
	}
	if len(view.fatal) != 0 {
		t.Fatalf("child load error should be non-fatal, got %v", view.fatal)
	}
	if len(view.status) == 0 || !strings.Contains(view.status[len(view.status)-1], "UIA root loaded, but child traversal failed:") {
		t.Fatalf("expected prominent child traversal status, got %v", view.status)
	}
	if strings.Contains(strings.ToLower(view.status[len(view.status)-1]), "loaded with warning") {
		t.Fatalf("child traversal error should not be masked as generic loaded warning, got %q", view.status[len(view.status)-1])
	}
	if !strings.Contains(strings.ToLower(view.status[len(view.status)-1]), "traversal") {
		t.Fatalf("expected traversal-specific warning status, got %q", view.status[len(view.status)-1])
	}
}

func TestOnWindowSelectedUIAOnlyChildFailureIsExplicitFailureState(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root"},
		nodeChildrenErr:    context.DeadlineExceeded,
		rootState: inspect.InspectModeState{
			RequestedMode: inspect.InspectModeUIAOnly,
			ActiveMode:    inspect.InspectModeUIAOnly,
		},
	}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.status) == 0 {
		t.Fatal("expected status updates")
	}
	last := view.status[len(view.status)-1]
	if !strings.Contains(last, "ERROR UIA root loaded, but child traversal failed:") {
		t.Fatalf("expected explicit failure state for UIA_ONLY child traversal error, got %q", last)
	}
	if !strings.Contains(last, "UIA-only failure: child traversal failed") {
		t.Fatalf("expected explicit UIA-only failure suffix, got %q", last)
	}
}

func TestViewerEventAdapter_TreeExpandStatusSet(t *testing.T) {
	svc := &fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"n1": {{NodeID: "c1"}}}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnTreeExpanded("n1", false)
	fn := <-mq.ch
	view.enterQueue()
	fn()
	view.exitQueue()
	if len(view.status) == 0 || view.status[len(view.status)-1] != "node expanded GetTreeRoot [n1]" {
		t.Fatalf("expected expand status, got %v", view.status)
	}
}

func TestOnTreeExpandedAddsChildren(t *testing.T) {
	svc := &fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"n1": {{NodeID: "c1"}}}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnTreeExpanded("n1", false)
	fn := <-mq.ch
	fn()
	if view.updates == 0 {
		t.Fatal("expected tree children update")
	}
}

func TestOnTreeSelectedUpdatesDetails(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnTreeSelected("node-1")
	fn := <-mq.ch
	fn()
	if view.updates == 0 {
		t.Fatal("expected details update")
	}
}

func TestOnWindowSelectedExpandsAndSelectsRoot(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.expanded) == 0 || view.expanded[0] != "root" {
		t.Fatalf("expected root expanded, got %v", view.expanded)
	}
	if view.expandCalls == 0 {
		t.Fatal("expected expand command path invocation")
	}
	if len(view.selected) == 0 || view.selected[0] != "root" {
		t.Fatalf("expected root selected, got %v", view.selected)
	}
}

func TestOnWindowSelectedExpansionOrderMatchesLoadedResults(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"root": {{NodeID: "child-1"}}, "child-1": {{NodeID: "grand-1"}}}}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.autoExpandDepth = func() int { return 2 }

	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()

	want := []string{"root", "child-1"}
	if len(view.expanded) != len(want) {
		t.Fatalf("expected deterministic expansion order %v, got %v", want, view.expanded)
	}
	for i := range want {
		if view.expanded[i] != want[i] {
			t.Fatalf("expected deterministic expansion order %v, got %v", want, view.expanded)
		}
	}
}

func TestOnWindowSelectedProgrammaticExpansionDoesNotTriggerDuplicateChildLoad(t *testing.T) {
	svc := &fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"n1": {{NodeID: "c1"}}}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnTreeExpanded("n1", false)
	fn := <-mq.ch
	fn()
	if len(svc.nodeChildrenReqs) != 1 {
		t.Fatalf("expected single child-load dispatch, got %d", len(svc.nodeChildrenReqs))
	}
}

func TestTreeSelectionTracksDisplayedDetails(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.selected) == 0 || view.selected[len(view.selected)-1] != "root" {
		t.Fatalf("expected selected root, got %v", view.selected)
	}

	adapter.OnTreeExpanded("root", false)
	fn = <-mq.ch
	fn()
	if len(view.selected) == 0 || view.selected[len(view.selected)-1] != "root" {
		t.Fatalf("expected root to remain selected after expansion, got %v", view.selected)
	}
}

func TestOnWindowSelectedUpdatesTreeRootEvenWhenDetailsFail(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root"},
		nodeDetailsErr:     context.DeadlineExceeded,
	}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.selected) == 0 || view.selected[0] != "root" {
		t.Fatalf("expected root selected despite details error, got %v", view.selected)
	}
	if view.updates == 0 {
		t.Fatal("expected tree root update")
	}
}

func TestOnWindowSelectedShowsFallbackStatus(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root"},
		nodeDetailsErr:     context.DeadlineExceeded,
	}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.status) == 0 {
		t.Fatal("expected status")
	}
	status := view.status[len(view.status)-1]
	if !strings.Contains(status, "partial load") || !strings.Contains(status, "GetNodeDetails") {
		t.Fatalf("expected partial fallback status, got %q", status)
	}
}

func TestOnWindowSelectedUpdatesChildrenForEveryAutoExpandLevel(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{
			childrenByNode: map[string][]inspect.TreeNodeDTO{
				"root": {{NodeID: "a"}, {NodeID: "b"}},
				"a":    {{NodeID: "a1"}},
				"b":    {},
			},
		},
		root: inspect.TreeNodeDTO{NodeID: "root"},
	}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.autoExpandDepth = func() int { return 2 }

	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()

	want := []string{"root", "child-1"}
	if len(view.childrenUpdated) != len(want) {
		t.Fatalf("expected UpdateNodeChildren calls for every expanded level %v, got %v", want, view.childrenUpdated)
	}
	for i := range want {
		if view.childrenUpdated[i] != want[i] {
			t.Fatalf("expected UpdateNodeChildren order %v, got %v", want, view.childrenUpdated)
		}
	}
}

func TestOnWindowSelectedAutoExpandsLoadedRootChildren(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"root": {{NodeID: "child"}}}}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 2)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)
	adapter.OnWindowSelected("0x2", false)
	fn := <-mq.ch
	fn()
	if len(view.expanded) == 0 || view.expanded[0] != "root" {
		t.Fatalf("expected root auto expansion, got %v", view.expanded)
	}
}
