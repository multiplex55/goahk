package actions

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type capture struct {
	order    []string
	ctxValue string
	deadline bool
}

func TestExecutor_OrderingAndShortCircuit(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("test.one", func(_ ActionContext, _ Step) error { return nil })
	_ = r.Register("test.fail", func(_ ActionContext, _ Step) error { return errors.New("boom") })
	_ = r.Register("test.never", func(_ ActionContext, _ Step) error { t.Fatal("should not run"); return nil })

	ex := NewExecutor(r)
	res := ex.Execute(ActionContext{Context: context.Background()}, Plan{
		{Name: "test.one"},
		{Name: "test.fail"},
		{Name: "test.never"},
	})
	if res.Success {
		t.Fatal("expected overall failure")
	}
	if len(res.Steps) != 2 {
		t.Fatalf("steps=%d want 2", len(res.Steps))
	}
	if res.Steps[0].Status != StepStatusSuccess || res.Steps[1].Status != StepStatusFailed {
		t.Fatalf("unexpected statuses: %#v", res.Steps)
	}
}

func TestExecutor_TimeoutAndMetadataPropagation(t *testing.T) {
	r := NewRegistry()
	cap := capture{}
	_ = r.Register("test.capture", func(ctx ActionContext, _ Step) error {
		cap.order = append(cap.order, "test.capture")
		cap.ctxValue = ctx.Metadata["trace"]
		_, cap.deadline = ctx.Context.Deadline()
		return nil
	})
	ex := NewExecutor(r)
	res := ex.Execute(ActionContext{Context: context.Background(), Timeout: 50 * time.Millisecond, Metadata: map[string]string{"trace": "abc"}}, Plan{{Name: "test.capture"}})
	if !res.Success {
		t.Fatalf("expected success: %#v", res)
	}
	if !cap.deadline {
		t.Fatal("expected timeout deadline on step context")
	}
	if !reflect.DeepEqual(cap.order, []string{"test.capture"}) || cap.ctxValue != "abc" {
		t.Fatalf("capture=%#v", cap)
	}
}
