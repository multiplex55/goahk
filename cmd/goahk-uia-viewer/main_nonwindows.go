//go:build !windows

package main

import (
	"fmt"
	"io"
	"os"
)

const windowsOnlyMessage = "goahk-uia-viewer is only supported on Windows"

func runNonWindows(out io.Writer) error {
	_, err := fmt.Fprintln(out, windowsOnlyMessage)
	return err
}

func main() {
	if err := runNonWindows(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write message: %v\n", err)
		os.Exit(1)
	}
}
