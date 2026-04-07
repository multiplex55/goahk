package boundary

import (
	"context"
	"io"
	"time"

	"goahk/internal/config"
	"goahk/internal/hotkey"
	"goahk/internal/uia"
	"goahk/internal/window"
)

// Clock abstracts wall-clock operations for deterministic testing.
type Clock interface {
	Now() time.Time
	Sleep(context.Context, time.Duration) error
}

// Logger abstracts structured logging side effects.
type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

// HotkeyProvider abstracts OS hotkey registration and event delivery.
type HotkeyProvider interface {
	RegisterHotkeys(context.Context, []config.HotkeyBinding) (io.Closer, error)
	Events() <-chan hotkey.TriggerEvent
}

// WindowProvider abstracts top-level window queries.
type WindowProvider interface {
	Active(context.Context) (window.Info, error)
	List(context.Context) ([]window.Info, error)
}

// ClipboardService abstracts clipboard operations.
type ClipboardService interface {
	ReadText(context.Context) (string, error)
	WriteText(context.Context, string) error
	AppendText(context.Context, string) error
	PrependText(context.Context, string) error
}

// UIATreeProvider abstracts UI Automation inspection operations.
type UIATreeProvider interface {
	Focused(context.Context) (uia.Element, error)
	UnderCursor(context.Context) (uia.Element, error)
	ActiveWindowTree(context.Context, int) (*uia.Node, error)
}
