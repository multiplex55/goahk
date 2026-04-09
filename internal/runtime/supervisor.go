package runtime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

type RuntimeControlCommand string

const (
	RuntimeControlStop     RuntimeControlCommand = "stop"
	RuntimeControlHardStop RuntimeControlCommand = "hard_stop"
	RuntimeControlSuspend  RuntimeControlCommand = "suspend"
	RuntimeControlReload   RuntimeControlCommand = "reload"
)

type runtimeControlEvent struct {
	BindingID string
	Command   RuntimeControlCommand
	Triggered hotkey.TriggerEvent
	Received  time.Time
}

type supervisorJobMeta struct {
	BindingID   string
	RunID       uint64
	StartedAt   time.Time
	PolicyState string
}

type activeJob struct {
	meta   supervisorJobMeta
	cancel context.CancelFunc
}

type supervisorJob struct {
	bindingID string
	trigger   hotkey.TriggerEvent
	plan      actions.Plan
}

type Supervisor struct {
	ctx      context.Context
	executor *actions.Executor
	base     actions.ActionContext
	logSink  DispatchLogSink

	controlCh chan runtimeControlEvent
	workCh    chan supervisorJob
	resultsCh chan DispatchResult

	onControl func(runtimeControlEvent)

	mu     sync.Mutex
	active map[string]activeJob

	nextRunID         atomic.Uint64
	hardStopRequested atomic.Bool
	closed            atomic.Bool
}

func NewSupervisor(ctx context.Context, executor *actions.Executor, base actions.ActionContext, logSink DispatchLogSink, onControl func(runtimeControlEvent)) *Supervisor {
	if logSink == nil {
		logSink = func(context.Context, DispatchLogEntry) {}
	}
	return &Supervisor{
		ctx:       ctx,
		executor:  executor,
		base:      base,
		logSink:   logSink,
		controlCh: make(chan runtimeControlEvent, 32),
		workCh:    make(chan supervisorJob, 128),
		resultsCh: make(chan DispatchResult, 32),
		onControl: onControl,
		active:    map[string]activeJob{},
	}
}

func (s *Supervisor) Results() <-chan DispatchResult { return s.resultsCh }

func (s *Supervisor) Start(workerCount int) {
	if workerCount <= 0 {
		workerCount = 4
	}
	go s.controlLoop()
	for i := 0; i < workerCount; i++ {
		go s.workerLoop()
	}
}

func (s *Supervisor) SubmitControl(ev runtimeControlEvent) {
	select {
	case s.controlCh <- ev:
	default:
		go func() {
			select {
			case s.controlCh <- ev:
			case <-s.ctx.Done():
			}
		}()
	}
}

func (s *Supervisor) SubmitWork(job supervisorJob) {
	select {
	case s.workCh <- job:
	default:
		go func() {
			select {
			case s.workCh <- job:
			case <-s.ctx.Done():
			}
		}()
	}
}

func (s *Supervisor) WaitForGraceful(timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if s.activeCount() == 0 {
			return true
		}
		select {
		case <-deadline.C:
			return false
		case <-tick.C:
		}
	}
}

func (s *Supervisor) HardStopRequested() bool {
	return s.hardStopRequested.Load()
}

func (s *Supervisor) ForceTerminateAll() {
	s.hardStopRequested.Store(true)
	now := time.Now().UTC()
	s.mu.Lock()
	for key, job := range s.active {
		job.cancel()
		s.logSink(s.ctx, DispatchLogEntry{Event: "job_forced_termination", BindingID: job.meta.BindingID, Timestamp: now})
		delete(s.active, key)
	}
	s.mu.Unlock()
	s.closeResults()
}

func (s *Supervisor) CloseWhenIdle(maxWait time.Duration) {
	go func() {
		if maxWait <= 0 {
			maxWait = 250 * time.Millisecond
		}
		deadline := time.NewTimer(maxWait)
		defer deadline.Stop()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			if s.activeCount() == 0 && len(s.workCh) == 0 {
				s.closeResults()
				return
			}
			select {
			case <-deadline.C:
				s.closeResults()
				return
			case <-ticker.C:
			}
		}
	}()
}

func (s *Supervisor) controlLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case ev := <-s.controlCh:
			s.logSink(s.ctx, DispatchLogEntry{Event: "control_command_received", BindingID: ev.BindingID, Timestamp: ev.Received})
			if ev.Command == RuntimeControlHardStop {
				s.hardStopRequested.Store(true)
			}
			if s.onControl != nil {
				s.onControl(ev)
			}
		}
	}
}

func (s *Supervisor) workerLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case job := <-s.workCh:
			s.runJob(job)
		}
	}
}

func (s *Supervisor) runJob(job supervisorJob) {
	runID := s.nextRunID.Add(1)
	jobCtx, cancel := context.WithCancel(s.ctx)
	meta := supervisorJobMeta{BindingID: job.bindingID, RunID: runID, StartedAt: time.Now().UTC(), PolicyState: "running"}
	key := s.jobKey(meta.BindingID, meta.RunID)

	s.mu.Lock()
	s.active[key] = activeJob{meta: meta, cancel: cancel}
	s.mu.Unlock()

	s.logSink(s.ctx, DispatchLogEntry{Event: "job_started", BindingID: job.bindingID, Timestamp: meta.StartedAt})
	actionCtx := s.base
	actionCtx.Context = jobCtx
	actionCtx.BindingID = job.bindingID
	actionCtx.TriggerText = job.trigger.Chord.String()

	execResult := s.executor.Execute(actionCtx, job.plan)
	envelope := buildDispatchResult(job.bindingID, job.plan, execResult)

	if !s.closed.Load() {
		select {
		case s.resultsCh <- envelope:
		default:
			select {
			case s.resultsCh <- envelope:
			case <-time.After(25 * time.Millisecond):
			}
		}
	}
	s.mu.Lock()
	delete(s.active, key)
	remaining := len(s.active)
	s.mu.Unlock()
	cancel()
	if jobCtx.Err() != nil {
		s.logSink(s.ctx, DispatchLogEntry{Event: "job_canceled", BindingID: job.bindingID, Timestamp: time.Now().UTC()})
	}
	if s.ctx.Err() != nil && remaining == 0 {
		s.closeResults()
	}
}

func (s *Supervisor) activeCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.active)
}

func (s *Supervisor) jobKey(bindingID string, runID uint64) string {
	return fmt.Sprintf("%s#%d", bindingID, runID)
}

func (s *Supervisor) closeResults() {
	if s.closed.CompareAndSwap(false, true) {
		close(s.resultsCh)
	}
}
