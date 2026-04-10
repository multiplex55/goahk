package inspect

import "context"

type highlightOverlay interface {
	Show(context.Context, Rect) error
	Clear(context.Context) error
	ScreenBounds(context.Context) (*Rect, error)
}

type noopHighlightOverlay struct{}

func (noopHighlightOverlay) Show(context.Context, Rect) error            { return nil }
func (noopHighlightOverlay) Clear(context.Context) error                 { return nil }
func (noopHighlightOverlay) ScreenBounds(context.Context) (*Rect, error) { return nil, nil }

func normalizeHighlightRect(rect *Rect, isOffscreen bool, screen *Rect) (Rect, bool) {
	if rect == nil || isOffscreen {
		return Rect{}, false
	}
	left := rect.Left
	top := rect.Top
	width := rect.Width
	height := rect.Height

	if width < 0 {
		left += width
		width = -width
	}
	if height < 0 {
		top += height
		height = -height
	}
	if width <= 0 || height <= 0 {
		return Rect{}, false
	}

	normalized := Rect{Left: left, Top: top, Width: width, Height: height}
	if screen == nil {
		return normalized, true
	}
	clipped, ok := intersectRect(normalized, *screen)
	if !ok {
		return Rect{}, false
	}
	return clipped, true
}

func intersectRect(a, b Rect) (Rect, bool) {
	aRight := a.Left + a.Width
	aBottom := a.Top + a.Height
	bRight := b.Left + b.Width
	bBottom := b.Top + b.Height

	left := maxInt(a.Left, b.Left)
	top := maxInt(a.Top, b.Top)
	right := minInt(aRight, bRight)
	bottom := minInt(aBottom, bBottom)
	if right <= left || bottom <= top {
		return Rect{}, false
	}
	return Rect{Left: left, Top: top, Width: right - left, Height: bottom - top}, true
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
