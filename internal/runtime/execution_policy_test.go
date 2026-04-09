package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestSupervisorExecutionPolicies(t *testing.T) {
	tests := []struct {
		name            string
		policy          string
		triggers        int
		wantResults     int
		wantStarts      int32
		wantPolicyEvent string
	}{
		{name: "serial", policy: "serial", triggers: 3, wantResults: 1, wantStarts: 1, wantPolicyEvent: "policy_serial_ignored_busy"},
		{name: "drop", policy: "drop", triggers: 3, wantResults: 1, wantStarts: 1, wantPolicyEvent: "policy_drop_ignored_busy"},
		{name: "parallel", policy: "parallel", triggers: 3, wantResults: 3, wantStarts: 3, wantPolicyEvent: "policy_parallel_admit"},
		{name: "queue-one", policy: "queue-one", triggers: 4, wantResults: 2, wantStarts: 2, wantPolicyEvent: "policy_queue_one_pending"},
		{name: "replace", policy: "replace", triggers: 3, wantResults: 3, wantStarts: 3, wantPolicyEvent: "policy_replace_cancel_running"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := actions.NewRegistry()
			release := make(chan struct{})
			var started atomic.Int32
			if err := reg.Register("test.block", func(ctx actions.ActionContext, _ actions.Step) error {
				started.Add(1)
				select {
				case <-release:
					return nil
				case <-ctx.Context.Done():
					return ctx.Context.Err()
				}
			}); err != nil {
				t.Fatalf("register action: %v", err)
			}

			exec := actions.NewExecutor(reg)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var mu sync.Mutex
			logs := []DispatchLogEntry{}
			s := NewSupervisor(ctx, map[string]actions.ExecutableBinding{
				"hk": {
					ID:     "hk",
					Kind:   actions.BindingKindPlan,
					Plan:   actions.Plan{{Name: "test.block"}},
					Policy: actions.BindingExecutionPolicy{Concurrency: tc.policy},
				},
			}, exec, actions.ActionContext{}, func(_ context.Context, entry DispatchLogEntry) {
				mu.Lock()
				defer mu.Unlock()
				logs = append(logs, entry)
			}, nil)
			s.Start(4)

			for i := 0; i < tc.triggers; i++ {
				s.SubmitWork(supervisorJob{bindingID: "hk", trigger: hotkey.TriggerEvent{BindingID: "hk"}})
			}

			time.Sleep(40 * time.Millisecond)
			close(release)
			s.CloseWhenIdle(500 * time.Millisecond)

			results := make([]DispatchResult, 0, tc.wantResults)
			deadline := time.After(2 * time.Second)
		loop:
			for {
				select {
				case r, ok := <-s.Results():
					if !ok {
						break loop
					}
					results = append(results, r)
				case <-deadline:
					t.Fatalf("timed out waiting for results; got=%d", len(results))
				}
			}

			if got := len(results); got != tc.wantResults {
				t.Fatalf("results = %d, want %d", got, tc.wantResults)
			}
			if got := started.Load(); got != tc.wantStarts {
				t.Fatalf("starts = %d, want %d", got, tc.wantStarts)
			}

			mu.Lock()
			defer mu.Unlock()
			found := false
			for _, entry := range logs {
				if entry.Event == tc.wantPolicyEvent {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("missing policy log %q in %#v", tc.wantPolicyEvent, logs)
			}
		})
	}
}
