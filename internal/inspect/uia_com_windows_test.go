//go:build windows
// +build windows

package inspect

import (
	"errors"
	"testing"

	"goahk/internal/window"
)

type fakeUIAClient struct{}

func (fakeUIAClient) ElementFromHWND(window.HWND) (*uiaBridgeElement, error)  { return nil, nil }
func (fakeUIAClient) FocusedElement() (*uiaBridgeElement, error)              { return nil, nil }
func (fakeUIAClient) ElementFromPoint(int, int) (*uiaBridgeElement, error)    { return nil, nil }
func (fakeUIAClient) ElementByRuntimeID(string) (*uiaBridgeElement, error)    { return nil, nil }
func (fakeUIAClient) Parent(*uiaBridgeElement) (*uiaBridgeElement, error)     { return nil, nil }
func (fakeUIAClient) Children(*uiaBridgeElement) ([]*uiaBridgeElement, error) { return nil, nil }

func TestNewWin32UIABridge_UsesNativeClientWhenAvailable(t *testing.T) {
	orig := newUIAComClient
	defer func() { newUIAComClient = orig }()
	newUIAComClient = func() (uiaAutomationClient, error) { return fakeUIAClient{}, nil }

	bridge := newWin32UIABridge().(*win32UIAComBridge)
	if bridge.initErr != nil {
		t.Fatalf("expected nil initErr, got %v", bridge.initErr)
	}
}

func TestNewWin32UIABridge_FallsBackWithInitError(t *testing.T) {
	orig := newUIAComClient
	defer func() { newUIAComClient = orig }()
	newUIAComClient = func() (uiaAutomationClient, error) { return nil, errors.New("com init failed") }

	bridge := newWin32UIABridge().(*win32UIAComBridge)
	if bridge.initErr == nil {
		t.Fatal("expected init error")
	}
	if _, err := bridge.ResolveRoot(0); err == nil || !errors.Is(err, bridge.initErr) {
		t.Fatalf("expected wrapped init error, got %v", err)
	}
}
