package flow

import (
	"context"
	"reflect"
	"testing"
)

type resolverFunc func(string) (ActionHandler, bool)

func (f resolverFunc) ResolveAction(name string) (ActionHandler, bool) { return f(name) }

func TestRunner_OrderedBranchAndRepeat(t *testing.T) {
	order := []string{}
	r := Runner{
		Actions: resolverFunc(func(name string) (ActionHandler, bool) {
			return func(_ context.Context, _ Step) error {
				order = append(order, name)
				return nil
			}, true
		}),
		Conditions: ConditionEvaluator{Windows: mockWindows{ok: true}},
	}
	def := Definition{Steps: []Step{
		{Action: "one"},
		{If: &IfBlock{Condition: Condition{WindowMatches: &WindowCondition{Matcher: "editor"}}, Then: []Step{{Action: "then"}}, Else: []Step{{Action: "else"}}}},
		{Repeat: &RepeatBlock{Times: 2, Steps: []Step{{Action: "loop"}}}},
	}}
	res := r.Run(context.Background(), def)
	if !res.Success {
		t.Fatalf("expected success: %#v", res)
	}
	want := []string{"one", "then", "loop", "loop"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order=%v want=%v", order, want)
	}
}
