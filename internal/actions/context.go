package actions

import (
	"context"
	"time"
)

type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

type NoopLogger struct{}

func (NoopLogger) Info(string, map[string]any)  {}
func (NoopLogger) Error(string, map[string]any) {}

type Services struct {
	MessageBox     func(context.Context, string, string) error
	ClipboardWrite func(context.Context, string) error
	ProcessLaunch  func(context.Context, string, []string) error
	WindowActivate func(context.Context, string) error
	InputSendText  func(context.Context, string) error
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
