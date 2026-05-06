//go:build windows
// +build windows

package inspect

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
)

var (
	modOle32         = syscall.NewLazyDLL("ole32.dll")
	procCoInitialize = modOle32.NewProc("CoInitialize")
	procCoUninit     = modOle32.NewProc("CoUninitialize")
)

type uiaCOMWorker struct {
	once sync.Once
	jobs chan func()
}

func newUIACOMWorker() *uiaCOMWorker {
	w := &uiaCOMWorker{jobs: make(chan func())}
	go w.loop()
	return w
}

func (w *uiaCOMWorker) loop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	_, _, _ = procCoInitialize.Call(0)
	defer procCoUninit.Call()
	for job := range w.jobs {
		job()
	}
}

func (w *uiaCOMWorker) do(op string, fn func() error) error {
	if w == nil {
		return &UIAComUnavailableError{Op: op, Err: fmt.Errorf("worker unavailable")}
	}
	errCh := make(chan error, 1)
	w.jobs <- func() { errCh <- fn() }
	return <-errCh
}
