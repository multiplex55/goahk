package uia

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// COMInitializer abstracts COM apartment management.
type COMInitializer interface {
	Initialize() error
	Uninitialize()
}

// Backend exposes native UIA operations behind a testable boundary.
type Backend interface {
	FocusedElement(ctx context.Context) (Element, error)
	ElementUnderCursor(ctx context.Context) (Element, error)
	ActiveWindowRootID(ctx context.Context) (string, error)
	Navigator() Navigator
}

// Client owns UIA access and optional COM init/uninit lifecycle.
type Client struct {
	backend Backend
	com     COMInitializer
	ownsCOM bool

	mu     sync.Mutex
	closed bool
}

func NewClient(backend Backend, com COMInitializer, ownsCOM bool) (*Client, error) {
	if backend == nil {
		return nil, errors.New("nil backend")
	}
	if ownsCOM {
		if com == nil {
			return nil, errors.New("ownsCOM requires COM initializer")
		}
		if err := com.Initialize(); err != nil {
			return nil, fmt.Errorf("initialize COM: %w", err)
		}
	}
	return &Client{backend: backend, com: com, ownsCOM: ownsCOM}, nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	if c.ownsCOM && c.com != nil {
		c.com.Uninitialize()
	}
}

func (c *Client) Focused(ctx context.Context) (Element, error) {
	if err := c.ensureOpen(); err != nil {
		return Element{}, err
	}
	return c.backend.FocusedElement(ctx)
}

func (c *Client) UnderCursor(ctx context.Context) (Element, error) {
	if err := c.ensureOpen(); err != nil {
		return Element{}, err
	}
	return c.backend.ElementUnderCursor(ctx)
}

func (c *Client) TreeFromActiveWindow(ctx context.Context, maxDepth int) (*Node, error) {
	if err := c.ensureOpen(); err != nil {
		return nil, err
	}
	rootID, err := c.backend.ActiveWindowRootID(ctx)
	if err != nil {
		return nil, err
	}
	return BuildTree(ctx, c.backend.Navigator(), rootID, TreeOptions{MaxDepth: maxDepth})
}

func (c *Client) ensureOpen() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("uia client closed")
	}
	return nil
}
