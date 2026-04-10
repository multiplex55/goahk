package uia

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("uia inspection is only supported on Windows")
var ErrInspectUnavailable = errors.New("uia inspection backend is unavailable")

type InspectProvider interface {
	Focused(context.Context) (Element, error)
	UnderCursor(context.Context) (Element, error)
	ActiveWindowTree(context.Context, int) (*Node, error)
}
