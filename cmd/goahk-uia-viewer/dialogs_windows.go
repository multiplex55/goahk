//go:build windows

package main

import "github.com/lxn/walk"

type walkDialogs struct{}

func (walkDialogs) PromptSetValue(defaultValue string) (string, bool, error) {
	value, ok := walk.InputBox(nil, "Set Value", "Enter new value", defaultValue)
	return value, ok, nil
}
