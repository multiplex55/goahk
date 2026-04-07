package flow

import "time"

type Definition struct {
	ID      string
	Timeout time.Duration
	Steps   []Step
}

type Step struct {
	Name      string
	Action    string
	Params    map[string]string
	Timeout   time.Duration
	If        *IfBlock
	WaitUntil *WaitUntilBlock
	Repeat    *RepeatBlock
}

type IfBlock struct {
	Condition Condition
	Then      []Step
	Else      []Step
}

type WaitUntilBlock struct {
	Condition Condition
	Timeout   time.Duration
	Interval  time.Duration
}

type RepeatBlock struct {
	Times int
	Steps []Step
}

func (s Step) kind() string {
	switch {
	case s.Action != "":
		return "action"
	case s.If != nil:
		return "if"
	case s.WaitUntil != nil:
		return "wait_until"
	case s.Repeat != nil:
		return "repeat"
	default:
		return "unknown"
	}
}
