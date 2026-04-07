package actions

import (
	"context"
	"errors"
	"testing"
	"time"

	"goahk/internal/uia"
)

type fakeUIAService struct {
	invokeErr error
	findErr   error
	findDiag  uia.ActionDiagnostics
}

func (f fakeUIAService) Find(context.Context, uia.Selector, time.Duration, time.Duration) (uia.Element, uia.ActionDiagnostics, error) {
	return uia.Element{ID: "x"}, f.findDiag, f.findErr
}
func (f fakeUIAService) Invoke(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{MissingPattern: "Invoke", SupportedPatterns: []string{"Value"}, RetryCount: 2}, f.invokeErr
}
func (f fakeUIAService) ValueSet(context.Context, uia.Selector, string, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f fakeUIAService) ValueGet(context.Context, uia.Selector, time.Duration, time.Duration) (string, uia.ActionDiagnostics, error) {
	return "ok", uia.ActionDiagnostics{}, nil
}
func (f fakeUIAService) Toggle(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f fakeUIAService) Expand(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}
func (f fakeUIAService) Select(context.Context, uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error) {
	return uia.ActionDiagnostics{}, nil
}

type captureLogger struct{ fields map[string]any }

func (c *captureLogger) Info(_ string, fields map[string]any) { c.fields = fields }
func (c *captureLogger) Error(string, map[string]any)         {}

func TestUIAInvoke_UnsupportedPattern(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("uia.invoke")
	err := h(ActionContext{
		Context: context.Background(),
		Services: Services{UIA: fakeUIAService{
			invokeErr: errors.New("unsupported pattern"),
		}},
	}, Step{Name: "uia.invoke", Params: map[string]string{"selector_json": `{"automationId":"submit"}`}})
	if err == nil || err.Error() != "unsupported pattern" {
		t.Fatalf("error=%v", err)
	}
}

func TestUIAFind_TimeoutPropagationAndSuccessPayload(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("uia.find")
	logger := &captureLogger{}
	ctx := ActionContext{
		Context:  context.Background(),
		Metadata: map[string]string{},
		Logger:   logger,
		Services: Services{UIA: fakeUIAService{findDiag: uia.ActionDiagnostics{RetryCount: 3, SupportedPatterns: []string{"Invoke"}}}},
	}
	err := h(ctx, Step{Name: "uia.find", Params: map[string]string{
		"selector_json":     `{"automationId":"submit"}`,
		"timeout_ms":        "1250",
		"retry_interval_ms": "100",
		"save_as":           "found",
	}})
	if err != nil {
		t.Fatalf("error=%v", err)
	}
	if ctx.Metadata["found"] == "" {
		t.Fatalf("metadata missing payload")
	}
	if logger.fields["retry_count"] != 3 {
		t.Fatalf("logger fields=%v", logger.fields)
	}
}

func TestUIAFind_FailurePayload(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("uia.find")
	ctx := ActionContext{
		Context:  context.Background(),
		Metadata: map[string]string{},
		Services: Services{UIA: fakeUIAService{findErr: errors.New("timeout")}},
	}
	err := h(ctx, Step{Name: "uia.find", Params: map[string]string{"selector_json": `{"name":"Missing"}`}})
	if err == nil || err.Error() != "timeout" {
		t.Fatalf("error=%v", err)
	}
}
