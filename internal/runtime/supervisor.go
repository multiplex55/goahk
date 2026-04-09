package runtime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
	"goahk/internal/program"
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

type bindingState struct {
	running  map[uint64]context.CancelFunc
	pending  *supervisorJob
	admitted int
}

type supervisorJob struct {
	bindingID string
	trigger   hotkey.TriggerEvent
}

type Supervisor struct {
	ctx      context.Context
	bindings map[string]actions.ExecutableBinding
	executor *actions.Executor
	base     actions.ActionContext
	logSink  DispatchLogSink

	controlCh chan runtimeControlEvent
	workCh    chan supervisorJob
	resultsCh chan DispatchResult

	onControl func(runtimeControlEvent)

	mu     sync.Mutex
	active map[string]activeJob
	state  map[string]*bindingState

	nextRunID         atomic.Uint64
	hardStopRequested atomic.Bool
	closed            atomic.Bool
}

func NewSupervisor(ctx context.Context, bindings map[string]actions.ExecutableBinding, executor *actions.Executor, base actions.ActionContext, logSink DispatchLogSink, onControl func(runtimeControlEvent)) *Supervisor {
	if logSink == nil {
		logSink = func(context.Context, DispatchLogEntry) {}
	}
	if bindings == nil {
		bindings = map[string]actions.ExecutableBinding{}
	}
	return &Supervisor{
		ctx:       ctx,
		bindings:  bindings,
		executor:  executor,
		base:      base,
		logSink:   logSink,
		controlCh: make(chan runtimeControlEvent, 32),
		workCh:    make(chan supervisorJob, 128),
		resultsCh: make(chan DispatchResult, 32),
		onControl: onControl,
		active:    map[string]activeJob{},
		state:     map[string]*bindingState{},
	}
}

func (s *Supervisor) Results() <-chan DispatchResult { return s.resultsCh }

func (s *Supervisor) Start(workerCount int) {
	go s.controlLoop()
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
	s.admitWork(job)
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

func (s *Supervisor) runJob(job supervisorJob) {
	runID := s.nextRunID.Add(1)
	jobCtx, cancel := context.WithCancel(s.ctx)
	meta := supervisorJobMeta{BindingID: job.bindingID, RunID: runID, StartedAt: time.Now().UTC(), PolicyState: "running"}
	key := s.jobKey(meta.BindingID, meta.RunID)

	s.mu.Lock()
	s.active[key] = activeJob{meta: meta, cancel: cancel}
	state := s.ensureBindingStateLocked(job.bindingID)
	state.running[runID] = cancel
	s.mu.Unlock()

	s.logSink(s.ctx, DispatchLogEntry{Event: "job_started", BindingID: job.bindingID, Timestamp: meta.StartedAt})
	actionCtx := s.base
	actionCtx.Context = jobCtx
	actionCtx.BindingID = job.bindingID
	actionCtx.TriggerText = job.trigger.Chord.String()

	binding, exists := s.bindings[job.bindingID]
	if !exists {
		s.logSink(s.ctx, DispatchLogEntry{Event: "dispatch_unknown_binding", BindingID: job.bindingID, Error: "binding descriptor not found", Timestamp: time.Now().UTC()})
		s.completeJob(key, runID, cancel, jobCtx, job.bindingID, buildDispatchResult(job.bindingID, actions.ExecutableBinding{ID: job.bindingID, Kind: actions.BindingKindPlan}, missingBindingExecutionResult(job.bindingID)))
		return
	}
	if err := validateExecutableBinding(binding); err != nil {
		s.completeJob(key, runID, cancel, jobCtx, job.bindingID, buildDispatchResult(job.bindingID, binding, invalidBindingExecutionResult(err)))
		return
	}

	execResult := s.executor.ExecuteBinding(actionCtx, binding)
	envelope := buildDispatchResult(job.bindingID, binding, execResult)
	s.completeJob(key, runID, cancel, jobCtx, job.bindingID, envelope)
}

func (s *Supervisor) completeJob(key string, runID uint64, cancel context.CancelFunc, jobCtx context.Context, bindingID string, envelope DispatchResult) {
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
	shouldStartPending := false
	var next supervisorJob
	if state := s.ensureBindingStateLocked(bindingID); state != nil {
		if state.admitted > 0 {
			state.admitted--
		}
		delete(state.running, runID)
		if len(state.running) == 0 && state.pending != nil {
			next = *state.pending
			state.pending = nil
			state.admitted++
			shouldStartPending = true
		}
	}
	remaining := len(s.active)
	s.mu.Unlock()
	cancel()
	if jobCtx.Err() != nil {
		s.logSink(s.ctx, DispatchLogEntry{Event: "job_canceled", BindingID: bindingID, Timestamp: time.Now().UTC()})
	}
	if s.ctx.Err() != nil && remaining == 0 {
		s.closeResults()
	}
	if shouldStartPending {
		s.logSink(s.ctx, DispatchLogEntry{Event: "policy_queue_one_dequeue", BindingID: bindingID, Timestamp: time.Now().UTC()})
		go s.runJob(next)
	}
}

func missingBindingExecutionResult(bindingID string) actions.ExecutionResult {
	msg := fmt.Sprintf("binding descriptor not found for %q", bindingID)
	return actions.ExecutionResult{StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC(), Success: false, Steps: []actions.StepResult{{Action: "binding", Kind: "binding", Status: actions.StepStatusFailed, Error: msg, ErrorChain: []string{msg}, StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC()}}}
}

func invalidBindingExecutionResult(err error) actions.ExecutionResult {
	now := time.Now().UTC()
	return actions.ExecutionResult{StartedAt: now, EndedAt: now, Success: false, Steps: []actions.StepResult{{Action: "binding", Kind: "binding", Status: actions.StepStatusFailed, Error: err.Error(), ErrorChain: []string{err.Error()}, StartedAt: now, EndedAt: now}}}
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

func (s *Supervisor) admitWork(job supervisorJob) {
	policy := string(program.DefaultConcurrencyPolicy())
	if binding, ok := s.bindings[job.bindingID]; ok {
		policy = bindingConcurrencyPolicy(binding)
	}
	now := time.Now().UTC()

	s.mu.Lock()
	state := s.ensureBindingStateLocked(job.bindingID)
	busy := state.admitted > 0
	switch policy {
	case "parallel":
		state.admitted++
		s.mu.Unlock()
		s.logSink(s.ctx, DispatchLogEntry{Event: "policy_parallel_admit", BindingID: job.bindingID, Timestamp: now})
		go s.runJob(job)
		return
	case "replace":
		if busy {
			for runID, cancel := range state.running {
				cancel()
				delete(state.running, runID)
			}
			s.logSink(s.ctx, DispatchLogEntry{Event: "policy_replace_cancel_running", BindingID: job.bindingID, Timestamp: now})
		}
		state.admitted++
		s.mu.Unlock()
		s.logSink(s.ctx, DispatchLogEntry{Event: "policy_replace_admit_latest", BindingID: job.bindingID, Timestamp: now})
		go s.runJob(job)
		return
	case "queue-one":
		if busy {
			cp := job
			state.pending = &cp
			s.mu.Unlock()
			s.logSink(s.ctx, DispatchLogEntry{Event: "policy_queue_one_pending", BindingID: job.bindingID, Timestamp: now})
			return
		}
		state.admitted++
		s.mu.Unlock()
		s.logSink(s.ctx, DispatchLogEntry{Event: "policy_queue_one_admit", BindingID: job.bindingID, Timestamp: now})
		go s.runJob(job)
		return
	case "drop":
		if busy {
			s.mu.Unlock()
			s.logSink(s.ctx, DispatchLogEntry{Event: "policy_drop_ignored_busy", BindingID: job.bindingID, Timestamp: now})
			return
		}
		state.admitted++
		s.mu.Unlock()
		s.logSink(s.ctx, DispatchLogEntry{Event: "policy_drop_admit", BindingID: job.bindingID, Timestamp: now})
		go s.runJob(job)
		return
	case "serial":
		fallthrough
	default:
		if busy {
			s.mu.Unlock()
			s.logSink(s.ctx, DispatchLogEntry{Event: "policy_serial_ignored_busy", BindingID: job.bindingID, Timestamp: now})
			return
		}
		state.admitted++
		s.mu.Unlock()
		s.logSink(s.ctx, DispatchLogEntry{Event: "policy_serial_admit", BindingID: job.bindingID, Timestamp: now})
		go s.runJob(job)
		return
	}
}

func (s *Supervisor) ensureBindingStateLocked(bindingID string) *bindingState {
	state, ok := s.state[bindingID]
	if !ok {
		state = &bindingState{running: map[uint64]context.CancelFunc{}}
		s.state[bindingID] = state
	}
	return state
}
