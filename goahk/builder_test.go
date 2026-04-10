package goahk

import "testing"

func TestOnSimpleKeysStored(t *testing.T) {
	a := NewApp()
	a.On("1").Do(Log("digit"))
	a.On("Escape").Do(Log("esc"))

	p := a.toProgram()
	if got, want := p.Bindings[0].Hotkey, "1"; got != want {
		t.Fatalf("hotkey[0] = %q, want %q", got, want)
	}
	if got, want := p.Bindings[1].Hotkey, "Escape"; got != want {
		t.Fatalf("hotkey[1] = %q, want %q", got, want)
	}
}

func TestOnModifierComboNormalizedAndStored(t *testing.T) {
	a := NewApp()
	a.On("shift+control+v").Do(Log("combo"))
	p := a.toProgram()
	if got, want := p.Bindings[0].Hotkey, "Ctrl+Shift+V"; got != want {
		t.Fatalf("hotkey = %q, want %q", got, want)
	}
}

func TestBindingBuilderPolicyHelpers_SetExpectedPolicy(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*BindingBuilder) *BindingBuilder
		wantPolicy string
	}{
		{name: "with_policy", configure: func(b *BindingBuilder) *BindingBuilder { return b.WithPolicy("replace") }, wantPolicy: "replace"},
		{name: "serial", configure: func(b *BindingBuilder) *BindingBuilder { return b.Serial() }, wantPolicy: "serial"},
		{name: "replace", configure: func(b *BindingBuilder) *BindingBuilder { return b.Replace() }, wantPolicy: "replace"},
		{name: "queue_one", configure: func(b *BindingBuilder) *BindingBuilder { return b.QueueOne() }, wantPolicy: "queue-one"},
		{name: "parallel", configure: func(b *BindingBuilder) *BindingBuilder { return b.Parallel() }, wantPolicy: "parallel"},
		{name: "drop", configure: func(b *BindingBuilder) *BindingBuilder { return b.Drop() }, wantPolicy: "drop"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := NewApp()
			tc.configure(a.On("Ctrl+H")).Do(Log("ok"))
			p := a.toProgram()
			if got := string(p.Bindings[0].ConcurrencyPolicy); got != tc.wantPolicy {
				t.Fatalf("policy = %q, want %q", got, tc.wantPolicy)
			}
		})
	}
}
