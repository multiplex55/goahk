package actions

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"goahk/internal/input"
)

type fakeInputService struct {
	texts  []string
	seqs   []input.Sequence
	chords []input.Chord
	opts   []input.SendOptions
	pos    input.MousePosition
	calls  []string
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

func (f *fakeInputService) MoveAbsolute(_ context.Context, _, _ int) error {
	f.calls = append(f.calls, "move_abs")
	return nil
}
func (f *fakeInputService) MoveRelative(_ context.Context, _, _ int) error {
	f.calls = append(f.calls, "move_rel")
	return nil
}
func (f *fakeInputService) Position(context.Context) (input.MousePosition, error) {
	f.calls = append(f.calls, "position")
	return f.pos, nil
}
func (f *fakeInputService) ButtonDown(_ context.Context, _ string) error {
	f.calls = append(f.calls, "button_down")
	return nil
}
func (f *fakeInputService) ButtonUp(_ context.Context, _ string) error {
	f.calls = append(f.calls, "button_up")
	return nil
}
func (f *fakeInputService) Click(_ context.Context, _ string) error {
	f.calls = append(f.calls, "click")
	return nil
}
func (f *fakeInputService) DoubleClick(_ context.Context, _ string) error {
	f.calls = append(f.calls, "double_click")
	return nil
}
func (f *fakeInputService) Wheel(_ context.Context, _ int) error {
	f.calls = append(f.calls, "wheel")
	return nil
}
func (f *fakeInputService) Drag(_ context.Context, _ string, _, _, _, _ int) error {
	f.calls = append(f.calls, "drag")
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

func TestInputMouseActions_DispatchToInputService(t *testing.T) {
	r := NewRegistry()
	in := &fakeInputService{pos: input.MousePosition{X: 10, Y: 20}}
	ctx := ActionContext{Context: context.Background(), Services: Services{Input: in}, Metadata: map[string]string{}, BindingID: "b1"}

	for _, step := range []Step{
		{Name: "input.mouse_move_absolute", Params: map[string]string{"x": "10", "y": "20"}},
		{Name: "input.mouse_move_relative", Params: map[string]string{"dx": "5", "dy": "6"}},
		{Name: "input.mouse_button_down", Params: map[string]string{"button": "left"}},
		{Name: "input.mouse_button_up", Params: map[string]string{"button": "left"}},
		{Name: "input.mouse_click", Params: map[string]string{"button": "right"}},
		{Name: "input.mouse_double_click", Params: map[string]string{"button": "left"}},
		{Name: "input.mouse_wheel", Params: map[string]string{"delta": "120"}},
		{Name: "input.mouse_drag", Params: map[string]string{"button": "left", "start_x": "1", "start_y": "2", "end_x": "3", "end_y": "4"}},
		{Name: "input.mouse_get_position", Params: map[string]string{"save_as": "cursor"}},
	} {
		h, _ := r.Lookup(step.Name)
		if err := h(ctx, step); err != nil {
			t.Fatalf("%s err: %v", step.Name, err)
		}
	}

	wantCalls := []string{"move_abs", "move_rel", "button_down", "button_up", "click", "double_click", "wheel", "drag", "position"}
	if !reflect.DeepEqual(in.calls, wantCalls) {
		t.Fatalf("calls=%v want=%v", in.calls, wantCalls)
	}
	if ctx.Metadata["cursor_x"] != "10" || ctx.Metadata["cursor_y"] != "20" {
		t.Fatalf("metadata=%v", ctx.Metadata)
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

func TestInputActions_MissingServiceErrors(t *testing.T) {
	r := NewRegistry()
	ctx := ActionContext{Context: context.Background(), BindingID: "b1"}
	cases := []Step{
		{Name: "input.send_text", Params: map[string]string{"text": "hi"}},
		{Name: "input.send_keys", Params: map[string]string{"sequence": "ctrl+c"}},
		{Name: "input.send_chord", Params: map[string]string{"chord": "alt+tab"}},
		{Name: "input.mouse_click", Params: map[string]string{"button": "left"}},
	}
	for _, step := range cases {
		h, _ := r.Lookup(step.Name)
		err := h(ctx, step)
		if err == nil {
			t.Fatalf("expected error for %s", step.Name)
		}
		if !strings.Contains(err.Error(), "input service unavailable") {
			t.Fatalf("%s err=%q", step.Name, err.Error())
		}
		if !strings.Contains(err.Error(), step.Name) {
			t.Fatalf("%s err=%q missing action identity", step.Name, err.Error())
		}
	}
}
