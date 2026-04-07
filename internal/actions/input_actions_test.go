package actions

import (
	"context"
	"reflect"
	"testing"

	"goahk/internal/input"
)

type fakeInputService struct {
	texts  []string
	seqs   []input.Sequence
	chords []input.Chord
	opts   []input.SendOptions
}

func (f *fakeInputService) SendText(_ context.Context, text string, opts input.SendOptions) error {
	f.texts = append(f.texts, text)
	f.opts = append(f.opts, opts)
	return nil
}

func (f *fakeInputService) SendKeys(_ context.Context, seq input.Sequence, opts input.SendOptions) error {
	f.seqs = append(f.seqs, seq)
	f.opts = append(f.opts, opts)
	return nil
}

func (f *fakeInputService) SendChord(_ context.Context, chord input.Chord, opts input.SendOptions) error {
	f.chords = append(f.chords, chord)
	f.opts = append(f.opts, opts)
	return nil
}

func TestInputActions_DispatchToInputService(t *testing.T) {
	r := NewRegistry()
	in := &fakeInputService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{Input: in}, Metadata: map[string]string{}, BindingID: "b1"}

	h, _ := r.Lookup("input.send_text")
	if err := h(ctx, Step{Name: "input.send_text", Params: map[string]string{"text": `hi\n🙂`, "decode_escapes": "true", "delay_ms": "5"}}); err != nil {
		t.Fatalf("send_text err: %v", err)
	}
	if len(in.texts) != 1 || in.texts[0] != "hi\n🙂" {
		t.Fatalf("texts=%v", in.texts)
	}
	if !in.opts[0].SuppressReentrancy {
		t.Fatal("expected suppress reentrancy by default")
	}

	h, _ = r.Lookup("input.send_keys")
	if err := h(ctx, Step{Name: "input.send_keys", Params: map[string]string{"sequence": "ctrl+c {enter}"}}); err != nil {
		t.Fatalf("send_keys err: %v", err)
	}
	if len(in.seqs) != 1 || len(in.seqs[0].Tokens) != 2 {
		t.Fatalf("seqs=%v", in.seqs)
	}

	h, _ = r.Lookup("input.send_chord")
	if err := h(ctx, Step{Name: "input.send_chord", Params: map[string]string{"chord": "alt+tab"}}); err != nil {
		t.Fatalf("send_chord err: %v", err)
	}
	if len(in.chords) != 1 || !reflect.DeepEqual(in.chords[0].Keys, []string{"alt", "tab"}) {
		t.Fatalf("chords=%v", in.chords)
	}
}

func TestInputActions_Validation(t *testing.T) {
	r := NewRegistry()
	ctx := ActionContext{Context: context.Background(), Services: Services{Input: &fakeInputService{}}, Metadata: map[string]string{}, BindingID: "b1"}

	cases := []Step{
		{Name: "input.send_keys", Params: map[string]string{}},
		{Name: "input.send_chord", Params: map[string]string{"chord": "ctrl+c alt+v"}},
		{Name: "input.send_text", Params: map[string]string{"delay_ms": "-1"}},
	}
	for _, step := range cases {
		h, _ := r.Lookup(step.Name)
		if err := h(ctx, step); err == nil {
			t.Fatalf("expected error for %#v", step)
		}
	}
}
