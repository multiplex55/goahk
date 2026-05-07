//go:build windows
// +build windows

package inspect

import (
	"errors"
	"sync"
	"testing"
)

func TestUIAWorkerDoSerializesJobs(t *testing.T) {
	w, err := newUIACOMWorkerWithInit(func(*uiaWorkerState) error { return nil })
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	defer w.Close()

	order := make([]int, 0, 3)
	var mu sync.Mutex
	for i := 0; i < 3; i++ {
		i := i
		if err := w.Do("job", func(*uiaWorkerState) error {
			mu.Lock()
			defer mu.Unlock()
			order = append(order, i)
			return nil
		}); err != nil {
			t.Fatalf("Do(%d): %v", i, err)
		}
	}
	for i := range order {
		if order[i] != i {
			t.Fatalf("unexpected order: %v", order)
		}
	}
}

func TestUIAWorkerDoPropagatesError(t *testing.T) {
	want := errors.New("boom")
	w, err := newUIACOMWorkerWithInit(func(*uiaWorkerState) error { return nil })
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	defer w.Close()
	if err := w.Do("job", func(*uiaWorkerState) error { return want }); !errors.Is(err, want) {
		t.Fatalf("expected propagated error, got %v", err)
	}
}

func TestUIAWorkerClosePreventsNewJobs(t *testing.T) {
	w, err := newUIACOMWorkerWithInit(func(*uiaWorkerState) error { return nil })
	if err != nil {
		t.Fatalf("new worker: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := w.Do("after-close", func(*uiaWorkerState) error { return nil }); err == nil {
		t.Fatal("expected error after close")
	}
}
