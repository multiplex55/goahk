package goahk

import (
	"context"

	"goahk/internal/actions"
	"goahk/internal/program"
)

type ActionFunc func(*Context) error

type Context struct {
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
	if c == nil || c.actionCtx == nil {
		return map[string]string{}
	}
	return c.actionCtx.Metadata
}

func (c *Context) Stop() {
	if c == nil || c.actionCtx == nil {
		return
	}
	actions.RequestRuntimeStop(c.actionCtx, "runtime.stop")
}

type callbackStep struct {
	fn ActionFunc
}

func Func(fn ActionFunc) callbackStep {
	return callbackStep{fn: fn}
}

func (c callbackStep) stepSpec() program.StepSpec {
	return program.StepSpec{Action: callbackActionPlaceholder, Params: map[string]any{}}
}
