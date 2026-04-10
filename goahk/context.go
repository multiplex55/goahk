package goahk

import (
	"context"
	"time"

	"goahk/internal/actions"
	"goahk/internal/input"
	"goahk/internal/process"
	"goahk/internal/window"
)

type ClipboardAPI interface {
	ReadText() (string, error)
	WriteText(text string) error
	AppendText(text string) error
	PrependText(text string) error
}

type InputAPI interface {
	SendText(text string) error
	SendKeys(keys ...string) error
	SendChord(keys ...string) error
	MouseMoveAbsolute(x, y int) error
	MouseMoveRelative(dx, dy int) error
	MousePosition() (input.MousePosition, error)
	MouseButtonDown(button string) error
	MouseButtonUp(button string) error
	MouseClick(button string) error
	MouseDoubleClick(button string) error
	MouseWheel(delta int) error
	MouseDrag(button string, startX, startY, endX, endY int) error
	Paste(text string) error
}

type WindowAPI interface {
	Active() (window.Info, error)
	List() ([]window.Info, error)
	Activate(matcher string) error
	ActivateMatch(matcher window.Matcher) error
	Bounds(hwnd window.HWND) (window.Rect, error)
	Move(hwnd window.HWND, x, y int) error
	MoveBy(hwnd window.HWND, dx, dy int) error
	Resize(hwnd window.HWND, width, height int) error
	ResizeBy(hwnd window.HWND, dw, dh int) error
	Center(hwnd window.HWND) error
	Minimize(hwnd window.HWND) error
	Maximize(hwnd window.HWND) error
	Restore(hwnd window.HWND) error
	Title() (string, error)
}

type ProcessAPI interface {
	Launch(executable string, args ...string) error
	Open(target string) error
}

type RuntimeAPI interface {
	Stop()
	Sleep(duration time.Duration) bool
}

// AutomationAPI aliases UIAService for ergonomic naming in callbacks.
type AutomationAPI = UIAService

// CallbackLogger is the callback-facing logging contract.
//
// It mirrors runtime logging and is safe to call from callback goroutines.
type CallbackLogger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

type Context struct {
	Clipboard  ClipboardAPI
	Input      InputAPI
	Window     WindowAPI
	Process    ProcessAPI
	UIA        UIAService
	Automation AutomationAPI
	Runtime    RuntimeAPI
	Vars       map[string]string
	AppState   StateStore

	actionCtx *actions.ActionContext
}

// Context returns the underlying callback context.
//
// Prefer Err and Sleep for most callback cancellation checks. Use Context()
// only for low-level integrations that require a context.Context.
func (c *Context) Context() context.Context {
	if c == nil || c.actionCtx == nil {
		return context.Background()
	}
	return c.actionCtx.Context
}

// Err reports callback cancellation/stop state.
//
// It returns nil while work may continue, and returns context cancellation
// errors (for example context.Canceled) after stop/cancel is observed.
// Err is safe to call concurrently.
func (c *Context) Err() error {
	return actions.NewCallbackContext(c.actionCtx).Err()
}

// Sleep waits for d or until callback cancellation/stop is observed.
//
// It returns true when the timer elapses normally, and false when the callback
// context is canceled before d completes.
// Sleep is safe to call concurrently.
func (c *Context) Sleep(d time.Duration) bool {
	return actions.NewCallbackContext(c.actionCtx).Sleep(d)
}

// Logger returns the callback logger configured by the runtime.
//
// If no logger is configured, a no-op logger is returned.
// The returned logger is safe for callback use across goroutines.
func (c *Context) Logger() CallbackLogger {
	return actions.NewCallbackContext(c.actionCtx).Log()
}

// Log emits an info-level callback log with optional key/value pairs.
//
// Odd trailing values are ignored. Non-string keys are skipped.
func (c *Context) Log(msg string, keyvals ...any) {
	fields := map[string]any{}
	for i := 0; i+1 < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		fields[key] = keyvals[i+1]
	}
	c.Logger().Info(msg, fields)
}

func (c *Context) BindingID() string {
	if c == nil || c.actionCtx == nil {
		return ""
	}
	return c.actionCtx.BindingID
}

// Binding returns the trigger binding identifier for this callback run.
func (c *Context) Binding() string { return c.BindingID() }

// Trigger returns the trigger text (hotkey chord) for this callback run.
func (c *Context) Trigger() string {
	if c == nil || c.actionCtx == nil {
		return ""
	}
	return c.actionCtx.TriggerText
}

func (c *Context) Metadata() map[string]string {
	if c == nil {
		return map[string]string{}
	}
	if c.Vars == nil {
		return map[string]string{}
	}
	return c.Vars
}

func (c *Context) Stop() {
	if c == nil || c.Runtime == nil {
		return
	}
	c.Runtime.Stop()
}

func newContext(actionCtx *actions.ActionContext, state StateStore) *Context {
	vars := copyVars(nil)
	if actionCtx != nil {
		vars = copyVars(actionCtx.Metadata)
	}
	ctx := &Context{actionCtx: actionCtx, Vars: vars, AppState: state}
	ctx.Clipboard = clipboardService{ctx: ctx}
	ctx.Input = inputService{ctx: ctx}
	ctx.Window = windowService{ctx: ctx}
	ctx.Process = processService{ctx: ctx}
	ctx.UIA = uiaService{ctx: ctx}
	ctx.Automation = ctx.UIA
	ctx.Runtime = runtimeService{ctx: ctx}
	return ctx
}

func copyVars(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func syncVarsToActionContext(c *Context) {
	if c == nil || c.actionCtx == nil {
		return
	}
	c.actionCtx.Metadata = copyVars(c.Vars)
}

func processRequestForOpen(target string) process.Request {
	return process.Request{OpenTarget: target}
}
