package testutil

import (
	"context"

	"goahk/internal/uia"
)

type FakeUIATreeProvider struct {
	FocusedElement   uia.Element
	UnderElement     uia.Element
	TreeNode         *uia.Node
	Err              error
	LastTreeMaxDepth int
}

func (f *FakeUIATreeProvider) Focused(context.Context) (uia.Element, error) {
	if f.Err != nil {
		return uia.Element{}, f.Err
	}
	return f.FocusedElement, nil
}

func (f *FakeUIATreeProvider) UnderCursor(context.Context) (uia.Element, error) {
	if f.Err != nil {
		return uia.Element{}, f.Err
	}
	return f.UnderElement, nil
}

func (f *FakeUIATreeProvider) ActiveWindowTree(_ context.Context, maxDepth int) (*uia.Node, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	f.LastTreeMaxDepth = maxDepth
	return f.TreeNode, nil
}
