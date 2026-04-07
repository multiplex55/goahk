package uia

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// Navigator provides tree traversal primitives for automation elements.
type Navigator interface {
	ElementByID(ctx context.Context, id string) (Element, error)
	ChildrenIDs(ctx context.Context, id string) ([]string, error)
}

type Node struct {
	Element  Element `json:"element"`
	Children []*Node `json:"children,omitempty"`
	Cycle    bool    `json:"cycle,omitempty"`
}

type TreeOptions struct {
	MaxDepth int
}

func BuildTree(ctx context.Context, nav Navigator, rootID string, opts TreeOptions) (*Node, error) {
	if nav == nil {
		return nil, errors.New("nil navigator")
	}
	if rootID == "" {
		return nil, errors.New("root id is required")
	}
	seen := map[string]bool{}
	return buildTreeNode(ctx, nav, rootID, 0, opts.MaxDepth, seen)
}

func buildTreeNode(ctx context.Context, nav Navigator, id string, depth, maxDepth int, seen map[string]bool) (*Node, error) {
	el, err := nav.ElementByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("resolve element %q: %w", id, err)
	}
	if el.ID == "" {
		el.ID = id
	}
	n := &Node{Element: el}

	if maxDepth >= 0 && depth >= maxDepth {
		return n, nil
	}

	if seen[id] {
		n.Cycle = true
		return n, nil
	}
	seen[id] = true
	defer delete(seen, id)

	childIDs, err := nav.ChildrenIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load children for %q: %w", id, err)
	}
	sort.Strings(childIDs)
	for _, childID := range childIDs {
		child, err := buildTreeNode(ctx, nav, childID, depth+1, maxDepth, seen)
		if err != nil {
			return nil, err
		}
		n.Children = append(n.Children, child)
	}
	return n, nil
}
