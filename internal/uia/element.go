package uia

import "fmt"

// Element represents a subset of UI Automation properties used for diagnostics.
type Element struct {
	ID           string   `json:"id,omitempty"`
	Name         *string  `json:"name,omitempty"`
	ClassName    *string  `json:"className,omitempty"`
	ControlType  *string  `json:"controlType,omitempty"`
	AutomationID *string  `json:"automationId,omitempty"`
	FrameworkID  *string  `json:"frameworkId,omitempty"`
	Patterns     []string `json:"patterns,omitempty"`
}

func ptrString(v *string) string {
	if v == nil || *v == "" {
		return "(missing)"
	}
	return *v
}

func (e Element) Summary() string {
	return fmt.Sprintf("id=%s name=%s controlType=%s class=%s", emptyOrUnknown(e.ID), ptrString(e.Name), ptrString(e.ControlType), ptrString(e.ClassName))
}

func emptyOrUnknown(v string) string {
	if v == "" {
		return "(unknown)"
	}
	return v
}
