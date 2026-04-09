package input

import (
	"context"
	"reflect"
	"testing"
)

type fakeMouseService struct {
	position MousePosition
	calls    []string
}

func (f *fakeMouseService) MoveAbsolute(_ context.Context, x, y int) error {
	f.position = MousePosition{X: x, Y: y}
	f.calls = append(f.calls, "move_abs")
	return nil
}

func (f *fakeMouseService) MoveRelative(_ context.Context, dx, dy int) error {
	f.position = MousePosition{X: f.position.X + dx, Y: f.position.Y + dy}
	f.calls = append(f.calls, "move_rel")
	return nil
}

func (f *fakeMouseService) Position(context.Context) (MousePosition, error) {
	f.calls = append(f.calls, "position")
	return f.position, nil
}

func (f *fakeMouseService) ButtonDown(context.Context, string) error {
	f.calls = append(f.calls, "button_down")
	return nil
}

func (f *fakeMouseService) ButtonUp(context.Context, string) error {
	f.calls = append(f.calls, "button_up")
	return nil
}

func (f *fakeMouseService) Click(context.Context, string) error {
	f.calls = append(f.calls, "click")
	return nil
}

func (f *fakeMouseService) DoubleClick(context.Context, string) error {
	f.calls = append(f.calls, "double_click")
	return nil
}

func (f *fakeMouseService) Wheel(context.Context, int) error {
	f.calls = append(f.calls, "wheel")
	return nil
}

func (f *fakeMouseService) Drag(context.Context, string, int, int, int, int) error {
	f.calls = append(f.calls, "drag")
	return nil
}

func TestMouse_MoveAndGetPositionRoundTrip(t *testing.T) {
	m := &fakeMouseService{}
	if err := m.MoveAbsolute(context.Background(), 200, 300); err != nil {
		t.Fatalf("MoveAbsolute() err = %v", err)
	}
	if err := m.MoveRelative(context.Background(), -10, 20); err != nil {
		t.Fatalf("MoveRelative() err = %v", err)
	}
	pos, err := m.Position(context.Background())
	if err != nil {
		t.Fatalf("Position() err = %v", err)
	}
	if pos != (MousePosition{X: 190, Y: 320}) {
		t.Fatalf("pos = %#v, want {190,320}", pos)
	}
}

func TestMouse_ButtonAndWheelDispatch(t *testing.T) {
	m := &fakeMouseService{}
	_ = m.ButtonDown(context.Background(), MouseButtonLeft)
	_ = m.ButtonUp(context.Background(), MouseButtonLeft)
	_ = m.Click(context.Background(), MouseButtonRight)
	_ = m.DoubleClick(context.Background(), MouseButtonLeft)
	_ = m.Wheel(context.Background(), 120)
	_ = m.Drag(context.Background(), MouseButtonLeft, 0, 0, 20, 40)

	want := []string{"button_down", "button_up", "click", "double_click", "wheel", "drag"}
	if !reflect.DeepEqual(m.calls, want) {
		t.Fatalf("calls = %v, want %v", m.calls, want)
	}
}
