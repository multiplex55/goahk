//go:build windows
// +build windows

package input

import (
	"context"
	"errors"
	"testing"
)

type fakeWinAPI struct {
	sendCount uint32
	sendErr   error
	x         int32
	y         int32
}

func (f *fakeWinAPI) sendInput(inputs []winInput) (uint32, error) {
	if f.sendErr != nil {
		return 0, f.sendErr
	}
	if f.sendCount == 0 {
		return uint32(len(inputs)), nil
	}
	return f.sendCount, nil
}
func (f *fakeWinAPI) getCursorPos() (int32, int32, error) { return f.x, f.y, nil }
func (f *fakeWinAPI) setCursorPos(x, y int32) error {
	f.x, f.y = x, y
	return nil
}

func TestSendInputBackend_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := windowsSendInputBackend{api: &fakeWinAPI{}}
	if err := b.sendChord(ctx, []string{"ctrl", "c"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled, got %v", err)
	}
}

func TestSendInputBackend_PartialSend(t *testing.T) {
	b := windowsSendInputBackend{api: &fakeWinAPI{sendCount: 1}}
	err := b.sendChord(context.Background(), []string{"ctrl", "c"})
	if err == nil {
		t.Fatal("expected partial send error")
	}
}
