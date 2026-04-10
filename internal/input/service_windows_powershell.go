//go:build windows
// +build windows

package input

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type windowsPowerShellBackend struct{}

func newWindowsPowerShellBackend() windowsBackend {
	return windowsPowerShellBackend{}
}

func (windowsPowerShellBackend) sendText(ctx context.Context, text string) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return runPowerShellSendKeys(ctx, escapeSendKeysLiteral(text), false)
}

func (windowsPowerShellBackend) sendChord(ctx context.Context, keys []string) error {
	combo := toSendKeysChord(strings.Join(keys, "+"))
	if strings.TrimSpace(combo) == "" {
		return nil
	}
	return runPowerShellSendKeys(ctx, combo, true)
}

func (windowsPowerShellBackend) moveAbsolute(ctx context.Context, x, y int) error {
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point(" + strconv.Itoa(x) + "," + strconv.Itoa(y) + ")",
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse move absolute")
}

func (windowsPowerShellBackend) moveRelative(ctx context.Context, dx, dy int) error {
	script := strings.Join([]string{
		"Add-Type -AssemblyName System.Windows.Forms",
		"$p=[System.Windows.Forms.Cursor]::Position",
		"[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point(($p.X+" + strconv.Itoa(dx) + "),($p.Y+" + strconv.Itoa(dy) + "))",
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse move relative")
}

func (windowsPowerShellBackend) position(ctx context.Context) (MousePosition, error) {
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

func (windowsPowerShellBackend) mouseButton(ctx context.Context, button, mode string) error {
	return runPowerShellMouseButton(ctx, button, mode)
}

func (windowsPowerShellBackend) wheel(ctx context.Context, delta int) error {
	script := strings.Join([]string{
		"Add-Type @'using System;using System.Runtime.InteropServices;public static class NativeMouse{[DllImport(\"user32.dll\")]public static extern void mouse_event(uint dwFlags,uint dx,uint dy,uint dwData,UIntPtr dwExtraInfo);}'@",
		"[NativeMouse]::mouse_event(0x0800,0,0," + strconv.Itoa(delta) + ",[UIntPtr]::Zero)",
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse wheel")
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
	return runPowerShellCommand(ctx, script, "send keys")
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

func runPowerShellMouseButton(ctx context.Context, button, mode string) error {
	down, up, err := mouseButtonFlags(button)
	if err != nil {
		return err
	}
	var body string
	switch mode {
	case "down":
		body = "[NativeMouse]::mouse_event(" + strconv.Itoa(int(down)) + ",0,0,0,[UIntPtr]::Zero)"
	case "up":
		body = "[NativeMouse]::mouse_event(" + strconv.Itoa(int(up)) + ",0,0,0,[UIntPtr]::Zero)"
	case "double":
		body = "[NativeMouse]::mouse_event(" + strconv.Itoa(int(down)) + ",0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event(" + strconv.Itoa(int(up)) + ",0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event(" + strconv.Itoa(int(down)) + ",0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event(" + strconv.Itoa(int(up)) + ",0,0,0,[UIntPtr]::Zero)"
	default:
		body = "[NativeMouse]::mouse_event(" + strconv.Itoa(int(down)) + ",0,0,0,[UIntPtr]::Zero);[NativeMouse]::mouse_event(" + strconv.Itoa(int(up)) + ",0,0,0,[UIntPtr]::Zero)"
	}
	script := strings.Join([]string{
		"Add-Type @'using System;using System.Runtime.InteropServices;public static class NativeMouse{[DllImport(\"user32.dll\")]public static extern void mouse_event(uint dwFlags,uint dx,uint dy,uint dwData,UIntPtr dwExtraInfo);}'@",
		body,
	}, ";")
	return runPowerShellCommand(ctx, script, "mouse "+mode)
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
