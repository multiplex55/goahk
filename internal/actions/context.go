package actions

import (
	"context"
	"time"

	"goahk/internal/input"
	"goahk/internal/process"
	"goahk/internal/services/messagebox"
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
	ActiveWindowTitle func(context.Context) (string, error)
	WindowList        func(context.Context) ([]window.Info, error)
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
