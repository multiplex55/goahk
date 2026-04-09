package actions

import (
	"context"
	"time"

	"goahk/internal/input"
	"goahk/internal/process"
	"goahk/internal/services/messagebox"
	"goahk/internal/shell/folders"
	"goahk/internal/uia"
	"goahk/internal/window"
)

type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

type NoopLogger struct{}

func (NoopLogger) Info(string, map[string]any)  {}
func (NoopLogger) Error(string, map[string]any) {}

type MessageBoxService interface {
	Show(context.Context, messagebox.Request) error
}

type ClipboardService interface {
	ReadText(context.Context) (string, error)
	WriteText(context.Context, string) error
	AppendText(context.Context, string) error
	PrependText(context.Context, string) error
}

type ProcessService interface {
	Launch(context.Context, process.Request) error
}

type Services struct {
	MessageBox        MessageBoxService
	Clipboard         ClipboardService
	Process           ProcessService
	WindowActivate    func(context.Context, string) error
	WindowActive      func(context.Context) (window.Info, error)
	ActiveWindowTitle func(context.Context) (string, error)
	WindowList        func(context.Context) ([]window.Info, error)
	WindowBounds      func(context.Context, window.HWND) (window.Rect, error)
	WindowMove        func(context.Context, window.HWND, int, int) error
	WindowResize      func(context.Context, window.HWND, int, int) error
	WindowMinimize    func(context.Context, window.HWND) error
	WindowMaximize    func(context.Context, window.HWND) error
	WindowRestore     func(context.Context, window.HWND) error
	FolderList        func(context.Context) ([]folders.FolderInfo, error)
	Input             input.Service
	UIA               uia.ActionService
}

type ActionContext struct {
	Context     context.Context
	Logger      Logger
	Services    Services
	Timeout     time.Duration
	Metadata    map[string]string
	BindingID   string
	TriggerText string
	Stop        func(string)

	stopRequested *bool
	stopNotified  *bool
}

type CallbackContext interface {
	Context() context.Context
	Done() <-chan struct{}
	Err() error
	IsCancelled() bool
	Stopped() bool
	Sleep(time.Duration) bool
	StopRuntime(reason string)
	BindingID() string
	TriggerText() string
	Window() Services
	Input() input.Service
	Clipboard() ClipboardService
	Log() Logger
	StateBag() map[string]string
}

type callbackContext struct {
	actionCtx *ActionContext
}

func NewCallbackContext(actionCtx *ActionContext) CallbackContext {
	return callbackContext{actionCtx: actionCtx}
}

func (c callbackContext) Context() context.Context {
	if c.actionCtx == nil {
		return context.Background()
	}
	return baseContext(c.actionCtx.Context)
}

func (c callbackContext) Done() <-chan struct{} { return c.Context().Done() }
func (c callbackContext) Err() error            { return c.Context().Err() }
func (c callbackContext) IsCancelled() bool     { return c.Err() != nil }
func (c callbackContext) Stopped() bool {
	return c.actionCtx != nil && c.actionCtx.isStopRequested()
}

func (c callbackContext) Sleep(d time.Duration) bool {
	if d <= 0 {
		return true
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-c.Done():
		return false
	}
}

func (c callbackContext) StopRuntime(reason string) {
	if c.actionCtx == nil {
		return
	}
	RequestRuntimeStop(c.actionCtx, reason)
}

func (c callbackContext) BindingID() string {
	if c.actionCtx == nil {
		return ""
	}
	return c.actionCtx.BindingID
}
func (c callbackContext) TriggerText() string {
	if c.actionCtx == nil {
		return ""
	}
	return c.actionCtx.TriggerText
}
func (c callbackContext) Window() Services {
	if c.actionCtx == nil {
		return Services{}
	}
	return c.actionCtx.Services
}
func (c callbackContext) Input() input.Service {
	if c.actionCtx == nil {
		return nil
	}
	return c.actionCtx.Services.Input
}
func (c callbackContext) Clipboard() ClipboardService {
	if c.actionCtx == nil {
		return nil
	}
	return c.actionCtx.Services.Clipboard
}
func (c callbackContext) Log() Logger {
	if c.actionCtx == nil || c.actionCtx.Logger == nil {
		return NoopLogger{}
	}
	return c.actionCtx.Logger
}
func (c callbackContext) StateBag() map[string]string {
	if c.actionCtx == nil {
		return map[string]string{}
	}
	if c.actionCtx.Metadata == nil {
		c.actionCtx.Metadata = map[string]string{}
	}
	return c.actionCtx.Metadata
}

func (c ActionContext) withContext(ctx context.Context) ActionContext {
	c.Context = ctx
	if c.Logger == nil {
		c.Logger = NoopLogger{}
	}
	if c.Metadata == nil {
		c.Metadata = map[string]string{}
	}
	if c.stopRequested == nil {
		c.stopRequested = new(bool)
	}
	if c.stopNotified == nil {
		c.stopNotified = new(bool)
	}
	return c
}

func (c ActionContext) requestStop() bool {
	if c.stopRequested == nil {
		c.stopRequested = new(bool)
	}
	if c.stopNotified == nil {
		c.stopNotified = new(bool)
	}
	if *c.stopRequested {
		return false
	}
	*c.stopRequested = true
	return true
}

func (c ActionContext) isStopRequested() bool {
	return c.stopRequested != nil && *c.stopRequested
}

func RequestRuntimeStop(ctx *ActionContext, reason string) {
	if ctx == nil {
		return
	}
	if ctx.stopRequested == nil {
		ctx.stopRequested = new(bool)
	}
	if ctx.stopNotified == nil {
		ctx.stopNotified = new(bool)
	}
	first := !*ctx.stopRequested
	*ctx.stopRequested = true
	if !first || *ctx.stopNotified {
		return
	}
	*ctx.stopNotified = true
	if ctx.Stop != nil {
		ctx.Stop(reason)
	}
}
