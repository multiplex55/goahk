package goahk

import (
	"time"

	"goahk/internal/uia"
)

type UIAElement = uia.Element
type UIADiagnostics = uia.ActionDiagnostics

type UIAService interface {
	Find(selector Selector, timeout, retryInterval time.Duration) (UIAElement, UIADiagnostics, error)
	Invoke(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error)
	ValueSet(selector Selector, value string, timeout, retryInterval time.Duration) (UIADiagnostics, error)
	ValueGet(selector Selector, timeout, retryInterval time.Duration) (string, UIADiagnostics, error)
	Toggle(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error)
	Expand(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error)
	Select(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error)
}

type uiaService struct {
	ctx *Context
}

func (s uiaService) Find(selector Selector, timeout, retryInterval time.Duration) (UIAElement, UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return UIAElement{}, UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.Find(s.ctx.Context(), selector.toInternal(), timeout, retryInterval)
}

func (s uiaService) Invoke(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.Invoke(s.ctx.Context(), selector.toInternal(), timeout, retryInterval)
}

func (s uiaService) ValueSet(selector Selector, value string, timeout, retryInterval time.Duration) (UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.ValueSet(s.ctx.Context(), selector.toInternal(), value, timeout, retryInterval)
}

func (s uiaService) ValueGet(selector Selector, timeout, retryInterval time.Duration) (string, UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return "", UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.ValueGet(s.ctx.Context(), selector.toInternal(), timeout, retryInterval)
}

func (s uiaService) Toggle(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.Toggle(s.ctx.Context(), selector.toInternal(), timeout, retryInterval)
}

func (s uiaService) Expand(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.Expand(s.ctx.Context(), selector.toInternal(), timeout, retryInterval)
}

func (s uiaService) Select(selector Selector, timeout, retryInterval time.Duration) (UIADiagnostics, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.UIA == nil {
		return UIADiagnostics{}, nil
	}
	return s.ctx.actionCtx.Services.UIA.Select(s.ctx.Context(), selector.toInternal(), timeout, retryInterval)
}
