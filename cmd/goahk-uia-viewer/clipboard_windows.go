//go:build windows

package main

import "github.com/lxn/walk"

type walkClipboard struct{}

func (walkClipboard) CopyText(v string) error {
	walk.Clipboard().SetText(v)
	return nil
}
