//go:build !windows
// +build !windows

package uia

import "context"

type OSInspectProvider struct{}

func NewOSInspectProvider() *OSInspectProvider { return &OSInspectProvider{} }

func (p *OSInspectProvider) Focused(context.Context) (Element, error) {
	return Element{}, ErrUnsupportedPlatform
}

func (p *OSInspectProvider) UnderCursor(context.Context) (Element, error) {
	return Element{}, ErrUnsupportedPlatform
}

func (p *OSInspectProvider) ActiveWindowTree(context.Context, int) (*Node, error) {
	return nil, ErrUnsupportedPlatform
}
