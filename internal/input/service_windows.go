//go:build windows
// +build windows

package input

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrInvalidInputArgument = errors.New("input: invalid argument")
	ErrSendKeysFailed       = errors.New("input: send keys failed")
)

type windowsService struct{}

type sendKeysRunner func(context.Context, string, bool) error

var platformSendKeys sendKeysRunner = runPowerShellSendKeys

func newPlatformService() Service {
	return windowsService{}
}

func (windowsService) SendText(ctx context.Context, text string, opts SendOptions) error {
	if err := validateSendOptions(opts); err != nil {
		return err
	}
	if err := sleepBefore(ctx, opts.DelayBefore); err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return mapSendError(platformSendKeys(ctx, escapeSendKeysLiteral(text), false))
}

func (windowsService) SendKeys(ctx context.Context, seq Sequence, opts SendOptions) error {
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
		combo := strings.TrimSpace(strings.Join(token.Keys, "+"))
		if combo == "" {
			continue
		}
		if err := mapSendError(platformSendKeys(ctx, toSendKeysChord(combo), true)); err != nil {
			return err
		}
	}
	return nil
}

func (windowsService) SendChord(ctx context.Context, chord Chord, opts SendOptions) error {
	if err := validateSendOptions(opts); err != nil {
		return err
	}
	if err := validateChord(chord); err != nil {
		return err
	}
	if err := sleepBefore(ctx, opts.DelayBefore); err != nil {
		return err
	}
	combo := strings.TrimSpace(strings.Join(chord.Keys, "+"))
	if combo == "" {
		return nil
	}
	return mapSendError(platformSendKeys(ctx, toSendKeysChord(combo), true))
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

func runPowerShellSendKeys(ctx context.Context, value string, wait bool) error {
	call := "$wshell.SendKeys($value)"
	if wait {
		call = "$wshell.SendKeys($value,$true)"
	}
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$wshell = New-Object -ComObject WScript.Shell",
		"$value = \"" + strings.ReplaceAll(value, "\"", "`\"") + "\"",
		call,
	}, ";")
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return fmt.Errorf("input send keys: %w", err)
		}
		return fmt.Errorf("input send keys: %w (%s)", err, trimmed)
	}
	return nil
}

func toSendKeysChord(combo string) string {
	parts := strings.Split(combo, "+")
	if len(parts) == 1 {
		return mapSendKeysToken(parts[0])
	}
	var mods strings.Builder
	for i := 0; i < len(parts)-1; i++ {
		switch strings.ToLower(strings.TrimSpace(parts[i])) {
		case "ctrl", "control":
			mods.WriteString("^")
		case "alt":
			mods.WriteString("%")
		case "shift":
			mods.WriteString("+")
		case "win", "lwin", "rwin", "meta":
			mods.WriteString("(")
			mods.WriteString("^{ESC}")
			mods.WriteString(")")
		default:
			mods.WriteString(mapSendKeysToken(parts[i]))
		}
	}
	last := mapSendKeysToken(parts[len(parts)-1])
	if mods.Len() == 0 {
		return last
	}
	return mods.String() + "(" + last + ")"
}

func mapSendKeysToken(key string) string {
	n := strings.ToLower(strings.TrimSpace(key))
	switch n {
	case "enter", "return":
		return "{ENTER}"
	case "esc", "escape":
		return "{ESC}"
	case "tab":
		return "{TAB}"
	case "space":
		return " "
	case "backspace":
		return "{BACKSPACE}"
	case "delete", "del":
		return "{DELETE}"
	case "up":
		return "{UP}"
	case "down":
		return "{DOWN}"
	case "left":
		return "{LEFT}"
	case "right":
		return "{RIGHT}"
	case "home":
		return "{HOME}"
	case "end":
		return "{END}"
	case "pgup", "pageup":
		return "{PGUP}"
	case "pgdn", "pagedown":
		return "{PGDN}"
	}
	if len(n) == 1 {
		return escapeSendKeysLiteral(n)
	}
	return "{" + strings.ToUpper(n) + "}"
}

func escapeSendKeysLiteral(s string) string {
	repl := strings.NewReplacer("+", "{+}", "^", "{^}", "%", "{%}", "~", "{~}", "(", "{(}", ")", "{)}", "[", "{[}", "]", "{]}", "{", "{{}", "}", "{}}")
	return repl.Replace(s)
}
