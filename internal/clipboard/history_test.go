package clipboard

import (
	"reflect"
	"testing"
)

func TestHistory_CapacityAndOrdering(t *testing.T) {
	h := NewHistory(3, false)
	h.Push("a")
	h.Push("b")
	h.Push("c")
	h.Push("d")

	got := h.Values()
	want := []string{"d", "c", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Values() = %#v want %#v", got, want)
	}
}

func TestHistory_DedupPolicy(t *testing.T) {
	h := NewHistory(5, true)
	h.Push("x")
	h.Push("x")
	h.Push("y")
	h.Push("y")
	h.Push("x")

	got := h.Values()
	want := []string{"x", "y", "x"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Values() = %#v want %#v", got, want)
	}
}
