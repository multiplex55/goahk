package inspect

import (
	"context"
	"sync"
)

type highlightController struct {
	overlay highlightOverlay

	mu                  sync.Mutex
	highlightedWindowID string
	highlightedNodeID   string
}

func newHighlightController(overlay highlightOverlay) *highlightController {
	if overlay == nil {
		overlay = noopHighlightOverlay{}
	}
	return &highlightController{overlay: overlay}
}

func (c *highlightController) ShowNode(ctx context.Context, nodeID string, element InspectElement, windowID string) (bool, error) {
	screen, err := c.overlay.ScreenBounds(ctx)
	if err != nil {
		return false, err
	}
	rect, ok := normalizeHighlightRect(element.BoundingRect, element.IsOffscreen, screen)
	if !ok {
		if err := c.Clear(ctx); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := c.overlay.Show(ctx, rect); err != nil {
		return false, err
	}
	c.mu.Lock()
	c.highlightedNodeID = nodeID
	c.highlightedWindowID = windowID
	c.mu.Unlock()
	return true, nil
}

func (c *highlightController) Clear(ctx context.Context) error {
	if err := c.overlay.Clear(ctx); err != nil {
		return err
	}
	c.mu.Lock()
	c.highlightedNodeID = ""
	c.highlightedWindowID = ""
	c.mu.Unlock()
	return nil
}

func (c *highlightController) ClearOnWindowSwitch(ctx context.Context, nextWindowID string) error {
	c.mu.Lock()
	current := c.highlightedWindowID
	c.mu.Unlock()
	if current == "" || nextWindowID == "" || current == nextWindowID {
		return nil
	}
	return c.Clear(ctx)
}

func (c *highlightController) ClearOnDeselection(ctx context.Context, selected InspectElement) error {
	if _, ok := normalizeHighlightRect(selected.BoundingRect, selected.IsOffscreen, nil); ok {
		return nil
	}
	return c.Clear(ctx)
}
