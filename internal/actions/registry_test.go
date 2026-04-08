package actions

import (
	"errors"
	"testing"
)

func TestRegistry_Lookup(t *testing.T) {
	r := NewRegistry()
	if _, ok := r.Lookup("system.log"); !ok {
		t.Fatal("expected built-in action system.log")
	}
	if _, ok := r.Lookup("runtime.stop"); !ok {
		t.Fatal("expected built-in action runtime.stop")
	}
	if _, ok := r.Lookup("system.open"); !ok {
		t.Fatal("expected built-in action system.open")
	}
	if _, ok := r.Lookup("window.list_open_applications"); !ok {
		t.Fatal("expected built-in action window.list_open_applications")
	}
}

func TestRegistry_DuplicateProtected(t *testing.T) {
	r := NewRegistry()
	err := r.Register("custom.x", func(ActionContext, Step) error { return nil })
	if err != nil {
		t.Fatalf("initial register failed: %v", err)
	}
	if err := r.Register("custom.x", func(ActionContext, Step) error { return errors.New("should not install") }); err == nil {
		t.Fatal("expected duplicate registration error")
	}
}
