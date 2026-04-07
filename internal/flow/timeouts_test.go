package flow

import (
	"testing"
	"time"
)

func TestEffectiveTimeout_Preference(t *testing.T) {
	if got := effectiveTimeout(2*time.Second, 500*time.Millisecond); got != 500*time.Millisecond {
		t.Fatalf("step should win when shorter; got %v", got)
	}
	if got := effectiveTimeout(500*time.Millisecond, 2*time.Second); got != 500*time.Millisecond {
		t.Fatalf("flow should cap longer step; got %v", got)
	}
	if got := effectiveTimeout(0, 2*time.Second); got != 2*time.Second {
		t.Fatalf("step should apply when flow unset; got %v", got)
	}
}
