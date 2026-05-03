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
	mu      sync.Mutex
	queued  bool
	busy    []bool
	status  []string
	updates int
}

func (v *guardedView) enterQueue() { v.mu.Lock(); v.queued = true; v.mu.Unlock() }
func (v *guardedView) exitQueue()  { v.mu.Lock(); v.queued = false; v.mu.Unlock() }
func (v *guardedView) assertQueued(t *testing.T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.queued {
		t.Fatalf("view mutation was not marshaled onto UI queue")
	}
}
func (v *guardedView) SetBusy(b bool)                                     { v.busy = append(v.busy, b) }
func (v *guardedView) SetStatus(s string)                                 { v.status = append(v.status, s) }
func (v *guardedView) UpdateWindowDetails(inspect.GetNodeDetailsResponse) { v.updates++ }
func (v *guardedView) UpdateNodeDetails(inspect.GetNodeDetailsResponse)   { v.updates++ }
func (v *guardedView) UpdateTreeRoot(inspect.TreeNodeDTO)                 { v.updates++ }
func (v *guardedView) UpdateNodeChildren(string, []inspect.TreeNodeDTO)   { v.updates++ }

func TestViewerEventAdapter_WindowSelectionPipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	mq := &queueMarshaller{ch: make(chan func(), 4)}
	view := &guardedView{}
	adapter := NewViewerEventAdapter(c, view, mq)

	adapter.OnWindowSelected("0x2")
	select {
	case fn := <-mq.ch:
		view.enterQueue()
		fn()
		view.exitQueue()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued callback")
	}
	if svc.inspectCalls != 1 || svc.treeRootCalls < 1 || svc.nodeDetailsCalls < 2 {
		t.Fatalf("pipeline incomplete: %+v", svc)
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
}
