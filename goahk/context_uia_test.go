package goahk

import (
	"context"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/uia"
)

type fakeContextUIAService struct {
	lastSelector uia.Selector
	findCalled   bool
}

func (f *fakeContextUIAService) Find(context.Context, uia.Selector, time.Duration, time.Duration) (uia.Element, uia.ActionDiagnostics, error) {
	f.findCalled = true
	return uia.Element{ID: "target"}, uia.ActionDiagnostics{RetryCount: 2}, nil
}

func (f *fakeContextUIAService) Invoke(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f *fakeContextUIAService) ValueSet(context.Context, uia.Selector, string, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f *fakeContextUIAService) ValueGet(context.Context, uia.Selector, time.Duration, time.Duration) (string, uia.ActionDiagnostics, error) {
	return "", uia.ActionDiagnostics{}, nil
}
func (f *fakeContextUIAService) Toggle(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f *fakeContextUIAService) Expand(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f *fakeContextUIAService) Select(_ context.Context, sel uia.Selector, _ time.Duration, _ time.Duration) (uia.ActionDiagnostics, error) {
	f.lastSelector = sel
	return uia.ActionDiagnostics{}, nil
}

func TestContext_UIAServiceExposed(t *testing.T) {
	svc := &fakeContextUIAService{}
	ctx := newContext(&actions.ActionContext{Context: context.Background(), Services: actions.Services{UIA: svc}}, nil)
	selector := SelectByAutomationID("submit").WithAncestors(SelectByName("Main").WithControlType("Window"))
	el, diag, err := ctx.UIA.Find(selector, time.Second, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if !svc.findCalled || el.ID != "target" || diag.RetryCount != 2 {
		t.Fatalf("unexpected find result: called=%t element=%+v diag=%+v", svc.findCalled, el, diag)
	}
	if _, err := ctx.Automation.Select(selector, time.Second, 10*time.Millisecond); err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if svc.lastSelector.AutomationID != "submit" || len(svc.lastSelector.Ancestors) != 1 || svc.lastSelector.Ancestors[0].Name != "Main" {
		t.Fatalf("selector mapping mismatch: %+v", svc.lastSelector)
	}
}

func TestSelectorBuilderHelpers(t *testing.T) {
	sel := SelectByName("Save").WithControlType("Button").WithAutomationID("save-btn")
	internal := sel.toInternal()
	if internal.Name != "Save" || internal.ControlType != "Button" || internal.AutomationID != "save-btn" {
		t.Fatalf("unexpected selector conversion: %+v", internal)
	}
}
