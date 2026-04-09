package window

import "context"

// Enumerator lists top-level windows from the host OS.
type Enumerator interface {
	EnumerateWindows(ctx context.Context) ([]Info, error)
}

// TopLevel returns visible top-level windows from the provider.
func TopLevel(ctx context.Context, e Enumerator) ([]Info, error) {
	return e.EnumerateWindows(ctx)
}

func Enumerate(ctx context.Context, e Enumerator) ([]Info, error) {
	return TopLevel(ctx, e)
}
