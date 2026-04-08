package goahk

import (
	"fmt"
	"testing"
)

func TestContextState_ConcurrentAccessIsSafe(t *testing.T) {
	t.Parallel()

	state := newAppState()
	const workers = 24
	const iterations = 100

	for i := 0; i < workers; i++ {
		i := i
		t.Run(fmt.Sprintf("worker_%d", i), func(t *testing.T) {
			t.Parallel()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("k_%d", j%5)
				state.Set(key, fmt.Sprintf("v_%d_%d", i, j))
				state.LoadOrStore(key, "fallback")
				state.Get(key)
			}
		})
	}
}
