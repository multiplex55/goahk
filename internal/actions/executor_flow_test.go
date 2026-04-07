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
