//go:build windows

package main

import "errors"

type walkDialogs struct{}

func (walkDialogs) PromptSetValue(defaultValue string) (string, bool, error) {
	_ = defaultValue
	return "", false, errors.New("set value dialog is not yet available in the walk shell")
}
