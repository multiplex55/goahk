package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"goahk/internal/inspect"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type followCursorEvent struct {
	EventID  int64               `json:"eventID"`
	WindowID string              `json:"windowID,omitempty"`
	Element  inspect.TreeNodeDTO `json:"element"`
}

type ViewerApp struct {
	service inspect.Service

	emitEvent  func(name string, payload any)
	runtimeCtx context.Context

	followMu       sync.Mutex
	followEnabled  bool
	followCtx      context.Context
	followCancel   context.CancelFunc
	followDone     chan struct{}
	followTicker   func() <-chan time.Time
	followInterval time.Duration
	followEventID  int64
	lastNodeID     string
}

func NewViewerApp(service inspect.Service) *ViewerApp {
	app := &ViewerApp{service: service, followInterval: 120 * time.Millisecond}
	app.followTicker = func() <-chan time.Time {
		ticker := time.NewTicker(app.followInterval)
		out := make(chan time.Time)
		go func() {
			defer close(out)
			defer ticker.Stop()
			for {
				select {
				case <-app.followCtx.Done():
					return
				case t := <-ticker.C:
					out <- t
				}
			}
		}()
		return out
	}
	return app
}

func (a *ViewerApp) OnStartup(ctx context.Context) {
	a.runtimeCtx = ctx
	a.SetEventEmitter(func(name string, payload any) {
		wailsRuntime.EventsEmit(a.runtimeContext(), name, payload)
	})
}

func (a *ViewerApp) OnShutdown(ctx context.Context) {
	_, _ = a.ToggleFollowCursor(inspect.ToggleFollowCursorRequest{Enabled: false})
	_, _ = a.service.ClearHighlight(a.runtimeContext(), inspect.ClearHighlightRequest{})
	a.SetEventEmitter(nil)
}

func (a *ViewerApp) SetEventEmitter(emitter func(name string, payload any)) {
	a.emitEvent = emitter
}

func (a *ViewerApp) ListWindows(req inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return a.service.ListWindows(a.runtimeContext(), req)
}

func (a *ViewerApp) InspectWindow(req inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	return a.service.InspectWindow(a.runtimeContext(), req)
}

func (a *ViewerApp) GetTreeRoot(req inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	return a.service.GetTreeRoot(a.runtimeContext(), req)
}

func (a *ViewerApp) GetNodeChildren(req inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	return a.service.GetNodeChildren(a.runtimeContext(), req)
}

func (a *ViewerApp) SelectNode(req inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return a.service.SelectNode(a.runtimeContext(), req)
}

func (a *ViewerApp) GetNodeDetails(req inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	return a.service.GetNodeDetails(a.runtimeContext(), req)
}

func (a *ViewerApp) GetFocusedElement(req inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	return a.service.GetFocusedElement(a.runtimeContext(), req)
}

func (a *ViewerApp) GetElementUnderCursor(req inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	return a.service.GetElementUnderCursor(a.runtimeContext(), req)
}

func (a *ViewerApp) HighlightNode(req inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	return a.service.HighlightNode(a.runtimeContext(), req)
}

func (a *ViewerApp) ClearHighlight(req inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	return a.service.ClearHighlight(a.runtimeContext(), req)
}

func (a *ViewerApp) CopyBestSelector(req inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return a.service.CopyBestSelector(a.runtimeContext(), req)
}

func (a *ViewerApp) GetPatternActions(req inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return a.service.GetPatternActions(a.runtimeContext(), req)
}

func (a *ViewerApp) InvokePattern(req inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	return a.service.InvokePattern(a.runtimeContext(), req)
}

func (a *ViewerApp) ActivateWindow(req inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	return a.service.ActivateWindow(a.runtimeContext(), req)
}

func (a *ViewerApp) ToggleFollowCursor(req inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	a.followMu.Lock()
	alreadyEnabled := a.followEnabled
	if req.Enabled == alreadyEnabled {
		a.followMu.Unlock()
		return inspect.ToggleFollowCursorResponse{Enabled: req.Enabled}, nil
	}

	if req.Enabled {
		a.followCtx, a.followCancel = context.WithCancel(a.runtimeContext())
		a.followDone = make(chan struct{})
		a.followEnabled = true
		a.lastNodeID = ""
		followCtx := a.followCtx
		followDone := a.followDone
		a.followMu.Unlock()
		go a.runFollowCursorLoop(followCtx, followDone)
		return inspect.ToggleFollowCursorResponse{Enabled: true}, nil
	}

	cancel := a.followCancel
	done := a.followDone
	a.followCancel = nil
	a.followDone = nil
	a.followEnabled = false
	a.followMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		waitCtx, cancelWait := context.WithTimeout(a.runtimeContext(), 2*time.Second)
		defer cancelWait()
		select {
		case <-done:
		case <-waitCtx.Done():
			return inspect.ToggleFollowCursorResponse{Enabled: false}, waitCtx.Err()
		}
	}

	return inspect.ToggleFollowCursorResponse{Enabled: false}, nil
}

func (a *ViewerApp) runFollowCursorLoop(loopCtx context.Context, done chan struct{}) {
	defer close(done)
	ticks := a.followTicker()
	for {
		select {
		case <-loopCtx.Done():
			return
		case <-ticks:
			resp, err := a.service.GetElementUnderCursor(loopCtx, inspect.GetElementUnderCursorRequest{})
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				a.emit("inspect:follow-cursor-error", map[string]string{"message": err.Error()})
				continue
			}
			nodeID := resp.Element.NodeID
			if nodeID == "" {
				continue
			}
			a.followMu.Lock()
			if nodeID == a.lastNodeID {
				a.followMu.Unlock()
				continue
			}
			a.lastNodeID = nodeID
			a.followEventID++
			eventID := a.followEventID
			a.followMu.Unlock()
			a.emit("inspect:follow-cursor", followCursorEvent{EventID: eventID, Element: resp.Element})
		}
	}
}

func (a *ViewerApp) emit(name string, payload any) {
	if a.emitEvent != nil {
		a.emitEvent(name, payload)
	}
}

func (a *ViewerApp) RefreshWindows(req inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	return a.service.RefreshWindows(a.runtimeContext(), req)
}

func (a *ViewerApp) runtimeContext() context.Context {
	if a.runtimeCtx != nil {
		return a.runtimeCtx
	}
	return context.Background()
}
