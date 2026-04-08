package actions

import (
	"context"
	"time"

	"goahk/internal/clipboard"
	"goahk/internal/input"
	"goahk/internal/process"
	"goahk/internal/services/messagebox"
	"goahk/internal/uia"
)

type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

type NoopLogger struct{}

func (NoopLogger) Info(string, map[string]any)  {}
func (NoopLogger) Error(string, map[string]any) {}

type Services struct {
	MessageBox        messagebox.Service
	Clipboard         clipboard.Service
	Process           process.Service
	WindowActivate    func(context.Context, string) error
	ActiveWindowTitle func(context.Context) (string, error)
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
	return c
}

func (c ActionContext) requestStop() {
	if c.stopRequested == nil {
		c.stopRequested = new(bool)
	}
	*c.stopRequested = true
}

func (c ActionContext) isStopRequested() bool {
	return c.stopRequested != nil && *c.stopRequested
}
