package program

import "time"

type Program struct {
	Bindings []BindingSpec
	Options  Options
}

type Options struct {
	Flows                        []FlowSpec
	UIASelectors                 map[string]UIASelectorSpec
	EnableImplicitEscapeControls bool
}

type FlowSpec struct {
	ID      string
	Timeout time.Duration
	Steps   []FlowStepSpec
}

type FlowStepSpec struct {
	Name      string
	Action    string
	Params    map[string]any
	Timeout   time.Duration
	If        *FlowIfSpec
	WaitUntil *FlowWaitUntilSpec
	Repeat    *FlowRepeatSpec
}

type FlowIfSpec struct {
	WindowMatches *string
	ElementExists *string
	Then          []FlowStepSpec
	Else          []FlowStepSpec
}

type FlowWaitUntilSpec struct {
	WindowMatches *string
	ElementExists *string
	Timeout       time.Duration
	Interval      time.Duration
}

type FlowRepeatSpec struct {
	Times int
	Steps []FlowStepSpec
}

type UIASelectorSpec struct {
	AutomationID string
	Name         string
	ControlType  string
	Ancestors    []UIASelectorSpec
}
