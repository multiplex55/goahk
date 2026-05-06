package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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

type WindowSelectionResult struct {
	Root              inspect.GetTreeRootResponse
	Children          []inspect.TreeNodeDTO
	Details           inspect.GetNodeDetailsResponse
	DetailsErr        error
	ChildLoadErr      error
	SelectErr         error
	HighlightErr      error
	RootRetryWarnings []error
}

type StatusUpdate struct {
	Text              string
	CaptureEnabled    bool
	HasLastACCPath    bool
	LastACCPathCopied bool
}

func NewController(ctx context.Context, svc inspect.Service) *Controller {
	c := &Controller{ctx: ctx, service: svc, followInterval: 120 * time.Millisecond, nodesByID: map[string]inspect.TreeNodeDTO{}, nodeChildren: map[string][]string{}, nodeExpanded: map[string]bool{}, nodeLoadFailed: map[string]error{}, statusText: "Click here to enable Acc path capturing (can't be used with UIA!)"}
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
func (c *Controller) SetMode(mode inspect.InspectMode) {
	c.mu.Lock()
	c.mode = mode
	c.mu.Unlock()
	log.Printf("uia.viewer mode_set mode=%s", mode)
}
func (c *Controller) Mode() inspect.InspectMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mode
}
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
func (c *Controller) SelectWindow(hwnd string, activate bool) (WindowSelectionResult, error) {
	log.Printf("uia.viewer select_window_start hwnd=%s activate=%t", hwnd, activate)
	_, _ = c.service.ClearHighlight(c.runtimeContext(), inspect.ClearHighlightRequest{})
	c.mu.Lock()
	mode := c.mode
	c.selectedWindowID = hwnd
	c.mu.Unlock()
	result := WindowSelectionResult{}
	if activate {
		if _, err := c.service.ActivateWindow(c.runtimeContext(), inspect.ActivateWindowRequest{HWND: hwnd}); err != nil {
			return WindowSelectionResult{}, fmt.Errorf("activate window hwnd=%s: %w", hwnd, err)
		}
	}
	root, retryWarnings, err := c.getTreeRootWithRetry(hwnd, mode)
	if err != nil {
		log.Printf("uia.viewer select_window_end hwnd=%s mode=%s err=%v", hwnd, mode, err)
		return WindowSelectionResult{}, fmt.Errorf("get tree root hwnd=%s: %w", hwnd, err)
	}
	result.Root = root
	result.RootRetryWarnings = append(result.RootRetryWarnings, retryWarnings...)
	rootNodeID := root.Root.NodeID
	log.Printf("uia.viewer inspect_root_ok hwnd=%s root_node=%s", hwnd, rootNodeID)
	c.mu.Lock()
	c.selectedNodeID = rootNodeID
	c.nodesByID[rootNodeID] = root.Root
	c.diagnostics = root.Diagnostics
	c.mu.Unlock()
	log.Printf("uia.viewer inspect_details_start hwnd=%s node=%s", hwnd, rootNodeID)
	details, err := c.service.GetNodeDetails(c.runtimeContext(), inspect.GetNodeDetailsRequest{NodeID: rootNodeID})
	if err != nil {
		result.DetailsErr = fmt.Errorf("get node details node=%s hwnd=%s: %w", rootNodeID, hwnd, err)
		result.Details = synthesizeRootDetails(root, result.DetailsErr)
		log.Printf("uia.viewer inspect_details_err hwnd=%s node=%s err=%v", hwnd, rootNodeID, err)
		log.Printf("uia.viewer inspect_details_fallback_ok hwnd=%s node=%s provider=%s", hwnd, rootNodeID, root.Source.Provider)
	} else {
		result.Details = details
		log.Printf("uia.viewer inspect_details_ok hwnd=%s node=%s properties=%d patterns=%d", hwnd, rootNodeID, len(details.Properties), len(details.Patterns))
	}
	c.mu.Lock()
	if c.accPathCaptureEnabled && strings.TrimSpace(result.Details.ACCPath) != "" {
		c.lastACCPath = result.Details.ACCPath
		c.statusText = "Path: " + result.Details.ACCPath
	}
	c.mu.Unlock()

	log.Printf("uia.viewer inspect_children_start hwnd=%s node=%s", hwnd, rootNodeID)
	childrenResp, childErr := c.ExpandNode(rootNodeID)
	if childErr != nil {
		result.ChildLoadErr = fmt.Errorf("load root children node=%s hwnd=%s: %w", rootNodeID, hwnd, childErr)
		log.Printf("uia.viewer inspect_children_err hwnd=%s node=%s err=%v", hwnd, rootNodeID, childErr)
	} else {
		result.Children = childrenResp.Children
		log.Printf("uia.viewer inspect_children_ok hwnd=%s node=%s children=%d", hwnd, rootNodeID, len(result.Children))
	}

	log.Printf("uia.viewer inspect_select_start hwnd=%s node=%s", hwnd, rootNodeID)
	if _, err := c.service.SelectNode(c.runtimeContext(), inspect.SelectNodeRequest{NodeID: rootNodeID}); err != nil {
		result.SelectErr = fmt.Errorf("select root node node=%s hwnd=%s: %w", rootNodeID, hwnd, err)
		log.Printf("uia.viewer inspect_select_err hwnd=%s node=%s err=%v", hwnd, rootNodeID, err)
	} else {
		log.Printf("uia.viewer inspect_select_ok hwnd=%s node=%s", hwnd, rootNodeID)
	}
	log.Printf("uia.viewer inspect_highlight_start hwnd=%s node=%s", hwnd, rootNodeID)
	if _, err := c.service.HighlightNode(c.runtimeContext(), inspect.HighlightNodeRequest{NodeID: rootNodeID}); err != nil {
		result.HighlightErr = fmt.Errorf("highlight root node node=%s hwnd=%s: %w", rootNodeID, hwnd, err)
		log.Printf("uia.viewer inspect_highlight_err hwnd=%s node=%s err=%v", hwnd, rootNodeID, err)
	} else {
		log.Printf("uia.viewer inspect_highlight_ok hwnd=%s node=%s", hwnd, rootNodeID)
		c.mu.Lock()
		c.selectedNodeID = rootNodeID
		c.mu.Unlock()
	}
	log.Printf("uia.viewer select_window_end hwnd=%s mode=%s provider=%s active_mode=%s fallback=%t err=nil", hwnd, mode, result.Root.Source.Provider, result.Root.State.ActiveMode, result.Root.State.FallbackUsed)
	return result, nil
}

func synthesizeRootDetails(root inspect.GetTreeRootResponse, detailsErr error) inspect.GetNodeDetailsResponse {
	status := "partial load: using root fallback details"
	if detailsErr != nil {
		status = fmt.Sprintf("partial load: details unavailable (%v); using root fallback details", detailsErr)
	}
	return inspect.GetNodeDetailsResponse{
		WindowInfo: inspect.WindowInfoDTO{
			Title: root.Root.Name,
			HWND:  root.Root.HWND,
			Class: firstNonEmpty(root.Root.ClassName, root.Root.DebugMeta.ClassName),
		},
		Element: inspect.ElementPropertiesDTO{
			NodeID:               root.Root.NodeID,
			HWND:                 firstNonEmpty(root.Root.HWND, root.Root.DebugMeta.HWND),
			Name:                 root.Root.Name,
			ControlType:          root.Root.ControlType,
			LocalizedControlType: root.Root.LocalizedControlType,
			AutomationID:         root.Root.DebugMeta.AutomationID,
			Status:               "partial",
		},
		StatusText: status,
		Source:     root.Source,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (c *Controller) getTreeRootWithRetry(hwnd string, mode inspect.InspectMode) (inspect.GetTreeRootResponse, []error, error) {
	delays := []time.Duration{0, 50 * time.Millisecond, 150 * time.Millisecond}
	warnings := make([]error, 0, len(delays)-1)
	var lastErr error
	for i, delay := range delays {
		if i > 0 {
			log.Printf("uia.viewer root_retry_delay hwnd=%s mode=%s attempt=%d delay_ms=%d", hwnd, mode, i+1, delay.Milliseconds())
			time.Sleep(delay)
		}
		log.Printf("uia.viewer root_attempt_start hwnd=%s mode=%s attempt=%d", hwnd, mode, i+1)
		root, err := c.service.GetTreeRoot(c.runtimeContext(), inspect.GetTreeRootRequest{HWND: hwnd, Refresh: true, Mode: mode})
		if err == nil {
			log.Printf("uia.viewer root_attempt_ok hwnd=%s mode=%s attempt=%d provider=%s active_mode=%s fallback=%t", hwnd, mode, i+1, root.Source.Provider, root.State.ActiveMode, root.State.FallbackUsed)
			return root, warnings, nil
		}
		log.Printf("uia.viewer root_attempt_err hwnd=%s mode=%s attempt=%d err=%v", hwnd, mode, i+1, err)
		lastErr = err
		if !isTransientInspectError(err) || i == len(delays)-1 {
			return inspect.GetTreeRootResponse{}, warnings, err
		}
		warnings = append(warnings, fmt.Errorf("attempt %d failed (%v): retrying transient root resolution", i+1, err))
	}
	return inspect.GetTreeRootResponse{}, warnings, lastErr
}

func isTransientInspectError(err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, inspect.ErrTransientFailure),
		errors.Is(err, inspect.ErrProviderTransientFailure),
		errors.Is(err, inspect.ErrStaleCache):
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "transient") ||
		strings.Contains(msg, "stale") ||
		strings.Contains(msg, "element not available") ||
		strings.Contains(msg, "nil element")
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
func (c *Controller) InvokeSetValue() (inspect.InvokePatternResponse, bool, error) {
	c.mu.Lock()
	nodeID := c.selectedNodeID
	c.mu.Unlock()
	if c.dialogs == nil {
		return inspect.InvokePatternResponse{}, false, errors.New("setValue dialog unavailable")
	}
	value, ok, err := c.dialogs.PromptSetValue("")
	if err != nil {
		return inspect.InvokePatternResponse{}, false, err
	}
	if !ok {
		return inspect.InvokePatternResponse{}, false, nil
	}
	resp, err := c.InvokePattern(nodeID, "setValue", map[string]any{"value": value})
	return resp, true, err
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
		c.statusText = "Click on path to copy to Clipboard"
	} else {
		c.statusText = "Click here to enable Acc path capturing (can't be used with UIA!)"
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
	if enabled && path != "" {
		if c.clipboard != nil {
			_ = c.clipboard.CopyText(path)
		}
		c.statusText = "Path: " + path
		c.mu.Unlock()
		return StatusUpdate{Text: "Path: " + path, CaptureEnabled: true, HasLastACCPath: true, LastACCPathCopied: true}
	}
	c.mu.Unlock()
	enabled = c.ToggleAccPathCapture()
	c.mu.Lock()
	defer c.mu.Unlock()
	path = strings.TrimSpace(c.lastACCPath)
	hasPath := path != ""
	text := c.statusText
	if enabled && hasPath {
		text = "Path: " + path
		c.statusText = text
	}
	return StatusUpdate{Text: text, CaptureEnabled: enabled, HasLastACCPath: hasPath}
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
