package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestSupervisor_ControlCommandsBypassWorkerBacklog(t *testing.T) {
	reg := actions.NewRegistry()
	block := make(chan struct{})
	if err := reg.Register("test.block", func(context actions.ActionContext, _ actions.Step) error {
		<-block
		return nil
	}); err != nil {
		t.Fatalf("register test.block: %v", err)
	}
	exec := actions.NewExecutor(reg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var controls atomic.Int32
	bindings := map[string]actions.ExecutableBinding{
		"job1": {ID: "job1", Kind: actions.BindingKindPlan, Plan: actions.Plan{{Name: "test.block"}}},
		"job2": {ID: "job2", Kind: actions.BindingKindPlan, Plan: actions.Plan{{Name: "test.block"}}},
	}
	s := NewSupervisor(ctx, bindings, exec, actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlStop {
			controls.Add(1)
		}
	})
	s.Start(1)
	s.SubmitWork(supervisorJob{bindingID: "job1", trigger: hotkey.TriggerEvent{BindingID: "job1"}})
	s.SubmitWork(supervisorJob{bindingID: "job2", trigger: hotkey.TriggerEvent{BindingID: "job2"}})

	s.SubmitControl(runtimeControlEvent{BindingID: "quit", Command: RuntimeControlStop, Received: time.Now().UTC()})
	deadline := time.After(200 * time.Millisecond)
	for controls.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("control command was delayed behind worker backlog")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	close(block)
	s.CloseWhenIdle(200 * time.Millisecond)
	for range s.Results() {
	}
}
