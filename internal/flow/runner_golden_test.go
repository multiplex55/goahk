package flow

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"goahk/internal/testutil"
)

func TestRunner_ExecutionTraceGolden(t *testing.T) {
	r := Runner{
		Actions: resolverFunc(func(name string) (ActionHandler, bool) {
			return func(context.Context, Step) error { return nil }, true
		}),
		Conditions: ConditionEvaluator{Windows: mockWindows{ok: true}},
	}
	res := r.Run(context.Background(), Definition{Steps: []Step{
		{Action: "one"},
		{If: &IfBlock{Condition: Condition{WindowMatches: &WindowCondition{Matcher: "editor"}}, Then: []Step{{Action: "then"}}}},
	}})
	normalizeResultTimes(&res)
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	testutil.AssertGolden(t, "testdata/golden/flow/execution_trace.json", string(b)+"\n")
}

func normalizeResultTimes(res *Result) {
	res.Duration = 0
	res.Started = time.Unix(0, 0).UTC()
	res.Ended = time.Unix(0, 0).UTC()
	for i := range res.Traces {
		normalizeTraceTimes(&res.Traces[i])
	}
}

func normalizeTraceTimes(tr *Trace) {
	tr.Duration = 0
	tr.Started = time.Unix(0, 0).UTC()
	tr.Ended = time.Unix(0, 0).UTC()
	for i := range tr.Nested {
		normalizeTraceTimes(&tr.Nested[i])
	}
}
