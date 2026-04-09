//go:build windows
// +build windows

package input

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
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
var (
	platformMouseMoveAbsolute = runPowerShellMouseMoveAbsolute
	platformMouseMoveRelative = runPowerShellMouseMoveRelative
	platformMousePosition     = runPowerShellMousePosition
	platformMouseButtonDown   = runPowerShellMouseButtonDown
	platformMouseButtonUp     = runPowerShellMouseButtonUp
	platformMouseClick        = runPowerShellMouseClick
	platformMouseDoubleClick  = runPowerShellMouseDoubleClick
	platformMouseWheel        = runPowerShellMouseWheel
)

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

func (windowsService) MoveAbsolute(ctx context.Context, x, y int) error {
	return platformMouseMoveAbsolute(ctx, x, y)
}

func (windowsService) MoveRelative(ctx context.Context, dx, dy int) error {
	return platformMouseMoveRelative(ctx, dx, dy)
}

func (windowsService) Position(ctx context.Context) (MousePosition, error) {
	return platformMousePosition(ctx)
}

func (windowsService) ButtonDown(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return platformMouseButtonDown(ctx, button)
}

func (windowsService) ButtonUp(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return platformMouseButtonUp(ctx, button)
}

func (windowsService) Click(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return platformMouseClick(ctx, button)
}

func (windowsService) DoubleClick(ctx context.Context, button string) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	return platformMouseDoubleClick(ctx, button)
}

func (windowsService) Wheel(ctx context.Context, delta int) error {
	return platformMouseWheel(ctx, delta)
}

func (svc windowsService) Drag(ctx context.Context, button string, startX, startY, endX, endY int) error {
	button, err := normalizeMouseButton(button)
	if err != nil {
		return err
	}
	if err := svc.MoveAbsolute(ctx, startX, startY); err != nil {
		return err
	}
	if err := svc.ButtonDown(ctx, button); err != nil {
		return err
	}
	if err := svc.MoveAbsolute(ctx, endX, endY); err != nil {
		return err
	}
	return svc.ButtonUp(ctx, button)
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

func runPowerShellMouseMoveAbsolute(ctx context.Context, x, y int) error {
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point(" + strconv.Itoa(x) + "," + strconv.Itoa(y) + ")",
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse move absolute")
}

func runPowerShellMouseMoveRelative(ctx context.Context, dx, dy int) error {
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$p=[System.Windows.Forms.Cursor]::Position",
		"[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point(($p.X+" + strconv.Itoa(dx) + "),($p.Y+" + strconv.Itoa(dy) + "))",
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse move relative")
}

func runPowerShellMousePosition(ctx context.Context) (MousePosition, error) {
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$p=[System.Windows.Forms.Cursor]::Position",
		"Write-Output ($p.X.ToString() + ',' + $p.Y.ToString())",
	}, ";")
	out, err := runPowerShellCommandOutput(ctx, script, "mouse position")
	if err != nil {
		return MousePosition{}, err
	}
	parts := strings.Split(strings.TrimSpace(out), ",")
	if len(parts) != 2 {
		return MousePosition{}, fmt.Errorf("%w: mouse position parse failed", ErrSendKeysFailed)
	}
	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return MousePosition{}, fmt.Errorf("%w: mouse position x parse failed", ErrSendKeysFailed)
	}
	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return MousePosition{}, fmt.Errorf("%w: mouse position y parse failed", ErrSendKeysFailed)
	}
	return MousePosition{X: x, Y: y}, nil
}

func runPowerShellMouseButtonDown(ctx context.Context, button string) error {
	return runPowerShellMouseButton(ctx, button, "down")
}

func runPowerShellMouseButtonUp(ctx context.Context, button string) error {
	return runPowerShellMouseButton(ctx, button, "up")
}

func runPowerShellMouseClick(ctx context.Context, button string) error {
	return runPowerShellMouseButton(ctx, button, "click")
}

func runPowerShellMouseDoubleClick(ctx context.Context, button string) error {
	return runPowerShellMouseButton(ctx, button, "double")
}

func runPowerShellMouseButton(ctx context.Context, button, mode string) error {
	mapped, err := mapPowerShellMouseButton(button)
	if err != nil {
		return err
	}
	call := "$wshell." + mode + "()"
	if mode == "click" {
		call = "$wshell.Click()"
	} else if mode == "double" {
		call = "$wshell.DoubleClick()"
	}
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$wshell = New-Object -ComObject WScript.Shell",
		"[System.Windows.Forms.Cursor]::Current = [System.Windows.Forms.Cursors]::Arrow",
		"$null = $wshell.AppActivate($PID)",
		"$null = [System.Windows.Forms.Cursor]::Position",
		"$mouse = New-Object -ComObject \"WScript.Shell\"",
		"$null = $mouse",
		"$btn = \"" + mapped + "\"",
		"$sh = New-Object -ComObject \"WScript.Shell\"",
		"$sh.SendKeys('{NUMLOCK}') > $null",
		"$wshell = New-Object -ComObject \"WScript.Shell\"",
		"Add-Type @'using System;using System.Runtime.InteropServices;public static class NativeMouse{[DllImport(\"user32.dll\")]public static extern void mouse_event(uint dwFlags,uint dx,uint dy,uint dwData,UIntPtr dwExtraInfo);}'@",
		"switch($btn){'left'{$down=0x0002;$up=0x0004};'right'{$down=0x0008;$up=0x0010};'middle'{$down=0x0020;$up=0x0040}}",
		"if('" + mode + "' -eq 'down'){[NativeMouse]::mouse_event($down,0,0,0,[UIntPtr]::Zero)}",
		"elseif('" + mode + "' -eq 'up'){[NativeMouse]::mouse_event($up,0,0,0,[UIntPtr]::Zero)}",
		"elseif('" + mode + "' -eq 'double'){[NativeMouse]::mouse_event($down,0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event($up,0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event($down,0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event($up,0,0,0,[UIntPtr]::Zero)}",
		"else{[NativeMouse]::mouse_event($down,0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event($up,0,0,0,[UIntPtr]::Zero)}",
		call,
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse "+mode)
}

func mapPowerShellMouseButton(button string) (string, error) {
	switch button {
	case MouseButtonLeft, MouseButtonRight, MouseButtonMiddle:
		return button, nil
	default:
		return "", fmt.Errorf("%w: unsupported mouse button %q", ErrInvalidInputArgument, button)
	}
}

func runPowerShellMouseWheel(ctx context.Context, delta int) error {
	script := strings.Join([]string{
		"Add-Type @'using System;using System.Runtime.InteropServices;public static class NativeMouse{[DllImport(\"user32.dll\")]public static extern void mouse_event(uint dwFlags,uint dx,uint dy,uint dwData,UIntPtr dwExtraInfo);}'@",
		"[NativeMouse]::mouse_event(0x0800,0,0," + strconv.Itoa(delta) + ",[UIntPtr]::Zero)",
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse wheel")
}

func runPowerShellCommand(ctx context.Context, script, action string) error {
	_, err := runPowerShellCommandOutput(ctx, script, action)
	return err
}

func runPowerShellCommandOutput(ctx context.Context, script, action string) (string, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return "", fmt.Errorf("input %s: %w", action, err)
	}
	return "", fmt.Errorf("input %s: %w (%s)", action, err, trimmed)
}
