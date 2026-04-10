//go:build windows
// +build windows

package input

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type windowsService struct {
	backend windowsBackend
}

type windowsBackend interface {
	sendText(context.Context, string) error
	sendChord(context.Context, []string) error
	moveAbsolute(context.Context, int, int) error
	moveRelative(context.Context, int, int) error
	position(context.Context) (MousePosition, error)
	mouseButton(context.Context, string, string) error
	wheel(context.Context, int) error
}

func newPlatformService() Service {
	backend := os.Getenv("GOAHK_WINDOWS_INPUT_BACKEND")
	if strings.EqualFold(strings.TrimSpace(backend), "powershell") {
		return windowsService{backend: newWindowsPowerShellBackend()}
	}
	return windowsService{backend: newWindowsSendInputBackend()}
}

func (s windowsService) SendText(ctx context.Context, text string, opts SendOptions) error {
	if err := validateSendOptions(opts); err != nil {
		return err
	}
	if err := sleepBefore(ctx, opts.DelayBefore); err != nil {
		return err
	}
	return mapSendError(s.backend.sendText(ctx, text))
}

func (s windowsService) SendKeys(ctx context.Context, seq Sequence, opts SendOptions) error {
	if err := validateSendOptions(opts); err != nil {
		return err
	}
	if err := validateSequence(seq); err != nil {
		return err
	}
	if err := sleepBefore(ctx, opts.DelayBefore); err != nil {
		return err
	}
	for _, token := range seq.Tokens {
		if err := mapSendError(s.backend.sendChord(ctx, token.Keys)); err != nil {
			return err
		}
	}
	return nil
}

func (s windowsService) SendChord(ctx context.Context, chord Chord, opts SendOptions) error {
	if err := validateSendOptions(opts); err != nil {
		return err
	}
	if err := validateChord(chord); err != nil {
		return err
	}
	if err := sleepBefore(ctx, opts.DelayBefore); err != nil {
		return err
	}
	return mapSendError(s.backend.sendChord(ctx, chord.Keys))
}

func validateSendOptions(opts SendOptions) error {
	if opts.DelayBefore < 0 {
		return fmt.Errorf("%w: delay_before must be >= 0", ErrInvalidInputArgument)
	}
	return nil
}

func validateSequence(seq Sequence) error {
	for _, token := range seq.Tokens {
		if len(token.Keys) == 0 {
			return fmt.Errorf("%w: sequence contains an empty token", ErrInvalidInputArgument)
		}
		for _, key := range token.Keys {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("%w: sequence contains an empty key", ErrInvalidInputArgument)
			}
		}
	}
	return nil
}

func validateChord(chord Chord) error {
	if len(chord.Keys) == 0 {
		return fmt.Errorf("%w: chord must contain at least one key", ErrInvalidInputArgument)
	}
	for _, key := range chord.Keys {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("%w: chord contains an empty key", ErrInvalidInputArgument)
		}
	}
	return nil
}

func mapSendError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrSendKeysFailed, err)
}

func sleepBefore(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (s windowsService) MoveAbsolute(ctx context.Context, x, y int) error {
	return s.backend.moveAbsolute(ctx, x, y)
}

func (s windowsService) MoveRelative(ctx context.Context, dx, dy int) error {
	return s.backend.moveRelative(ctx, dx, dy)
}

func (s windowsService) Position(ctx context.Context) (MousePosition, error) {
	return s.backend.position(ctx)
}

func (s windowsService) ButtonDown(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return mapSendError(s.backend.mouseButton(ctx, button, "down"))
}

func (s windowsService) ButtonUp(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return mapSendError(s.backend.mouseButton(ctx, button, "up"))
}

func (s windowsService) Click(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return mapSendError(s.backend.mouseButton(ctx, button, "click"))
}

func (s windowsService) DoubleClick(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return mapSendError(s.backend.mouseButton(ctx, button, "double"))
}

func (s windowsService) Wheel(ctx context.Context, delta int) error {
	return mapSendError(s.backend.wheel(ctx, delta))
}

func (s windowsService) Drag(ctx context.Context, button string, startX, startY, endX, endY int) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	if err := s.MoveAbsolute(ctx, startX, startY); err != nil {
		return err
	}
	if err := s.ButtonDown(ctx, button); err != nil {
		return err
	}
	if err := s.MoveAbsolute(ctx, endX, endY); err != nil {
		return err
	}
	return s.ButtonUp(ctx, button)
}

func normalizeMouseButton(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", MouseButtonLeft:
		return MouseButtonLeft, nil
	case MouseButtonRight:
		return MouseButtonRight, nil
	case MouseButtonMiddle:
		return MouseButtonMiddle, nil
	default:
		return "", fmt.Errorf("%w: mouse button must be left/right/middle", ErrInvalidInputArgument)
	}
}
