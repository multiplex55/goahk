//go:build windows
// +build windows

package input

import (
	"context"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type windowsSendInputBackend struct {
	api windowsAPI
}

func newWindowsSendInputBackend() windowsBackend {
	return windowsSendInputBackend{api: nativeWindowsAPI{}}
}

type windowsAPI interface {
	sendInput([]winInput) (uint32, error)
	getCursorPos() (int32, int32, error)
	setCursorPos(int32, int32) error
}

type nativeWindowsAPI struct{}

type winInput struct {
	typ uint32
	ki  winKeyboardInput
	mi  winMouseInput
}

type winKeyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type winMouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type winPoint struct {
	x int32
	y int32
}

const (
	inputMouse    uint32 = 0
	inputKeyboard uint32 = 1
)

var (
	user32Proc       = windows.NewLazySystemDLL("user32.dll")
	procSendInput    = user32Proc.NewProc("SendInput")
	procGetCursorPos = user32Proc.NewProc("GetCursorPos")
	procSetCursorPos = user32Proc.NewProc("SetCursorPos")
	sizeOfWinInput   = uint32(unsafe.Sizeof(winInput{}))
)

func (nativeWindowsAPI) sendInput(inputs []winInput) (uint32, error) {
	if len(inputs) == 0 {
		return 0, nil
	}
	r1, _, e1 := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(sizeOfWinInput),
	)
	if r1 == 0 {
		if e1 != nil && e1 != windows.ERROR_SUCCESS {
			return 0, e1
		}
		return 0, windows.GetLastError()
	}
	return uint32(r1), nil
}

func (nativeWindowsAPI) getCursorPos() (int32, int32, error) {
	var p winPoint
	r1, _, e1 := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	if r1 == 0 {
		if e1 != nil && e1 != windows.ERROR_SUCCESS {
			return 0, 0, e1
		}
		return 0, 0, windows.GetLastError()
	}
	return p.x, p.y, nil
}

func (nativeWindowsAPI) setCursorPos(x, y int32) error {
	r1, _, e1 := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if r1 == 0 {
		if e1 != nil && e1 != windows.ERROR_SUCCESS {
			return e1
		}
		return windows.GetLastError()
	}
	return nil
}

func (b windowsSendInputBackend) sendText(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	inputs := buildTextKeyInputs(text)
	return b.sendKeyInputs(ctx, inputs)
}

func (b windowsSendInputBackend) sendChord(ctx context.Context, keys []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	inputs, err := buildChordKeyInputs(keys)
	if err != nil {
		return err
	}
	return b.sendKeyInputs(ctx, inputs)
}

func (b windowsSendInputBackend) sendKeyInputs(ctx context.Context, keys []keyInput) error {
	if len(keys) == 0 {
		return nil
	}
	inputs := make([]winInput, 0, len(keys))
	for _, k := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		inputs = append(inputs, winInput{typ: inputKeyboard, ki: winKeyboardInput{wVk: k.vk, wScan: k.scan, dwFlags: k.flags}})
	}
	sent, err := b.api.sendInput(inputs)
	if err != nil {
		return fmt.Errorf("send input keyboard: %w", err)
	}
	if sent != uint32(len(inputs)) {
		return fmt.Errorf("send input keyboard: partial send %d/%d", sent, len(inputs))
	}
	return nil
}

func (b windowsSendInputBackend) moveAbsolute(ctx context.Context, x, y int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.api.setCursorPos(int32(x), int32(y))
}

func (b windowsSendInputBackend) moveRelative(ctx context.Context, dx, dy int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	x, y, err := b.api.getCursorPos()
	if err != nil {
		return err
	}
	return b.api.setCursorPos(x+int32(dx), y+int32(dy))
}

func (b windowsSendInputBackend) position(ctx context.Context) (MousePosition, error) {
	if err := ctx.Err(); err != nil {
		return MousePosition{}, err
	}
	x, y, err := b.api.getCursorPos()
	if err != nil {
		return MousePosition{}, err
	}
	return MousePosition{X: int(x), Y: int(y)}, nil
}

func (b windowsSendInputBackend) mouseButton(ctx context.Context, button string, mode string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	events, err := buildMouseButtonInputs(button, mode)
	if err != nil {
		return err
	}
	inputs := make([]winInput, 0, len(events))
	for _, ev := range events {
		inputs = append(inputs, winInput{typ: inputMouse, mi: winMouseInput{dwFlags: ev.flags, mouseData: ev.data}})
	}
	sent, err := b.api.sendInput(inputs)
	if err != nil {
		return err
	}
	if sent != uint32(len(inputs)) {
		return fmt.Errorf("send input mouse: partial send %d/%d", sent, len(inputs))
	}
	return nil
}

func (b windowsSendInputBackend) wheel(ctx context.Context, delta int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	in := []winInput{{typ: inputMouse, mi: winMouseInput{dwFlags: mouseeventfWheel, mouseData: uint32(delta)}}}
	sent, err := b.api.sendInput(in)
	if err != nil {
		return err
	}
	if sent != 1 {
		return fmt.Errorf("send input wheel: partial send %d/1", sent)
	}
	return nil
}
