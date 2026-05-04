//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
)

var errUnsupportedPlatform = errors.New("goahk-uia-viewer is only supported on Windows")

func runNonWindows() error {
	return errUnsupportedPlatform
}

func main() {
	if err := runNonWindows(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
