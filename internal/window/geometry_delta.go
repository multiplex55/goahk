package window

import "fmt"

// TranslateMoveBy returns absolute coordinates after applying a delta.
func TranslateMoveBy(bounds Rect, dx, dy int) (int, int) {
	return bounds.Left + dx, bounds.Top + dy
}

// TranslateResizeBy returns absolute size after applying a size delta.
func TranslateResizeBy(bounds Rect, dw, dh int) (int, int, error) {
	width := bounds.Width() + dw
	height := bounds.Height() + dh
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("resulting window size must be positive, got width=%d height=%d", width, height)
	}
	return width, height, nil
}

// CenterPosition returns top-left coordinates for centering bounds inside workArea.
func CenterPosition(bounds, workArea Rect) (int, int) {
	return workArea.Left + (workArea.Width()-bounds.Width())/2, workArea.Top + (workArea.Height()-bounds.Height())/2
}
