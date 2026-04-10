package goahk

import "errors"

var (
	ErrWindowServiceUnavailable    = errors.New("window service unavailable")
	ErrInputServiceUnavailable     = errors.New("input service unavailable")
	ErrClipboardServiceUnavailable = errors.New("clipboard service unavailable")
)
