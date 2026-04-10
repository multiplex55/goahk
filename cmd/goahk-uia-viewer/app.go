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

	emitEvent func(name string, payload any)

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
	a.SetEventEmitter(func(name string, payload any) {
		wailsRuntime.EventsEmit(ctx, name, payload)
	})
}

func (a *ViewerApp) OnShutdown(ctx context.Context) {
	_, _ = a.ToggleFollowCursor(ctx, inspect.ToggleFollowCursorRequest{Enabled: false})
	a.SetEventEmitter(nil)
}

func (a *ViewerApp) SetEventEmitter(emitter func(name string, payload any)) {
	a.emitEvent = emitter
}

func (a *ViewerApp) ListWindows(ctx context.Context, req inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return a.service.ListWindows(ctx, req)
}

func (a *ViewerApp) InspectWindow(ctx context.Context, req inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	return a.service.InspectWindow(ctx, req)
}

func (a *ViewerApp) GetTreeRoot(ctx context.Context, req inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	return a.service.GetTreeRoot(ctx, req)
}

func (a *ViewerApp) GetNodeChildren(ctx context.Context, req inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	return a.service.GetNodeChildren(ctx, req)
}

func (a *ViewerApp) SelectNode(ctx context.Context, req inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return a.service.SelectNode(ctx, req)
}

func (a *ViewerApp) GetFocusedElement(ctx context.Context, req inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	return a.service.GetFocusedElement(ctx, req)
}

func (a *ViewerApp) GetElementUnderCursor(ctx context.Context, req inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	return a.service.GetElementUnderCursor(ctx, req)
}

func (a *ViewerApp) HighlightNode(ctx context.Context, req inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	return a.service.HighlightNode(ctx, req)
}

func (a *ViewerApp) ClearHighlight(ctx context.Context, req inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	return a.service.ClearHighlight(ctx, req)
}

func (a *ViewerApp) CopyBestSelector(ctx context.Context, req inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return a.service.CopyBestSelector(ctx, req)
}

func (a *ViewerApp) GetPatternActions(ctx context.Context, req inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return a.service.GetPatternActions(ctx, req)
}

func (a *ViewerApp) InvokePattern(ctx context.Context, req inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	return a.service.InvokePattern(ctx, req)
}

func (a *ViewerApp) ToggleFollowCursor(ctx context.Context, req inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	a.followMu.Lock()
	alreadyEnabled := a.followEnabled
	if req.Enabled == alreadyEnabled {
		a.followMu.Unlock()
		return inspect.ToggleFollowCursorResponse{Enabled: req.Enabled}, nil
	}

	if req.Enabled {
		a.followCtx, a.followCancel = context.WithCancel(context.Background())
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
		select {
		case <-done:
		case <-ctx.Done():
			return inspect.ToggleFollowCursorResponse{Enabled: false}, ctx.Err()
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

func (a *ViewerApp) RefreshWindows(ctx context.Context, req inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	return a.service.RefreshWindows(ctx, req)
}
