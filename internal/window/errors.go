package window

import "errors"

// ErrUnsupportedPlatform indicates window operations are unavailable on non-Windows platforms.
var ErrUnsupportedPlatform = errors.New("window provider is only supported on Windows")
