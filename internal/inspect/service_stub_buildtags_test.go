//go:build !windows
// +build !windows

package inspect

import (
	"context"
	"errors"
	"testing"
)

func TestBuildTags_StubProviderSelected(t *testing.T) {
	if _, ok := newWindowsProvider().(unsupportedProvider); !ok {
		t.Fatalf("expected unsupportedProvider implementation")
	}
}

func TestBuildTags_NewServiceUsesStubProviderOnNonWindows(t *testing.T) {
	svc := NewService()
	_, err := svc.ListWindows(context.Background(), ListWindowsRequest{})
	if !errors.Is(err, ErrUnsupportedPatternAction) {
		t.Fatalf("expected ErrUnsupportedPatternAction, got %v", err)
	}
}
