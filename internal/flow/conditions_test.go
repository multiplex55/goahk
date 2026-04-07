package flow

import (
	"context"
	"errors"
	"testing"
)

type mockWindows struct {
	ok  bool
	err error
}

func (m mockWindows) WindowMatches(context.Context, string) (bool, error) { return m.ok, m.err }

type mockElements struct {
	ok  bool
	err error
}

func (m mockElements) ElementExists(context.Context, string) (bool, error) { return m.ok, m.err }

func TestConditionEvaluator_WindowAndElement(t *testing.T) {
	e := ConditionEvaluator{Windows: mockWindows{ok: true}, Elements: mockElements{ok: false}}
	ok, err := e.Evaluate(context.Background(), Condition{WindowMatches: &WindowCondition{Matcher: "editor"}})
	if err != nil || !ok {
		t.Fatalf("window condition got ok=%v err=%v", ok, err)
	}
	ok, err = e.Evaluate(context.Background(), Condition{ElementExists: &ElementCondition{Selector: "submit"}})
	if err != nil || ok {
		t.Fatalf("element condition got ok=%v err=%v", ok, err)
	}
}

func TestConditionEvaluator_ErrorPropagation(t *testing.T) {
	expected := errors.New("boom")
	e := ConditionEvaluator{Windows: mockWindows{err: expected}}
	_, err := e.Evaluate(context.Background(), Condition{WindowMatches: &WindowCondition{Matcher: "x"}})
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
