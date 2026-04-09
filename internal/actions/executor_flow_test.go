package actions

import (
	"context"
	"testing"

	"goahk/internal/flow"
)

func TestExecutor_ExecuteFlow_InteropAndTrace(t *testing.T) {
	reg := NewRegistry()
	calls := 0
	_ = reg.Register("test.flow", func(_ ActionContext, _ Step) error {
		calls++
		return nil
	})
	ex := NewExecutor(reg)

	def := flow.Definition{Steps: []flow.Step{{Action: "test.flow"}, {Repeat: &flow.RepeatBlock{Times: 2, Steps: []flow.Step{{Action: "test.flow"}}}}}}
	res := ex.ExecuteFlow(ActionContext{Context: context.Background()}, def, flow.ConditionEvaluator{})
	if !res.Success {
		t.Fatalf("expected success: %#v", res)
	}
	if calls != 3 {
		t.Fatalf("calls=%d want=3", calls)
	}
	if len(res.Steps) != 2 || len(res.Steps[1].Nested) != 2 {
		t.Fatalf("unexpected nested trace shape: %#v", res.Steps)
	}
}

func TestExecutor_ExecuteBinding_CancellationParityAcrossPlanFlowAndCallback(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register("test.cancelable", func(ctx ActionContext, _ Step) error {
		<-ctx.Context.Done()
		return ctx.Context.Err()
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	if err := reg.RegisterCallback("named", func(ctx CallbackContext) error {
		<-ctx.Done()
		return ctx.Err()
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}
	exec := NewExecutor(reg)

	cases := []ExecutableBinding{
		{ID: "plan", Kind: BindingKindPlan, Plan: Plan{{Name: "test.cancelable"}}},
		{ID: "flow", Kind: BindingKindFlow, Flow: &flow.Definition{ID: "f", Steps: []flow.Step{{Action: "test.cancelable"}}}},
		{ID: "callback", Kind: BindingKindCallback, Policy: BindingExecutionPolicy{CallbackRef: "named"}},
	}
	for _, binding := range cases {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		res := exec.ExecuteBinding(ActionContext{Context: ctx, BindingID: binding.ID}, binding)
		if res.Success {
			t.Fatalf("binding %q should fail when canceled: %#v", binding.ID, res)
		}
		if len(res.Steps) == 0 || res.Steps[0].Error == "" || res.Steps[0].Error != context.Canceled.Error() {
			t.Fatalf("binding %q should report cancellation error: %#v", binding.ID, res)
		}
	}
}
