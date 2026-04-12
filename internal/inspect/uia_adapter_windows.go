//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"strings"
)

var errUIAElementNotAvailable = errors.New("inspect: uia element not available")
var errUIANilElement = errors.New("inspect: uia nil element")

type windowsUIADeps interface {
	ResolveWindowRoot(context.Context, string) (*uiaElement, error)
	GetFocusedElement(context.Context) (*uiaElement, error)
	GetCursorPosition(context.Context) (int, int, error)
	ElementFromPoint(context.Context, int, int) (*uiaElement, error)
	GetElementByRef(context.Context, string) (*uiaElement, error)
	GetParent(context.Context, string) (*uiaElement, error)
	GetChildren(context.Context, string) ([]*uiaElement, error)
	GetChildCount(context.Context, string) (int, bool, error)
	Invoke(context.Context, string) error
	Select(context.Context, string) error
	SetValue(context.Context, string, string) error
	DoDefaultAction(context.Context, string) error
	Toggle(context.Context, string) error
	Expand(context.Context, string) error
	Collapse(context.Context, string) error
}

type windowsUIAAdapter struct{ deps windowsUIADeps }

func newUIAAdapter(deps windowsUIADeps) uiaAdapter {
	if deps == nil {
		return newUnsupportedUIAAdapter()
	}
	return &windowsUIAAdapter{deps: deps}
}

func (a *windowsUIAAdapter) ResolveWindowRoot(ctx context.Context, hwnd string) (*uiaElement, error) {
	el, err := a.deps.ResolveWindowRoot(ctx, hwnd)
	if err != nil {
		return nil, mapUIAError(err)
	}
	if el == nil {
		return nil, errNilElementReference
	}
	return a.normalizeElement(el), nil
}

func (a *windowsUIAAdapter) GetFocusedElement(ctx context.Context) (*uiaElement, error) {
	el, err := a.deps.GetFocusedElement(ctx)
	if err != nil {
		return nil, mapUIAError(err)
	}
	if el == nil {
		return nil, errNilElementReference
	}
	return a.normalizeElement(el), nil
}

func (a *windowsUIAAdapter) GetCursorPosition(ctx context.Context) (int, int, error) {
	x, y, err := a.deps.GetCursorPosition(ctx)
	if err != nil {
		return 0, 0, mapUIAError(err)
	}
	return x, y, nil
}

func (a *windowsUIAAdapter) ElementFromPoint(ctx context.Context, x, y int) (*uiaElement, error) {
	el, err := a.deps.ElementFromPoint(ctx, x, y)
	if err != nil {
		return nil, mapUIAError(err)
	}
	if el == nil {
		return nil, errNilElementReference
	}
	return a.normalizeElement(el), nil
}

func (a *windowsUIAAdapter) GetElementByRef(ctx context.Context, ref string) (*uiaElement, error) {
	el, err := a.deps.GetElementByRef(ctx, ref)
	if err != nil {
		return nil, mapUIAError(err)
	}
	if el == nil {
		return nil, errNilElementReference
	}
	return a.normalizeElement(el), nil
}

func (a *windowsUIAAdapter) GetParent(ctx context.Context, ref string) (*uiaElement, error) {
	el, err := a.deps.GetParent(ctx, ref)
	if err != nil {
		return nil, mapUIAError(err)
	}
	if el == nil {
		return nil, errNilElementReference
	}
	return a.normalizeElement(el), nil
}

func (a *windowsUIAAdapter) GetChildren(ctx context.Context, ref string) ([]*uiaElement, error) {
	children, err := a.deps.GetChildren(ctx, ref)
	if err != nil {
		return nil, mapUIAError(err)
	}
	return children, nil
}

func (a *windowsUIAAdapter) GetChildCount(ctx context.Context, ref string) (int, bool, error) {
	count, ok, err := a.deps.GetChildCount(ctx, ref)
	if err != nil {
		return 0, false, mapUIAError(err)
	}
	return count, ok, nil
}

func (a *windowsUIAAdapter) Invoke(ctx context.Context, ref string) error {
	return mapUIAError(a.deps.Invoke(ctx, ref))
}
func (a *windowsUIAAdapter) Select(ctx context.Context, ref string) error {
	return mapUIAError(a.deps.Select(ctx, ref))
}
func (a *windowsUIAAdapter) SetValue(ctx context.Context, ref, value string) error {
	return mapUIAError(a.deps.SetValue(ctx, ref, value))
}
func (a *windowsUIAAdapter) DoDefaultAction(ctx context.Context, ref string) error {
	return mapUIAError(a.deps.DoDefaultAction(ctx, ref))
}
func (a *windowsUIAAdapter) Toggle(ctx context.Context, ref string) error {
	return mapUIAError(a.deps.Toggle(ctx, ref))
}
func (a *windowsUIAAdapter) Expand(ctx context.Context, ref string) error {
	return mapUIAError(a.deps.Expand(ctx, ref))
}
func (a *windowsUIAAdapter) Collapse(ctx context.Context, ref string) error {
	return mapUIAError(a.deps.Collapse(ctx, ref))
}

func (a *windowsUIAAdapter) normalizeElement(el *uiaElement) *uiaElement {
	if el == nil {
		return nil
	}
	if el.UnsupportedProps == nil {
		el.UnsupportedProps = map[string]bool{}
	}
	el.RuntimeID = strings.TrimSpace(el.RuntimeID)
	el.Name = strings.TrimSpace(el.Name)
	el.LocalizedControlType = strings.TrimSpace(el.LocalizedControlType)
	el.ControlType = strings.TrimSpace(el.ControlType)
	el.AutomationID = strings.TrimSpace(el.AutomationID)
	el.ClassName = strings.TrimSpace(el.ClassName)
	el.FrameworkID = strings.TrimSpace(el.FrameworkID)
	el.HWND = strings.TrimSpace(el.HWND)
	return el
}

func mapUIAError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, errUIANilElement):
		return errNilElementReference
	case errors.Is(err, errUIAElementNotAvailable):
		return errStaleElementReference
	default:
		return err
	}
}
