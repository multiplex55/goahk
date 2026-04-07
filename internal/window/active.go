package window

import "context"

// ActiveFinder resolves the currently active/foreground window.
type ActiveFinder interface {
	ActiveWindow(ctx context.Context) (Info, error)
}

func Active(ctx context.Context, finder ActiveFinder) (Info, error) {
	return finder.ActiveWindow(ctx)
}
