package goahk

import "testing"

func TestBuilderCallback_MixedPipelineOrderPreserved(t *testing.T) {
	t.Parallel()

	a := NewApp()
	a.Bind("Ctrl+H", Log("one"), Func(func(*Context) error { return nil }), Log("three"))

	p, _, _ := a.runtimeArtifacts()
	got := p.Bindings[0].Steps
	if len(got) != 3 {
		t.Fatalf("steps len = %d, want 3", len(got))
	}
	if got[0].Action != "system.log" {
		t.Fatalf("step[0].Action = %q, want system.log", got[0].Action)
	}
	if got[1].Action != callbackActionName(0, 1) {
		t.Fatalf("step[1].Action = %q, want %q", got[1].Action, callbackActionName(0, 1))
	}
	if got[2].Action != "system.log" {
		t.Fatalf("step[2].Action = %q, want system.log", got[2].Action)
	}
}
