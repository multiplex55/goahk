package goahk

import (
	"context"
	"time"

	"goahk/internal/actions"
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
	Paste(text string) error
}

type WindowAPI interface {
	Active() (window.Info, error)
	List() ([]window.Info, error)
	Activate(matcher string) error
	Bounds(hwnd window.HWND) (window.Rect, error)
	Move(hwnd window.HWND, x, y int) error
	Resize(hwnd window.HWND, width, height int) error
	Title() (string, error)
}

type ProcessAPI interface {
	Launch(executable string, args ...string) error
	Open(target string) error
}

type RuntimeAPI interface {
	Stop()
	Sleep(duration time.Duration)
}

type Context struct {
	Clipboard ClipboardAPI
	Input     InputAPI
	Window    WindowAPI
	Process   ProcessAPI
	Runtime   RuntimeAPI
	Vars      map[string]string
	AppState  StateStore

	actionCtx *actions.ActionContext
}

func (c *Context) Context() context.Context {
	if c == nil || c.actionCtx == nil {
		return context.Background()
	}
	return c.actionCtx.Context
}

func (c *Context) BindingID() string {
	if c == nil || c.actionCtx == nil {
		return ""
	}
	return c.actionCtx.BindingID
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
