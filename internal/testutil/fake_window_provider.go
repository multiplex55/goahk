package testutil

import (
	"context"

	"goahk/internal/window"
)

type FakeWindowProvider struct {
	ActiveInfo window.Info
	ListInfo   []window.Info
	Err        error
}

func (f *FakeWindowProvider) Active(context.Context) (window.Info, error) {
	if f.Err != nil {
		return window.Info{}, f.Err
	}
	return f.ActiveInfo, nil
}

func (f *FakeWindowProvider) List(context.Context) ([]window.Info, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return append([]window.Info(nil), f.ListInfo...), nil
}
