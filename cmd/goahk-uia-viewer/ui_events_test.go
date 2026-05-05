package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"goahk/internal/inspect"
)

type queueMarshaller struct{ ch chan func() }

func (m *queueMarshaller) Queue(fn func()) { m.ch <- fn }

type guardedView struct {
	mu       sync.Mutex
	queued   bool
	busy     []bool
	status   []string
	updates  int
	expanded []string
	selected []string
	fatal    []string
}

func (v *guardedView) enterQueue()                                        { v.mu.Lock(); v.queued = true; v.mu.Unlock() }
func (v *guardedView) exitQueue()                                         { v.mu.Lock(); v.queued = false; v.mu.Unlock() }
func (v *guardedView) SetBusy(b bool)                                     { v.mu.Lock(); v.busy = append(v.busy, b); v.mu.Unlock() }
func (v *guardedView) SetStatus(s string)                                 { v.mu.Lock(); v.status = append(v.status, s); v.mu.Unlock() }
func (v *guardedView) ShowFatal(s string)                                 { v.mu.Lock(); v.fatal = append(v.fatal, s); v.mu.Unlock() }
func (v *guardedView) UpdateWindowDetails(inspect.GetNodeDetailsResponse) { v.updates++ }
func (v *guardedView) UpdateNodeDetails(inspect.GetNodeDetailsResponse)   { v.updates++ }
func (v *guardedView) UpdateTreeRoot(inspect.TreeNodeDTO)                 { v.updates++ }
func (v *guardedView) UpdateNodeChildren(string, []inspect.TreeNodeDTO)   { v.updates++ }
func (v *guardedView) ExpandTreeNode(nodeID string)                       { v.expanded = append(v.expanded, nodeID) }
func (v *guardedView) SelectTreeNode(nodeID string)                       { v.selected = append(v.selected, nodeID) }

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
	if svc.inspectCalls != 1 || svc.treeRootCalls < 1 || svc.nodeDetailsCalls < 1 {
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
	if len(view.status) == 0 || view.status[len(view.status)-1] != "window loaded InspectWindow [0x2]: properties=0 patterns=0 children=1" {
		t.Fatalf("expected success status with counts, got %v", view.status)
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

func TestOnWindowSelectedShowsErrorWhenInspectFails(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, inspectErr: context.DeadlineExceeded}
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
	if len(view.selected) == 0 || view.selected[0] != "root" {
		t.Fatalf("expected root selected, got %v", view.selected)
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
