package input

import "context"

type MousePosition struct {
	X int
	Y int
}

const (
	MouseButtonLeft   = "left"
	MouseButtonRight  = "right"
	MouseButtonMiddle = "middle"
)

type MouseService interface {
	MoveAbsolute(ctx context.Context, x, y int) error
	MoveRelative(ctx context.Context, dx, dy int) error
	Position(ctx context.Context) (MousePosition, error)
	ButtonDown(ctx context.Context, button string) error
	ButtonUp(ctx context.Context, button string) error
	Click(ctx context.Context, button string) error
	DoubleClick(ctx context.Context, button string) error
	Wheel(ctx context.Context, delta int) error
	Drag(ctx context.Context, button string, startX, startY, endX, endY int) error
}
