package input

import "errors"

var (
	ErrInvalidInputArgument = errors.New("input: invalid argument")
	ErrSendKeysFailed       = errors.New("input: send keys failed")
)
