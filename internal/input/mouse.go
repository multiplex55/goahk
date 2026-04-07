package input

import "context"

// MouseService is optional and intentionally separate from keyboard/text input.
type MouseService interface {
	Click(ctx context.Context, button string) error
}
