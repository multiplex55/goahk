package goahk

import (
	"fmt"
	"strings"

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

func (s windowService) ActivateMatch(matcher window.Matcher) error {
	return s.Activate(encodeMatcher(matcher))
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

func (s windowService) MoveBy(hwnd window.HWND, dx, dy int) error {
	bounds, err := s.Bounds(hwnd)
	if err != nil {
		return err
	}
	x, y := window.TranslateMoveBy(bounds, dx, dy)
	return s.Move(hwnd, x, y)
}

func (s windowService) Resize(hwnd window.HWND, width, height int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowResize == nil {
		return fmt.Errorf("window resize: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowResize(s.ctx.Context(), hwnd, width, height)
}

func (s windowService) ResizeBy(hwnd window.HWND, dw, dh int) error {
	bounds, err := s.Bounds(hwnd)
	if err != nil {
		return err
	}
	width, height, err := window.TranslateResizeBy(bounds, dw, dh)
	if err != nil {
		return fmt.Errorf("window resize_by: %w", err)
	}
	return s.Resize(hwnd, width, height)
}

func (s windowService) Center(hwnd window.HWND) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowWorkArea == nil {
		return fmt.Errorf("window center: %w", ErrWindowServiceUnavailable)
	}
	bounds, err := s.Bounds(hwnd)
	if err != nil {
		return err
	}
	workArea, err := s.ctx.actionCtx.Services.WindowWorkArea(s.ctx.Context(), hwnd)
	if err != nil {
		return fmt.Errorf("window center: %w", err)
	}
	x, y := window.CenterPosition(bounds, workArea)
	return s.Move(hwnd, x, y)
}

func (s windowService) Minimize(hwnd window.HWND) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowMinimize == nil {
		return fmt.Errorf("window minimize: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowMinimize(s.ctx.Context(), hwnd)
}

func (s windowService) Maximize(hwnd window.HWND) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowMaximize == nil {
		return fmt.Errorf("window maximize: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowMaximize(s.ctx.Context(), hwnd)
}

func (s windowService) Restore(hwnd window.HWND) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.WindowRestore == nil {
		return fmt.Errorf("window restore: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.WindowRestore(s.ctx.Context(), hwnd)
}

func (s windowService) Title() (string, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.ActiveWindowTitle == nil {
		return "", fmt.Errorf("window title: %w", ErrWindowServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.ActiveWindowTitle(s.ctx.Context())
}

func encodeMatcher(m window.Matcher) string {
	parts := []string{}
	if value := strings.TrimSpace(m.TitleContains); value != "" {
		parts = append(parts, "title:"+value)
	}
	if value := strings.TrimSpace(m.TitleExact); value != "" {
		parts = append(parts, "title_exact:"+value)
	}
	if value := strings.TrimSpace(m.TitleRegex); value != "" {
		parts = append(parts, "title_regex:"+value)
	}
	if value := strings.TrimSpace(m.ClassName); value != "" {
		parts = append(parts, "class:"+value)
	}
	if value := strings.TrimSpace(m.ExeName); value != "" {
		parts = append(parts, "exe:"+value)
	}
	if m.ActiveOnly {
		parts = append(parts, "active:true")
	}
	return strings.Join(parts, ",")
}
