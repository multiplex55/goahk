package window

import (
	"context"
	"errors"
	"testing"
)

type fakeGeometryProvider struct {
	windows      []Info
	rects        map[HWND]Rect
	active       Info
	moveErr      error
	resizeErr    error
	boundsErr    error
	moved        HWND
	resized      HWND
	movePosition [2]int
	resizeSize   [2]int
}

func (f *fakeGeometryProvider) EnumerateWindows(context.Context) ([]Info, error) {
	return f.windows, nil
}
func (f *fakeGeometryProvider) ActiveWindow(context.Context) (Info, error) { return f.active, nil }
func (f *fakeGeometryProvider) WindowBounds(_ context.Context, hwnd HWND) (Rect, error) {
	if f.boundsErr != nil {
		return Rect{}, f.boundsErr
	}
	return f.rects[hwnd], nil
}
func (f *fakeGeometryProvider) MoveWindow(_ context.Context, hwnd HWND, x, y int) error {
	if f.moveErr != nil {
		return f.moveErr
	}
	f.moved = hwnd
	f.movePosition = [2]int{x, y}
	return nil
}
func (f *fakeGeometryProvider) ResizeWindow(_ context.Context, hwnd HWND, width, height int) error {
	if f.resizeErr != nil {
		return f.resizeErr
	}
	f.resized = hwnd
	f.resizeSize = [2]int{width, height}
	return nil
}

func TestBoundsForMatcher_UsesMatcherResolution(t *testing.T) {
	p := &fakeGeometryProvider{
		windows: []Info{{HWND: 10, Title: "Editor A"}, {HWND: 20, Title: "Editor B"}},
		rects:   map[HWND]Rect{20: {Left: 100, Top: 50, Right: 900, Bottom: 700}},
	}
	info, bounds, err := BoundsForMatcher(context.Background(), p, Matcher{TitleExact: "Editor B"}, ActivationPolicy{RequireSingleMatch: true})
	if err != nil {
		t.Fatalf("BoundsForMatcher() error = %v", err)
	}
	if info.HWND != 20 {
		t.Fatalf("resolved hwnd = %v, want 20", info.HWND)
	}
	if bounds.Width() != 800 || bounds.Height() != 650 {
		t.Fatalf("unexpected bounds: %#v", bounds)
	}
}

func TestInfoBoundsPositionAndSizeHelpers(t *testing.T) {
	info := Info{}
	if _, ok := info.Bounds(); ok {
		t.Fatal("Bounds() should report missing rect")
	}
	if _, _, ok := info.Position(); ok {
		t.Fatal("Position() should report missing rect")
	}
	if _, _, ok := info.Size(); ok {
		t.Fatal("Size() should report missing rect")
	}

	info.Rect = &Rect{Left: 10, Top: 12, Right: 210, Bottom: 312}
	bounds, ok := info.Bounds()
	if !ok {
		t.Fatal("Bounds() should report rect present")
	}
	if bounds.Width() != 200 || bounds.Height() != 300 {
		t.Fatalf("unexpected bounds size: %#v", bounds)
	}
	x, y, ok := info.Position()
	if !ok || x != 10 || y != 12 {
		t.Fatalf("unexpected position: x=%d y=%d ok=%t", x, y, ok)
	}
	w, h, ok := info.Size()
	if !ok || w != 200 || h != 300 {
		t.Fatalf("unexpected size: w=%d h=%d ok=%t", w, h, ok)
	}
}

func TestMoveAndResizeForMatcher_DelegatesToProvider(t *testing.T) {
	p := &fakeGeometryProvider{
		windows: []Info{{HWND: 44, Title: "Terminal"}},
	}
	_, err := MoveForMatcher(context.Background(), p, Matcher{TitleContains: "term"}, 10, 20, ActivationPolicy{})
	if err != nil {
		t.Fatalf("MoveForMatcher() error = %v", err)
	}
	_, err = ResizeForMatcher(context.Background(), p, Matcher{TitleContains: "term"}, 1200, 700, ActivationPolicy{})
	if err != nil {
		t.Fatalf("ResizeForMatcher() error = %v", err)
	}
	if p.moved != 44 || p.movePosition != [2]int{10, 20} {
		t.Fatalf("move call mismatch: hwnd=%v pos=%v", p.moved, p.movePosition)
	}
	if p.resized != 44 || p.resizeSize != [2]int{1200, 700} {
		t.Fatalf("resize call mismatch: hwnd=%v size=%v", p.resized, p.resizeSize)
	}
}

func TestActiveGeometryHelpers(t *testing.T) {
	p := &fakeGeometryProvider{
		active: Info{HWND: 99, Title: "Active"},
		rects:  map[HWND]Rect{99: {Left: 0, Top: 0, Right: 1920, Bottom: 1080}},
	}
	_, bounds, err := ActiveBounds(context.Background(), p)
	if err != nil {
		t.Fatalf("ActiveBounds() error = %v", err)
	}
	if bounds.Width() != 1920 {
		t.Fatalf("width = %d, want 1920", bounds.Width())
	}
	if _, err := MoveActive(context.Background(), p, 5, 6); err != nil {
		t.Fatalf("MoveActive() error = %v", err)
	}
	if _, err := ResizeActive(context.Background(), p, 1000, 900); err != nil {
		t.Fatalf("ResizeActive() error = %v", err)
	}
}

func TestGeometryHelperErrorsAreWrapped(t *testing.T) {
	boom := errors.New("boom")
	p := &fakeGeometryProvider{
		windows:   []Info{{HWND: 1, Title: "x"}},
		moveErr:   boom,
		resizeErr: boom,
	}
	if _, err := MoveForMatcher(context.Background(), p, Matcher{TitleExact: "x"}, 1, 2, ActivationPolicy{}); !errors.Is(err, boom) {
		t.Fatalf("MoveForMatcher() error = %v, want wrapped boom", err)
	}
	if _, err := ResizeForMatcher(context.Background(), p, Matcher{TitleExact: "x"}, 10, 20, ActivationPolicy{}); !errors.Is(err, boom) {
		t.Fatalf("ResizeForMatcher() error = %v, want wrapped boom", err)
	}
}
