package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"goahk/internal/inspect"
)

type Clipboard interface{ CopyText(string) error }
type Dialogs interface {
	PromptSetValue(defaultValue string) (string, bool, error)
}

type Controller struct {
	service   inspect.Service
	ctx       context.Context
	clipboard Clipboard
	dialogs   Dialogs

	mu                    sync.Mutex
	selectedWindowID      string
	selectedNodeID        string
	visibleOnly           bool
	titleOnly             bool
	activateOnSelect      bool
	mode                  inspect.InspectMode
	nodesByID             map[string]inspect.TreeNodeDTO
	nodeChildren          map[string][]string
	nodeExpanded          map[string]bool
	nodeLoadFailed        map[string]error
	followEnabled         bool
	followPaused          bool
	followLocked          bool
	lastFollowNode        string
	followCtx             context.Context
	followCancel          context.CancelFunc
	followDone            chan struct{}
	followTicker          func() <-chan time.Time
	followInterval        time.Duration
	accPathCaptureEnabled bool
	lastACCPath           string
	statusText            string
	lastError             string
	diagnostics           *inspect.InspectDiagnostics
	onFollowElement       []func(inspect.TreeNodeDTO)
	onFollowError         []func(error)
}

type StatusUpdate struct {
	Text              string
	CaptureEnabled    bool
	HasLastACCPath    bool
	LastACCPathCopied bool
}

func NewController(ctx context.Context, svc inspect.Service) *Controller {
	c := &Controller{ctx: ctx, service: svc, followInterval: 120 * time.Millisecond, nodesByID: map[string]inspect.TreeNodeDTO{}, nodeChildren: map[string][]string{}, nodeExpanded: map[string]bool{}, nodeLoadFailed: map[string]error{}, statusText: "Click status to enable ACC path capture"}
	c.followTicker = func() <-chan time.Time {
		t := time.NewTicker(c.followInterval)
		out := make(chan time.Time)
		go func() {
			defer close(out)
			defer t.Stop()
			for {
				select {
				case <-c.followCtx.Done():
					return
				case at := <-t.C:
					out <- at
				}
			}
		}()
		return out
	}
	return c
}
func (c *Controller) WithClipboard(cb Clipboard) *Controller { c.clipboard = cb; return c }
func (c *Controller) WithDialogs(d Dialogs) *Controller      { c.dialogs = d; return c }
func (c *Controller) SetClipboard(cb Clipboard)              { c.clipboard = cb }
func (c *Controller) SetDialogs(d Dialogs)                   { c.dialogs = d }
func (c *Controller) runtimeContext() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}
func (c *Controller) OnFollowCursorElement(cb func(inspect.TreeNodeDTO)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onFollowElement = append(c.onFollowElement, cb)
}
func (c *Controller) OnFollowCursorError(cb func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onFollowError = append(c.onFollowError, cb)
}
func (c *Controller) RefreshWindows(filter string, visibleOnly, titleOnly bool) (inspect.RefreshWindowsResponse, error) {
	_, _ = c.service.ClearHighlight(c.runtimeContext(), inspect.ClearHighlightRequest{})
	c.mu.Lock()
	c.visibleOnly = visibleOnly
	c.titleOnly = titleOnly
	c.mu.Unlock()
	return c.service.RefreshWindows(c.runtimeContext(), inspect.RefreshWindowsRequest{Filter: filter, VisibleOnly: visibleOnly, TitleOnly: titleOnly})
}
func (c *Controller) SelectWindow(hwnd string) error {
	_, _ = c.service.ClearHighlight(c.runtimeContext(), inspect.ClearHighlightRequest{})
	c.mu.Lock()
	activate := c.activateOnSelect
	mode := c.mode
	c.selectedWindowID = hwnd
	c.mu.Unlock()
	if activate {
		if _, err := c.service.ActivateWindow(c.runtimeContext(), inspect.ActivateWindowRequest{HWND: hwnd}); err != nil {
			return err
		}
	}
	if _, err := c.service.InspectWindow(c.runtimeContext(), inspect.InspectWindowRequest{HWND: hwnd, Mode: mode}); err != nil {
		return err
	}
	root, err := c.service.GetTreeRoot(c.runtimeContext(), inspect.GetTreeRootRequest{HWND: hwnd, Refresh: true, Mode: mode})
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.nodesByID[root.Root.NodeID] = root.Root
	c.diagnostics = root.Diagnostics
	c.mu.Unlock()
	_, err = c.service.GetNodeDetails(c.runtimeContext(), inspect.GetNodeDetailsRequest{NodeID: root.Root.NodeID})
	return err
}
func (c *Controller) LoadTreeRoot() (inspect.GetTreeRootResponse, error) {
	c.mu.Lock()
	hwnd := c.selectedWindowID
	mode := c.mode
	c.mu.Unlock()
	return c.service.GetTreeRoot(c.runtimeContext(), inspect.GetTreeRootRequest{HWND: hwnd, Refresh: true, Mode: mode})
}
func (c *Controller) ExpandNode(nodeID string) (inspect.GetNodeChildrenResponse, error) {
	resp, err := c.service.GetNodeChildren(c.runtimeContext(), inspect.GetNodeChildrenRequest{NodeID: nodeID})
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.nodeLoadFailed[nodeID] = err
		return resp, err
	}
	ids := make([]string, 0, len(resp.Children))
	for _, ch := range resp.Children {
		c.nodesByID[ch.NodeID] = ch
		ids = append(ids, ch.NodeID)
	}
	c.nodeChildren[nodeID] = ids
	c.nodeExpanded[nodeID] = true
	delete(c.nodeLoadFailed, nodeID)
	return resp, nil
}
func (c *Controller) SelectNode(nodeID string) error {
	if _, err := c.service.SelectNode(c.runtimeContext(), inspect.SelectNodeRequest{NodeID: nodeID}); err != nil {
		return err
	}
	details, err := c.service.GetNodeDetails(c.runtimeContext(), inspect.GetNodeDetailsRequest{NodeID: nodeID})
	if err != nil {
		return err
	}
	_, err = c.service.HighlightNode(c.runtimeContext(), inspect.HighlightNodeRequest{NodeID: nodeID})
	if err == nil {
		c.mu.Lock()
		c.selectedNodeID = nodeID
		if c.accPathCaptureEnabled && strings.TrimSpace(details.ACCPath) != "" {
			c.lastACCPath = details.ACCPath
			c.statusText = "Path: " + details.ACCPath
		}
		c.mu.Unlock()
	}
	return err
}
func (c *Controller) RefreshSelectedNodeDetails() (inspect.GetNodeDetailsResponse, error) {
	c.mu.Lock()
	nodeID := c.selectedNodeID
	c.mu.Unlock()
	return c.service.GetNodeDetails(c.runtimeContext(), inspect.GetNodeDetailsRequest{NodeID: nodeID})
}
func (c *Controller) InvokePattern(nodeID, action string, payload map[string]any) (inspect.InvokePatternResponse, error) {
	return c.service.InvokePattern(c.runtimeContext(), inspect.InvokePatternRequest{NodeID: nodeID, Action: action, Payload: payload})
}
func (c *Controller) InvokePatternForSelection(action string) (inspect.InvokePatternResponse, error) {
	c.mu.Lock()
	nodeID := c.selectedNodeID
	c.mu.Unlock()
	return c.InvokePattern(nodeID, action, nil)
}
func (c *Controller) InvokeSetValue() (inspect.InvokePatternResponse, error) {
	c.mu.Lock()
	nodeID := c.selectedNodeID
	c.mu.Unlock()
	if c.dialogs == nil {
		return inspect.InvokePatternResponse{}, errors.New("setValue dialog unavailable")
	}
	value, ok, err := c.dialogs.PromptSetValue("")
	if err != nil {
		return inspect.InvokePatternResponse{}, err
	}
	if !ok {
		return inspect.InvokePatternResponse{}, nil
	}
	return c.InvokePattern(nodeID, "setValue", map[string]any{"value": value})
}
func (c *Controller) CopyProperty(v string) string {
	if c.clipboard != nil {
		_ = c.clipboard.CopyText(v)
	}
	return v
}
func (c *Controller) CopySelectedText(v string) error {
	if c.clipboard == nil {
		return nil
	}
	return c.clipboard.CopyText(v)
}
func (c *Controller) CopyBestSelector(nodeID string) (inspect.CopyBestSelectorResponse, error) {
	return c.service.CopyBestSelector(c.runtimeContext(), inspect.CopyBestSelectorRequest{NodeID: nodeID})
}
func (c *Controller) ToggleAccPathCapture() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accPathCaptureEnabled = !c.accPathCaptureEnabled
	if c.accPathCaptureEnabled {
		c.statusText = "ACC path capture enabled"
	} else {
		c.statusText = "ACC path capture paused"
	}
	return c.accPathCaptureEnabled
}
func (c *Controller) OnStatusInteraction() string {
	return c.OnStatusInteractionUpdate().Text
}

func (c *Controller) OnStatusInteractionUpdate() StatusUpdate {
	c.mu.Lock()
	path := strings.TrimSpace(c.lastACCPath)
	enabled := c.accPathCaptureEnabled
	c.mu.Unlock()
	if !enabled || path == "" {
		enabled = c.ToggleAccPathCapture()
		c.mu.Lock()
		defer c.mu.Unlock()
		return StatusUpdate{Text: c.statusText, CaptureEnabled: enabled, HasLastACCPath: strings.TrimSpace(c.lastACCPath) != ""}
	}
	if c.clipboard != nil {
		_ = c.clipboard.CopyText(path)
	}
	c.mu.Lock()
	c.statusText = "ACC path copied"
	c.mu.Unlock()
	return StatusUpdate{Text: "ACC path copied", CaptureEnabled: true, HasLastACCPath: true, LastACCPathCopied: true}
}
func (c *Controller) PauseFollowCursor() {
	c.mu.Lock()
	c.followPaused = true
	c.mu.Unlock()
	_, _ = c.service.PauseFollowCursor(c.runtimeContext(), inspect.PauseFollowCursorRequest{})
}
func (c *Controller) ResumeFollowCursor() {
	c.mu.Lock()
	c.followPaused = false
	c.mu.Unlock()
	_, _ = c.service.ResumeFollowCursor(c.runtimeContext(), inspect.ResumeFollowCursorRequest{})
}
func (c *Controller) LockFollowCursor() {
	c.mu.Lock()
	c.followLocked = true
	node := c.selectedNodeID
	c.mu.Unlock()
	_, _ = c.service.LockFollowCursor(c.runtimeContext(), inspect.LockFollowCursorRequest{NodeID: node})
}
func (c *Controller) UnlockFollowCursor() {
	c.mu.Lock()
	c.followLocked = false
	c.mu.Unlock()
	_, _ = c.service.UnlockFollowCursor(c.runtimeContext(), inspect.UnlockFollowCursorRequest{})
}
func (c *Controller) ToggleFollowCursor(enabled bool) error {
	c.mu.Lock()
	if enabled == c.followEnabled {
		c.mu.Unlock()
		return nil
	}
	if enabled {
		c.followCtx, c.followCancel = context.WithCancel(c.runtimeContext())
		c.followDone = make(chan struct{})
		c.followEnabled = true
		c.lastFollowNode = ""
		loopCtx := c.followCtx
		done := c.followDone
		c.mu.Unlock()
		go c.runFollow(loopCtx, done)
		_, _ = c.service.ToggleFollowCursor(c.runtimeContext(), inspect.ToggleFollowCursorRequest{Enabled: true})
		return nil
	}
	cancel := c.followCancel
	done := c.followDone
	c.followCancel = nil
	c.followDone = nil
	c.followEnabled = false
	c.followPaused = false
	c.followLocked = false
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	_, _ = c.service.ToggleFollowCursor(c.runtimeContext(), inspect.ToggleFollowCursorRequest{Enabled: false})
	return nil
}
func (c *Controller) runFollow(loopCtx context.Context, done chan struct{}) {
	defer close(done)
	ticks := c.followTicker()
	for {
		select {
		case <-loopCtx.Done():
			return
		case <-ticks:
			c.mu.Lock()
			paused := c.followPaused
			locked := c.followLocked
			c.mu.Unlock()
			if paused || locked {
				continue
			}
			resp, err := c.service.GetElementUnderCursor(loopCtx, inspect.GetElementUnderCursorRequest{})
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				c.emitFollowErr(err)
				continue
			}
			if resp.Element.NodeID == "" {
				continue
			}
			c.mu.Lock()
			if resp.Element.NodeID == c.lastFollowNode {
				c.mu.Unlock()
				continue
			}
			c.lastFollowNode = resp.Element.NodeID
			c.mu.Unlock()
			c.emitFollowElement(resp.Element)
		}
	}
}
func (c *Controller) emitFollowElement(n inspect.TreeNodeDTO) {
	c.mu.Lock()
	cbs := append([]func(inspect.TreeNodeDTO){}, c.onFollowElement...)
	c.mu.Unlock()
	for _, cb := range cbs {
		cb(n)
	}
}
func (c *Controller) emitFollowErr(err error) {
	c.mu.Lock()
	c.lastError = normalizeInspectError(err)
	cbs := append([]func(error){}, c.onFollowError...)
	c.mu.Unlock()
	for _, cb := range cbs {
		cb(err)
	}
}
func (c *Controller) Shutdown() {
	_ = c.ToggleFollowCursor(false)
	_, _ = c.service.ClearHighlight(c.runtimeContext(), inspect.ClearHighlightRequest{})
	c.mu.Lock()
	c.followCancel = nil
	c.followDone = nil
	c.followEnabled = false
	c.followPaused = false
	c.followLocked = false
	c.lastFollowNode = ""
	c.selectedNodeID = ""
	c.mu.Unlock()
}
func normalizeInspectError(err error) string {
	switch {
	case err == nil:
		return "none"
	case errors.Is(err, inspect.ErrProviderActionUnsupported):
		return "ErrProviderActionUnsupported"
	case errors.Is(err, inspect.ErrStaleCache):
		return "ErrStaleCache"
	case errors.Is(err, inspect.ErrInvalidNodeID):
		return "ErrInvalidNodeID"
	case errors.Is(err, inspect.ErrTransientFailure), errors.Is(err, inspect.ErrProviderTransientFailure):
		return "ErrTransientFailure"
	default:
		return fmt.Sprintf("%T", err)
	}
}
