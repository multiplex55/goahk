package actions

import "time"

type StepStatus string

const (
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
	StepStatusSkipped StepStatus = "skipped"
)

type StepResult struct {
	Action     string
	Kind       string
	Status     StepStatus
	StartedAt  time.Time
	EndedAt    time.Time
	Duration   time.Duration
	Error      string
	ErrorChain []string
	Nested     []StepResult
}

type ExecutionResult struct {
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
	Success   bool
	Steps     []StepResult
}
