package actions

import (
	"context"
	"time"

	"goahk/internal/clipboard"
	"goahk/internal/input"
)

type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

type NoopLogger struct{}

func (NoopLogger) Info(string, map[string]any)  {}
func (NoopLogger) Error(string, map[string]any) {}

type Services struct {
	MessageBox        func(context.Context, string, string) error
	Clipboard         clipboard.Service
	ProcessLaunch     func(context.Context, string, []string) error
	WindowActivate    func(context.Context, string) error
	ActiveWindowTitle func(context.Context) (string, error)
	Input             input.Service
}

type ActionContext struct {
	Context     context.Context
	Logger      Logger
	Services    Services
	Timeout     time.Duration
	Metadata    map[string]string
	BindingID   string
	TriggerText string
}

func (c ActionContext) withContext(ctx context.Context) ActionContext {
	c.Context = ctx
	if c.Logger == nil {
		c.Logger = NoopLogger{}
	}
	if c.Metadata == nil {
		c.Metadata = map[string]string{}
	}
	return c
}
