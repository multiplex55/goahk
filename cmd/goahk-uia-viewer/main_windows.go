//go:build windows

package main

import (
	"fmt"
	"io"
	"os"
)

func runWindows(out io.Writer) error {
	_, err := fmt.Fprintln(out, "goahk-uia-viewer Windows entrypoint is under migration")
	return err
}

func main() {
	if err := runWindows(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write message: %v\n", err)
		os.Exit(1)
	}
}
