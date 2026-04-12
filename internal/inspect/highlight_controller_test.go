package inspect

import (
	"context"
	"sync"
	"testing"
)

type mockOverlay struct {
	mu         sync.Mutex
	showCalls  []Rect
	clearCalls int
	screen     *Rect
	screenErr  error
	showErr    error
	clearErr   error
}

func (m *mockOverlay) Show(_ context.Context, rect Rect) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.showCalls = append(m.showCalls, rect)
	return m.showErr
}

func (m *mockOverlay) Clear(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clearCalls++
	return m.clearErr
}

func (m *mockOverlay) ScreenBounds(context.Context) (*Rect, error) {
	return m.screen, m.screenErr
}

func TestNormalizeHighlightRect(t *testing.T) {
	t.Run("normalizes negative dimensions and clips to screen", func(t *testing.T) {
		rect, ok := normalizeHighlightRect(&Rect{Left: 100, Top: 100, Width: -30, Height: -20}, false, &Rect{Left: 50, Top: 50, Width: 60, Height: 60})
		if !ok {
			t.Fatalf("expected valid rectangle")
		}
		if rect != (Rect{Left: 70, Top: 80, Width: 30, Height: 20}) {
			t.Fatalf("unexpected normalized rectangle: %+v", rect)
		}
	})

	t.Run("rejects zero-sized and fully offscreen rectangles", func(t *testing.T) {
		if _, ok := normalizeHighlightRect(&Rect{Left: 0, Top: 0, Width: 0, Height: 20}, false, nil); ok {
			t.Fatalf("expected zero-width rectangle to be rejected")
		}
		if _, ok := normalizeHighlightRect(&Rect{Left: 1000, Top: 1000, Width: 10, Height: 10}, false, &Rect{Left: 0, Top: 0, Width: 100, Height: 100}); ok {
			t.Fatalf("expected offscreen rectangle to be rejected")
		}
	})
}

func TestHighlightController_Lifecycle(t *testing.T) {
	t.Run("show on selection", func(t *testing.T) {
		overlay := &mockOverlay{screen: &Rect{Left: 0, Top: 0, Width: 500, Height: 500}}
		controller := newHighlightController(overlay)
		highlighted, err := controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 10, Top: 20, Width: 100, Height: 50}}, "0x1")
		if err != nil {
			t.Fatalf("ShowNode error: %v", err)
		}
		if !highlighted {
			t.Fatalf("expected highlighted true")
		}
		if len(overlay.showCalls) != 1 {
			t.Fatalf("expected one show call, got %d", len(overlay.showCalls))
		}
	})

	t.Run("clear on refresh/switch", func(t *testing.T) {
		overlay := &mockOverlay{screen: &Rect{Left: 0, Top: 0, Width: 500, Height: 500}}
		controller := newHighlightController(overlay)
		_, _ = controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 10, Top: 20, Width: 100, Height: 50}}, "0x1")
		if err := controller.ClearOnWindowSwitch(context.Background(), "0x2"); err != nil {
			t.Fatalf("ClearOnWindowSwitch error: %v", err)
		}
		if overlay.clearCalls == 0 {
			t.Fatalf("expected clear call on switch")
		}
	})

	t.Run("clear before rehighlighting a different node", func(t *testing.T) {
		overlay := &mockOverlay{screen: &Rect{Left: 0, Top: 0, Width: 500, Height: 500}}
		controller := newHighlightController(overlay)

		_, _ = controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 10, Top: 20, Width: 100, Height: 50}}, "0x1")
		_, err := controller.ShowNode(context.Background(), "node:2", InspectElement{BoundingRect: &Rect{Left: 20, Top: 30, Width: 80, Height: 40}}, "0x1")
		if err != nil {
			t.Fatalf("ShowNode error: %v", err)
		}
		if overlay.clearCalls != 1 {
			t.Fatalf("expected clear-before-rehighlight, got %d clear calls", overlay.clearCalls)
		}
		if len(overlay.showCalls) != 2 {
			t.Fatalf("expected two show calls, got %d", len(overlay.showCalls))
		}
	})

	t.Run("selection change to no selectable rect clears highlight", func(t *testing.T) {
		overlay := &mockOverlay{}
		controller := newHighlightController(overlay)
		_, _ = controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 10, Top: 20, Width: 100, Height: 50}}, "0x1")

		err := controller.ClearOnDeselection(context.Background(), InspectElement{})
		if err != nil {
			t.Fatalf("ClearOnDeselection error: %v", err)
		}
		if overlay.clearCalls == 0 {
			t.Fatalf("expected clear call on deselection")
		}
	})

	t.Run("refresh-triggered clear is idempotent", func(t *testing.T) {
		overlay := &mockOverlay{}
		controller := newHighlightController(overlay)
		_, _ = controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 10, Top: 20, Width: 100, Height: 50}}, "0x1")

		if err := controller.Clear(context.Background()); err != nil {
			t.Fatalf("first Clear error: %v", err)
		}
		if err := controller.Clear(context.Background()); err != nil {
			t.Fatalf("second Clear error: %v", err)
		}
		if overlay.clearCalls != 2 {
			t.Fatalf("expected two clear calls, got %d", overlay.clearCalls)
		}
	})

	t.Run("ignore invalid bounds safely", func(t *testing.T) {
		overlay := &mockOverlay{screen: &Rect{Left: 0, Top: 0, Width: 100, Height: 100}}
		controller := newHighlightController(overlay)
		highlighted, err := controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 300, Top: 300, Width: 5, Height: 5}}, "0x1")
		if err != nil {
			t.Fatalf("ShowNode error: %v", err)
		}
		if highlighted {
			t.Fatalf("expected invalid bounds to avoid highlight")
		}
		if len(overlay.showCalls) != 0 {
			t.Fatalf("expected no show calls for invalid bounds")
		}
		if overlay.clearCalls == 0 {
			t.Fatalf("expected clear to be called for invalid bounds")
		}
	})
}

func TestHighlightController_NoopOverlaySafety(t *testing.T) {
	controller := newHighlightController(nil)

	highlighted, err := controller.ShowNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 1, Top: 1, Width: 10, Height: 10}}, "0x1")
	if err != nil {
		t.Fatalf("ShowNode with noop overlay error: %v", err)
	}
	if !highlighted {
		t.Fatalf("expected noop overlay path to report highlighted")
	}
	if err := controller.Clear(context.Background()); err != nil {
		t.Fatalf("Clear with noop overlay error: %v", err)
	}
}

func TestHighlightController_ConcurrencySafety(t *testing.T) {
	overlay := &mockOverlay{screen: &Rect{Left: 0, Top: 0, Width: 1000, Height: 1000}}
	controller := newHighlightController(overlay)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = controller.ShowNode(context.Background(), "node", InspectElement{BoundingRect: &Rect{Left: i, Top: i, Width: 10, Height: 10}}, "0x1")
			_ = controller.ClearOnWindowSwitch(context.Background(), "0x2")
		}(i)
	}
	wg.Wait()
}

func TestHighlightController_SyncSelectedNode_ValidInvalidTransitions(t *testing.T) {
	overlay := &mockOverlay{screen: &Rect{Left: 0, Top: 0, Width: 400, Height: 400}}
	controller := newHighlightController(overlay)

	highlighted, err := controller.SyncSelectedNode(context.Background(), "node:1", InspectElement{BoundingRect: &Rect{Left: 10, Top: 10, Width: 40, Height: 40}}, "0x1")
	if err != nil || !highlighted {
		t.Fatalf("expected initial sync highlight, highlighted=%v err=%v", highlighted, err)
	}
	if len(overlay.showCalls) != 1 {
		t.Fatalf("expected one show call, got %d", len(overlay.showCalls))
	}

	highlighted, err = controller.SyncSelectedNode(context.Background(), "node:2", InspectElement{}, "0x1")
	if err != nil {
		t.Fatalf("expected invalid rect transition to clear, err=%v", err)
	}
	if highlighted {
		t.Fatalf("expected invalid rect to not be highlighted")
	}
	if overlay.clearCalls == 0 {
		t.Fatalf("expected clear call after invalid rect transition")
	}
}
