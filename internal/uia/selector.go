package uia

import (
	"fmt"
	"strings"
)

type Selector struct {
	AutomationID string     `json:"automationId,omitempty"`
	Name         string     `json:"name,omitempty"`
	ControlType  string     `json:"controlType,omitempty"`
	Ancestors    []Selector `json:"ancestors,omitempty"`
}

func (s Selector) Validate() error {
	if strings.TrimSpace(s.AutomationID) == "" && strings.TrimSpace(s.Name) == "" && strings.TrimSpace(s.ControlType) == "" {
		return fmt.Errorf("selector requires automationId, name, or controlType")
	}
	for i, anc := range s.Ancestors {
		if err := anc.Validate(); err != nil {
			return fmt.Errorf("ancestor[%d]: %w", i, err)
		}
	}
	return nil
}
