//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	modOle32                = syscall.NewLazyDLL("ole32.dll")
	modOleAut32             = syscall.NewLazyDLL("oleaut32.dll")
	procCoInitializeEx      = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize      = modOle32.NewProc("CoUninitialize")
	procCoCreateInstance    = modOle32.NewProc("CoCreateInstance")
	procSafeArrayDestroy    = modOleAut32.NewProc("SafeArrayDestroy")
	procSafeArrayGetLBound  = modOleAut32.NewProc("SafeArrayGetLBound")
	procSafeArrayGetUBound  = modOleAut32.NewProc("SafeArrayGetUBound")
	procSafeArrayGetElement = modOleAut32.NewProc("SafeArrayGetElement")
	procVariantClear        = modOleAut32.NewProc("VariantClear")
	procSysAllocString      = modOleAut32.NewProc("SysAllocString")
	procSysFreeString       = modOleAut32.NewProc("SysFreeString")
	clsidCUIAutomation      = syscall.GUID{Data1: 0xff48dba4, Data2: 0x60ef, Data3: 0x4201, Data4: [8]byte{0xaa, 0x87, 0x54, 0x10, 0x3e, 0xef, 0x59, 0x4e}}
	iidIUIAutomation        = syscall.GUID{Data1: 0x30cbe57d, Data2: 0xd9d0, Data3: 0x452a, Data4: [8]byte{0xab, 0x13, 0x7a, 0xc5, 0xac, 0x48, 0x25, 0xee}}
)

const (
	coInitApartmentThreaded = 0x2
	clsctxInprocServer      = 0x1
)

type uiaWorkerState struct {
	automation uintptr
	trueCond   uintptr
	treeWalker uintptr
}

type uiaWorkerJob struct {
	op string
	fn func(*uiaWorkerState) error
	dn chan error
}

type uiaCOMWorker struct {
	jobs     chan uiaWorkerJob
	closed   chan struct{}
	once     sync.Once
	closeMux sync.Mutex
	isClosed bool
}

var uiaWorkerJobObserver func(op string)

func newUIACOMWorker() (*uiaCOMWorker, error) {
	return newUIACOMWorkerWithInit(workerCOMInit)
}

func newUIACOMWorkerWithInit(initFn func(*uiaWorkerState) error) (*uiaCOMWorker, error) {
	w := &uiaCOMWorker{jobs: make(chan uiaWorkerJob), closed: make(chan struct{})}
	ready := make(chan error, 1)
	go w.loop(ready, initFn)
	if err := <-ready; err != nil {
		return nil, err
	}
	return w, nil
}

func (w *uiaCOMWorker) loop(ready chan<- error, initFn func(*uiaWorkerState) error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	state := &uiaWorkerState{}
	if err := initFn(state); err != nil {
		log.Printf("inspect.uia.worker_start status=error backend=native-com err=%v", err)
		ready <- err
		close(w.closed)
		return
	}
	log.Printf("inspect.uia.worker_start status=ok backend=native-com")
	ready <- nil
	for job := range w.jobs {
		if uiaWorkerJobObserver != nil {
			uiaWorkerJobObserver(job.op)
		}
		job.dn <- job.fn(state)
	}
	releaseWorkerState(state)
	log.Printf("inspect.uia.worker_stop status=ok backend=native-com")
	close(w.closed)
}

func (w *uiaCOMWorker) Do(op string, fn func(*uiaWorkerState) error) error {
	if w == nil {
		return &UIAComUnavailableError{Op: op, Err: errors.New("worker unavailable")}
	}
	w.closeMux.Lock()
	if w.isClosed {
		w.closeMux.Unlock()
		return &UIAComUnavailableError{Op: op, Err: errors.New("worker closed")}
	}
	res := make(chan error, 1)
	w.jobs <- uiaWorkerJob{op: op, fn: fn, dn: res}
	w.closeMux.Unlock()
	return <-res
}

func (w *uiaCOMWorker) Close() error {
	if w == nil {
		return nil
	}
	w.once.Do(func() {
		w.closeMux.Lock()
		w.isClosed = true
		close(w.jobs)
		w.closeMux.Unlock()
		<-w.closed
	})
	return nil
}

func workerCOMInit(state *uiaWorkerState) error {
	hr, _, _ := procCoInitializeEx.Call(0, uintptr(coInitApartmentThreaded))
	if int32(hr) < 0 {
		return fmt.Errorf("CoInitializeEx failed: hr=0x%x", hr)
	}
	var automation uintptr
	hr, _, _ = procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidCUIAutomation)),
		0,
		uintptr(clsctxInprocServer),
		uintptr(unsafe.Pointer(&iidIUIAutomation)),
		uintptr(unsafe.Pointer(&automation)),
	)
	if int32(hr) < 0 {
		procCoUninitialize.Call()
		return fmt.Errorf("CoCreateInstance(IUIAutomation) failed: hr=0x%x", hr)
	}
	state.automation = automation
	trueCond, err := uiaCreateTrueCondition(automation)
	if err != nil {
		comRelease(automation)
		procCoUninitialize.Call()
		return err
	}
	if trueCond == 0 {
		comRelease(automation)
		procCoUninitialize.Call()
		return errors.New("CreateTrueCondition returned nil condition")
	}
	treeWalker, err := uiaGetRawViewWalker(automation)
	if err != nil {
		comRelease(trueCond)
		comRelease(automation)
		procCoUninitialize.Call()
		return err
	}
	if treeWalker == 0 {
		comRelease(trueCond)
		comRelease(automation)
		procCoUninitialize.Call()
		return errors.New("get_RawViewWalker returned nil walker")
	}
	state.trueCond = trueCond
	state.treeWalker = treeWalker
	return nil
}

func releaseWorkerState(state *uiaWorkerState) {
	if state == nil {
		return
	}
	comRelease(state.treeWalker)
	comRelease(state.trueCond)
	comRelease(state.automation)
	procCoUninitialize.Call()
}
