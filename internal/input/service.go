package input

import "context"

// NewService returns the platform implementation for synthetic input.
func NewService() Service {
	return newPlatformService()
}

// UnsupportedServiceError reports unsupported input operations on non-Windows platforms.
type UnsupportedServiceError interface {
	error
	Unsupported() bool
}

// IsUnsupported reports whether err indicates unsupported platform behavior.
func IsUnsupported(err error) bool {
	type unsupported interface{ Unsupported() bool }
	if err == nil {
		return false
	}
	u, ok := err.(unsupported)
	return ok && u.Unsupported()
}

// NopContext allows callers to pass nil contexts safely in tests/integration glue.
func NopContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
