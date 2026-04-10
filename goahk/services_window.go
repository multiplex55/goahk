package goahk

import (
	"fmt"

	"goahk/internal/window"
)

type windowService struct {
	ctx *Context
}

func (s windowService) Active() (window.Info, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil {
		return window.Info{}, fmt.Errorf("window active: %w", ErrWindowServiceUnavailable)
	}
	if s.ctx.actionCtx.Services.WindowActive != nil {
		return s.ctx.actionCtx.Services.WindowActive(s.ctx.Context())
	}
	items, err := s.List()
	if err != nil {
		return window.Info{}, err
	}
	for _, item := range items {
		if item.Active {
			return item, nil
		}
	}
	return window.Info{}, fmt.Errorf("no active window")
}

func (s windowService) List() ([]window.Info, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowList == nil {
		return nil, fmt.Errorf("window list: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowList(s.ctx.Context())
}

func (s windowService) Activate(matcher string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowActivate == nil {
		return fmt.Errorf("window activate: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowActivate(s.ctx.Context(), matcher)
}

func (s windowService) Bounds(hwnd window.HWND) (window.Rect, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowBounds == nil {
		return window.Rect{}, fmt.Errorf("window bounds: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowBounds(s.ctx.Context(), hwnd)
}

func (s windowService) Move(hwnd window.HWND, x, y int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowMove == nil {
		return fmt.Errorf("window move: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowMove(s.ctx.Context(), hwnd, x, y)
}

func (s windowService) Resize(hwnd window.HWND, width, height int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowResize == nil {
		return fmt.Errorf("window resize: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowResize(s.ctx.Context(), hwnd, width, height)
}

func (s windowService) Title() (string, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.ActiveWindowTitle == nil {
		return "", fmt.Errorf("window title: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.ActiveWindowTitle(s.ctx.Context())
}
